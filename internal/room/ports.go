package room

import (
	"fmt"
	"sort"
	"strconv"
)

type EprPorts struct {
	Min uint16
	Max uint16
}

func (manager *RoomManagerCtx) allocatePorts(sum uint16) (EprPorts, error) {
	if sum < 1 {
		return EprPorts{}, fmt.Errorf("unable to allocate 0 ports")
	}

	min := manager.config.EprMin
	max := manager.config.EprMax

	epr := EprPorts{
		Min: min,
		Max: min + sum - 1,
	}

	ports, err := manager.getUsedPorts()
	if err != nil {
		return epr, err
	}

	for _, port := range ports {
		if (epr.Min >= port.Min && epr.Min <= port.Max) || (epr.Max >= port.Min && epr.Max <= port.Max) {
			epr.Min = port.Max + 1
			epr.Max = port.Max + sum
		}
	}

	if epr.Min > max || epr.Max > max {
		return epr, fmt.Errorf("unable to allocate ports: not enough ports")
	}

	return epr, nil
}

func (manager *RoomManagerCtx) getUsedPorts() ([]EprPorts, error) {
	containers, err := manager.listContainers()
	if err != nil {
		return nil, err
	}

	result := []EprPorts{}
	for _, container := range containers {
		epr, err := manager.getEprFromLabels(container.Labels)
		if err != nil {
			return nil, err
		}

		result = append(result, epr)
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Min < result[j].Min
	})

	return result, nil
}

func (manager *RoomManagerCtx) getEprFromLabels(labels map[string]string) (EprPorts, error) {
	var err error
	epr := EprPorts{}

	eprMinStr, ok := labels["m1k1o.neko_rooms.epr.min"]
	if !ok {
		return epr, fmt.Errorf("damaged container labels: epr.min not found")
	}

	eprMin, err := strconv.ParseUint(eprMinStr, 10, 16)
	if err != nil {
		return epr, err
	}

	eprMaxStr, ok := labels["m1k1o.neko_rooms.epr.max"]
	if !ok {
		return epr, fmt.Errorf("damaged container labels: epr.max not found")
	}

	eprMax, err := strconv.ParseUint(eprMaxStr, 10, 16)
	if err != nil {
		return epr, err
	}

	epr.Min = uint16(eprMin)
	epr.Max = uint16(eprMax)
	return epr, nil
}
