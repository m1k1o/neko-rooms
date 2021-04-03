package types

import (
	"net/http"

	"github.com/go-chi/chi"
)

type ApiManager interface {
	Mount(r chi.Router)

	RoomLobby(w http.ResponseWriter, r *http.Request)
}
