package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/att-cloudnative-labs/khan/cmd/registry/config"
	"github.com/att-cloudnative-labs/khan/cmd/registry/routes"
	"github.com/att-cloudnative-labs/khan/internal/registry"

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

	controller := registry.NewController(informerFactory.Core().V1().Pods(),
		informerFactory.Core().V1().Services(), informerFactory.Core().V1().Nodes())

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
