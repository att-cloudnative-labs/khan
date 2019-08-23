package agent

import (
	"context"
	"egbitbucket.dtvops.net/com/goatt/pkg/yawl"
	"egbitbucket.dtvops.net/com/goatt/pkg/yawl/log"
	"encoding/json"
	"fmt"
	"github.com/cloud-native-labs/khan/agent/internal/platform/netclient"
	"net/http"
)

// Here is where you define your business logic and your data structure.

var (
	statusDebug                = yawl.NewStatus("DEBUG_MESSAGE", 700, "Example of a debug message.")
	statusErrorExampleFunction = yawl.NewStatus("EXAMPLE_FUNCTION_ERROR", 4001, "The example function failed.")
)

// HelloWorld is an example struct
type HelloWorld struct {
	Hello string `json:"Hello"`
}

// MetricWelcomeCount is an example of a custom metric.
// Doesn't do anything but is set to do something from cmd/webapp/metrics
var MetricWelcomeCount = func(_ string) {
	// NOOP - deliberately empty to avoid nil checks on this function
}

// Welcome returns Welcome!
func Welcome() HelloWorld {
	MetricWelcomeCount("labelValueExample")
	hello := HelloWorld{Hello: "World!"}
	return hello
}

// ExampleCall is an example of a downstream call
func ExampleCall(ctx context.Context, url string) (HelloWorld, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		err = fmt.Errorf("failed to make request object: %v", err)
		return HelloWorld{}, err
	}

	req = req.WithContext(ctx)
	client := netclient.ClientDME("example_call", "sandbox-test.fakeservice", "1")

	resp, err := client.Do(req)
	if err != nil {
		return HelloWorld{}, err
	}
	defer resp.Body.Close()

	var hello HelloWorld
	err = json.NewDecoder(resp.Body).Decode(&hello)
	if err != nil {
		err = fmt.Errorf("failed to decode response body as json: %v", err)
		return hello, err
	}

	return hello, err
}

func exampleUnexportedFunction() {
	err := fmt.Errorf("forcing this function to fail")
	log.Error(statusErrorExampleFunction, yawl.Err(err))
	log.Debug(statusDebug, yawl.Str("extra_message", "Example of adding more information."))
}
