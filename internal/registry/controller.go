package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/att-cloudnative-labs/khan/pkg/hosts"
	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	informers "k8s.io/client-go/informers/core/v1"
	listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

// AtomicHostCache atomic HostCache
type AtomicHostCache struct {
	sync.RWMutex
	internal hosts.HostCache
}

// NewCache create new Cache
func NewCache() *AtomicHostCache {
	return &AtomicHostCache{
		internal: make(map[string]hosts.Host),
	}
}

var hostCache = NewCache()

// GetCache get the current Cache
func (p *AtomicHostCache) GetCache() hosts.HostCache {
	p.RLock()
	defer p.RUnlock()
	return p.internal
}

// PutCache Add entry to Cache
func (p *AtomicHostCache) PutCache(ip string, host hosts.Host) {
	p.Lock()
	defer p.Unlock()
	p.internal[ip] = host
}

// Controller initializes and updates host cache
type Controller struct {
	podLister      listers.PodLister
	podsSynced     cache.InformerSynced
	serviceLister  listers.ServiceLister
	servicesSynced cache.InformerSynced
	nodeLister     listers.NodeLister
	nodesSynced    cache.InformerSynced
}

// NewController new controller
func NewController(podInformer informers.PodInformer,
	serviceInformer informers.ServiceInformer,
	nodeInformer informers.NodeInformer) *Controller {
	return &Controller{
		podLister:      podInformer.Lister(),
		podsSynced:     podInformer.Informer().HasSynced,
		serviceLister:  serviceInformer.Lister(),
		servicesSynced: serviceInformer.Informer().HasSynced,
		nodeLister:     nodeInformer.Lister(),
		nodesSynced:    serviceInformer.Informer().HasSynced,
	}
}

// Start run controller
func (c *Controller) Start(stopCh <-chan struct{}) {
	glog.Info("waiting for caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.podsSynced); !ok {
		panic(fmt.Errorf("failed to wait for caches to sync"))
	}
	if err := c.BuildCache(); err != nil {
		glog.Error(err.Error())
	}
	ticker := time.NewTicker(time.Duration(30) * time.Second)
	for {
		select {
		case <-ticker.C:
			go func() {
				err := c.BuildCache()
				if err != nil {
					glog.Error(err.Error())
				}
			}()
		case <-stopCh:
			ticker.Stop()
			return
		}
	}
}

// BuildCache build updated Cache
func (c *Controller) BuildCache() error {
	pods, err := c.podLister.Pods(metav1.NamespaceAll).List(labels.Everything())
	if err != nil {
		return fmt.Errorf("error listing pods: %s", err.Error())
	}
	for _, pod := range pods {

		hostCache.internal[pod.Status.PodIP] = hosts.Host{
			Type:      "Pod",
			Namespace: pod.Namespace,
			Name:      pod.Name,
			App:       pod.GetLabels()["app"],
			NodeIP:    pod.Status.HostIP,
		}

	}
	services, err := c.serviceLister.Services(metav1.NamespaceAll).List(labels.Everything())
	if err != nil {
		return fmt.Errorf("error listing services: %s", err.Error())
	}
	for _, service := range services {
		if len(service.Spec.ClusterIP) > 0 {
			hostCache.internal[service.Spec.ClusterIP] = hosts.Host{
				Type:      "Service",
				Namespace: service.Namespace,
				Name:      service.Name,
				App:       service.GetLabels()["app"],
			}
		}
	}
	nodes, err := c.nodeLister.List(labels.Everything())
	if err != nil {
		return fmt.Errorf("error listig nodes: %s", err.Error())
	}
	for _, node := range nodes {
		hostCache.internal[node.Status.Addresses[0].Address] = hosts.Host{
			Type:   "Node",
			NodeIP: node.Status.Addresses[0].Address,
		}
	}
	glog.Infof("hosts synced")
	return nil
}

// RequestHandler rest request handler
func RequestHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		glog.Errorf("Error parsing request for NodeMapping: %s", err.Error())
		return
	}
	err = json.NewEncoder(w).Encode(hostCache.GetCache())
	if err != nil {
		glog.Errorf("Error encoding PodCache: %s", err.Error())
	}
}
