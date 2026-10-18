//go:build windows && amd64

package localapi_test

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"hopvista/internal/engine"
	"hopvista/internal/localapi"
	"hopvista/internal/probe/windowsipv4"
	"hopvista/internal/result"
	"hopvista/internal/run"
)

func TestLiveEngineThroughLocalAPI(t *testing.T) {
	if os.Getenv("HOPVISTA_LIVE_ENGINE") != "1" {
		t.Skip("set HOPVISTA_LIVE_ENGINE=1 for opt-in local integration")
	}
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) }))
	defer targetServer.Close()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	handler, err := localapi.NewHandler(listener.Addr().String(), run.NewManager(engine.New(windowsipv4.Adapter{})))
	if err != nil {
		t.Fatal(err)
	}
	apiServer := &http.Server{Handler: handler}
	defer apiServer.Close()
	go func() { _ = apiServer.Serve(listener) }()
	base := "http://" + listener.Addr().String()
	client := &http.Client{Timeout: 15 * time.Second}
	response, err := client.Get(base + "/api/v1/session")
	if err != nil {
		t.Fatal(err)
	}
	var session struct {
		Token string `json:"mutation_token"`
	}
	if err := json.NewDecoder(response.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	request, err := http.NewRequest(http.MethodPost, base+"/api/v1/runs", bytes.NewBufferString(`{"target_url":"`+targetServer.URL+`/","max_hops":4}`))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Origin", base)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-HopVista-Token", session.Token)
	response, err = client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	var started struct {
		RunID     string `json:"run_id"`
		StatusURL string `json:"status_url"`
	}
	if err := json.NewDecoder(response.Body).Decode(&started); err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusAccepted || started.RunID == "" {
		t.Fatalf("start status %d: %+v", response.StatusCode, started)
	}
	until := time.Now().Add(10 * time.Second)
	for time.Now().Before(until) {
		response, err = client.Get(base + started.StatusURL)
		if err != nil {
			t.Fatal(err)
		}
		var snapshot result.Run
		if err := json.NewDecoder(response.Body).Decode(&snapshot); err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if !snapshot.State.Terminal() {
			time.Sleep(25 * time.Millisecond)
			continue
		}
		if snapshot.State != result.RunCompleted || len(snapshot.Families) != 2 || len(snapshot.Families[0].Exchanges) != 1 || len(snapshot.Families[0].Paths) != 1 || len(snapshot.Findings) != 1 {
			t.Fatalf("terminal run: %+v", snapshot)
		}
		x := snapshot.Families[0].Exchanges[0]
		p := snapshot.Families[0].Paths[0]
		if x.ResponseStatus == nil || *x.ResponseStatus != 503 || x.HTTP.Outcome != result.Completed || p.Route.Outcome != result.Completed || p.Target.Replied != 5 || p.Trace.Outcome != result.DestinationReplied || p.Trace.TargetRole != "final_host" {
			t.Fatalf("integrated evidence: exchange=%+v path=%+v", x, p)
		}
		return
	}
	t.Fatal("diagnostic run did not reach terminal state")
}
