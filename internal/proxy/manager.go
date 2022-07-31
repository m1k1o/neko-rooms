package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"
	"sync"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerClient "github.com/docker/docker/client"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/m1k1o/neko-rooms/internal/config"
)

type ProxyManagerCtx struct {
	logger   zerolog.Logger
	mu       sync.RWMutex
	ctx      context.Context
	cancel   func()
	prefix   string
	client   *dockerClient.Client
	config   *config.Room
	handlers map[string]*httputil.ReverseProxy
}

func New(client *dockerClient.Client, config *config.Room) *ProxyManagerCtx {
	logger := log.With().Str("module", "proxy").Logger()

	prefix := config.PathPrefix
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	return &ProxyManagerCtx{
		logger:   logger,
		prefix:   prefix,
		client:   client,
		config:   config,
		handlers: map[string]*httputil.ReverseProxy{},
	}
}

func (p *ProxyManagerCtx) Start() {
	p.ctx, p.cancel = context.WithCancel(context.Background())

	go func() {
		err := p.Refresh()
		if err != nil {
			p.logger.Err(err).Msg("unable to refresh containers")
		}

		msgs, errs := p.client.Events(p.ctx, dockerTypes.EventsOptions{
			Filters: filters.NewArgs(
				filters.Arg("type", "container"),
				filters.Arg("label", fmt.Sprintf("m1k1o.neko_rooms.instance=%s", p.config.InstanceName)),
				filters.Arg("event", "create"),
				filters.Arg("event", "start"),
				//filters.Arg("event", "health_status"),
				filters.Arg("event", "stop"),
				filters.Arg("event", "destroy"),
			),
		})

		for {
			select {
			case err := <-errs:
				p.logger.Info().Interface("err", err).Msg("eee")
			case msg := <-msgs:
				host := msg.ID[:12]
				name, port, ok := p.parseLabels(msg.Actor.Attributes)
				if !ok {
					break
				}

				p.logger.Info().
					Str("action", msg.Action).
					Str("name", name).
					Str("host", host).
					Msg("new docker event")

				p.mu.Lock()
				switch msg.Action {
				case "create":
					p.handlers[name] = nil
				case "start":
					p.handlers[name] = httputil.NewSingleHostReverseProxy(&url.URL{
						Scheme: "http",
						Host:   host + ":" + port,
					})
				case "stop":
					p.handlers[name] = nil
				case "destroy":
					delete(p.handlers, name)
				}
				p.mu.Unlock()
			}
		}
	}()
}

func (p *ProxyManagerCtx) Shutdown() error {
	p.cancel()
	return p.ctx.Err()
}

func (p *ProxyManagerCtx) Refresh() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	containers, err := p.client.ContainerList(p.ctx, dockerTypes.ContainerListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("m1k1o.neko_rooms.instance=%s", p.config.InstanceName)),
		),
	})

	if err != nil {
		return err
	}

	p.handlers = map[string]*httputil.ReverseProxy{}

	for _, cont := range containers {
		name, port, ok := p.parseLabels(cont.Labels)
		if !ok {
			continue
		}

		if cont.State == "running" {
			host := cont.ID[:12]
			p.handlers[name] = httputil.NewSingleHostReverseProxy(&url.URL{
				Scheme: "http",
				Host:   host + ":" + port,
			})
		} else {
			p.handlers[name] = nil
		}
	}

	return nil
}

func (p *ProxyManagerCtx) parseLabels(labels map[string]string) (name string, port string, ok bool) {
	name, ok = labels["m1k1o.neko_rooms.name"]
	if !ok {
		return
	}

	// TODO: Do not use trafik for this.
	port, ok = labels["traefik.http.services.neko-rooms-"+name+"-frontend.loadbalancer.server.port"]
	return
}

func (p *ProxyManagerCtx) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// only if has prefix
	if !strings.HasPrefix(r.URL.Path, p.prefix) {
		http.NotFound(w, r)
		return
	}

	// remove prefix and leading /
	roomPath := strings.TrimPrefix(r.URL.Path, p.prefix)
	roomPath = strings.TrimLeft(roomPath, "/")
	if roomPath == "" {
		http.NotFound(w, r)
		return
	}

	// get room name
	roomName, doRedir := roomPath, false
	if i := strings.Index(roomPath, "/"); i != -1 {
		roomName = roomPath[:i]
	} else {
		doRedir = true
	}

	// get proxy by room name
	p.mu.RLock()
	proxy, ok := p.handlers[roomName]
	p.mu.RUnlock()

	// if room not found
	if !ok {
		RoomNotFound(w, r)
		return
	}

	// redirect to room ending with /
	if doRedir {
		r.URL.Path += "/"
		http.Redirect(w, r, r.URL.String(), http.StatusTemporaryRedirect)
		return
	}

	// if room not running
	if proxy == nil {
		RoomNotRunning(w, r)
		return
	}

	// handle not ready room
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		if r.URL.Path == "/" {
			RoomNotReady(w, r)
		}

		p.logger.Err(err).Str("room", roomName).Msg("proxying error")
	}

	// strip prefix from proxy
	pathPrefix := path.Join(p.prefix, roomName)
	handler := http.StripPrefix(pathPrefix, proxy)

	// handle by proxy
	handler.ServeHTTP(w, r)
}
