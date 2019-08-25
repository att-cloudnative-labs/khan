package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/att-cloudnative-labs/khan/pkg/hosts"
	"github.com/golang/glog"
)

var hostCache = make(map[string]hosts.Host)

// Controller builds and serves host cache
type Controller struct {
	NodeName     string
	RegistryURL  string
	UpdatePeriod int
}

// NewController create new Controller
func NewController(nodeName string, registryURL string, updatePeriod int) Controller {
	return Controller{
		NodeName:     nodeName,
		RegistryURL:  registryURL,
		UpdatePeriod: updatePeriod,
	}
}

// Start Controller
func (c *Controller) Start(stopCh chan struct{}) {
	ticker := time.NewTicker(time.Duration(c.UpdatePeriod) * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				glog.Info("Retrieving host cache")
				func() {
					resp, err := http.Get(fmt.Sprintf("%s?node=%s", c.RegistryURL, c.NodeName))
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

			case <-stopCh:
				ticker.Stop()
				return
			}
		}
	}()
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
