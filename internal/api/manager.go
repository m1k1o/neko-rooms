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
	r.Get("/config/rooms", manager.roomsConfig)

	r.Get("/rooms", manager.roomsList)
	r.Post("/rooms", manager.roomCreate)

	r.Route("/rooms/{roomId}", func(r chi.Router) {
		r.Get("/", manager.roomGetEntry)
		r.Delete("/", manager.roomRemove)

		r.Get("/settings", manager.roomGetSettings)
		r.Get("/stats", manager.roomGetStats)

		r.Post("/start", manager.roomGenericAction(manager.rooms.Start))
		r.Post("/stop", manager.roomGenericAction(manager.rooms.Stop))
		r.Post("/restart", manager.roomGenericAction(manager.rooms.Restart))
	})
}
