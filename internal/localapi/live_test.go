package localapi_test

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"hopvista/internal/localapi"
	"hopvista/internal/run"
)

// Run only where loopback sockets are permitted: HOPVISTA_LIVE_LOOPBACK=1 go
// test ./internal/localapi -run TestLiveLoopback. In-memory handler tests remain
// the default in the restricted development sandbox.
func TestLiveLoopback(t *testing.T) {
	if os.Getenv("HOPVISTA_LIVE_LOOPBACK") != "1" {
		t.Skip("opt-in live loopback socket check")
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	h, err := localapi.NewHandler(listener.Addr().String(), run.NewManager(run.Placeholder{}))
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: h, ReadHeaderTimeout: 5 * time.Second}
	go server.Serve(listener)
	defer server.Close()
	base := "http://" + listener.Addr().String()
	client := &http.Client{Timeout: 3 * time.Second, Transport: &http.Transport{Proxy: nil}}
	resp, err := client.Get(base + "/api/v1/session")
	if err != nil {
		t.Fatal(err)
	}
	var session struct {
		MutationToken string `json:"mutation_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 || session.MutationToken == "" {
		t.Fatal("session unavailable")
	}
	r, err := http.NewRequest("POST", base+"/api/v1/runs", bytes.NewBufferString(`{"target_url":"https://example.com/"}`))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Origin", base)
	r.Header.Set("X-HopVista-Token", session.MutationToken)
	r.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 202 {
		t.Fatalf("start returned %d", resp.StatusCode)
	}
}
