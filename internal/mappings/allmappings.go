package mappings

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/cloud-native-labs/khan/pkg/mappings"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	informers "k8s.io/client-go/informers/core/v1"
	listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

// TargetCache IP to App cache
type TargetCache struct {
	sync.RWMutex
	internal map[string]App
}

// NewTargetCache create new TargetCache
func NewTargetCache() *TargetCache {
	return &TargetCache{
		internal: make(map[string]mappings.Target),
	}
}

var targetCache = TargetCache()

// GetCache get the current TargetCache
func (p *TargetCache) GetCache() map[string]mappings.Target {
	p.RLock()
	defer p.RUnlock()
	return p.internal
}

// PutCache Add entry to TargetCache
func (p *TargetCache) PutCache(ip string, target mappings.Target) {
	p.Lock()
	defer p.Unlock()
	p.internal[ip] = target
}

// TargetCacheController initializes and updates target cache
type TargetCacheController struct {
	podLister  listers.PodLister
	podsSynced cache.InformerSynced
}

// NewController new controller
func NewController(podInformer informers.PodInformer) (*TargetCacheController, error) {
	return &TargetCacheController{
		podLister:  podInformer.Lister(),
		podsSynced: podInformer.Informer().HasSynced,
	}, nil
}

// Start run controller
func (c *TargetCacheController) Start(stopCh <-chan struct{}) {
	fmt.Println("waiting for caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.podsSynced); !ok {
		panic(fmt.Errorf("failed to wait for caches to sync"))
	}
	c.BuildCache()
	ticker := time.NewTicker(time.Duration(30) * time.Second)
	for {
		select {
		case <-ticker.C:
			go c.BuildCache()
		case <-stopCh:
			ticker.Stop()
			return
		}
	}
}

// BuildCache build updated TargetCache
func (c *TargetCacheController) BuildCache() {
	pods, err := c.podLister.Pods(metav1.NamespaceAll).List(labels.Everything())
	if err != nil {
		panic(fmt.Errorf("error listing pods: %s", err.Error()))
	}
	for _, pod := range pods {

		podCache.internal[pod.Status.PodIP] = App{
			Namespace: pod.Namespace,
			AppName:   pod.GetLabels()["app"],
			PodName:   pod.Name,
			NodeIp:    pod.Status.HostIP,
		}

	}
	fmt.Println("appmapping synced")
}

// RequestHandler rest request handler
func RequestHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Printf("Error parsing request for NodeMapping: %s", err.Error())
		return
	}
	err = json.NewEncoder(w).Encode(podCache.GetCache())
	if err != nil {
		fmt.Printf("Error encoding PodCache: %s", err.Error())
	}
}
