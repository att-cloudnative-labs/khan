package routes

import (
	"github.com/cloud-native-labs/khan/agent/internal/agent/appmapping"
	"github.com/cloud-native-labs/khan/agent/internal/agent/conntrack"
	"github.com/go-chi/chi"
)

// Set is used to set the routes
func Set(httpservice *chi.Mux) {
	httpservice.Post("/appmapping", appmapping.SetCache)
	httpservice.Get("/appmapping", appmapping.GetFullCache)
	httpservice.Get("/connections", conntrack.GetConnections)
}
