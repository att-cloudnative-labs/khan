package connections

import(
	"net/http"
)

type Connections string

var currentConnections string = ""

func ClearConnections() {
	currentConnections = ""
}

func AddConnections(connections string) {
	currentConnections = currentConnections + "\n" + connections
}

func ConnectionsCallHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(currentConnections))
}