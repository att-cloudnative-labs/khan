package main

import (
	"fmt"

	"github.com/cloud-native-labs/khan/agent/cmd/webapp/config"
	"github.com/cloud-native-labs/khan/agent/cmd/webapp/metrics"
	"github.com/cloud-native-labs/khan/agent/cmd/webapp/routes"
	"github.com/cloud-native-labs/khan/agent/internal/agent/appmapping"
	"github.com/cloud-native-labs/khan/agent/internal/agent/conntrack"
	"github.com/cloud-native-labs/khan/agent/internal/platform/netclient"

	"egbitbucket.dtvops.net/com/goatt"
)

var (
	goattVersion = "0.0.0"
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

	netclient.Preset = config.Registry.GetString("PRESET")

	goattService := goatt.NewGoattService(port)
	enableProfiling := config.Registry.GetBool("ENABLE_PROFILING")
	goattService.Httpservice.EnableProfiling(enableProfiling)
	// add to /info endpoint
	goattService.Httpservice.ExposeInfo("goattVersion", goattVersion)
	metrics.Set(goattService.Metrics)
	routes.Set(goattService.Httpservice)

	// start appmapping updater
	appmapper := appmapping.NewAppmappingController(nodeName, appmappingUrl, 20)

	// start conntrack updater
	stopCh := make(chan struct{})
	appmapper.Start(stopCh)
	conntrack.StartUpdateTimer(nodeName, conntrackScript, connUpdatePeriod, stopCh)

	fmt.Printf("Starting application on port %s\n", port)
	err := goattService.Httpservice.Start()
	if err != nil {
		panic(err)
	}
}
