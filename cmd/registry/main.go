package main

import (
	"flag"
	"github.com/att-cloudnative-labs/khan/internal/registry"
	"github.com/golang/glog"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	flag.Parse()

	v := viper.New()
	v.SetDefault("SERVER_ADDR", ":8080")
	v.SetDefault("CACHE_SYNC_PERIOD", 30)
	v.AutomaticEnv()

	addr := v.GetString("SERVER_ADDR")
	apiserver := v.GetString("APISERVER")
	kubeconfig := v.GetString("KUBECONFIG")
	cacheSyncPeriod := v.GetInt("CACHE_SYNC_PERIOD")

	stopCh := make(chan os.Signal)
	var wg sync.WaitGroup
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	cacheBuilder := registry.NewCacheBuilder(apiserver, kubeconfig, cacheSyncPeriod, &wg)
	server := registry.NewServer(addr, &wg)

	wg.Add(2)
	go cacheBuilder.StartCacheBuilder()
	go server.StartServer()

	<-stopCh

	cacheBuilder.StopCacheBuilder()
	server.StopServer()

	wg.Wait()
	glog.Info("graceful shutdown")
}
