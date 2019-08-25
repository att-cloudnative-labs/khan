package routes

import (
	"github.com/att-cloudnative-labs/khan/internal/agent"
	"github.com/go-chi/chi"
)

// Set is used to set the routes
func Set(httpService *chi.Mux) {
	httpService.Post("/cache", agent.SetCache)
	httpService.Get("/cache", agent.GetCache)
	httpService.Get("/connections", agent.GetConnections)
}
