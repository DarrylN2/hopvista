//go:build windows && amd64

package appsmoke

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"hopvista/internal/result"
)

// Exercise the real packaged executable/API, real loopback HTTP and native IPv4
// probes. Each restart launches the same isolated binary from the same directory.
func TestPackagedRestartPrivacy(t *testing.T) {
	executable := copyPackagedExecutable(t)
	image, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(image)
	assertDirectory := func() {
		t.Helper()
		files, err := os.ReadDir(filepath.Dir(executable))
		if err != nil || len(files) != 1 || files[0].Name() != "hopvista.exe" {
			t.Fatal("packaged executable created unexpected working-directory files")
		}
		after, err := os.ReadFile(executable)
		if err != nil || sha256.Sum256(after) != digest {
			t.Fatal("packaged executable changed on disk")
		}
	}
	assertDirectory()
	const canary = "hopvista-privacy-canary-20261018"
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, canary) {
			t.Error("fixture did not receive the synthetic target query")
		}
		switch r.URL.Path {
		case "/start/" + canary:
			w.Header().Set("Location", "/final/"+canary+"?record="+canary)
			w.WriteHeader(http.StatusFound)
		case "/abort/" + canary:
			conn, _, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Error(err)
				return
			}
			conn.Close()
		default:
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}))
	defer target.Close()
	client := &http.Client{Timeout: 5 * time.Second, Transport: &http.Transport{Proxy: nil}}
	defer client.CloseIdleConnections()
	request := func(p *packagedProcess, method, path, body, token string) (int, []byte) {
		t.Helper()
		req, err := http.NewRequest(method, p.base+path, strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		if method == http.MethodPost {
			req.Header.Set("Origin", p.base)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-HopVista-Token", token)
		}
		response, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		data, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		if response.Header.Get("Cache-Control") != "no-store" {
			t.Fatal("API response permits caching")
		}
		return response.StatusCode, data
	}
	session := func(p *packagedProcess) string {
		t.Helper()
		code, data := request(p, http.MethodGet, "/api/v1/session", "", "")
		var s struct {
			Token string `json:"mutation_token"`
		}
		if err := json.Unmarshal(data, &s); err != nil || code != http.StatusOK || len(s.Token) != 64 {
			t.Fatal("session unavailable")
		}
		return s.Token
	}
	assertError := func(p *packagedProcess, method, path, body, token string, wantStatus int, wantCode string) {
		t.Helper()
		code, data := request(p, method, path, body, token)
		var e struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if err := json.Unmarshal(data, &e); err != nil || code != wantStatus || e.Error.Code != wantCode || bytes.Contains(data, []byte(canary)) {
			t.Fatalf("unexpected API rejection: status=%d wanted=%d code=%s", code, wantStatus, e.Error.Code)
		}
	}
	measure := func(p *packagedProcess, token, path string, wantOutcome result.Outcome) (string, string) {
		t.Helper()
		url := target.URL + path + canary + "?secret=" + canary
		body, _ := json.Marshal(map[string]any{"target_url": url, "max_hops": 1})
		code, data := request(p, http.MethodPost, "/api/v1/runs", string(body), token)
		var start struct {
			RunID     string `json:"run_id"`
			StatusURL string `json:"status_url"`
		}
		if err := json.Unmarshal(data, &start); err != nil || code != http.StatusAccepted || start.RunID == "" || start.StatusURL != "/api/v1/runs/"+start.RunID {
			t.Fatal("diagnostic run did not start")
		}
		until := time.Now().Add(15 * time.Second)
		for time.Now().Before(until) {
			code, data = request(p, http.MethodGet, start.StatusURL, "", "")
			var snapshot result.Run
			if err := json.Unmarshal(data, &snapshot); err != nil || code != http.StatusOK {
				t.Fatal("diagnostic run could not be read")
			}
			if snapshot.State.Terminal() {
				if snapshot.State != result.RunCompleted || snapshot.TargetURL != url || len(snapshot.Families) != 2 || snapshot.Families[0].Outcome != wantOutcome || len(snapshot.Findings) != 1 || !bytes.Contains(data, []byte(canary)) {
					t.Fatal("controlled run result/evidence missing")
				}
				if wantOutcome == result.Completed {
					x := snapshot.Families[0].Exchanges
					if len(x) != 2 || x[0].ResponseStatus == nil || *x[0].ResponseStatus != 302 || x[1].ResponseStatus == nil || *x[1].ResponseStatus != 503 {
						t.Fatal("completed redirect response evidence missing")
					}
				} else if snapshot.Families[0].Error == nil || snapshot.Families[0].Error.Code != "http_exchange_failed" {
					t.Fatal("controlled incomplete HTTP response missing")
				}
				return start.StatusURL, start.RunID
			}
			time.Sleep(25 * time.Millisecond)
		}
		t.Fatal("controlled diagnostic run did not finish")
		return "", ""
	}
	assertOutput := func(p *packagedProcess, secrets ...string) {
		t.Helper()
		p.stop(t)
		stdout, stderr := p.stdout.String(), p.stderr.buffer.String()
		for _, secret := range append(secrets, canary) {
			if strings.Contains(stdout, secret) || strings.Contains(stderr, secret) {
				t.Fatal("private diagnostic data appeared in default process output")
			}
		}
		lines := strings.Split(strings.TrimSpace(stderr), "\n")
		suffix := " HopVista: " + p.base + "/"
		if stdout != "" || len(lines) != 1 || !strings.HasSuffix(strings.TrimSuffix(lines[0], "\r"), suffix) {
			t.Fatalf("unexpected default output: stdout_bytes=%d stderr_lines=%d", len(stdout), len(lines))
		}
		prefix := strings.TrimSuffix(strings.TrimSuffix(lines[0], "\r"), suffix)
		if _, err := time.Parse("2006/01/02 15:04:05", prefix); err != nil {
			t.Fatal("startup line contains unexpected data")
		}
		assertDirectory()
	}

	first := startPackagedProcess(t, executable)
	firstToken := session(first)
	firstStatus, firstID := measure(first, firstToken, "/start/", result.Completed)
	assertError(first, http.MethodGet, "/"+canary, "", "", http.StatusNotFound, "not_found")
	assertError(first, http.MethodPost, "/api/v1/runs", `{"target_url":"http://user:`+canary+`@127.0.0.1/"}`, firstToken, http.StatusBadRequest, "invalid_target_url")
	assertError(first, http.MethodPost, "/api/v1/runs", `{"target_url":"`+canary, firstToken, http.StatusBadRequest, "invalid_request")
	assertOutput(first, firstToken, firstID)
	t.Log("launch 1: completed redirect/503 retained; rejected inputs checked; stdout empty; stderr contains only listener announcement; working directory unchanged")

	second := startPackagedProcess(t, executable)
	secondToken := session(second)
	if secondToken == firstToken {
		t.Fatal("restart reused the mutation token")
	}
	assertError(second, http.MethodGet, firstStatus, "", "", http.StatusNotFound, "not_found")
	assertError(second, http.MethodPost, firstStatus+"/cancel", "", secondToken, http.StatusNotFound, "not_found")
	assertError(second, http.MethodPost, "/api/v1/runs", `{"target_url":"`+target.URL+`/"}`, firstToken, http.StatusForbidden, "invalid_token")
	secondStatus, secondID := measure(second, secondToken, "/abort/", result.Error)
	assertOutput(second, firstToken, secondToken, firstID, secondID)
	t.Log("launch 2: prior result absent before a new run; old token rejected; fresh run produced retained incomplete HTTP evidence; output and directory checks passed")

	third := startPackagedProcess(t, executable)
	thirdToken := session(third)
	if thirdToken == firstToken || thirdToken == secondToken {
		t.Fatal("second restart reused a mutation token")
	}
	for _, status := range []string{firstStatus, secondStatus} {
		assertError(third, http.MethodGet, status, "", "", http.StatusNotFound, "not_found")
	}
	assertOutput(third, firstToken, secondToken, thirdToken, firstID, secondID)
	t.Log("launch 3: both prior run IDs absent; token changed again; output and directory checks passed")
}
