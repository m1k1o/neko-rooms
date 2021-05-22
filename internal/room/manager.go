package room

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerMount "github.com/docker/docker/api/types/mount"
	network "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	dockerClient "github.com/docker/docker/client"
	dockerNames "github.com/docker/docker/daemon/names"
	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"m1k1o/neko_rooms/internal/config"
	"m1k1o/neko_rooms/internal/types"
	"m1k1o/neko_rooms/internal/utils"
)

const (
	frontendPort        = 8080
	templateStoragePath = "./templates"
	privateStoragePath  = "./rooms"
	privateStorageUid   = 1000
	privateStorageGid   = 1000
)

func New(config *config.Room) *RoomManagerCtx {
	logger := log.With().Str("module", "room").Logger()

	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv)
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

func (manager *RoomManagerCtx) FindByName(name string) (*types.RoomEntry, error) {
	container, err := manager.containerByName(name)
	if err != nil {
		return nil, err
	}

	return manager.containerToEntry(*container)
}

func (manager *RoomManagerCtx) Create(settings types.RoomSettings) (string, error) {
	if settings.Name != "" && !dockerNames.RestrictedNamePattern.MatchString(settings.Name) {
		return "", fmt.Errorf("invalid container name, must match " + dockerNames.RestrictedNameChars)
	}

	if !manager.config.StorageEnabled && len(settings.Mounts) > 0 {
		return "", fmt.Errorf("mounts cannot be specified, because storage is disabled or unavailable")
	}

	if in, _ := utils.ArrayIn(settings.NekoImage, manager.config.NekoImages); !in {
		return "", fmt.Errorf("invalid neko image")
	}

	// TODO: Check if path name exists.
	roomName := settings.Name
	if roomName == "" {
		var err error
		roomName, err = utils.NewUID(8)
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

	instanceUrl := manager.config.InstanceUrl
	if instanceUrl == "" {
		urlProto := "http"
		if manager.config.TraefikCertresolver != "" {
			urlProto = "https"
		}

		// deprecated
		port := ""
		if manager.config.TraefikPort != "" {
			port = ":" + manager.config.TraefikPort
		}

		instanceUrl = urlProto + "://" + manager.config.TraefikDomain + port + "/"
	} else if !strings.HasSuffix(instanceUrl, "/") {
		instanceUrl = instanceUrl + "/"
	}

	labels := map[string]string{
		// Set internal labels
		"m1k1o.neko_rooms.name":       roomName,
		"m1k1o.neko_rooms.url":        instanceUrl + roomName + "/",
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

	env := []string{
		fmt.Sprintf("NEKO_BIND=:%d", frontendPort),
		fmt.Sprintf("NEKO_EPR=%d-%d", epr.Min, epr.Max),
		"NEKO_ICELITE=true",
	}

	// optional nat mapping
	if len(manager.config.NAT1To1IPs) > 0 {
		env = append(env, fmt.Sprintf("NEKO_NAT1TO1=%s", strings.Join(manager.config.NAT1To1IPs, ",")))
	}

	mounts := []dockerMount.Mount{}
	for _, mount := range settings.Mounts {
		readOnly := false

		hostPath := filepath.Clean(mount.HostPath)
		containerPath := filepath.Clean(mount.ContainerPath)

		if !filepath.IsAbs(hostPath) || !filepath.IsAbs(containerPath) {
			return "", fmt.Errorf("mount paths must be absolute")
		}

		// private container's data
		if mount.Type == types.MountPrivate {
			// ensure that target exists
			internalPath := path.Join(manager.config.StorageInternal, privateStoragePath, roomName, hostPath)
			if err := os.MkdirAll(internalPath, os.ModePerm); err != nil {
				return "", err
			}

			if err := utils.ChownR(internalPath, privateStorageUid, privateStorageGid); err != nil {
				return "", err
			}

			// prefix host path
			hostPath = path.Join(manager.config.StorageExternal, privateStoragePath, roomName, hostPath)
		} else if mount.Type == types.MountTemplate {
			// readonly template data
			readOnly = true

			// prefix host path
			hostPath = path.Join(manager.config.StorageExternal, templateStoragePath, hostPath)
		} else if mount.Type == types.MountPublic {
			// public whitelisted mounts
			var isAllowed = false
			for _, path := range manager.config.MountsWhitelist {
				if strings.HasPrefix(hostPath, path) {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				return "", fmt.Errorf("mount path is not whitelisted in config")
			}
		} else {
			return "", fmt.Errorf("unknown mount type %q", mount.Type)
		}

		mounts = append(mounts,
			dockerMount.Mount{
				Type:        dockerMount.TypeBind,
				Source:      hostPath,
				Target:      containerPath,
				ReadOnly:    readOnly,
				Consistency: dockerMount.ConsistencyDefault,

				BindOptions: &dockerMount.BindOptions{
					Propagation:  dockerMount.PropagationRPrivate,
					NonRecursive: false,
				},
			},
		)
	}

	config := &container.Config{
		// Hostname
		Hostname: containerName,
		// Domainname is preventing from running container on LXC (Proxmox)
		// https://www.gitmemory.com/issue/docker/for-linux/743/524569376
		// Domainname: containerName,
		// List of exposed ports
		ExposedPorts: exposedPorts,
		// List of environment variable to set in the container
		Env: append(env, settings.ToEnv()...),
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
			Name: "unless-stopped",
		},
		// List of kernel capabilities to add to the container
		CapAdd: strslice.StrSlice{
			"SYS_ADMIN",
		},
		// Total shm memory usage
		ShmSize: 2 * 10e9,
		// Mounts specs used by the container
		Mounts: mounts,
	}

	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			manager.config.TraefikNetwork: {},
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
	container, err := manager.containerById(id)
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
		return nil, fmt.Errorf("damaged container labels: name not found")
	}

	nekoImage, ok := container.Config.Labels["m1k1o.neko_rooms.neko_image"]
	if !ok {
		return nil, fmt.Errorf("damaged container labels: neko_image not found")
	}

	epr, err := manager.getEprFromLabels(container.Config.Labels)
	if err != nil {
		return nil, err
	}

	privateStorageRoot := path.Join(manager.config.StorageExternal, privateStoragePath, roomName)
	templateStorageRoot := path.Join(manager.config.StorageExternal, templateStoragePath)

	mounts := []types.RoomMount{}
	for _, mount := range container.Mounts {
		mountType := types.MountPublic
		hostPath := mount.Source

		if strings.HasPrefix(hostPath, privateStorageRoot) {
			mountType = types.MountPrivate
			hostPath = strings.TrimPrefix(hostPath, privateStorageRoot)
		} else if strings.HasPrefix(hostPath, templateStorageRoot) {
			mountType = types.MountTemplate
			hostPath = strings.TrimPrefix(hostPath, templateStorageRoot)
		}

		mounts = append(mounts, types.RoomMount{
			Type:          mountType,
			HostPath:      hostPath,
			ContainerPath: mount.Destination,
		})
	}

	settings := types.RoomSettings{
		Name:           roomName,
		NekoImage:      nekoImage,
		MaxConnections: epr.Max - epr.Min + 1,
		Mounts:         mounts,
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
