package api

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"

	"github.com/m1k1o/neko-rooms/internal/utils"
)

func (manager *ApiManagerCtx) RoomLobby(w http.ResponseWriter, r *http.Request) {
	if !manager.conf.Lobby {
		http.NotFound(w, r)
	}

	roomName := chi.URLParam(r, "roomName")
	response, err := manager.rooms.FindByName(roomName)

	if err != nil || response.Name != roomName {
		manager.roomNotFound(w, r)
		return
	}

	if !response.Running {
		manager.roomNotRunning(w, r)
		return
	}

	if strings.Contains(response.Status, "starting") {
		manager.roomNotReady(w, r)
		return
	}

	manager.roomReady(w, r)
}

func (manager *ApiManagerCtx) roomNotFound(w http.ResponseWriter, r *http.Request) {
	utils.Swal2Response(w, `
		<div class="swal2-header">
			<div class="swal2-icon swal2-error">
				<div class="swal2-icon-content">X</div>
			</div>
			<h2 class="swal2-title">Room not found!</h2>
		</div>
		<div class="swal2-content">
			<div>The room you are trying to join does not exist.</div>
		</div>
	`)
}

func (manager *ApiManagerCtx) roomNotRunning(w http.ResponseWriter, r *http.Request) {
	utils.Swal2Response(w, `
		<div class="swal2-header">
			<div class="swal2-icon swal2-warning">
				<div class="swal2-icon-content">!</div>
			</div>
			<h2 class="swal2-title">Room is not running!</h2>
		</div>
		<div class="swal2-content">
			<div>The room you are trying to join is not running.</div>
		</div>
	`)
}

func (manager *ApiManagerCtx) roomNotReady(w http.ResponseWriter, r *http.Request) {
	utils.Swal2Response(w, `
		<meta http-equiv="refresh" content="2">

		<div class="swal2-header">
			<div class="swal2-icon swal2-info">
				<div class="swal2-icon-content">i</div>
			</div>
			<h2 class="swal2-title">Room is not ready, yet!</h2>
		</div>
		<div class="swal2-content">
			<div>Please wait, until this room is ready so you can join. This should happen any second now.</div>
		</div>
		<div class="swal2-actions">
			<div class="swal2-loader"></div>
			<button type="button" onclick="location = location" class="swal2-confirm swal2-styled" style="margin-top: 1.25em">Reload</button>
		</div>
	`)
}

func (manager *ApiManagerCtx) roomReady(w http.ResponseWriter, r *http.Request) {
	utils.Swal2Response(w, `
		<div class="swal2-header">
			<div class="swal2-icon swal2-success swal2-icon-show" style="display: flex;">
				<div class="swal2-success-circular-line-left" style="background-color: rgb(47, 49, 54);"></div>
				<span class="swal2-success-line-tip"></span> <span class="swal2-success-line-long"></span>
				<div class="swal2-success-ring"></div> <div class="swal2-success-fix" style="background-color: rgb(47, 49, 54);"></div>
				<div class="swal2-success-circular-line-right" style="background-color: rgb(47, 49, 54);"></div>
			</div>
			<h2 class="swal2-title">Room is ready!</h2>
		</div>
		<div class="swal2-content">
			<div>Requested room is ready, you can join now.</div>
			<div style="padding-top: .5em;">Try to reload page.</div>
		</div>
		<div class="swal2-actions">
			<button type="button" onclick="location = location" class="swal2-confirm swal2-styled">Go to room</button>
		</div>
		<div class="swal2-content swal2-actions">
			<small>If you see this page after refresh, <br /> it can mean misconfiguration on your side.</small>
		</div>
	`)
}
