package api

import (
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"m1k1o/neko_rooms/internal/config"
)

type ApiManagerCtx struct {
	logger zerolog.Logger
	conf   *config.API
}

func New(conf *config.API) *ApiManagerCtx {
	return &ApiManagerCtx{
		logger: log.With().Str("module", "router").Logger(),
		conf:   conf,
	}
}

func (a *ApiManagerCtx) Mount(r chi.Router) {

}
