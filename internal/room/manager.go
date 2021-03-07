package room

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"m1k1o/neko_rooms/internal/config"
)

func New(config *config.Room) *RoomManagerCtx {
	logger := log.With().Str("module", "room").Logger()

	return &RoomManagerCtx{
		logger: logger,
		config: config,
	}
}

type RoomManagerCtx struct {
	logger zerolog.Logger
	config *config.Room
}
