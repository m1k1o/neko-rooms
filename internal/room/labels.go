package room

import (
	"fmt"
	"strconv"
)

type RoomLabels struct {
	Name      string
	URL       string
	Epr       EprPorts
	NekoImage string
}

func (manager *RoomManagerCtx) extractLabels(labels map[string]string) (*RoomLabels, error) {
	name, ok := labels["m1k1o.neko_rooms.name"]
	if !ok {
		return nil, fmt.Errorf("damaged container labels: name not found")
	}

	url, ok := labels["m1k1o.neko_rooms.url"]
	if !ok {
		return nil, fmt.Errorf("damaged container labels: url not found")
	}

	nekoImage, ok := labels["m1k1o.neko_rooms.neko_image"]
	if !ok {
		return nil, fmt.Errorf("damaged container labels: neko_image not found")
	}

	eprMinStr, ok := labels["m1k1o.neko_rooms.epr.min"]
	if !ok {
		return nil, fmt.Errorf("damaged container labels: epr.min not found")
	}

	eprMin, err := strconv.ParseUint(eprMinStr, 10, 16)
	if err != nil {
		return nil, err
	}

	eprMaxStr, ok := labels["m1k1o.neko_rooms.epr.max"]
	if !ok {
		return nil, fmt.Errorf("damaged container labels: epr.max not found")
	}

	eprMax, err := strconv.ParseUint(eprMaxStr, 10, 16)
	if err != nil {
		return nil, err
	}

	return &RoomLabels{
		Name:      name,
		URL:       url,
		NekoImage: nekoImage,
		Epr: EprPorts{
			Min: uint16(eprMin),
			Max: uint16(eprMax),
		},
	}, nil
}

func (manager *RoomManagerCtx) serializeLabels(labels RoomLabels) map[string]string {
	return map[string]string{
		"m1k1o.neko_rooms.name":       labels.Name,
		"m1k1o.neko_rooms.url":        labels.URL,
		"m1k1o.neko_rooms.instance":   manager.config.InstanceName,
		"m1k1o.neko_rooms.epr.min":    fmt.Sprintf("%d", labels.Epr.Min),
		"m1k1o.neko_rooms.epr.max":    fmt.Sprintf("%d", labels.Epr.Max),
		"m1k1o.neko_rooms.neko_image": labels.NekoImage,
	}
}
