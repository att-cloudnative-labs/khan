package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"egbitbucket.dtvops.net/com/agent/internal/platform/netclient"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWelcome(t *testing.T) {
	result := Welcome()

	if result.Hello != "World!" {
		t.Errorf("Welcome did not return the expected results. Expected World! got %s", result.Hello)
	}
}

// testServer takes in a fileLocation and returns that data for the response body
func testServer(fileLocation string) (*httptest.Server, error) {
	var ts = &httptest.Server{}
	content, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		return ts, err
	}

	ts = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, string(content))
	}))

	return ts, nil
}

func TestExampleClientCall(t *testing.T) {
	ts, err := testServer("testdata/example_response1.json")
	if err != nil {
		t.Errorf("failed to start the server! %v", err)
		return
	}

	client := ts.Client() // test server client works with insecure cert that test server creates
	netclient.SetDME("example_call", client)

	ctx := context.Background()
	hello, err := ExampleCall(ctx, ts.URL)
	if err != nil {
		t.Errorf("ExampleCall has failed: %v", err)
		return
	}

	b, err := json.Marshal(hello)
	if err != nil {
		t.Errorf("failed to marshall ExampleCall response: %v", err)
	}

	if string(b) != `{"Hello":"World!"}` {
		t.Errorf("Incorrect response. Expecting {\"Hello\":\"World!\"} got %s", b)
	}
}