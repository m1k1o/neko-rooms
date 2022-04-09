package server

import (
	"context"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/m1k1o/neko-rooms/internal/config"
	"github.com/m1k1o/neko-rooms/internal/types"
)

type ServerManagerCtx struct {
	logger zerolog.Logger
	router *chi.Mux
	server *http.Server
	config *config.Server
}

func New(ApiManager types.ApiManager, pathPrefix string, config *config.Server) *ServerManagerCtx {
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

	// Basic CORS
	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	router.Route("/api", ApiManager.Mount)

	// serve static files
	if config.Static != "" {
		fs := http.FileServer(http.Dir(config.Static))
		router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			if _, err := os.Stat(config.Static + r.URL.Path); !os.IsNotExist(err) {
				fs.ServeHTTP(w, r)
			} else {
				http.NotFound(w, r)
			}
		})
	}

	// add simple lobby room
	router.Get(path.Join("/", pathPrefix, "{roomName}"), ApiManager.RoomLobby)
	router.Get(path.Join("/", pathPrefix, "{roomName}")+"/", ApiManager.RoomLobby)

	// mount pprof endpoint
	if config.PProf {
		withPProf(router)
		logger.Info().Msgf("with pprof endpoint at %s", pprofPath)
	}

	// we could use custom 404
	router.NotFound(http.NotFound)

	return &ServerManagerCtx{
		logger: logger,
		router: router,
		server: &http.Server{
			Addr:    config.Bind,
			Handler: router,
		},
		config: config,
	}
}

func (s *ServerManagerCtx) Start() {
	if s.config.Cert != "" && s.config.Key != "" {
		go func() {
			if err := s.server.ListenAndServeTLS(s.config.Cert, s.config.Key); err != http.ErrServerClosed {
				s.logger.Panic().Err(err).Msg("unable to start https server")
			}
		}()
		s.logger.Info().Msgf("https listening on %s", s.server.Addr)
	} else {
		go func() {
			if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
				s.logger.Panic().Err(err).Msg("unable to start http server")
			}
		}()
		s.logger.Info().Msgf("http listening on %s", s.server.Addr)
	}
}

func (s *ServerManagerCtx) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}
