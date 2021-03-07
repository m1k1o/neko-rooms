package room

import (
	"context"
	"fmt"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	network "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	dockerClient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"m1k1o/neko_rooms/internal/config"
	"m1k1o/neko_rooms/internal/types"
)

const (
	nekoImage       = "m1k1o/neko:latest"
	containerPrefix = "neko-room-"
	frontendPort    = 8080
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
	containers, err := manager.client.ContainerList(context.Background(), dockerTypes.ContainerListOptions{
		All: true,
	})
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
	// configs
	pathName := "foobar"
	eprStart := uint(52135)
	eprEnd := uint(52145)

	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{
		nat.Port(fmt.Sprintf("%d/udp", frontendPort)): struct{}{},
	}

	for port := eprStart; port <= eprEnd; port++ {
		portKey := nat.Port(fmt.Sprintf("%d/udp", port))

		portBindings[portKey] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: fmt.Sprintf("%d", port),
			},
		}

		exposedPorts[portKey] = struct{}{}
	}

	containerName := containerPrefix + pathName

	labels := map[string]string{
		// Set internal labels
		"m1k1o.neko_rooms":     "true",
		"m1k1o.neko_rooms.epr": fmt.Sprintf("%d-%d", eprStart, eprEnd),

		// Set traefik labels
		"traefik.enable": "true",
		"traefik.http.services." + containerName + "-frontend.loadbalancer.server.port": fmt.Sprintf("%d", frontendPort),
		"traefik.http.routers." + containerName + ".entrypoints":                        manager.config.TraefikEntrypoint,
		"traefik.http.routers." + containerName + ".rule":                               "Host(`" + manager.config.TraefikDomain + "`) && PathPrefix(`/" + pathName + "`)",
		"traefik.http.middlewares." + containerName + "-rdr.redirectregex.regex":        "/" + pathName + "$$",
		"traefik.http.middlewares." + containerName + "-rdr.redirectregex.replacement":  "/" + pathName + "/",
		"traefik.http.middlewares." + containerName + "-prf.stripprefix.prefixes":       "/" + pathName + "/",
		"traefik.http.routers." + containerName + ".middlewares":                        containerName + "-rdr," + containerName + "-prf",
	}

	// optional HTTPS
	if manager.config.TraefikCertresolver != "" {
		labels["traefik.http.routers."+containerName+".tls"] = "true"
		labels["traefik.http.routers."+containerName+".tls.certresolver"] = manager.config.TraefikCertresolver
	}

	config := &container.Config{
		// Hostname
		Hostname: containerName,
		// Domainname
		Domainname: containerName,
		// List of exposed ports
		ExposedPorts: exposedPorts,
		// List of environment variable to set in the container
		Env: append([]string{
			fmt.Sprintf("NEKO_BIND=%d", frontendPort),
		}, settings.Env(eprStart, eprEnd, manager.config.NAT1To1IPs)...),
		// Name of the image as it was passed by the operator (e.g. could be symbolic)
		Image: nekoImage,
		// List of labels set to this container
		Labels: labels,
	}

	hostConfig := &container.HostConfig{
		// Port mapping between the exposed port (container) and the host
		PortBindings: portBindings,
		// Configuration of the logs for this container
		LogConfig: container.LogConfig{
			Type:   "json-file",
			Config: map[string]string{},
		},
		// Restart policy to be used for the container
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		// List of kernel capabilities to add to the container
		CapAdd: strslice.StrSlice{
			"SYS_ADMIN",
		},
		// Total shm memory usage
		ShmSize: 2 * 10e9,
	}

	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			manager.config.TraefikNetwork: &network.EndpointSettings{},
		},
	}

	// Creating the actual container
	cont, err := manager.client.ContainerCreate(
		context.Background(),
		config,
		hostConfig,
		networkingConfig,
		nil,
		containerName,
	)

	if err != nil {
		return nil, err
	}

	// Run the actual container
	err = manager.client.ContainerStart(context.Background(), cont.ID, dockerTypes.ContainerStartOptions{})

	if err != nil {
		return nil, err
	}

	return &types.RoomData{
		ID:           cont.ID,
		RoomSettings: settings,
	}, nil
}

func (manager *RoomManagerCtx) Get(id string) (*types.RoomData, error) {
	return &types.RoomData{
		ID:           "foo",
		RoomSettings: types.RoomSettings{},
	}, nil
}

func (manager *RoomManagerCtx) Update(id string, settings types.RoomSettings) error {
	return nil
}

func (manager *RoomManagerCtx) Remove(id string) error {
	return nil
}
