package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloud-native-labs/khan/cmd/registry/routes"
	"github.com/cloud-native-labs/khan/cmd/registry/config"
	"github.com/cloud-native-labs/khan/internal/mappings"

	"github.com/go-chi/chi"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	config.Set()
	port := config.Registry.GetString("SERVER_PORT")
	apiserver := config.Registry.GetString("APISERVER")
	kubeconfig := config.Registry.GetString("KUBECONFIG")

	r := chi.NewRouter()

	routes.Set(r)

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

	err = http.ListenAndServe(port, r)

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
