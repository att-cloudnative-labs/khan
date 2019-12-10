package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/att-cloudnative-labs/khan/internal/agent"
)

func main() {
	flag.Parse()

	v := viper.New()
	v.SetDefault("SERVER_ADDR", ":8080")
	v.SetDefault("CONNTRACK_PERIOD", 30)
	v.SetDefault("HOST_PERIOD", 30)
	v.SetDefault("REGISTRY_URL", "http://registry/cache")
	v.AutomaticEnv()

	addr := v.GetString("SERVER_ADDR")
	hostPeriod := v.GetInt("HOST_PERIOD")
	conntrackPeriod := v.GetInt("CONNTRACK_PERIOD")
	registryURL := v.GetString("REGISTRY_URL")
	nodeName := v.GetString("NODE_NAME")

	stopCh := make(chan os.Signal)
	var wg sync.WaitGroup
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	hostCacheUpdater := agent.NewHostCacheUpdater(nodeName, registryURL, hostPeriod, &wg)
	conntrackUpdater := agent.NewConntrackUpdater(nodeName, conntrackPeriod, &wg)
	server := agent.NewServer(addr, &wg)

	wg.Add(3)
	go hostCacheUpdater.StartHostCacheUpdater()
	go conntrackUpdater.StartConntrackUpdater()
	go server.StartServer()

	<-stopCh

	hostCacheUpdater.StopHostCacheUpdater()
	conntrackUpdater.StopConntrackUpdater()
	server.StopServer()

	wg.Wait()
	glog.Infof("graceful shutdown")
}
