package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerClient "github.com/docker/docker/client"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ProxyManagerCtx struct {
	logger zerolog.Logger
	mu     sync.RWMutex
	ctx    context.Context
	cancel func()

	client       *dockerClient.Client
	instanceName string
	handlers     *prefixHandler
}

func New(client *dockerClient.Client, instanceName string) *ProxyManagerCtx {
	return &ProxyManagerCtx{
		logger: log.With().Str("module", "proxy").Logger(),

		client:       client,
		instanceName: instanceName,
		handlers:     &prefixHandler{},
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
				filters.Arg("label", fmt.Sprintf("m1k1o.neko_rooms.instance=%s", p.instanceName)),
				filters.Arg("label", "m1k1o.neko_rooms.proxy=true"),
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
				path, port, ok := p.parseLabels(msg.Actor.Attributes)
				if !ok {
					break
				}

				p.logger.Info().
					Str("action", msg.Action).
					Str("path", path).
					Str("port", port).
					Msg("new docker event")

				p.mu.Lock()
				switch msg.Action {
				case "create":
					p.handlers.Set(path, nil)
				case "start":
					proxy := httputil.NewSingleHostReverseProxy(&url.URL{
						Scheme: "http",
						Host:   msg.ID[:12] + ":" + port,
					})

					p.handlers.Set(path, proxy)
				case "stop":
					p.handlers.Set(path, nil)
				case "destroy":
					p.handlers.Remove(path)
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
			filters.Arg("label", fmt.Sprintf("m1k1o.neko_rooms.instance=%s", p.instanceName)),
		),
	})

	if err != nil {
		return err
	}

	p.handlers = &prefixHandler{}

	for _, cont := range containers {
		path, port, ok := p.parseLabels(cont.Labels)
		if !ok {
			continue
		}

		var proxy *httputil.ReverseProxy
		if cont.State == "running" {
			proxy = httputil.NewSingleHostReverseProxy(&url.URL{
				Scheme: "http",
				Host:   cont.ID[:12] + ":" + port,
			})
		}

		p.handlers.Set(path, proxy)
	}

	return nil
}

func (p *ProxyManagerCtx) parseLabels(labels map[string]string) (path string, port string, ok bool) {
	path, ok = labels["m1k1o.neko_rooms.proxy.path"]
	if !ok {
		return
	}

	port, ok = labels["m1k1o.neko_rooms.proxy.port"]
	return
}

func (p *ProxyManagerCtx) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get proxy by room name
	p.mu.RLock()
	proxy, prefix, ok := p.handlers.Match(r.URL.Path)
	p.mu.RUnlock()

	// if room not found
	if !ok {
		RoomNotFound(w, r)
		return
	}

	// redirect to room ending with /
	if r.URL.Path == prefix {
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

		p.logger.Err(err).Str("prefix", prefix).Msg("proxying error")
	}

	// strip prefix from proxy
	handler := http.StripPrefix(prefix, proxy)

	// handle by proxy
	handler.ServeHTTP(w, r)
}
