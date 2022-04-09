package server

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/m1k1o/neko-rooms/internal/config"
)

type Manager struct {
	logger zerolog.Logger
	config *config.Server
	router *chi.Mux
	server *http.Server
}

func New(config *config.Server) *Manager {
	logger := log.With().Str("module", "server").Logger()

	router := chi.NewRouter()
	router.Use(middleware.RequestID) // Create a request ID for each request

	// get real users ip
	if config.Proxy {
		router.Use(middleware.RealIP)
	}

	// add http logger
	router.Use(middleware.RequestLogger(&logformatter{logger}))
	router.Use(middleware.Recoverer) // Recover from panics without crashing server

	// serve static files
	if config.Static != "" {
		fs := http.FileServer(http.Dir(config.Static))
		router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			if _, err := os.Stat(config.Static + r.RequestURI); os.IsNotExist(err) {
				http.StripPrefix(r.RequestURI, fs).ServeHTTP(w, r)
			} else {
				fs.ServeHTTP(w, r)
			}
		})
	}

	// mount pprof endpoint
	if config.PProf {
		withPProf(router)
		logger.Info().Msgf("with pprof endpoint at %s", pprofPath)
	}

	// we could use custom 404
	router.NotFound(http.NotFound)

	return &Manager{
		logger: logger,
		config: config,
		router: router,
		server: &http.Server{
			Addr:    config.Bind,
			Handler: router,
		},
	}
}

func (s *Manager) Start() {
	if s.config.SSLCert != "" && s.config.SSLKey != "" {
		go func() {
			if err := s.server.ListenAndServeTLS(s.config.SSLCert, s.config.SSLKey); err != http.ErrServerClosed {
				s.logger.Panic().Err(err).Msg("unable to start https server")
			}
		}()
		s.logger.Info().Msgf("https server listening on %s", s.server.Addr)
	} else {
		go func() {
			if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
				s.logger.Panic().Err(err).Msg("unable to start http server")
			}
		}()
		s.logger.Info().Msgf("http server listening on %s", s.server.Addr)
	}
}

func (s *Manager) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}

func (s *Manager) Mount(fn func(r *chi.Mux)) {
	fn(s.router)
}

func (s *Manager) Handle(pattern string, fn http.Handler) {
	s.router.Handle(pattern, fn)
}
