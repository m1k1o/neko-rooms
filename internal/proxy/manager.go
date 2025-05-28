package proxy

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/m1k1o/neko-rooms/internal/room"
	"github.com/m1k1o/neko-rooms/internal/types"
	"github.com/m1k1o/neko-rooms/pkg/prefix"
)

type entry struct {
	id      string
	running bool
	ready   bool
	paused  bool
	handler http.Handler
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

	rooms    *room.RoomManagerCtx
	handlers prefix.Tree[*entry]
}

func New(rooms *room.RoomManagerCtx, waitEnabled bool) *ProxyManagerCtx {
	return &ProxyManagerCtx{
		logger:    log.With().Str("module", "proxy").Logger(),
		waitChans: map[string]*wait{},

		rooms:       rooms,
		waitEnabled: waitEnabled,
		handlers:    prefix.NewTree[*entry](),
	}
}

func (p *ProxyManagerCtx) Start() {
	p.ctx, p.cancel = context.WithCancel(context.Background())

	go func() {
		err := p.Refresh()
		if err != nil {
			p.logger.Err(err).Msg("unable to refresh containers")
		}

		msgs, errs := p.rooms.Events(p.ctx)

		for {
			select {
			case err, ok := <-errs:
				if !ok {
					return
				}

				p.logger.Err(err).Msg("room event error")
			case msg, ok := <-msgs:
				enabled, path, port, ok := p.parseLabels(msg.ContainerLabels)
				if !ok {
					break
				}

				host := msg.ID + ":" + port

				p.logger.Info().
					Str("action", string(msg.Action)).
					Str("path", path).
					Str("host", host).
					Msg("got room event")

				// terminate waiting for room ready event
				if p.waitEnabled && msg.Action == types.RoomEventReady {
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
				case types.RoomEventCreated:
					p.handlers.Insert(path, &entry{
						id:      msg.ID,
						running: false,
					})
				case types.RoomEventStarted:
					p.handlers.Insert(path, &entry{
						id:      msg.ID,
						running: true,
						ready:   false,
					})
				case types.RoomEventReady:
					e := &entry{
						id:      msg.ID,
						running: true,
						ready:   true,
					}

					// if proxying is disabled
					if enabled {
						e.handler = p.newProxyHandler(path, host)
					}

					p.handlers.Insert(path, e)
				case types.RoomEventStopped:
					p.handlers.Insert(path, &entry{
						id:      msg.ID,
						running: false,
					})
				case types.RoomEventPaused:
					p.handlers.Insert(path, &entry{
						id:      msg.ID,
						running: false,
						paused:  true,
					})
				case types.RoomEventDestroyed:
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

	rooms, err := p.rooms.List(p.ctx, nil)
	if err != nil {
		return err
	}

	p.handlers = prefix.NewTree[*entry]()

	for _, room := range rooms {
		enabled, path, port, ok := p.parseLabels(room.ContainerLabels)
		if !ok {
			continue
		}

		host := room.ID + ":" + port

		entry := &entry{
			id:      room.ID,
			running: room.Running,
			ready:   room.IsReady,
			paused:  room.Paused,
		}

		// if proxying is enabled and room is ready
		if enabled && room.IsReady {
			entry.handler = p.newProxyHandler(path, host)
		}

		p.handlers.Insert(path, entry)
	}

	return nil
}

func (p *ProxyManagerCtx) parseLabels(labels map[string]string) (enabled bool, path string, port string, ok bool) {
	var enabledStr string
	enabledStr, ok = labels["m1k1o.neko_rooms.proxy.enabled"]
	if !ok {
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

func (p *ProxyManagerCtx) newProxyHandler(prefix, host string) http.Handler {
	handler := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   host,
	})
	handler.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		p.logger.Err(err).Str("prefix", prefix).Msg("proxy error")
		http.Error(w, "unable to connect to room", http.StatusBadGateway)
	}
	return http.StripPrefix(prefix, handler)
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

	// if room is not ready
	if !ok || !proxy.running || !proxy.ready {
		// blocking until room is ready
		if r.URL.Query().Has("wait") && p.waitEnabled {
			p.waitForPath(w, r, cleanPath)
			return
		}

		if !ok {
			RoomNotFound(w, r, p.waitEnabled)
		} else if proxy.paused {
			RoomPaused(w, r, p.waitEnabled)
		} else if !proxy.running {
			RoomNotRunning(w, r, p.waitEnabled)
		} else {
			RoomNotReady(w, r, p.waitEnabled)
		}
		return
	}

	// redirect to room ending with /
	if cleanPath == prefix && !strings.HasSuffix(r.URL.Path, "/") {
		r.URL.Path = cleanPath + "/"
		http.Redirect(w, r, r.URL.String(), http.StatusTemporaryRedirect)
		return
	}

	// if not proxying, just return room ready
	if proxy.handler == nil {
		RoomReady(w, r)
		return
	}

	// handle by proxy
	proxy.handler.ServeHTTP(w, r)
}
