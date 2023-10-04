package room

import (
	"context"
	"fmt"
	"sort"
)

type EprPorts struct {
	Min uint16
	Max uint16
}

func (manager *RoomManagerCtx) allocatePorts(ctx context.Context, sum uint16) (EprPorts, error) {
	if sum < 1 {
		return EprPorts{}, fmt.Errorf("unable to allocate 0 ports")
	}

	min := manager.config.EprMin
	max := manager.config.EprMax

	epr := EprPorts{
		Min: min,
		Max: min + sum - 1,
	}

	ports, err := manager.getUsedPorts(ctx)
	if err != nil {
		return epr, err
	}

	for _, port := range ports {
		if (epr.Min >= port.Min && epr.Min <= port.Max) ||
			(epr.Max >= port.Min && epr.Max <= port.Max) {
			epr.Min = port.Max + 1
			epr.Max = port.Max + sum
		}
	}

	if epr.Min > max || epr.Max > max {
		return epr, fmt.Errorf("unable to allocate ports: not enough ports")
	}

	return epr, nil
}

func (manager *RoomManagerCtx) getUsedPorts(ctx context.Context) ([]EprPorts, error) {
	containers, err := manager.listContainers(ctx, nil)
	if err != nil {
		return nil, err
	}

	result := []EprPorts{}
	for _, container := range containers {
		labels, err := manager.extractLabels(container.Labels)
		if err != nil {
			return nil, err
		}

		result = append(result, labels.Epr)
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Min < result[j].Min
	})

	return result, nil
}
