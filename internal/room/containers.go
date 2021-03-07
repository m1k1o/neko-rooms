package room

import (
	"context"
	"fmt"

	dockerTypes "github.com/docker/docker/api/types"
)

func (manager *RoomManagerCtx) listContainers() ([]dockerTypes.Container, error) {
	containers, err := manager.client.ContainerList(context.Background(), dockerTypes.ContainerListOptions{ All: true })
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

func (manager *RoomManagerCtx) inspectContainer(id string) (*dockerTypes.ContainerJSON, error) {
	container, err := manager.client.ContainerInspect(context.Background(), id)
	if err != nil {
		return nil, err
	}

	val, ok := container.Config.Labels["m1k1o.neko_rooms.instance"]
	if !ok || val != manager.config.InstanceName {
		return nil, fmt.Errorf("This container does not belong to neko_rooms.")
	}

	return &container, nil
}
