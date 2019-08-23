package routes

import (
	"github.com/cloud-native-labs/khan/agent/internal/agent/appmapping"
	"github.com/cloud-native-labs/khan/agent/internal/agent/conntrack"

	"egbitbucket.dtvops.net/com/goatt/pkg/service"
)

// Set is used to set the routes
func Set(httpservice *service.Service) {
	httpservice.Router.Post("/appmapping", appmapping.SetCache)
	httpservice.Router.Get("/appmapping", appmapping.GetFullCache)
	httpservice.Router.Get("/connections", conntrack.GetConnections)
}
