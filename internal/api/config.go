package api

import (
	"encoding/json"
	"net/http"
)

func (manager *ApiManagerCtx) configRooms(w http.ResponseWriter, r *http.Request) {
	response := manager.rooms.Config()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
