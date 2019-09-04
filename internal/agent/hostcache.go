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

type HostCache struct {
	sync.RWMutex
	internal map[string]hosts.Host
}

func NewHostCache() *HostCache {
	return &HostCache{
		internal: make(map[string]hosts.Host),
	}
}

var hostCache = NewHostCache()

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
				var incomingHostCache map[string]hosts.Host
				if err = json.NewDecoder(resp.Body).Decode(&incomingHostCache); err != nil {
					glog.Errorf("error decoding host cache: %s", err.Error())
					return
				}
				hostCache.Lock()
				defer hostCache.Unlock()
				hostCache.internal = incomingHostCache
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
	hostCache.Lock()
	defer hostCache.Unlock()
	hostCache.internal = newCache
	fmt.Println("cache was set :)")
}

// GetCache returns entire cache
func GetCache(w http.ResponseWriter, _ *http.Request) {
	hostCache.RLock()
	defer hostCache.RUnlock()
	mapCopy := make(map[string]hosts.Host)
	for k, v := range hostCache.internal {
		mapCopy[k] = v
	}
	if err := json.NewEncoder(w).Encode(mapCopy); err != nil {
		glog.Errorf("error serving host cache: %s", err.Error())
	}
}

// Get returns Host for a given IP
func GetHost(ip string) hosts.Host {
	hostCache.RLock()
	defer hostCache.RUnlock()
	return hostCache.internal[ip]
}
