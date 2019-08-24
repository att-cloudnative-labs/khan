package mappings

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/att-cloudnative-labs/khan/pkg/mappings"
)

var localTargetCache = make(map[string]mappings.Target)

// LocalTargetCacheController controller for retrieving the local cache
type LocalTargetCacheController struct {
	NodeName     string
	RegistryURL  string
	UpdatePeriod int
}

// NewLocalTargetCacheController create new LocalTargetCacheController
func NewLocalTargetCacheController(nodeName string, registryURL string, updatePeriod int) LocalTargetCacheController {
	return LocalTargetCacheController{
		NodeName:     nodeName,
		RegistryURL:  registryURL,
		UpdatePeriod: updatePeriod,
	}
}

// Start LocalTargetCacheController
func (c *LocalTargetCacheController) Start(stopCh chan struct{}) {
	ticker := time.NewTicker(time.Duration(c.UpdatePeriod) * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Println("Retrieving app cache")
				resp, err := http.Get(fmt.Sprintf("%s?node=%s", c.RegistryURL, c.NodeName))
				if err != nil {
					panic(fmt.Errorf("Error retrieving appmappings: %s", err.Error()))
				}
				defer resp.Body.Close()
				json.NewDecoder(resp.Body).Decode(&localTargetCache)
				fmt.Println("Successfully retrieved app cache")
			case <-stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

// SetCache sets the entire cache
func SetCache(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var newcache map[string]mappings.Target
	if err := decoder.Decode(&newcache); err != nil {
		fmt.Println("Error processing cache update")
	}
	localTargetCache = newcache
	fmt.Println("cache was set :)")
}

// GetFullCache returns full cache
func GetFullCache(w http.ResponseWriter, _ *http.Request) {
	json.NewEncoder(w).Encode(localTargetCache)
}

// Get returns target for a given IP
func Get(ip string) mappings.Target {
	return localTargetCache[ip]
}
