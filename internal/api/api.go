package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/m1k1o/neko-rooms/internal/types"
)

type ApiManagerCtx struct {
	logger zerolog.Logger
	rooms  types.RoomManager
	pull   types.PullManager
}

func New(rooms types.RoomManager, pull types.PullManager) *ApiManagerCtx {
	return &ApiManagerCtx{
		logger: log.With().Str("module", "api").Logger(),
		rooms:  rooms,
		pull:   pull,
	}
}

func (manager *ApiManagerCtx) Mount(r chi.Router) {
	//
	// config
	//

	r.Get("/config/rooms", manager.configRooms)

	//
	// pull
	//

	r.Route("/pull", func(r chi.Router) {
		r.Get("/", manager.pullStatus)
		r.Get("/sse", manager.pullStatusSSE)
		r.Post("/", manager.pullStart)
		r.Delete("/", manager.pullStop)
	})

	//
	// rooms
	//

	r.Get("/rooms", manager.roomsList)
	r.Post("/rooms", manager.roomCreate)

	r.Route("/rooms/{roomId}", func(r chi.Router) {
		r.Get("/", manager.roomGetEntry)
		r.Get("/by-name", manager.roomGetEntryByName)

		r.Get("/settings", manager.roomGetSettings)
		r.Get("/stats", manager.roomGetStats)

		r.Delete("/", manager.roomGenericAction(manager.rooms.Remove))
		r.Post("/start", manager.roomGenericAction(manager.rooms.Start))
		r.Post("/stop", manager.roomGenericAction(manager.rooms.Stop))
		r.Post("/restart", manager.roomGenericAction(manager.rooms.Restart))
		r.Post("/pause", manager.roomGenericAction(manager.rooms.Pause))
		r.Post("/recreate", manager.roomRecreate)
	})

	r.Get("/docker-compose.yaml", manager.dockerCompose)

	//
	// events
	//

	r.Get("/events", manager.events)
}
