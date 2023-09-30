package types

import (
	"github.com/go-chi/chi/v5"
)

type ApiManager interface {
	Mount(r chi.Router)
}
