package routes

import (
	"egbitbucket.dtvops.net/com/controller/internal/controller"
	"egbitbucket.dtvops.net/com/controller/internal/controller/appmappings"
	"egbitbucket.dtvops.net/com/controller/internal/controller/injest"
	"egbitbucket.dtvops.net/com/controller/internal/controller/connections"
	"encoding/json"
	"fmt"
	"net/http"

	"egbitbucket.dtvops.net/com/goatt/pkg/service"
)

// Set is used to set the routes
func Set(httpservice *service.Service) {
	httpservice.Router.Get("/welcome", serviceCallHandler)
	httpservice.Router.Post("/injest", injest.InjestCallHandler)
	httpservice.Router.Get("/connections", connections.ConnectionsCallHandler)
	httpservice.Router.Get("/appmapping", appmappings.NodeMappingRequestHandler)
}

func serviceCallHandler(w http.ResponseWriter, _ *http.Request) {
	// the business logic is done outside of cmd/webapp. This function should
	// do the bare minimum required where most of the work is done by
	// other packages.
	hello := controller.Welcome()
	b, err := json.Marshal(hello)
	if err != nil {
		err = fmt.Errorf("failed to marshal hello: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}