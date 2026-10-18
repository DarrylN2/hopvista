package localapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"hopvista/internal/localapi"
	"hopvista/internal/result"
	"hopvista/internal/run"
)

type executorFunc func(context.Context, run.Input, func([]result.Family)) ([]result.Family, []result.Finding, error)

func (f executorFunc) Run(ctx context.Context, in run.Input, publish func([]result.Family)) ([]result.Family, []result.Finding, error) {
	return f(ctx, in, publish)
}

type memoryServer struct{ URL string }
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func startServer(t *testing.T, executor run.Executor) (*memoryServer, *http.Client, string) {
	t.Helper()
	s := &memoryServer{URL: "http://127.0.0.1:8123"}
	h, err := localapi.NewHandler("127.0.0.1:8123", run.NewManager(executor))
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w.Result(), nil
	})}
	r, err := client.Get(s.URL + "/api/v1/session")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Body.Close()
	var session struct {
		MutationToken string `json:"mutation_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&session); err != nil || session.MutationToken == "" {
		t.Fatalf("session: %+v %v", session, err)
	}
	return s, client, session.MutationToken
}

func request(t *testing.T, client *http.Client, method, url, body, origin, token string) (int, map[string]any, http.Header) {
	t.Helper()
	r, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	if token != "" {
		r.Header.Set("X-HopVista-Token", token)
	}
	resp, err := client.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var data map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("decode %s %s: %v", method, url, err)
	}
	return resp.StatusCode, data, resp.Header
}

func errorCode(v map[string]any) string {
	e, _ := v["error"].(map[string]any)
	code, _ := e["code"].(string)
	return code
}

func TestMutationSecurityAndLocalHost(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	t.Cleanup(func() { once.Do(func() { close(release) }) })
	s, client, token := startServer(t, executorFunc(func(_ context.Context, _ run.Input, _ func([]result.Family)) ([]result.Family, []result.Finding, error) {
		close(entered)
		<-release
		return []result.Family{}, []result.Finding{}, nil
	}))
	body := `{"target_url":"https://example.com/"}`
	for _, tc := range []struct{ origin, token, code string }{
		{"", token, "invalid_origin"}, {"https://evil.example", token, "invalid_origin"}, {s.URL, "", "invalid_token"}, {s.URL, "wrong", "invalid_token"},
	} {
		status, data, _ := request(t, client, "POST", s.URL+"/api/v1/runs", body, tc.origin, tc.token)
		if status != 403 || errorCode(data) != tc.code {
			t.Fatalf("origin %q token %q: %d %+v", tc.origin, tc.token, status, data)
		}
	}
	status, data, headers := request(t, client, "POST", s.URL+"/api/v1/runs", body, s.URL, token)
	if status != 202 || headers.Get("Access-Control-Allow-Origin") != "" || headers.Get("Cache-Control") != "no-store" {
		t.Fatalf("start: %d %+v", status, data)
	}
	id, ok := data["run_id"].(string)
	if !ok || len(id) != 32 || data["status_url"] != "/api/v1/runs/"+id || data["cancel_url"] != "/api/v1/runs/"+id+"/cancel" {
		t.Fatalf("start links: %+v", data)
	}
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("executor did not start")
	}
	status, data, _ = request(t, client, "POST", s.URL+"/api/v1/runs", body, s.URL, token)
	if status != 409 || errorCode(data) != "busy" {
		t.Fatalf("busy: %d %+v", status, data)
	}
	for _, path := range []string{"/api/v1/session", "/api/v1/runs/" + id, "/"} {
		r, err := http.NewRequest("GET", s.URL+path, nil)
		if err != nil {
			t.Fatal(err)
		}
		r.Host = "rebound.example"
		resp, err := client.Do(r)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != 403 {
			t.Fatalf("rebound host %s: %d", path, resp.StatusCode)
		}
	}
	status, data, _ = request(t, client, "POST", s.URL+"/api/v1/runs/"+id+"/cancel", "", s.URL, token)
	if status != 200 || data["state"] != "stopping" || data["ended_at"] != nil || data["stop_requested_at"] == nil {
		t.Fatalf("stopping: %d %+v", status, data)
	}
	status, data, _ = request(t, client, "POST", s.URL+"/api/v1/runs/"+id+"/cancel", "", s.URL, token)
	if status != 200 || data["state"] != "stopping" {
		t.Fatal("repeat cancel changed state")
	}
	status, data, _ = request(t, client, "POST", s.URL+"/api/v1/runs", body, s.URL, token)
	if status != 409 || errorCode(data) != "busy" {
		t.Fatal("stopping released busy slot")
	}
	once.Do(func() { close(release) })
	until := time.After(2 * time.Second)
	for {
		status, data, _ = request(t, client, "GET", s.URL+"/api/v1/runs/"+id, "", "", "")
		if status == 200 && data["state"] == "cancelled" {
			break
		}
		select {
		case <-until:
			t.Fatalf("cancel never quiesced: %d %+v", status, data)
		case <-time.After(time.Millisecond):
		}
	}
	if data["ended_at"] == nil || data["findings"] == nil {
		t.Fatal("terminal snapshot incomplete")
	}
}

func TestValidationAndMethods(t *testing.T) {
	s, client, token := startServer(t, run.Placeholder{})
	for _, tc := range []struct{ body, code string }{
		{`{"target_url":"ftp://example.com"}`, "invalid_target_url"},
		{`{"target_url":"https://user:pass@example.com"}`, "invalid_target_url"},
		{`{"target_url":"http://[::1]/"}`, "ipv6_not_supported_in_v1"},
		{`{"target_url":"http://[::ffff:192.0.2.1]/"}`, "ipv6_not_supported_in_v1"},
		{`{"target_url":"https://example.com/","max_hops":0}`, "invalid_max_hops"},
		{`{"target_url":"https://example.com/","max_hops":-1}`, "invalid_max_hops"},
		{`{"target_url":"https://example.com/","max_hops":65}`, "invalid_max_hops"},
		{`{"target_url":"https://example.com/","max_hops":2.5}`, "invalid_max_hops"},
		{`{"target_url":"https://example.com/","max_hops":"30"}`, "invalid_max_hops"},
		{`{"target_url":"https://example.com/","max_hops":null}`, "invalid_max_hops"},
		{`{"target_url":"https://example.com/","extra":1}`, "invalid_request"},
		{`{"target_url":"https://example.com/"} {}`, "invalid_request"},
	} {
		status, data, _ := request(t, client, "POST", s.URL+"/api/v1/runs", tc.body, s.URL, token)
		if status != 400 || errorCode(data) != tc.code {
			t.Fatalf("%s: %d %+v", tc.body, status, data)
		}
	}
	status, data, _ := request(t, client, "GET", s.URL+"/api/v1/runs/00000000000000000000000000000000", "", "", "")
	if status != 404 || errorCode(data) != "not_found" {
		t.Fatalf("unknown id: %d %+v", status, data)
	}
	status, data, headers := request(t, client, "OPTIONS", s.URL+"/api/v1/runs", "", s.URL, token)
	if status != 405 || headers.Get("Access-Control-Allow-Origin") != "" || errorCode(data) != "method_not_allowed" {
		t.Fatalf("preflight: %d %+v", status, data)
	}
	status, data, _ = request(t, client, "POST", s.URL+"/api/v1/runs", `{"target_url":"http://127.0.0.1:8080/path#fragment","max_hops":1}`, s.URL, token)
	if status != 202 {
		t.Fatalf("explicit loopback target: %d %+v", status, data)
	}
	id := data["run_id"].(string)
	until := time.After(2 * time.Second)
	for {
		_, data, _ = request(t, client, "GET", s.URL+"/api/v1/runs/"+id, "", "", "")
		if data["state"] == "completed" {
			break
		}
		select {
		case <-until:
			t.Fatal("placeholder did not finish")
		case <-time.After(time.Millisecond):
		}
	}
	if data["target_url"] != "http://127.0.0.1:8080/path" || data["max_hops"] != float64(1) {
		t.Fatalf("input not snapped/fragment sent: %+v", data)
	}
	family := data["families"].([]any)[0].(map[string]any)
	if family["outcome"] != "unavailable" || family["error"].(map[string]any)["code"] != "measurement_not_connected" {
		t.Fatalf("placeholder invented measurements: %+v", family)
	}
	status, data, _ = request(t, client, "POST", s.URL+"/api/v1/runs", `{"target_url":"http://127.0.0.1/"}`, s.URL, token)
	if status != 202 {
		t.Fatalf("default hop start: %d %+v", status, data)
	}
	_, defaultRun, _ := request(t, client, "GET", s.URL+"/api/v1/runs/"+data["run_id"].(string), "", "", "")
	if defaultRun["max_hops"] != float64(30) {
		t.Fatalf("default hops: %+v", defaultRun)
	}
	defaultID := data["run_id"].(string)
	until = time.After(2 * time.Second)
	for defaultRun["state"] != "completed" {
		select {
		case <-until:
			t.Fatal("default run did not finish")
		case <-time.After(time.Millisecond):
		}
		_, defaultRun, _ = request(t, client, "GET", s.URL+"/api/v1/runs/"+defaultID, "", "", "")
	}
	status, data, _ = request(t, client, "POST", s.URL+"/api/v1/runs", `{"target_url":"http://127.0.0.1/","max_hops":64}`, s.URL, token)
	if status != 202 {
		t.Fatalf("64-hop start: %d %+v", status, data)
	}
	_, maxRun, _ := request(t, client, "GET", s.URL+"/api/v1/runs/"+data["run_id"].(string), "", "", "")
	if maxRun["max_hops"] != float64(64) {
		t.Fatalf("64-hop setting: %+v", maxRun)
	}
}

func TestStaticAssetsSessionAndBindingChecks(t *testing.T) {
	s, client, _ := startServer(t, run.Placeholder{})
	for _, path := range []string{"/", "/app.js", "/style.css"} {
		resp, err := client.Get(s.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil || resp.StatusCode != 200 || len(body) == 0 {
			t.Fatalf("asset %s: %d %v", path, resp.StatusCode, err)
		}
		if resp.Header.Get("Access-Control-Allow-Origin") != "" || resp.Header.Get("Cache-Control") != "no-store" {
			t.Fatal("unexpected CORS/cache header")
		}
		if path == "/" && (!bytes.Contains(body, []byte("id=\"root\"")) || !bytes.Contains(body, []byte("/app.js")) || !bytes.Contains(body, []byte("/style.css")) || resp.Header.Get("Content-Security-Policy") == "") {
			t.Fatal("production UI HTML missing assets or security contract")
		}
	}
	for _, address := range []string{"0.0.0.0:1234", "[::1]:1234", "127.0.0.1:0", "example.com:1234"} {
		if _, err := localapi.NewHandler(address, run.NewManager(run.Placeholder{})); err == nil {
			t.Fatalf("accepted listener %q", address)
		}
	}
}
