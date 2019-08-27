package agent

import (
	"encoding/json"
	"fmt"
	"github.com/att-cloudnative-labs/khan/pkg/hosts"
	"github.com/golang/glog"
	"net/http"
	"sync"
	"time"
)

var hostCache = make(map[string]hosts.Host)

// HostCacheUpdater builds and serves host cache
type HostCacheUpdater struct {
	NodeName     string
	RegistryURL  string
	UpdatePeriod int
	stopCh       chan bool
	wg           *sync.WaitGroup
}

// NewHostCacheUpdater create new HostCacheUpdater
func NewHostCacheUpdater(nodeName string, registryURL string, updatePeriod int, wg *sync.WaitGroup) HostCacheUpdater {
	return HostCacheUpdater{
		NodeName:     nodeName,
		RegistryURL:  registryURL,
		UpdatePeriod: updatePeriod,
		stopCh:       make(chan bool),
		wg:           wg,
	}
}

// StartHostCacheUpdater HostCacheUpdater
func (c *HostCacheUpdater) StartHostCacheUpdater() {
	defer c.wg.Done()
	glog.Info("starting HostCacheUpdater")
	ticker := time.NewTicker(time.Duration(c.UpdatePeriod) * time.Second)
	for {
		select {
		case <-ticker.C:
			glog.Info("Retrieving host cache")
			func() {
				resp, err := http.Get(c.RegistryURL)
				if err != nil {
					glog.Errorf("error retrieving host cache: %s", err.Error())
					return
				}
				defer func() {
					if err := resp.Body.Close(); err != nil {
						glog.Errorf("error closing response reader")
					}
				}()
				if err = json.NewDecoder(resp.Body).Decode(&hostCache); err != nil {
					glog.Errorf("error decoding host cache: %s", err.Error())
					return
				}
				glog.Info("Successfully retrieved host cache")
			}()

		case <-c.stopCh:
			glog.Info("shutting down HostCacheUpdater")
			ticker.Stop()
			return
		}
	}
}

// StopHostCacheUpdater stopCh
func (c *HostCacheUpdater) StopHostCacheUpdater() {
	close(c.stopCh)
}

// SetCache sets the entire cache
func SetCache(_ http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var newCache hosts.HostCache
	if err := decoder.Decode(&newCache); err != nil {
		fmt.Println("Error processing cache update")
	}
	hostCache = newCache
	fmt.Println("cache was set :)")
}

// GetCache returns entire cache
func GetCache(w http.ResponseWriter, _ *http.Request) {
	if err := json.NewEncoder(w).Encode(hostCache); err != nil {
		glog.Errorf("error serving host cache: %s", err.Error())
	}
}

// Get returns Host for a given IP
func GetHost(ip string) hosts.Host {
	return hostCache[ip]
}
