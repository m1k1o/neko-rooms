package room

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/docker/cli/opts"
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

	"github.com/m1k1o/neko-rooms/internal/config"
	"github.com/m1k1o/neko-rooms/internal/policies"
	"github.com/m1k1o/neko-rooms/internal/types"
	"github.com/m1k1o/neko-rooms/internal/utils"
)

const (
	frontendPort        = 8080
	templateStoragePath = "./templates"
	privateStoragePath  = "./rooms"
	privateStorageUid   = 1000
	privateStorageGid   = 1000
)

func New(client *dockerClient.Client, config *config.Room) *RoomManagerCtx {
	logger := log.With().Str("module", "room").Logger()

	return &RoomManagerCtx{
		logger: logger,
		config: config,
		client: client,
	}
}

type RoomManagerCtx struct {
	logger zerolog.Logger
	config *config.Room
	client *dockerClient.Client
}

func (manager *RoomManagerCtx) Config() types.RoomsConfig {
	return types.RoomsConfig{
		Connections:    manager.config.EprMax - manager.config.EprMin + 1,
		NekoImages:     manager.config.NekoImages,
		StorageEnabled: manager.config.StorageEnabled,
		UsesMux:        manager.config.Mux,
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
		return "", fmt.Errorf("invalid container name, must match %s", dockerNames.RestrictedNameChars)
	}

	if !manager.config.StorageEnabled && len(settings.Mounts) > 0 {
		return "", fmt.Errorf("mounts cannot be specified, because storage is disabled or unavailable")
	}

	if in, _ := utils.ArrayIn(settings.NekoImage, manager.config.NekoImages); !in {
		return "", fmt.Errorf("invalid neko image")
	}

	isPrivilegedImage, _ := utils.ArrayIn(settings.NekoImage, manager.config.NekoPrivilegedImages)

	// TODO: Check if path name exists.
	roomName := settings.Name
	if roomName == "" {
		var err error
		roomName, err = utils.NewUID(8)
		if err != nil {
			return "", err
		}
	}

	containerName := manager.config.InstanceName + "-" + roomName

	//
	// Allocate ports
	//

	portsNeeded := settings.MaxConnections
	if manager.config.Mux {
		portsNeeded = 1
	}

	epr, err := manager.allocatePorts(portsNeeded)
	if err != nil {
		return "", err
	}

	portBindings := nat.PortMap{}
	for port := epr.Min; port <= epr.Max; port++ {
		portBindings[nat.Port(fmt.Sprintf("%d/udp", port))] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: fmt.Sprintf("%d", port),
			},
		}

		// expose TCP port as well when using mux
		if manager.config.Mux {
			portBindings[nat.Port(fmt.Sprintf("%d/tcp", port))] = []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: fmt.Sprintf("%d", port),
				},
			}
		}
	}

	exposedPorts := nat.PortSet{
		nat.Port(fmt.Sprintf("%d/tcp", frontendPort)): struct{}{},
	}

	for port := range portBindings {
		exposedPorts[port] = struct{}{}
	}

	//
	// Set internal labels
	//

	var browserPolicyLabels *BrowserPolicyLabels
	if settings.BrowserPolicy != nil {
		browserPolicyLabels = &BrowserPolicyLabels{
			Type: settings.BrowserPolicy.Type,
			Path: settings.BrowserPolicy.Path,
		}
	}

	labels := manager.serializeLabels(RoomLabels{
		Name:      roomName,
		Mux:       manager.config.Mux,
		Epr:       epr,
		NekoImage: settings.NekoImage,

		BrowserPolicy: browserPolicyLabels,
	})

	//
	// Set traefik labels
	//

	pathPrefix := path.Join("/", manager.config.PathPrefix, roomName)

	if t := manager.config.Traefik; t.Enabled {
		// create traefik rule
		traefikRule := "PathPrefix(`" + pathPrefix + "`)"
		if t.Domain != "" && t.Domain != "*" {
			// match *.domain.tld as subdomain
			if strings.HasPrefix(t.Domain, "*.") {
				traefikRule = fmt.Sprintf(
					"Host(`%s.%s`)",
					roomName,
					strings.TrimPrefix(t.Domain, "*."),
				)
			} else {
				traefikRule += " && Host(`" + t.Domain + "`)"
			}
		} else {
			traefikRule += " && HostRegexp(`{host:.+}`)"
		}

		labels["traefik.enable"] = "true"
		labels["traefik.http.services."+containerName+"-frontend.loadbalancer.server.port"] = fmt.Sprintf("%d", frontendPort)
		labels["traefik.http.routers."+containerName+".entrypoints"] = t.Entrypoint
		labels["traefik.http.routers."+containerName+".rule"] = traefikRule
		labels["traefik.http.middlewares."+containerName+"-rdr.redirectregex.regex"] = pathPrefix + "$$"
		labels["traefik.http.middlewares."+containerName+"-rdr.redirectregex.replacement"] = pathPrefix + "/"
		labels["traefik.http.middlewares."+containerName+"-prf.stripprefix.prefixes"] = pathPrefix + "/"
		labels["traefik.http.routers."+containerName+".middlewares"] = containerName + "-rdr," + containerName + "-prf"
		labels["traefik.http.routers."+containerName+".service"] = containerName + "-frontend"

		// optional HTTPS
		if t.Certresolver != "" {
			labels["traefik.http.routers."+containerName+".tls"] = "true"
			labels["traefik.http.routers."+containerName+".tls.certresolver"] = t.Certresolver
		}
	} else {
		labels["m1k1o.neko_rooms.proxy.enabled"] = "true"
		labels["m1k1o.neko_rooms.proxy.path"] = pathPrefix
		labels["m1k1o.neko_rooms.proxy.port"] = fmt.Sprintf("%d", frontendPort)
	}

	// add custom labels
	for _, label := range manager.config.Labels {
		// replace dynamic values in labels
		label = strings.Replace(label, "{containerName}", containerName, -1)
		label = strings.Replace(label, "{roomName}", roomName, -1)

		if t := manager.config.Traefik; t.Enabled {
			label = strings.Replace(label, "{traefikEntrypoint}", t.Entrypoint, -1)
			label = strings.Replace(label, "{traefikCertresolver}", t.Certresolver, -1)
		}

		v := strings.SplitN(label, "=", 2)
		if len(v) != 2 {
			manager.logger.Warn().Str("label", label).Msg("invalid custom label")
			continue
		}

		key, val := v[0], v[1]
		labels[key] = val
	}

	//
	// Set environment variables
	//

	env := settings.ToEnv(manager.config, types.PortSettings{
		FrontendPort: frontendPort,
		EprMin:       epr.Min,
		EprMax:       epr.Max,
	})

	//
	// Set browser policies
	//

	if settings.BrowserPolicy != nil {
		if !manager.config.StorageEnabled {
			return "", fmt.Errorf("policies cannot be specified, because storage is disabled or unavailable")
		}

		policyJson, err := policies.Generate(settings.BrowserPolicy.Content, settings.BrowserPolicy.Type)
		if err != nil {
			return "", err
		}

		// create policy path (+ also get host path)
		policyPath := fmt.Sprintf("/%s-%s-policy.json", roomName, settings.BrowserPolicy.Type)
		templateInternalPath := path.Join(manager.config.StorageInternal, templateStoragePath)
		policyInternalPath := path.Join(templateInternalPath, policyPath)

		// create dir if does not exist
		if _, err := os.Stat(templateInternalPath); os.IsNotExist(err) {
			if err := os.MkdirAll(templateInternalPath, os.ModePerm); err != nil {
				return "", err
			}
		}

		// write policy to file
		if err := os.WriteFile(policyInternalPath, []byte(policyJson), 0644); err != nil {
			return "", err
		}

		// mount policy file
		settings.Mounts = append(settings.Mounts, types.RoomMount{
			Type:          types.MountTemplate,
			HostPath:      policyPath,
			ContainerPath: settings.BrowserPolicy.Path,
		})
	}

	//
	// Set container mounts
	//

	paths := map[string]bool{}
	mounts := []dockerMount.Mount{}
	for _, mount := range settings.Mounts {
		// ignore duplicates
		if _, ok := paths[mount.ContainerPath]; ok {
			continue
		}

		readOnly := false

		hostPath := filepath.Clean(mount.HostPath)
		containerPath := filepath.Clean(mount.ContainerPath)

		if !filepath.IsAbs(hostPath) || !filepath.IsAbs(containerPath) {
			return "", fmt.Errorf("mount paths must be absolute")
		}

		// private container's data
		if mount.Type == types.MountPrivate {
			// ensure that target exists with correct permissions
			internalPath := path.Join(manager.config.StorageInternal, privateStoragePath, roomName, hostPath)
			if _, err := os.Stat(internalPath); os.IsNotExist(err) {
				if err := os.MkdirAll(internalPath, os.ModePerm); err != nil {
					return "", err
				}

				if err := utils.ChownR(internalPath, privateStorageUid, privateStorageGid); err != nil {
					return "", err
				}
			}

			// prefix host path
			hostPath = path.Join(manager.config.StorageExternal, privateStoragePath, roomName, hostPath)
		} else if mount.Type == types.MountTemplate {
			// readonly template data
			readOnly = true

			// prefix host path
			hostPath = path.Join(manager.config.StorageExternal, templateStoragePath, hostPath)
		} else if mount.Type == types.MountProtected || mount.Type == types.MountPublic {
			// readonly if mount type is protected
			readOnly = mount.Type == types.MountProtected

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

		paths[mount.ContainerPath] = true
	}

	//
	// Set container device requests
	//

	var deviceRequests []container.DeviceRequest

	if len(settings.Resources.Gpus) > 0 {
		gpuOpts := opts.GpuOpts{}

		// convert to csv
		var buf bytes.Buffer
		w := csv.NewWriter(&buf)
		if err := w.Write(settings.Resources.Gpus); err != nil {
			return "", err
		}
		w.Flush()

		// set GPU opts
		if err := gpuOpts.Set(buf.String()); err != nil {
			return "", err
		}

		deviceRequests = append(deviceRequests, gpuOpts.Value()...)
	}

	//
	// Set container configs
	//

	config := &container.Config{
		// Hostname
		Hostname: containerName,
		// Domainname is preventing from running container on LXC (Proxmox)
		// https://www.gitmemory.com/issue/docker/for-linux/743/524569376
		// Domainname: containerName,
		// List of exposed ports
		ExposedPorts: exposedPorts,
		// List of environment variable to set in the container
		Env: env,
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
		ShmSize: settings.Resources.ShmSize,
		// Mounts specs used by the container
		Mounts: mounts,
		// Resources contains container's resources (cgroups config, ulimits...)
		Resources: container.Resources{
			CPUShares:      settings.Resources.CPUShares,
			NanoCPUs:       settings.Resources.NanoCPUs,
			Memory:         settings.Resources.Memory,
			DeviceRequests: deviceRequests,
		},
		// Privileged
		Privileged: isPrivilegedImage,
	}

	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			manager.config.InstanceNetwork: {},
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

	labels, err := manager.extractLabels(container.Config.Labels)
	if err != nil {
		return nil, err
	}

	privateStorageRoot := path.Join(manager.config.StorageExternal, privateStoragePath, labels.Name)
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

	var browserPolicy *types.BrowserPolicy
	if labels.BrowserPolicy != nil {
		browserPolicy = &types.BrowserPolicy{
			Type: labels.BrowserPolicy.Type,
			Path: labels.BrowserPolicy.Path,
		}

		var policyMount *types.RoomMount
		for _, mount := range mounts {
			if mount.ContainerPath == labels.BrowserPolicy.Path {
				policyMount = &mount
				break
			}
		}

		// TODO: Refactor.
		if policyMount != nil && policyMount.Type == types.MountTemplate {
			templateInternalPath := path.Join(manager.config.StorageInternal, templateStoragePath, policyMount.HostPath)
			if _, err := os.Stat(templateInternalPath); !os.IsNotExist(err) {
				if data, err := os.ReadFile(templateInternalPath); err == nil {
					if content, err := policies.Parse(string(data), labels.BrowserPolicy.Type); err == nil {
						browserPolicy.Content = *content
					}
				}
			}
		}
	}

	var roomResources types.RoomResources
	if container.HostConfig != nil {
		gpus := []string{}
		for _, req := range container.HostConfig.DeviceRequests {
			var isGpu bool
			var caps []string
			for _, cc := range req.Capabilities {
				for _, c := range cc {
					if c == "gpu" {
						isGpu = true
						continue
					}
					caps = append(caps, c)
				}
			}
			if !isGpu {
				continue
			}

			if req.Count > 1 {
				gpus = append(gpus, fmt.Sprintf("count=%d", req.Count))
			} else if req.Count == -1 {
				gpus = append(gpus, "all")
			}
			if req.Driver != "" {
				gpus = append(gpus, fmt.Sprintf("driver=%s", req.Driver))
			}
			if len(req.DeviceIDs) > 0 {
				gpus = append(gpus, fmt.Sprintf("device=%s", strings.Join(req.DeviceIDs, ",")))
			}
			if len(caps) > 0 {
				gpus = append(gpus, fmt.Sprintf("capabilities=%s", strings.Join(caps, ",")))
			}
			var opts []string
			for key, val := range req.Options {
				opts = append(opts, fmt.Sprintf("%s=%s", key, val))
			}
			if len(opts) > 0 {
				gpus = append(gpus, fmt.Sprintf("options=%s", strings.Join(opts, ",")))
			}
		}

		roomResources = types.RoomResources{
			CPUShares: container.HostConfig.CPUShares,
			NanoCPUs:  container.HostConfig.NanoCPUs,
			ShmSize:   container.HostConfig.ShmSize,
			Memory:    container.HostConfig.Memory,
			Gpus:      gpus,
		}
	}

	settings := types.RoomSettings{
		Name:           labels.Name,
		NekoImage:      labels.NekoImage,
		MaxConnections: labels.Epr.Max - labels.Epr.Min + 1,
		Mounts:         mounts,
		BrowserPolicy:  browserPolicy,
		Resources:      roomResources,
	}

	if labels.Mux {
		settings.MaxConnections = 0
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
