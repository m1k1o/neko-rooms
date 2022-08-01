package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"sync"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerClient "github.com/docker/docker/client"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type entry struct {
	running bool
	handler *httputil.ReverseProxy
}

type ProxyManagerCtx struct {
	logger zerolog.Logger
	mu     sync.RWMutex
	ctx    context.Context
	cancel func()

	client       *dockerClient.Client
	instanceName string
	handlers     *prefixHandler[*entry]
}

func New(client *dockerClient.Client, instanceName string) *ProxyManagerCtx {
	return &ProxyManagerCtx{
		logger: log.With().Str("module", "proxy").Logger(),

		client:       client,
		instanceName: instanceName,
		handlers:     &prefixHandler[*entry]{},
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
				enabled, path, port, ok := p.parseLabels(msg.Actor.Attributes)
				if !ok {
					break
				}

				host := msg.ID[:12] + ":" + port

				p.logger.Info().
					Str("action", msg.Action).
					Str("path", path).
					Str("host", host).
					Msg("got docker event")

				p.mu.Lock()
				switch msg.Action {
				case "create":
					p.handlers.Set(path, &entry{
						running: false,
					})
				case "start":
					e := &entry{
						running: true,
					}

					// if proxying is disabled
					if enabled {
						e.handler = httputil.NewSingleHostReverseProxy(&url.URL{
							Scheme: "http",
							Host:   host,
						})
					}

					p.handlers.Set(path, e)
				case "stop":
					p.handlers.Set(path, &entry{
						running: false,
					})
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

	p.handlers = &prefixHandler[*entry]{}

	for _, cont := range containers {
		enabled, path, port, ok := p.parseLabels(cont.Labels)
		if !ok {
			continue
		}

		host := cont.ID[:12] + ":" + port

		entry := &entry{}
		if cont.State == "running" {
			entry.running = true

			// if proxying is disabled
			if enabled {
				entry.handler = httputil.NewSingleHostReverseProxy(&url.URL{
					Scheme: "http",
					Host:   host,
				})
			}
		}

		p.handlers.Set(path, entry)
	}

	return nil
}

func (p *ProxyManagerCtx) parseLabels(labels map[string]string) (enabled bool, path string, port string, ok bool) {
	enabledStr, found := labels["m1k1o.neko_rooms.proxy.enabled"]
	if !found {
		//
		// workaround for legacy traefik labels
		//

		// get room name
		var roomName string
		roomName, ok = labels["m1k1o.neko_rooms.name"]
		if !ok {
			return
		}

		// get container name
		containerName := p.instanceName + "-" + roomName

		// get path
		path, ok = labels["traefik.http.middlewares."+containerName+"-prf.stripprefix.prefixes"]
		if !ok {
			return
		}

		// remove last /
		path = strings.TrimSuffix(path, "/")

		// get port
		port, ok = labels["traefik.http.services."+containerName+"-frontend.loadbalancer.server.port"]
		if !ok {
			return
		}

		//
		// workaround for legacy traefik labels
		//

		return
	}

	enabledBool, err := strconv.ParseBool(enabledStr)
	enabled = enabledBool && err == nil

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
	if !proxy.running {
		RoomNotRunning(w, r)
		return
	}

	// if not proxying
	if proxy.handler == nil {
		RoomReady(w, r)
		return
	}

	// handle not ready room
	proxy.handler.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		if r.URL.Path == "/" {
			RoomNotReady(w, r)
		}

		p.logger.Err(err).Str("prefix", prefix).Msg("proxying error")
	}

	// strip prefix from proxy
	handler := http.StripPrefix(prefix, proxy.handler)

	// handle by proxy
	handler.ServeHTTP(w, r)
}
