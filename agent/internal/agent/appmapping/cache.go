package appmapping

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// App contains fields for tags for an app
type App struct {
	Namespace string `json:"namespace"`
	AppName   string `json:"appName"`
	PodName   string `json:"podName"`
}

var appmappingCache = make(map[string]App)

type AppmappingController struct {
	NodeName      string
	AppmappingUrl string
	UpdatePeriod  int
}

func NewAppmappingController(nodeName string, appmappingUrl string, updatePeriod int) AppmappingController {
	return AppmappingController{
		NodeName:      nodeName,
		AppmappingUrl: appmappingUrl,
		UpdatePeriod:  updatePeriod,
	}
}

func (c *AppmappingController) Start(stopCh chan struct{}) {
	ticker := time.NewTicker(time.Duration(c.UpdatePeriod) * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Println("Retrieving app cache")
				resp, err := http.Get(fmt.Sprintf("%s?node=%s", c.AppmappingUrl, c.NodeName))
				if err != nil {
					panic(fmt.Errorf("Error retrieving appmappings: %s", err.Error()))
				}
				defer resp.Body.Close()
				json.NewDecoder(resp.Body).Decode(&appmappingCache)
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
	var newcache map[string]App
	if err := decoder.Decode(&newcache); err != nil {
		fmt.Println("Error processing cache update")
	}
	appmappingCache = newcache
	fmt.Println("cache was set :)")
}

// GetFullCache returns full cache
func GetFullCache(w http.ResponseWriter, _ *http.Request) {
	json.NewEncoder(w).Encode(appmappingCache)
}

// Get returns app for a given podip
func Get(podIP string) App {
	return appmappingCache[podIP]
}
