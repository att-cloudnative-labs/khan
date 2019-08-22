package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"egbitbucket.dtvops.net/com/controller/cmd/webapp/config"
	"egbitbucket.dtvops.net/com/controller/cmd/webapp/metrics"
	"egbitbucket.dtvops.net/com/controller/cmd/webapp/routes"
	"egbitbucket.dtvops.net/com/controller/internal/controller/appmappings"
	"egbitbucket.dtvops.net/com/controller/internal/platform/netclient"
	"egbitbucket.dtvops.net/com/goatt"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	goattVersion = "0.0.0"
)

func main() {
	config.Set()
	port := config.Registry.GetString("SERVER_PORT")
	apiserver := config.Registry.GetString("APISERVER")
	kubeconfig := config.Registry.GetString("KUBECONFIG")
	netclient.Preset = config.Registry.GetString("PRESET")

	goattService := goatt.NewGoattService(port)
	enableProfiling := config.Registry.GetBool("ENABLE_PROFILING")
	goattService.Httpservice.EnableProfiling(enableProfiling)
	// add to /info endpoint
	goattService.Httpservice.ExposeInfo("goattVersion", goattVersion)
	metrics.Set(goattService.Metrics)
	routes.Set(goattService.Httpservice)

	// appmapping
	cfg, err := clientcmd.BuildConfigFromFlags(apiserver, kubeconfig)
	if err != nil {
		panic(fmt.Errorf("error creating controller: %s", err.Error()))
	}
	stdClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(fmt.Errorf("error creating controller: %s", err.Error()))
	}
	
	stopCh := setupSignalHandler()
	informerFactory := informers.NewSharedInformerFactory(stdClient, time.Second*30)

	controller, err := appmappings.NewController(informerFactory.Core().V1().Pods())
	if err != nil {
		panic(fmt.Errorf("error creating controller: %s", err.Error()))
	}

	go informerFactory.Start(stopCh)
	go controller.Start(stopCh)

	fmt.Printf("Starting application on port %s\n", port)
	err = goattService.Httpservice.Start()
	if err != nil {
		panic(err)
	}
}

var onlyOneSignalHandler = make(chan struct{})
var shutdownHandler chan os.Signal

func setupSignalHandler() <-chan struct{} {
	close(onlyOneSignalHandler) // panics when called twice

	shutdownHandler = make(chan os.Signal, 2)

	stop := make(chan struct{})
	signal.Notify(shutdownHandler, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-shutdownHandler
		close(stop)
		<-shutdownHandler
		os.Exit(1) // second signal. Exit directly
	}()
	return stop
}
