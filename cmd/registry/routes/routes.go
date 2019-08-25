package routes

import (
	"github.com/att-cloudnative-labs/khan/internal/registry"
	"github.com/go-chi/chi"
)

// Set is used to set the routes
func Set(httpService *chi.Mux) {
	httpService.Get("/cache", registry.RequestHandler)
}
