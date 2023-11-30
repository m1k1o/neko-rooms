package room

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/m1k1o/neko-rooms/internal/config"
	"github.com/m1k1o/neko-rooms/internal/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerClient "github.com/docker/docker/client"
)

type roomReady struct {
	id     string
	labels map[string]string
}

type events struct {
	wg sync.WaitGroup

	logger zerolog.Logger
	config *config.Room
	client *dockerClient.Client

	roomsReadyCh chan roomReady
	roomsReadyMu sync.Mutex
	roomsReady   map[string]struct{}

	ctx    context.Context
	cancel context.CancelFunc

	listeners   []chan types.RoomEvent
	listenersMu sync.Mutex
}

func newEvents(config *config.Room, client *dockerClient.Client) *events {
	return &events{
		logger: log.With().Str("module", "events").Logger(),
		config: config,
		client: client,

		roomsReadyCh: make(chan roomReady),
		roomsReady:   make(map[string]struct{}),
	}
}

func (e *events) Start() {
	e.ctx, e.cancel = context.WithCancel(context.Background())

	msgs, errs := e.client.Events(e.ctx, dockerTypes.EventsOptions{
		Filters: filters.NewArgs(
			filters.Arg("type", "container"),
			filters.Arg("label", fmt.Sprintf("m1k1o.neko_rooms.instance=%s", e.config.InstanceName)),
			filters.Arg("event", "create"),
			filters.Arg("event", "start"),
			filters.Arg("event", "health_status"),
			filters.Arg("event", "stop"),
			filters.Arg("event", "destroy"),
		),
	})

	e.wg.Add(1)
	go func() {
		defer e.wg.Done()

		for {
			select {
			case <-e.ctx.Done():
				e.logger.Info().Msg("docker event context closed")
				return
			case err, ok := <-errs:
				if !ok {
					e.logger.Error().Msg("docker event error channel closed")
					return
				}

				e.logger.Err(err).Msg("got docker event error")
				return
			case room := <-e.roomsReadyCh:
				e.logger.Info().Str("id", room.id).Msg("room ready")

				// ignore if room was already ready
				if !e.setRoomReady(room.id) {
					continue
				}

				e.broadcast(types.RoomEvent{
					ID:     room.id,
					Action: "ready",
					Labels: room.labels,
				})
			case msg := <-msgs:
				roomId := msg.Actor.ID[:12]
				labels := msg.Actor.Attributes

				e.logger.Info().
					Str("id", roomId).
					Str("action", msg.Action).
					Msg("got docker event")

				action := ""
				switch msg.Action {
				case "create":
					action = "created"
				case "start":
					action = "started"
					e.waitForRoomReady(roomId, labels)
				case "health_status: healthy":
					action = "ready"
					// ignore if room was already ready
					if !e.setRoomReady(roomId) {
						continue
					}
				case "stop":
					action = "stopped"
					e.setRoomNotReady(roomId)
				case "destroy":
					action = "destroyed"
				}

				e.broadcast(types.RoomEvent{
					ID:     roomId,
					Action: action,
					Labels: labels,
				})
			}
		}
	}()
}

func (e *events) Shutdown() error {
	e.cancel()
	close(e.roomsReadyCh)
	e.wg.Wait()
	return nil
}

//
// room ready
//

func (e *events) waitForRoomReady(roomId string, labels map[string]string) {
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()

		// check if room is ready
		exec, err := e.client.ContainerExecCreate(e.ctx, roomId, dockerTypes.ExecConfig{
			AttachStdout: true,
			Cmd: []string{
				"/bin/bash", "-c",
				fmt.Sprintf(`for ((a=1; a<=5; a++)); do (echo > /dev/tcp/localhost/%d) >/dev/null && echo -n OK && exit; sleep 1; done; exit`, frontendPort),
			},
		})
		if err != nil {
			e.logger.Err(err).Msg("failed to create exec")
			return
		}

		conn, err := e.client.ContainerExecAttach(e.ctx, exec.ID, dockerTypes.ExecStartCheck{})
		if err != nil {
			e.logger.Err(err).Msg("failed to attach exec")
			return
		}
		defer conn.Close()

		data, err := io.ReadAll(conn.Reader)
		if err != nil {
			e.logger.Err(err).Msg("failed to read exec")
			return
		}

		if strings.HasSuffix(string(data), "OK") {
			e.logger.Info().Str("id", roomId).Msg("room ready")
			e.roomsReadyCh <- roomReady{
				id:     roomId,
				labels: labels,
			}
			return
		}

		e.logger.Info().Str("id", roomId).Str("data", string(data)).Msg("room not ready")
	}()
}

func (e *events) setRoomReady(roomId string) bool {
	e.roomsReadyMu.Lock()
	defer e.roomsReadyMu.Unlock()

	_, ok := e.roomsReady[roomId]
	e.roomsReady[roomId] = struct{}{}
	return !ok
}

func (e *events) setRoomNotReady(roomId string) {
	e.roomsReadyMu.Lock()
	defer e.roomsReadyMu.Unlock()

	delete(e.roomsReady, roomId)
}

func (e *events) IsRoomReady(roomId string) bool {
	e.roomsReadyMu.Lock()
	defer e.roomsReadyMu.Unlock()

	_, ok := e.roomsReady[roomId]
	return ok
}

//
// events
//

func (e *events) broadcast(event types.RoomEvent) {
	e.listenersMu.Lock()
	for _, listener := range e.listeners {
		listener <- event
	}
	e.listenersMu.Unlock()
}

func (e *events) Events(ctx context.Context) (<-chan types.RoomEvent, <-chan error) {
	messages := make(chan types.RoomEvent)
	errs := make(chan error, 1)

	// add listener
	e.listenersMu.Lock()
	e.listeners = append(e.listeners, messages)
	e.listenersMu.Unlock()

	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		defer close(errs)

		select {
		case <-ctx.Done():
			errs <- ctx.Err()
		case <-e.ctx.Done():
			errs <- fmt.Errorf("room events shutdown")
			return
		}

		// remove listener
		e.listenersMu.Lock()
		for i, listener := range e.listeners {
			if listener == messages {
				e.listeners = append(e.listeners[:i], e.listeners[i+1:]...)
				break
			}
		}
		e.listenersMu.Unlock()
	}()

	return messages, errs
}
