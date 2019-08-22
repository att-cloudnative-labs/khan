package routes

import (
	"egbitbucket.dtvops.net/com/agent/internal/agent/appmapping"
	"egbitbucket.dtvops.net/com/agent/internal/agent/conntrack"

	"egbitbucket.dtvops.net/com/goatt/pkg/service"
)

// Set is used to set the routes
func Set(httpservice *service.Service) {
	httpservice.Router.Post("/appmapping", appmapping.SetCache)
	httpservice.Router.Get("/appmapping", appmapping.GetFullCache)
	httpservice.Router.Get("/connections", conntrack.GetConnections)
}
