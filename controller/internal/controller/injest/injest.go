package injest

import (
	"net/http"
)

func InjestCallHandler(w http.ResponseWriter, r *http.Request) {
	// body, err := ioutil.ReadAll(r.Body)
	// if err != nil {
	// 	fmt.Printf("Error reading body: %s", err.Error())
	// 	return
	// }
	// newconnections := string(body)
	// connections.AddConnections(newconnections)
	// w.Header().Set("Content-Type", "application/json")
	// w.Write([]byte("{}"))
}
