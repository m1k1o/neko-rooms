package pull

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	dockerTypes "github.com/docker/docker/api/types"
	dockerClient "github.com/docker/docker/client"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/m1k1o/neko-rooms/internal/types"
	"github.com/m1k1o/neko-rooms/internal/utils"
)

type PullManagerCtx struct {
	logger zerolog.Logger
	client *dockerClient.Client
	images []string

	mu     sync.Mutex
	cancel func()
	status types.PullStatus
	layers map[string]int
}

func New(client *dockerClient.Client, nekoImages []string) *PullManagerCtx {
	return &PullManagerCtx{
		logger: log.With().Str("module", "pull").Logger(),
		client: client,
		images: nekoImages,
	}
}

func (manager *PullManagerCtx) tryInitialize(cancel func()) bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if manager.status.Active {
		cancel()
		return false
	}

	now := time.Now()
	manager.cancel = cancel

	manager.status = types.PullStatus{
		Active:  true,
		Started: &now,
		Layers:  []types.PullLayer{},
		Status:  []string{},
	}

	manager.layers = map[string]int{}

	return true
}

func (manager *PullManagerCtx) setDone() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	now := time.Now()
	manager.status.Active = false
	manager.status.Finished = &now
}

func (manager *PullManagerCtx) Start(request types.PullStart) error {
	if in, _ := utils.ArrayIn(request.NekoImage, manager.images); !in {
		return fmt.Errorf("unknown neko image")
	}

	ctx, cancel := context.WithCancel(context.Background())
	if !manager.tryInitialize(cancel) {
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
			return err
		}

		opts = dockerTypes.ImagePullOptions{
			RegistryAuth: base64.URLEncoding.EncodeToString(encodedJSON),
		}
	}

	reader, err := manager.client.ImagePull(ctx, request.NekoImage, opts)

	if err != nil {
		manager.setDone()
		return err
	}

	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			data := scanner.Bytes()

			layer := types.PullLayer{}
			if err := json.Unmarshal(data, &layer); err != nil {
				manager.status.Status = append(
					manager.status.Status,
					fmt.Sprintf("Error while parsing pull response: %s", err),
				)
				continue
			}

			if layer.ProgressDetail != nil {
				// map layer id to slice index
				if index, ok := manager.layers[layer.ID]; ok {
					manager.status.Layers[index] = layer
				} else {
					manager.layers[layer.ID] = len(manager.layers)
					manager.status.Layers = append(manager.status.Layers, layer)
				}
			} else {
				manager.status.Status = append(
					manager.status.Status,
					layer.Status,
				)
			}
		}

		if err := scanner.Err(); err != nil {
			manager.status.Status = append(
				manager.status.Status,
				fmt.Sprintf("Error while reading pull response: %s", err),
			)
		}

		reader.Close()
		manager.setDone()
	}()

	return nil
}

func (manager *PullManagerCtx) Stop() error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if !manager.status.Active {
		return fmt.Errorf("pull is not in progess")

	}

	manager.cancel()
	return nil
}

func (manager *PullManagerCtx) Status() types.PullStatus {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.status
}
