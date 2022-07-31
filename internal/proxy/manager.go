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
	handlers map[string]http.Handler
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
		handlers: map[string]http.Handler{},
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

				switch msg.Action {
				case "create":
					p.logger.Info().
						Str("id", msg.ID).
						Str("name", name).
						Str("host", host).
						Msg("container created")

				case "start":
					p.Add(name, host+":"+port)

					p.logger.Info().
						Str("id", msg.ID).
						Str("name", name).
						Str("host", host).
						Msg("container started")

				case "stop":
					p.Remove(name)

					p.logger.Info().
						Str("id", msg.ID).
						Str("name", name).
						Str("host", host).
						Msg("container stopped")

				case "destroy":
					p.logger.Info().
						Str("id", msg.ID).
						Str("name", name).
						Str("host", host).
						Msg("container destroyed")
				}
			}
		}
	}()
}

func (p *ProxyManagerCtx) Shutdown() error {
	p.Clear()

	p.cancel()
	return p.ctx.Err()
}

func (p *ProxyManagerCtx) refresh() error {
	containers, err := p.client.ContainerList(p.ctx, dockerTypes.ContainerListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("m1k1o.neko_rooms.instance=%s", p.config.InstanceName)),
		),
	})

	if err != nil {
		return err
	}

	p.clear()

	for _, cont := range containers {
		name, port, ok := p.parseLabels(cont.Labels)
		if ok {
			host := cont.ID[:12]
			p.add(name, host+":"+port)
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

func (p *ProxyManagerCtx) add(name, host string) {
	p.handlers[name] = httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   host,
	})

	p.logger.Debug().Str("name", name).Str("host", host).Msg("add handler")
}

func (p *ProxyManagerCtx) remove(name string) {
	delete(p.handlers, name)
	p.logger.Debug().Str("name", name).Msg("remove handler")
}

func (p *ProxyManagerCtx) clear() {
	p.handlers = map[string]http.Handler{}
	p.logger.Debug().Msg("clear handlers")
}

func (p *ProxyManagerCtx) Refresh() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.refresh()
}

func (p *ProxyManagerCtx) Add(name, host string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.add(name, host)
}

func (p *ProxyManagerCtx) Remove(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.remove(name)
}

func (p *ProxyManagerCtx) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.clear()
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
		http.NotFound(w, r)
		return
	}

	// redirect to room ending with /
	if doRedir {
		r.URL.Path += "/"
		http.Redirect(w, r, r.URL.String(), http.StatusTemporaryRedirect)
		return
	}

	// strip prefix from proxy
	pathPrefix := path.Join(p.prefix, roomName)
	proxy = http.StripPrefix(pathPrefix, proxy)

	// handle by proxy
	proxy.ServeHTTP(w, r)
}
