package server

import (
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi"
)

const pprofPath = "/debug/pprof/"

func withPProf(router *chi.Mux) {
	router.Route(pprofPath, func(r chi.Router) {
		r.Get("/", pprof.Index)

		r.Get("/{action}", func(w http.ResponseWriter, r *http.Request) {
			action := chi.URLParam(r, "action")

			switch action {
			case "cmdline":
				pprof.Cmdline(w, r)
			case "profile":
				pprof.Profile(w, r)
			case "symbol":
				pprof.Symbol(w, r)
			case "trace":
				pprof.Trace(w, r)
			default:
				pprof.Handler(action).ServeHTTP(w, r)
			}
		})
	})
}
