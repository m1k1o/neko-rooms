package room

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerClient "github.com/docker/docker/client"

	"github.com/m1k1o/neko-rooms/internal/types"
)

func (manager *RoomManagerCtx) containerToEntry(container dockerTypes.Container) (*types.RoomEntry, error) {
	labels, err := manager.extractLabels(container.Labels)
	if err != nil {
		return nil, err
	}

	roomId := container.ID[:12]

	entry := &types.RoomEntry{
		ID:             roomId,
		URL:            labels.URL,
		Name:           labels.Name,
		NekoImage:      labels.NekoImage,
		IsOutdated:     labels.NekoImage != container.Image,
		MaxConnections: labels.Epr.Max - labels.Epr.Min + 1,
		Running:        container.State == "running",
		IsReady:        manager.events.IsRoomReady(roomId) || strings.Contains(container.Status, "healthy"),
		Status:         container.Status,
		Created:        time.Unix(container.Created, 0),
		Labels:         labels.UserDefined,

		ContainerLabels: container.Labels,
	}

	if labels.Mux {
		entry.MaxConnections = 0
	}

	return entry, nil
}

func (manager *RoomManagerCtx) listContainers(ctx context.Context, labels map[string]string) ([]dockerTypes.Container, error) {
	args := filters.NewArgs(
		filters.Arg("label", fmt.Sprintf("m1k1o.neko_rooms.instance=%s", manager.config.InstanceName)),
	)

	for key, val := range labels {
		args.Add("label", fmt.Sprintf("m1k1o.neko_rooms.x-%s=%s", key, val))
	}

	return manager.client.ContainerList(ctx, dockerTypes.ContainerListOptions{
		All:     true,
		Filters: args,
	})
}

func (manager *RoomManagerCtx) containerFilter(ctx context.Context, args filters.Args) (*dockerTypes.Container, error) {
	args.Add("label", fmt.Sprintf("m1k1o.neko_rooms.instance=%s", manager.config.InstanceName))

	containers, err := manager.client.ContainerList(ctx, dockerTypes.ContainerListOptions{
		All:     true,
		Filters: args,
	})

	if err != nil {
		return nil, err
	}

	if len(containers) == 0 {
		return nil, types.ErrRoomNotFound
	}

	container := containers[0]
	return &container, nil
}

func (manager *RoomManagerCtx) containerById(ctx context.Context, id string) (*dockerTypes.Container, error) {
	return manager.containerFilter(ctx, filters.NewArgs(
		filters.Arg("id", id),
	))
}

func (manager *RoomManagerCtx) containerByName(ctx context.Context, name string) (*dockerTypes.Container, error) {
	return manager.containerFilter(ctx, filters.NewArgs(
		filters.Arg("label", fmt.Sprintf("m1k1o.neko_rooms.name=%s", name)),
	))
}

func (manager *RoomManagerCtx) inspectContainer(ctx context.Context, id string) (*dockerTypes.ContainerJSON, error) {
	container, err := manager.client.ContainerInspect(ctx, id)
	if err != nil {
		if dockerClient.IsErrNotFound(err) {
			return nil, types.ErrRoomNotFound
		}
		return nil, err
	}

	val, ok := container.Config.Labels["m1k1o.neko_rooms.instance"]
	if !ok || val != manager.config.InstanceName {
		return nil, fmt.Errorf("this container does not belong to neko_rooms")
	}

	return &container, nil
}

func (manager *RoomManagerCtx) containerExec(ctx context.Context, id string, cmd []string) (string, error) {
	exec, err := manager.client.ContainerExecCreate(ctx, id, dockerTypes.ExecConfig{
		AttachStderr: true,
		AttachStdin:  true,
		AttachStdout: true,
		Cmd:          cmd,
		Tty:          true,
		Detach:       false,
	})
	if err != nil {
		if dockerClient.IsErrNotFound(err) {
			return "", types.ErrRoomNotFound
		}
		return "", err
	}

	conn, err := manager.client.ContainerExecAttach(ctx, exec.ID, dockerTypes.ExecStartCheck{
		Detach: false,
		Tty:    true,
	})
	if err != nil {
		return "", err
	}
	defer conn.Close()

	data, err := io.ReadAll(conn.Reader)
	return string(data), err
}
