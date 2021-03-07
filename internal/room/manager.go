package room

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/docker/docker/client"

	"m1k1o/neko_rooms/internal/types"
	"m1k1o/neko_rooms/internal/config"
)

func New(config *config.Room) *RoomManagerCtx {
	logger := log.With().Str("module", "room").Logger()

	cli, err := client.NewEnvClient()
	if err != nil {
		logger.Panic().Err(err).Msg("unable to connect to docker client")
	} else {
		logger.Info().Msg("successfully connected to docker client")
	}

	return &RoomManagerCtx{
		logger: logger,
		config: config,
		client: cli,
	}
}

type RoomManagerCtx struct {
	logger zerolog.Logger
	config *config.Room
	client *client.Client
}

func (manager *RoomManagerCtx) List() []types.RoomData {
	return []types.RoomData{}
}

func (manager *RoomManagerCtx) Create(settings types.RoomSettings) (string, error) {
	return "id", nil
}

func (manager *RoomManagerCtx) Update(id string, settings types.RoomSettings) error {
	return nil
}

func (manager *RoomManagerCtx) Remove(id string) error {
	return nil
}
