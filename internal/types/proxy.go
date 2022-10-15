package types

import "net/http"

type ProxyManager interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	Shutdown() error
}
