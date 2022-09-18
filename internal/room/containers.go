package room

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/m1k1o/neko-rooms/internal/types"
)

func (manager *RoomManagerCtx) containerToEntry(container dockerTypes.Container) (*types.RoomEntry, error) {
	labels, err := manager.extractLabels(container.Labels)
	if err != nil {
		return nil, err
	}

	entry := &types.RoomEntry{
		ID:             container.ID,
		URL:            labels.URL,
		Name:           labels.Name,
		NekoImage:      labels.NekoImage,
		IsOutdated:     labels.NekoImage != container.Image,
		MaxConnections: labels.Epr.Max - labels.Epr.Min + 1,
		Running:        container.State == "running",
		Status:         container.Status,
		Created:        time.Unix(container.Created, 0),
	}

	if labels.Mux {
		entry.MaxConnections = 0
	}

	return entry, nil
}

func (manager *RoomManagerCtx) listContainers() ([]dockerTypes.Container, error) {
	args := filters.NewArgs(
		filters.Arg("label", "m1k1o.neko_rooms.instance"),
	)

	containers, err := manager.client.ContainerList(context.Background(), dockerTypes.ContainerListOptions{
		All:     true,
		Filters: args,
	})

	if err != nil {
		return nil, err
	}

	result := []dockerTypes.Container{}
	for _, container := range containers {
		val, ok := container.Labels["m1k1o.neko_rooms.instance"]
		if !ok || val != manager.config.InstanceName {
			continue
		}

		result = append(result, container)
	}

	return result, nil
}

func (manager *RoomManagerCtx) containerFilter(args filters.Args) (*dockerTypes.Container, error) {
	args.Add("label", "m1k1o.neko_rooms.instance")

	containers, err := manager.client.ContainerList(context.Background(), dockerTypes.ContainerListOptions{
		All:     true,
		Filters: args,
	})

	if err != nil {
		return nil, err
	}

	if len(containers) == 0 {
		return nil, fmt.Errorf("container not found")
	}

	container := containers[0]

	val, ok := container.Labels["m1k1o.neko_rooms.instance"]
	if !ok || val != manager.config.InstanceName {
		return nil, fmt.Errorf("this container does not belong to neko_rooms")
	}

	return &container, nil
}

func (manager *RoomManagerCtx) containerById(id string) (*dockerTypes.Container, error) {
	return manager.containerFilter(filters.NewArgs(
		filters.Arg("id", id),
	))
}

func (manager *RoomManagerCtx) containerByName(name string) (*dockerTypes.Container, error) {
	return manager.containerFilter(filters.NewArgs(
		filters.Arg("name", manager.config.InstanceName+"-"+name),
	))
}

func (manager *RoomManagerCtx) inspectContainer(id string) (*dockerTypes.ContainerJSON, error) {
	container, err := manager.client.ContainerInspect(context.Background(), id)
	if err != nil {
		return nil, err
	}

	val, ok := container.Config.Labels["m1k1o.neko_rooms.instance"]
	if !ok || val != manager.config.InstanceName {
		return nil, fmt.Errorf("this container does not belong to neko_rooms")
	}

	return &container, nil
}

func (manager *RoomManagerCtx) containerExec(id string, cmd []string) ([]byte, error) {
	exec, err := manager.client.ContainerExecCreate(context.Background(), id, dockerTypes.ExecConfig{
		AttachStderr: true,
		AttachStdin:  true,
		AttachStdout: true,
		Cmd:          cmd,
		Tty:          true,
		Detach:       false,
	})
	if err != nil {
		return nil, err
	}

	conn, err := manager.client.ContainerExecAttach(context.Background(), exec.ID, dockerTypes.ExecStartCheck{
		Detach: false,
		Tty:    true,
	})
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return io.ReadAll(conn.Reader)
}

func (manager *RoomManagerCtx) containerExecDialer(id string) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		fmt.Println("dialer", network, addr)

		// split addr
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}

		exec, err := manager.client.ContainerExecCreate(ctx, id, dockerTypes.ExecConfig{
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Cmd:          []string{"wget", "-O-", "http://" + host + ":" + port + ""},
			Tty:          false,
			Detach:       false,
		})
		if err != nil {
			return nil, err
		}

		c, err := manager.client.ContainerExecAttach(ctx, exec.ID, dockerTypes.ExecStartCheck{
			Detach: false,
			Tty:    false,
		})
		if err != nil {
			return nil, err
		}

		// Read all
		// io.ReadAll(c.Reader)
		dat, err := io.ReadAll(c.Reader)
		if err != nil {
			return nil, err
		}

		fmt.Println("dialer", string(dat))

		// new pipe
		stderr := stdcopy.NewStdWriter(c.Conn, stdcopy.Stderr)
		stdout := stdcopy.NewStdWriter(c.Conn, stdcopy.Stdout)

		return &cmdDialer{
			stdin:  c.Conn,
			stdout: c.Conn,
		}, nil
	}
}
