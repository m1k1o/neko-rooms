package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/m1k1o/neko-rooms/internal/types"
)

func (manager *ApiManagerCtx) pullStart(w http.ResponseWriter, r *http.Request) {
	request := types.PullStart{}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	err := manager.pull.Start(request)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	response := manager.pull.Status()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (manager *ApiManagerCtx) pullStatus(w http.ResponseWriter, r *http.Request) {
	response := manager.pull.Status()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (manager *ApiManagerCtx) pullStatusSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Connection does not support streaming", http.StatusBadRequest)
		return
	}

	sseChan := make(chan string)
	unsubscribe := manager.pull.Subscribe(sseChan)

	for {
		select {
		case <-r.Context().Done():
			manager.logger.Debug().Msg("sse context done")
			unsubscribe()
			return
		case data, ok := <-sseChan:
			if !ok {
				manager.logger.Debug().Msg("sse channel closed")
				return
			}

			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

func (manager *ApiManagerCtx) pullStop(w http.ResponseWriter, r *http.Request) {
	err := manager.pull.Stop()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
