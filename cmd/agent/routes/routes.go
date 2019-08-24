package routes

import (
	"github.com/cloud-native-labs/khan/internal/conntrack"
	"github.com/cloud-native-labs/khan/internal/mappings"
	"github.com/go-chi/chi"
)

// Set is used to set the routes
func Set(httpservice *chi.Mux) {
	httpservice.Post("/appmapping", mappings.SetCache)
	httpservice.Get("/appmapping", mappings.GetFullCache)
	httpservice.Get("/connections", conntrack.GetConnections)
}
