package routes

import (
	"github.com/cloud-native-labs/khan/controller/internal/controller/appmappings"
	"github.com/cloud-native-labs/khan/controller/internal/controller/connections"
	"github.com/go-chi/chi"
)

// Set is used to set the routes
func Set(httpservice *chi.Mux) {

	httpservice.Get("/connections", connections.ConnectionsCallHandler)
	httpservice.Get("/appmapping", appmappings.NodeMappingRequestHandler)
}
