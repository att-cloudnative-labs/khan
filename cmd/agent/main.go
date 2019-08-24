package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/cloud-native-labs/khan/cmd/agent/routes"
	"github.com/cloud-native-labs/khan/cmd/agent/config"
	"github.com/cloud-native-labs/khan/internal/mappings"
	"github.com/cloud-native-labs/khan/internal/conntrack"
)

func main() {
	config.Set()

	config.Registry.SetDefault("CONN_UPDATE_PERIOD", "30")
	config.Registry.SetDefault("CONNTRACK_SCRIPT", "/tmp/conntrackScript.sh")
	config.Registry.SetDefault("APPMAPPING_URL", "http://controller/appmapping")

	port := config.Registry.GetString("SERVER_PORT")
	conntrackScript := config.Registry.GetString("CONNTRACK_SCRIPT")
	connUpdatePeriod := config.Registry.GetInt("CONN_UPDATE_PERIOD")
	appmappingUrl := config.Registry.GetString("APPMAPPING_URL")
	nodeName := config.Registry.GetString("NODE_NAME")

	r := chi.NewRouter()

	routes.Set(r)

	// start appmapping updater
	mapper := mappings.NewLocalTargetCacheController(nodeName, appmappingUrl, 20)

	// start conntrack updater
	stopCh := make(chan struct{})
	mapper.Start(stopCh)
	conntrack.StartUpdateTimer(nodeName, conntrackScript, connUpdatePeriod, stopCh)

	fmt.Printf("Starting application on port %s\n", port)

	err := http.ListenAndServe(port, r)

	if err != nil {
		panic(err)
	}
}
