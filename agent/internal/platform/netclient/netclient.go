package netclient

import (
	"net/http"
	"sync"
	"time"

	"egbitbucket.dtvops.net/com/goatt"
)

var dmeClients map[string]Httpclient
var clients map[string]Httpclient
var lock sync.RWMutex
var throttleLimit = 1000

// Preset is used to define the preset in the client, value is set elsewhere such as cmd/webapp/main.go
var Preset string

func init() {
	dmeClients = make(map[string]Httpclient)
	clients = make(map[string]Httpclient)
}

//Httpclient is used to pass client
type Httpclient interface {
	Do(req *http.Request) (*http.Response, error)
}

// ClientDME returns the client or makes a generic one if nothing is set. dmeName and dmeVersion
// is only used if a client hasn't been made yet.
func ClientDME(callAlias, dmeName, dmeVersion string) Httpclient {
	lock.RLock()
	if client, ok := dmeClients[callAlias]; ok {
		lock.RUnlock()
		return client
	}
	lock.RUnlock()

	lock.Lock()
	if client, ok := dmeClients[callAlias]; ok {
		lock.Unlock()
		return client
	}
	client := goatt.NewRestClient(callAlias, 5000*time.Millisecond).
		Preset(Preset).
		WithDME(dmeName, dmeVersion).
		Build()
	dmeClients[callAlias] = client
	lock.Unlock()

	return client
}

// Client returns the client or makes a generic one if nothing is set.
// Generic clients that are not DME.
func Client(callAlias string) Httpclient {
	lock.RLock()
	if client, ok := clients[callAlias]; ok {
		lock.RUnlock()
		return client
	}
	lock.RUnlock()

	lock.Lock()
	if client, ok := clients[callAlias]; ok {
		lock.Unlock()
		return client
	}
	client := goatt.NewRestClient(callAlias, 5000*time.Millisecond).
		Preset("OV").
		Build()
	clients[callAlias] = client
	lock.Unlock()

	return client
}

// Set is used to set a custom client instead of the default one Client creates.
func Set(callAlias string, client Httpclient) {
	lock.Lock()
	clients[callAlias] = client
	lock.Unlock()
}

// SetDME is used to set a custom client instead of the default one Client creates.
func SetDME(callAlias string, client Httpclient) {
	lock.Lock()
	dmeClients[callAlias] = client
	lock.Unlock()
}
