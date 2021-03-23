package room

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
	"m1k1o/neko_rooms/internal/utils"
)

const (
	frontendPort = 8080
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

func (manager *RoomManagerCtx) Config() types.RoomsConfig {
	return types.RoomsConfig{
		Connections: manager.config.EprMax - manager.config.EprMin + 1,
		NekoImages:  manager.config.NekoImages,
	}
}

func (manager *RoomManagerCtx) List() ([]types.RoomEntry, error) {
	containers, err := manager.listContainers()
	if err != nil {
		return nil, err
	}

	result := []types.RoomEntry{}
	for _, container := range containers {
		entry, err := manager.containerToEntry(container)
		if err != nil {
			return nil, err
		}

		result = append(result, *entry)
	}

	return result, nil
}

func (manager *RoomManagerCtx) Create(settings types.RoomSettings) (string, error) {
	if in, _ := utils.ArrayIn(settings.NekoImage, manager.config.NekoImages); !in {
		return "", fmt.Errorf("invalid neko image")
	}

	// TODO: Check if path name exists.
	roomName := settings.Name
	if roomName == "" {
		var err error
		roomName, err = utils.NewUID(32)
		if err != nil {
			return "", err
		}
	}

	epr, err := manager.allocatePorts(settings.MaxConnections)
	if err != nil {
		return "", err
	}

	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{
		nat.Port(fmt.Sprintf("%d/udp", frontendPort)): struct{}{},
	}

	for port := epr.Min; port <= epr.Max; port++ {
		portKey := nat.Port(fmt.Sprintf("%d/udp", port))

		portBindings[portKey] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: fmt.Sprintf("%d", port),
			},
		}

		exposedPorts[portKey] = struct{}{}
	}

	containerName := manager.config.InstanceName + "-" + roomName

	urlProto := "http"
	if manager.config.TraefikCertresolver != "" {
		urlProto = "https"
	}

	port := ""
	if manager.config.TraefikPort != "" {
		port = ":" + manager.config.TraefikPort
	}

	labels := map[string]string{
		// Set internal labels
		"m1k1o.neko_rooms.name":       roomName,
		"m1k1o.neko_rooms.url":        urlProto + "://" + manager.config.TraefikDomain + port + "/" + roomName + "/",
		"m1k1o.neko_rooms.instance":   manager.config.InstanceName,
		"m1k1o.neko_rooms.epr.min":    fmt.Sprintf("%d", epr.Min),
		"m1k1o.neko_rooms.epr.max":    fmt.Sprintf("%d", epr.Max),
		"m1k1o.neko_rooms.neko_image": settings.NekoImage,

		// Set traefik labels
		"traefik.enable": "true",
		"traefik.http.services." + containerName + "-frontend.loadbalancer.server.port": fmt.Sprintf("%d", frontendPort),
		"traefik.http.routers." + containerName + ".entrypoints":                        manager.config.TraefikEntrypoint,
		"traefik.http.routers." + containerName + ".rule":                               "Host(`" + manager.config.TraefikDomain + "`) && PathPrefix(`/" + roomName + "`)",
		"traefik.http.middlewares." + containerName + "-rdr.redirectregex.regex":        "/" + roomName + "$$",
		"traefik.http.middlewares." + containerName + "-rdr.redirectregex.replacement":  "/" + roomName + "/",
		"traefik.http.middlewares." + containerName + "-prf.stripprefix.prefixes":       "/" + roomName + "/",
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
			fmt.Sprintf("NEKO_BIND=:%d", frontendPort),
			fmt.Sprintf("NEKO_EPR=%d-%d", epr.Min, epr.Max),
			fmt.Sprintf("NEKO_NAT1TO1=%s", strings.Join(manager.config.NAT1To1IPs, ",")),
			"NEKO_ICELITE=true",
		}, settings.ToEnv()...),
		// Name of the image as it was passed by the operator (e.g. could be symbolic)
		Image: settings.NekoImage,
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
	container, err := manager.client.ContainerCreate(
		context.Background(),
		config,
		hostConfig,
		networkingConfig,
		nil,
		containerName,
	)

	if err != nil {
		return "", err
	}

	// Start the actual container
	err = manager.client.ContainerStart(context.Background(), container.ID, dockerTypes.ContainerStartOptions{})

	if err != nil {
		return "", err
	}

	return container.ID, nil
}

func (manager *RoomManagerCtx) GetEntry(id string) (*types.RoomEntry, error) {
	container, err := manager.containerInfo(id)
	if err != nil {
		return nil, err
	}

	return manager.containerToEntry(*container)
}

func (manager *RoomManagerCtx) Remove(id string) error {
	_, err := manager.inspectContainer(id)
	if err != nil {
		return err
	}

	// Stop the actual container
	err = manager.client.ContainerStop(context.Background(), id, nil)

	if err != nil {
		return err
	}

	// Remove the actual container
	err = manager.client.ContainerRemove(context.Background(), id, dockerTypes.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})

	return err
}

func (manager *RoomManagerCtx) GetSettings(id string) (*types.RoomSettings, error) {
	container, err := manager.inspectContainer(id)
	if err != nil {
		return nil, err
	}

	roomName, ok := container.Config.Labels["m1k1o.neko_rooms.name"]
	if !ok {
		return nil, fmt.Errorf("Damaged container labels: name not found.")
	}

	nekoImage, ok := container.Config.Labels["m1k1o.neko_rooms.neko_image"]
	if !ok {
		return nil, fmt.Errorf("Damaged container labels: neko_image not found.")
	}

	epr, err := manager.getEprFromLabels(container.Config.Labels)
	if err != nil {
		return nil, err
	}

	settings := types.RoomSettings{
		Name:           roomName,
		NekoImage:      nekoImage,
		MaxConnections: epr.Max - epr.Min + 1,
	}

	err = settings.FromEnv(container.Config.Env)
	return &settings, err
}

func (manager *RoomManagerCtx) GetStats(id string) (*types.RoomStats, error) {
	container, err := manager.inspectContainer(id)
	if err != nil {
		return nil, err
	}

	settings := types.RoomSettings{}
	err = settings.FromEnv(container.Config.Env)
	if err != nil {
		return nil, err
	}

	output, err := manager.containerExec(id, []string{
		"wget", "-q", "-O-", "http://127.0.0.1:8080/stats?pwd=" + settings.AdminPass,
	})
	if err != nil {
		return nil, err
	}

	var stats types.RoomStats
	if err := json.Unmarshal([]byte(output), &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

func (manager *RoomManagerCtx) Start(id string) error {
	_, err := manager.inspectContainer(id)
	if err != nil {
		return err
	}

	// Start the actual container
	return manager.client.ContainerStart(context.Background(), id, dockerTypes.ContainerStartOptions{})
}

func (manager *RoomManagerCtx) Stop(id string) error {
	_, err := manager.inspectContainer(id)
	if err != nil {
		return err
	}

	// Stop the actual container
	return manager.client.ContainerStop(context.Background(), id, nil)
}

func (manager *RoomManagerCtx) Restart(id string) error {
	_, err := manager.inspectContainer(id)
	if err != nil {
		return err
	}

	// Restart the actual container
	return manager.client.ContainerRestart(context.Background(), id, nil)
}
