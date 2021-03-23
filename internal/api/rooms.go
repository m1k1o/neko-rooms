package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"

	"m1k1o/neko_rooms/internal/types"
)

func (manager *ApiManagerCtx) roomsConfig(w http.ResponseWriter, r *http.Request) {
	response := manager.rooms.Config()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

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
	request := types.RoomSettings{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	ID, err := manager.rooms.Create(request)
	if err != nil {
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

func (manager *ApiManagerCtx) roomRemove(w http.ResponseWriter, r *http.Request) {
	roomId := chi.URLParam(r, "roomId")

	err := manager.rooms.Remove(roomId)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
