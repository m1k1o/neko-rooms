package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/m1k1o/neko-rooms/internal/types"
)

func (manager *ApiManagerCtx) roomsList(w http.ResponseWriter, r *http.Request) {
	response, err := manager.rooms.List()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (manager *ApiManagerCtx) roomCreate(w http.ResponseWriter, r *http.Request) {
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

	ID, err := manager.rooms.Create(request)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if err := manager.rooms.Start(ID); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	response, err := manager.rooms.GetEntry(ID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (manager *ApiManagerCtx) roomRecreate(w http.ResponseWriter, r *http.Request) {
	roomId := chi.URLParam(r, "roomId")

	entry, err := manager.rooms.GetEntry(roomId)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	settings, err := manager.rooms.GetSettings(roomId)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if err := manager.rooms.Remove(roomId); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	ID, err := manager.rooms.Create(*settings)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if entry.Running {
		if err := manager.rooms.Start(ID); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}

	response, err := manager.rooms.GetEntry(ID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (manager *ApiManagerCtx) roomGetEntry(w http.ResponseWriter, r *http.Request) {
	roomId := chi.URLParam(r, "roomId")

	response, err := manager.rooms.GetEntry(roomId)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (manager *ApiManagerCtx) roomGetSettings(w http.ResponseWriter, r *http.Request) {
	roomId := chi.URLParam(r, "roomId")

	response, err := manager.rooms.GetSettings(roomId)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (manager *ApiManagerCtx) roomGetStats(w http.ResponseWriter, r *http.Request) {
	roomId := chi.URLParam(r, "roomId")

	response, err := manager.rooms.GetStats(roomId)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (manager *ApiManagerCtx) roomGenericAction(action func(id string) error) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		roomId := chi.URLParam(r, "roomId")

		err := action(roomId)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
