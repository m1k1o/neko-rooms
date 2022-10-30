package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
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
	id      string
	running bool
	handler *httputil.ReverseProxy
}

type wait struct {
	subs   int
	signal chan struct{}
}

type ProxyManagerCtx struct {
	logger zerolog.Logger
	mu     sync.RWMutex
	ctx    context.Context
	cancel func()

	waitMu      sync.RWMutex
	waitChans   map[string]*wait
	waitEnabled bool

	client       *dockerClient.Client
	instanceName string
	handlers     *prefixHandler[*entry]
}

func New(client *dockerClient.Client, instanceName string, waitEnabled bool) *ProxyManagerCtx {
	return &ProxyManagerCtx{
		logger:    log.With().Str("module", "proxy").Logger(),
		waitChans: map[string]*wait{},

		client:       client,
		instanceName: instanceName,
		waitEnabled:  waitEnabled,
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
			case err, ok := <-errs:
				if !ok {
					p.logger.Fatal().Msg("docker event error channel closed")
					return
				}

				p.logger.Err(err).Msg("got docker event error")
			case msg, ok := <-msgs:
				if !ok {
					p.logger.Fatal().Msg("docker event channel closed")
					return
				}

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

				// terminate waiting for any events
				if p.waitEnabled {
					p.waitMu.Lock()
					ch, ok := p.waitChans[path]
					if ok {
						close(ch.signal)
						delete(p.waitChans, path)
					}
					p.waitMu.Unlock()
				}

				p.mu.Lock()
				switch msg.Action {
				case "create":
					p.handlers.Set(path, &entry{
						id:      msg.ID,
						running: false,
					})
				case "start":
					e := &entry{
						id:      msg.ID,
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
						id:      msg.ID,
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

		entry := &entry{
			id: cont.ID,
		}

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

func (p *ProxyManagerCtx) waitForPath(w http.ResponseWriter, r *http.Request, path string) {
	p.logger.Debug().Str("path", path).Msg("adding new wait handler")

	p.waitMu.Lock()
	ch, ok := p.waitChans[path]
	if !ok {
		ch = &wait{
			subs:   1,
			signal: make(chan struct{}),
		}
		p.waitChans[path] = ch
	} else {
		p.waitChans[path].subs += 1
	}
	p.waitMu.Unlock()

	select {
	case <-ch.signal:
		w.Write([]byte("ready"))
	case <-r.Context().Done():
		http.Error(w, r.Context().Err().Error(), http.StatusRequestTimeout)

		p.waitMu.Lock()
		ch.subs -= 1
		if ch.subs <= 0 {
			delete(p.waitChans, path)
		}
		p.waitMu.Unlock()
		p.logger.Debug().Str("path", path).Msg("wait handler removed")
	case <-p.ctx.Done():
		w.Write([]byte("shutdown"))
	}
}

func (p *ProxyManagerCtx) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cleanPath := path.Clean(r.URL.Path)

	// get proxy by room name
	p.mu.RLock()
	proxy, prefix, ok := p.handlers.Match(cleanPath)
	p.mu.RUnlock()

	// if room not found
	if !ok {
		// blocking until room is created
		if r.URL.Query().Has("wait") && p.waitEnabled {
			p.waitForPath(w, r, cleanPath)
			return
		}

		RoomNotFound(w, r, p.waitEnabled)
		return
	}

	// redirect to room ending with /
	if cleanPath == prefix && !strings.HasSuffix(r.URL.Path, "/") {
		r.URL.Path = cleanPath + "/"
		http.Redirect(w, r, r.URL.String(), http.StatusTemporaryRedirect)
		return
	}

	// if room not running
	if !proxy.running {
		// blocking until room is running
		if r.URL.Query().Has("wait") && p.waitEnabled {
			p.waitForPath(w, r, cleanPath)
			return
		}

		RoomNotRunning(w, r, p.waitEnabled)
		return
	}

	// if not proxying, check for container readiness status
	if proxy.handler == nil {
		containers, err := p.client.ContainerList(p.ctx, dockerTypes.ContainerListOptions{
			All: true,
			Filters: filters.NewArgs(
				filters.Arg("label", fmt.Sprintf("m1k1o.neko_rooms.instance=%s", p.instanceName)),
				filters.Arg("id", proxy.id),
			),
		})

		if err != nil || len(containers) != 1 {
			p.logger.Err(err).Msg("error while getting container ready status")
			RoomNotReady(w, r)
			return
		}

		container := containers[0]
		if strings.Contains(container.Status, "starting") {
			RoomNotReady(w, r)
		} else {
			RoomReady(w, r)
		}

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
