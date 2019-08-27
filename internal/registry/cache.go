package registry

import (
	"encoding/json"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/att-cloudnative-labs/khan/pkg/hosts"
	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
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

// CacheBuilder initializes and updates host cache
type CacheBuilder struct {
	informerFactory informers.SharedInformerFactory
	podLister       listers.PodLister
	podsSynced      cache.InformerSynced
	serviceLister   listers.ServiceLister
	servicesSynced  cache.InformerSynced
	nodeLister      listers.NodeLister
	nodesSynced     cache.InformerSynced
	updatePeriod    int
	stopCh          chan struct{}
	wg              *sync.WaitGroup
}

// NewCacheBuilder new CacheBuilder
func NewCacheBuilder(apiserver string, kubeconfig string, updatePeriod int, wg *sync.WaitGroup) *CacheBuilder {

	cfg, err := clientcmd.BuildConfigFromFlags(apiserver, kubeconfig)
	if err != nil {
		panic(fmt.Errorf("error creating controller: %s", err.Error()))
	}
	stdClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(fmt.Errorf("error creating controller: %s", err.Error()))
	}
	informerFactory := informers.NewSharedInformerFactory(stdClient, time.Second*30)
	return &CacheBuilder{
		informerFactory: informerFactory,
		podLister:       informerFactory.Core().V1().Pods().Lister(),
		podsSynced:      informerFactory.Core().V1().Pods().Informer().HasSynced,
		serviceLister:   informerFactory.Core().V1().Services().Lister(),
		servicesSynced:  informerFactory.Core().V1().Services().Informer().HasSynced,
		nodeLister:      informerFactory.Core().V1().Nodes().Lister(),
		nodesSynced:     informerFactory.Core().V1().Nodes().Informer().HasSynced,
		updatePeriod:    updatePeriod,
		stopCh:          make(chan struct{}),
		wg:              wg,
	}
}

// StartCacheBuilder run CacheBuilder
func (r *CacheBuilder) StartCacheBuilder() {
	defer r.wg.Done()
	glog.Info("Starting CacheBuilder")
	r.informerFactory.Start(r.stopCh)
	glog.Info("waiting for informer to sync")
	if ok := cache.WaitForCacheSync(r.stopCh, r.podsSynced); !ok {
		panic(fmt.Errorf("failed to wait for caches to sync"))
	}
	if err := r.BuildCache(); err != nil {
		glog.Error(err.Error())
	}
	ticker := time.NewTicker(time.Duration(r.updatePeriod) * time.Second)
	for {
		select {
		case <-ticker.C:
			go func() {
				err := r.BuildCache()
				if err != nil {
					glog.Error(err.Error())
				}
			}()
		case <-r.stopCh:
			ticker.Stop()
			return
		}
	}
}

// StopCacheBuilder stop CacheBuilder
func (r *CacheBuilder) StopCacheBuilder() {
	glog.Info("shutting down CacheBuilder")
	close(r.stopCh)
}

// BuildCache build updated Cache
func (r *CacheBuilder) BuildCache() error {
	pods, err := r.podLister.Pods(metav1.NamespaceAll).List(labels.Everything())
	if err != nil {
		return fmt.Errorf("error listing pods: %s", err.Error())
	}
	for _, pod := range pods {

		hostCache.internal[pod.Status.PodIP] = hosts.Host{
			Type:      "pod",
			Namespace: pod.Namespace,
			Name:      pod.Name,
			App:       pod.GetLabels()["app"],
			NodeIP:    pod.Status.HostIP,
		}

	}
	services, err := r.serviceLister.Services(metav1.NamespaceAll).List(labels.Everything())
	if err != nil {
		return fmt.Errorf("error listing services: %s", err.Error())
	}
	for _, service := range services {
		if len(service.Spec.ClusterIP) > 0 && service.Spec.ClusterIP != "None" {
			hostCache.internal[service.Spec.ClusterIP] = hosts.Host{
				Type:      "service",
				Namespace: service.Namespace,
				Name:      service.Name,
				App:       service.GetLabels()["app"],
			}
		}
	}
	nodes, err := r.nodeLister.List(labels.Everything())
	if err != nil {
		return fmt.Errorf("error listig nodes: %s", err.Error())
	}
	for _, node := range nodes {
		hostCache.internal[node.Status.Addresses[0].Address] = hosts.Host{
			Type:   "node",
			Name:   node.Name,
			NodeIP: node.Status.Addresses[0].Address,
		}
		podCIDR := node.Spec.PodCIDR
		if len(podCIDR) > 0 {
			ip, _, err := net.ParseCIDR(podCIDR)
			if err != nil {
				glog.Errorf("Error parsing PodCIDR: %s", err.Error())
			}
			ipString := ip.To4().String()
			hostCache.internal[ipString] = hosts.Host{
				Type:   "network",
				Name:   node.Name,
				NodeIP: node.Status.Addresses[0].Address,
			}
			cniIP := ip.To4()
			cniIP[3]++
			hostCache.internal[cniIP.String()] = hosts.Host{
				Type:   "gateway",
				Name:   node.Name,
				NodeIP: node.Status.Addresses[0].Address,
			}
		}
	}
	glog.Infof("hosts cache build complete")
	return nil
}

// GetCacheRequestHandler rest request handler
func GetCacheRequestHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		glog.Errorf("Error parsing request for GetCache: %s", err.Error())
		return
	}
	err = json.NewEncoder(w).Encode(hostCache.GetCache())
	if err != nil {
		glog.Errorf("Error encoding HostCache: %s", err.Error())
	}
}
