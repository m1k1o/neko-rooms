package server

import (
	"context"
	"net/http"
	"path"
	"strings"
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

func New(ApiManager types.ApiManager, roomConfig *config.Room, config *config.Server, proxyHandler http.Handler) *ServerManagerCtx {
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

	// mount pprof endpoint
	if config.PProf {
		withPProf(router)
		logger.Info().Msgf("with pprof endpoint at %s", pprofPath)
	}

	//
	// rooms page
	//

	router.Route(roomConfig.PathPrefix, func(r chi.Router) {
		if !roomConfig.Traefik.Enabled {
			r.Handle("/*", proxyHandler)
		} else {
			// add simple lobby room
			r.Get("/{roomName}", ApiManager.RoomLobby)
			r.Get("/{roomName}/", ApiManager.RoomLobby)
		}
	})

	//
	// admin page
	//

	// in v1 default location was at / with traefik overriding
	// the actual room address. in order to keep this setting
	// we set new default path prefix only without traefik
	if !roomConfig.Traefik.Enabled && config.Admin.PathPrefix == "" {
		config.Admin.PathPrefix = "/admin"
	}

	router.Group(func(r chi.Router) {
		// handle authorization
		if config.Admin.Password != "" {
			r.Use(middleware.BasicAuth("neko-rooms admin", map[string]string{
				"admin": config.Admin.Password,
			}))
		}

		// bind API
		apiPath := path.Join("/", config.Admin.PathPrefix, "/api")
		router.Route(apiPath, ApiManager.Mount)

		// serve static files
		if config.Admin.Static != "" {
			prefix := config.Admin.PathPrefix
			dir := http.Dir(config.Admin.Static)

			fs := http.FileServer(dir)
			fs = http.StripPrefix(prefix, fs)

			router.Handle(prefix+"/", fs)
		}
	})

	// redirect / to admin path prefix
	if config.Admin.PathPrefix != "/" {
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, strings.TrimPrefix(config.Admin.PathPrefix, "/"), http.StatusTemporaryRedirect)
		})
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
