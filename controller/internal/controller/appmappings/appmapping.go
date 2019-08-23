package appmappings

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	informers "k8s.io/client-go/informers/core/v1"
	listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

// App contains fields for tags for an app
type App struct {
	Namespace string `json:"namespace"`
	AppName   string `json:"appName"`
	PodName   string `json:"podName"`
	NodeIp    string `json:"nodeIP"`
}

// PodCache IP to App cache
type PodCache struct {
	sync.RWMutex
	internal map[string]App
}

func NewPodCache() *PodCache {
	return &PodCache{
		internal: make(map[string]App),
	}
}

var podCache = NewPodCache()

func (p *PodCache) GetCache() map[string]App {
	p.RLock()
	defer p.RUnlock()
	return p.internal
}

func (p *PodCache) PutCache(podIP string, app App) {
	p.Lock()
	defer p.Unlock()
	p.internal[podIP] = app
}

// AppmappingController initializes and updates appmapping cache
type AppmappingController struct {
	podLister  listers.PodLister
	podsSynced cache.InformerSynced
}

func NewController(podInformer informers.PodInformer) (*AppmappingController, error) {
	return &AppmappingController{
		podLister:  podInformer.Lister(),
		podsSynced: podInformer.Informer().HasSynced,
	}, nil
}

func (c *AppmappingController) Start(stopCh <-chan struct{}) {
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

func (c *AppmappingController) BuildCache() {
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

func NodeMappingRequestHandler(w http.ResponseWriter, r *http.Request) {
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
