package room

import (
	"context"
	"fmt"

	"github.com/m1k1o/neko-rooms/internal/types"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

func (manager *RoomManagerCtx) EventsLoopStart() {
	ctx, cancel := context.WithCancel(context.Background())
	manager.eventsLoopCtx = ctx
	manager.eventsLoopCancel = cancel

	msgs, errs := manager.client.Events(ctx, dockerTypes.EventsOptions{
		Filters: filters.NewArgs(
			filters.Arg("type", "container"),
			filters.Arg("label", fmt.Sprintf("m1k1o.neko_rooms.instance=%s", manager.config.InstanceName)),
			filters.Arg("event", "create"),
			filters.Arg("event", "start"),
			//filters.Arg("event", "health_status"),
			filters.Arg("event", "stop"),
			filters.Arg("event", "destroy"),
		),
	})

	manager.eventsWg.Add(1)
	go func() {
		defer manager.eventsWg.Done()

		for {
			select {
			case err, ok := <-errs:
				if !ok {
					manager.logger.Fatal().Msg("docker event error channel closed")
					return
				}

				manager.logger.Err(err).Msg("got docker event error")
				return
			case msg, ok := <-msgs:
				if !ok {
					manager.logger.Fatal().Msg("docker event message channel closed")
					return
				}

				e := types.RoomEvent{
					ID:     msg.Actor.ID[:12],
					Action: msg.Action,
					Labels: msg.Actor.Attributes,
				}

				manager.logger.Info().
					Str("id", e.ID).
					Str("action", e.Action).
					Msg("got docker event")

				// broadcast event
				manager.eventsMu.Lock()
				for _, listener := range manager.eventsListeners {
					listener <- e
				}
				manager.eventsMu.Unlock()
			}
		}
	}()
}

func (manager *RoomManagerCtx) EventsLoopStop() error {
	manager.eventsLoopCancel()
	manager.eventsWg.Wait()
	return nil
}

func (manager *RoomManagerCtx) Events(ctx context.Context) (<-chan types.RoomEvent, <-chan error) {
	messages := make(chan types.RoomEvent)
	errs := make(chan error, 1)

	// add listener
	manager.eventsMu.Lock()
	manager.eventsListeners = append(manager.eventsListeners, messages)
	manager.eventsMu.Unlock()

	manager.eventsWg.Add(1)
	go func() {
		defer manager.eventsWg.Done()
		defer close(errs)

		select {
		case <-ctx.Done():
			errs <- ctx.Err()
		case <-manager.eventsLoopCtx.Done():
			errs <- fmt.Errorf("room manager shutdown")
			return
		}

		// remove listener
		manager.eventsMu.Lock()
		for i, listener := range manager.eventsListeners {
			if listener == messages {
				manager.eventsListeners = append(manager.eventsListeners[:i], manager.eventsListeners[i+1:]...)
				break
			}
		}
		manager.eventsMu.Unlock()
	}()

	return messages, errs
}
