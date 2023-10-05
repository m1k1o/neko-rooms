package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/m1k1o/neko-rooms/internal/room"
	"github.com/m1k1o/neko-rooms/internal/types"
)

func (manager *ApiManagerCtx) roomsList(w http.ResponseWriter, r *http.Request) {
	labelsMap := map[string]string{}
	for key, value := range r.URL.Query() {
		key = strings.ToLower(key)

		if !room.CheckLabelKey(key) {
			http.Error(w, "invalid label name, allowed characters: [a-z0-9.-]", 400)
			return
		}

		labelsMap[key] = value[0]
	}

	response, err := manager.rooms.List(r.Context(), labelsMap)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (manager *ApiManagerCtx) roomCreate(w http.ResponseWriter, r *http.Request) {
	var start = true // default value
	if s := r.URL.Query().Get("start"); s != "" {
		var err error
		start, err = strconv.ParseBool(s)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
	}

	// Default values
	request := types.RoomSettings{
		MaxConnections: 10,
		Resources: types.RoomResources{
			ShmSize: 2 * 1e9,
		},
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	ID, err := manager.rooms.Create(r.Context(), request)
	if err != nil {
		manager.logger.Error().Err(err).Msg("create: failed to create room")
		http.Error(w, err.Error(), 500)
		return
	}

	if start {
		if err := manager.rooms.Start(r.Context(), ID); err != nil {
			manager.logger.Error().Err(err).Msg("create: failed to start room")
			http.Error(w, err.Error(), 500)
			return
		}
	}

	response, err := manager.rooms.GetEntry(r.Context(), ID)
	if err != nil {
		manager.logger.Error().Err(err).Msg("create: failed to get room entry")
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (manager *ApiManagerCtx) roomRecreate(w http.ResponseWriter, r *http.Request) {
	roomId := chi.URLParam(r, "roomId")

	var start bool
	if s := r.URL.Query().Get("start"); s != "" {
		var err error
		start, err = strconv.ParseBool(s)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
	} else {
		entry, err := manager.rooms.GetEntry(r.Context(), roomId)
		if err != nil {
			if errors.Is(err, types.ErrRoomNotFound) {
				http.Error(w, err.Error(), 404)
			} else {
				manager.logger.Error().Err(err).Msg("recreate: failed to get room entry")
				http.Error(w, err.Error(), 500)
			}
			return
		}

		start = entry.Running
	}

	settings, err := manager.rooms.GetSettings(r.Context(), roomId)
	if err != nil {
		if errors.Is(err, types.ErrRoomNotFound) {
			http.Error(w, err.Error(), 404)
		} else {
			manager.logger.Error().Err(err).Msg("recreate: failed to get room settings")
			http.Error(w, err.Error(), 500)
		}
		return
	}

	// optional settings payload
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, err.Error(), 400)
		return
	}

	if err := manager.rooms.Remove(r.Context(), roomId); err != nil {
		manager.logger.Error().Err(err).Msg("recreate: failed to remove room")
		http.Error(w, err.Error(), 500)
		return
	}

	ID, err := manager.rooms.Create(r.Context(), *settings)
	if err != nil {
		manager.logger.Error().Err(err).Msg("recreate: failed to create room")
		http.Error(w, err.Error(), 500)
		return
	}

	if start {
		if err := manager.rooms.Start(r.Context(), ID); err != nil {
			manager.logger.Error().Err(err).Msg("recreate: failed to start room")
			http.Error(w, err.Error(), 500)
			return
		}
	}

	response, err := manager.rooms.GetEntry(r.Context(), ID)
	if err != nil {
		manager.logger.Error().Err(err).Msg("recreate: failed to get room entry")
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (manager *ApiManagerCtx) roomGetEntry(w http.ResponseWriter, r *http.Request) {
	roomId := chi.URLParam(r, "roomId")

	response, err := manager.rooms.GetEntry(r.Context(), roomId)
	if err != nil {
		if errors.Is(err, types.ErrRoomNotFound) {
			http.Error(w, err.Error(), 404)
		} else {
			http.Error(w, err.Error(), 500)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (manager *ApiManagerCtx) roomGetEntryByName(w http.ResponseWriter, r *http.Request) {
	// roomId is actually room name here
	roomName := chi.URLParam(r, "roomId")

	response, err := manager.rooms.GetEntryByName(r.Context(), roomName)
	if err != nil {
		if errors.Is(err, types.ErrRoomNotFound) {
			http.Error(w, err.Error(), 404)
		} else {
			http.Error(w, err.Error(), 500)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (manager *ApiManagerCtx) roomGetSettings(w http.ResponseWriter, r *http.Request) {
	roomId := chi.URLParam(r, "roomId")

	response, err := manager.rooms.GetSettings(r.Context(), roomId)
	if err != nil {
		if errors.Is(err, types.ErrRoomNotFound) {
			http.Error(w, err.Error(), 404)
		} else {
			http.Error(w, err.Error(), 500)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (manager *ApiManagerCtx) roomGetStats(w http.ResponseWriter, r *http.Request) {
	roomId := chi.URLParam(r, "roomId")

	response, err := manager.rooms.GetStats(r.Context(), roomId)
	if err != nil {
		if errors.Is(err, types.ErrRoomNotFound) {
			http.Error(w, err.Error(), 404)
		} else {
			http.Error(w, err.Error(), 500)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (manager *ApiManagerCtx) roomGenericAction(action func(ctx context.Context, id string) error) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		roomId := chi.URLParam(r, "roomId")

		err := action(r.Context(), roomId)
		if err != nil {
			if errors.Is(err, types.ErrRoomNotFound) {
				http.Error(w, err.Error(), 404)
			} else {
				http.Error(w, err.Error(), 500)
			}
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (manager *ApiManagerCtx) dockerCompose(w http.ResponseWriter, r *http.Request) {
	response, err := manager.rooms.ExportAsDockerCompose(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "text/yaml")
	w.Write(response)
}
