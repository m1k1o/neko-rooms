package api

import (
	"encoding/json"
	"net/http"

	"github.com/m1k1o/neko-rooms/internal/types"
)

func (manager *ApiManagerCtx) pullStart(w http.ResponseWriter, r *http.Request) {
	request := types.PullStart{}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	err := manager.rooms.PullStart(request)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	response := manager.rooms.PullStatus()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (manager *ApiManagerCtx) pullStatus(w http.ResponseWriter, r *http.Request) {
	response := manager.rooms.PullStatus()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (manager *ApiManagerCtx) pullStop(w http.ResponseWriter, r *http.Request) {
	err := manager.rooms.PullStop()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
