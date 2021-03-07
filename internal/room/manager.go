package room

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	dockerTypes "github.com/docker/docker/api/types"
	dockerClient "github.com/docker/docker/client"

	"m1k1o/neko_rooms/internal/types"
	"m1k1o/neko_rooms/internal/config"
)

func New(config *config.Room) *RoomManagerCtx {
	logger := log.With().Str("module", "room").Logger()

	cli, err := dockerClient.NewEnvClient()
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
	client *dockerClient.Client
}

func (manager *RoomManagerCtx) List() (*[]types.RoomData, error) {
	containers, err := manager.client.ContainerList(context.Background(), dockerTypes.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	for _, container := range containers {
		manager.logger.Info().
			Str("id", container.ID[:10]).
			Str("image", container.Image).
			Msg("container")
	}

	return &[]types.RoomData{}, nil
}

func (manager *RoomManagerCtx) Create(settings types.RoomSettings) (*types.RoomData, error) {
	return &types.RoomData{
		ID: "foo",
		RoomSettings: settings,
	}, nil
}

func (manager *RoomManagerCtx) Get(id string) (*types.RoomData, error) {
	return &types.RoomData{
		ID: "foo",
		RoomSettings: types.RoomSettings{},
	}, nil
}

func (manager *RoomManagerCtx) Update(id string, settings types.RoomSettings) error {
	return nil
}

func (manager *RoomManagerCtx) Remove(id string) error {
	return nil
}
