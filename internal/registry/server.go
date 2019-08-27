package registry

import (
	"context"
	"github.com/go-chi/chi"
	"github.com/golang/glog"
	"net/http"
	"sync"
	"time"
)

// Server http server wrapper
type Server struct {
	server *http.Server
	ctx    context.Context
	cancel context.CancelFunc
	stopCh chan bool
	wg     *sync.WaitGroup
}

// Set is used to set the routes
func SetRoutes(httpService *chi.Mux) {
	httpService.Get("/cache", GetCacheRequestHandler)
}

// NewServer new server
func NewServer(addr string, wg *sync.WaitGroup) *Server {
	r := chi.NewRouter()
	SetRoutes(r)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	return &Server{
		server: &http.Server{
			Addr:    addr,
			Handler: r,
		},
		ctx:    ctx,
		cancel: cancel,
		wg:     wg,
	}
}

// StartServer start server
func (s *Server) StartServer() {
	defer s.cancel()
	defer s.wg.Done()
	glog.Infof("Starting server on %s\n", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		glog.Fatalf("error serving http: %s", err.Error())
	}
}

// StopServer stop server
func (s *Server) StopServer() {
	glog.Info("shutting down server")
	if err := s.server.Shutdown(s.ctx); err != nil {
		glog.Errorf("error during server shutdown: %s", err.Error())
	}
}
