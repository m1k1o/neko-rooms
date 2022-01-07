package room

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"m1k1o/neko_rooms/internal/types"
	"m1k1o/neko_rooms/internal/utils"
	"sync"
	"time"

	dockerTypes "github.com/docker/docker/api/types"
)

type PullConfig struct {
	mu     sync.Mutex
	cancel func()
	status types.PullStatus
	layers map[string]int
}

func (pull *PullConfig) TryInitialize(cancel func()) bool {
	pull.mu.Lock()
	defer pull.mu.Unlock()

	if pull.status.Active {
		cancel()
		return false
	}

	now := time.Now()
	pull.cancel = cancel

	pull.status = types.PullStatus{
		Active:  true,
		Started: &now,
		Layers:  []types.PullLayer{},
		Status:  []string{},
	}

	pull.layers = map[string]int{}

	return true
}

func (pull *PullConfig) SetDone() {
	pull.mu.Lock()
	defer pull.mu.Unlock()

	now := time.Now()
	pull.status.Active = false
	pull.status.Finished = &now
}

func (pull *PullConfig) Stop() bool {
	pull.mu.Lock()
	defer pull.mu.Unlock()

	if pull.status.Active {
		pull.cancel()
	}

	return pull.status.Active
}

func (pull *PullConfig) Get() types.PullStatus {
	pull.mu.Lock()
	defer pull.mu.Unlock()

	return pull.status
}

func (manager *RoomManagerCtx) PullStart(request types.PullStart) error {
	if in, _ := utils.ArrayIn(request.NekoImage, manager.config.NekoImages); !in {
		return fmt.Errorf("unknown neko image")
	}

	ctx, cancel := context.WithCancel(context.Background())
	if !manager.pull.TryInitialize(cancel) {
		return fmt.Errorf("pull is already in progess")
	}

	// handle registry auth
	var opts dockerTypes.ImagePullOptions
	if request.RegistryUser != "" && request.RegistryPass != "" {
		authConfig := dockerTypes.AuthConfig{
			Username: request.RegistryUser,
			Password: request.RegistryPass,
		}

		encodedJSON, err := json.Marshal(authConfig)
		if err != nil {
			panic(err)
		}

		opts = dockerTypes.ImagePullOptions{
			RegistryAuth: base64.URLEncoding.EncodeToString(encodedJSON),
		}
	}

	reader, err := manager.client.ImagePull(ctx, request.NekoImage, opts)

	if err != nil {
		manager.pull.SetDone()
		return err
	}

	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			data := scanner.Bytes()

			layer := types.PullLayer{}
			if err := json.Unmarshal(data, &layer); err != nil {
				manager.pull.status.Status = append(
					manager.pull.status.Status,
					fmt.Sprintf("Error while parsing pull response: %s", err),
				)

				continue
			}

			if layer.ProgressDetail != nil {
				// map layer id to slice index
				if index, ok := manager.pull.layers[layer.ID]; ok {
					manager.pull.status.Layers[index] = layer
				} else {
					manager.pull.layers[layer.ID] = len(manager.pull.layers)
					manager.pull.status.Layers = append(manager.pull.status.Layers, layer)
				}
			} else {
				manager.pull.status.Status = append(
					manager.pull.status.Status,
					layer.Status,
				)
			}
		}

		if err := scanner.Err(); err != nil {
			manager.pull.status.Status = append(
				manager.pull.status.Status,
				fmt.Sprintf("Error while reading pull response: %s", err),
			)
		}

		reader.Close()
		manager.pull.SetDone()
	}()

	return nil
}

func (manager *RoomManagerCtx) PullStop() error {
	if !manager.pull.Stop() {
		return fmt.Errorf("pull is not in progess")
	}

	return nil
}

func (manager *RoomManagerCtx) PullStatus() types.PullStatus {
	return manager.pull.Get()
}
