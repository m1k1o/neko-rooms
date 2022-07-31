package api

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/m1k1o/neko-rooms/internal/proxy"
)

func (manager *ApiManagerCtx) RoomLobby(w http.ResponseWriter, r *http.Request) {
	if !manager.conf.Lobby {
		http.NotFound(w, r)
	}

	roomName := chi.URLParam(r, "roomName")
	response, err := manager.rooms.FindByName(roomName)

	if err != nil || response.Name != roomName {
		proxy.RoomNotFound(w, r)
		return
	}

	if !response.Running {
		proxy.RoomNotRunning(w, r)
		return
	}

	if strings.Contains(response.Status, "starting") {
		proxy.RoomNotReady(w, r)
		return
	}

	proxy.RoomReady(w, r)
}
