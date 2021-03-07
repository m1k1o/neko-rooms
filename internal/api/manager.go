package api

import (
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"m1k1o/neko_rooms/internal/config"
	"m1k1o/neko_rooms/internal/types"
)

type ApiManagerCtx struct {
	logger zerolog.Logger
	rooms  types.RoomManager
	conf   *config.API
}

func New(roomManager types.RoomManager, conf *config.API) *ApiManagerCtx {
	return &ApiManagerCtx{
		logger: log.With().Str("module", "router").Logger(),
		rooms:  roomManager,
		conf:   conf,
	}
}

func (manager *ApiManagerCtx) Mount(r chi.Router) {
	r.Get("/rooms", manager.roomsList)
	r.Post("/rooms", manager.roomCreate)

	r.Route("/rooms/{roomId}", func(r chi.Router) {
		r.Get("/", manager.roomGet)
		r.Post("/", manager.roomUpdate)
		r.Delete("/", manager.roomRemove)
	})
}
