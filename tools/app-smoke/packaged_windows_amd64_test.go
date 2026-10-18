//go:build windows && amd64

package appsmoke

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"hopvista/internal/result"
)

// Build cmd/hopvista first, then set HOPVISTA_PACKAGED_EXE to its absolute
// path. The copied child executable launches without the repository or a Go /
// Node executable on PATH. This checks one local Windows environment only.
func TestPackagedExecutableLoopback(t *testing.T) {
	executable := copyPackagedExecutable(t)
	process := startPackagedProcess(t, executable)
	base := process.base
	client := &http.Client{Timeout: 5 * time.Second, Transport: &http.Transport{Proxy: nil}}
	get := func(path string) *http.Response {
		t.Helper()
		response, err := client.Get(base + path)
		if err != nil {
			t.Fatal(err)
		}
		return response
	}
	for _, asset := range []struct{ path, contains string }{{"/", "id=\"root\""}, {"/app.js", "/api/v1/runs"}, {"/style.css", "body"}} {
		response := get(asset.path)
		body, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil || response.StatusCode != 200 || !bytes.Contains(body, []byte(asset.contains)) || response.Header.Get("Cache-Control") != "no-store" {
			t.Fatalf("asset %s: status=%d error=%v", asset.path, response.StatusCode, err)
		}
	}
	response := get("/api/v1/session")
	var session struct {
		Token string `json:"mutation_token"`
	}
	if err := json.NewDecoder(response.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != 200 || len(session.Token) != 64 {
		t.Fatal("local session token unavailable")
	}
	badHost, err := http.NewRequest(http.MethodGet, base+"/api/v1/session", nil)
	if err != nil {
		t.Fatal(err)
	}
	badHost.Host = "example.invalid"
	response, err = client.Do(badHost)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusForbidden || response.Header.Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("packaged Host boundary: %d", response.StatusCode)
	}

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.Host, "localhost:") {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer target.Close()
	missingOrigin, err := http.NewRequest(http.MethodPost, base+"/api/v1/runs", bytes.NewBufferString(`{"target_url":"`+target.URL+`/"}`))
	if err != nil {
		t.Fatal(err)
	}
	missingOrigin.Header.Set("Content-Type", "application/json")
	missingOrigin.Header.Set("X-HopVista-Token", session.Token)
	response, err = client.Do(missingOrigin)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusForbidden || response.Header.Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("packaged Origin boundary: %d", response.StatusCode)
	}
	for _, rejected := range []struct {
		name, body, token, code string
		status                  int
	}{
		{"wrong token", `{"target_url":"` + target.URL + `/"}`, "wrong", "invalid_token", http.StatusForbidden},
		{"IPv6 literal", `{"target_url":"http://[::1]/"}`, session.Token, "ipv6_not_supported_in_v1", http.StatusBadRequest},
	} {
		req, err := http.NewRequest(http.MethodPost, base+"/api/v1/runs", bytes.NewBufferString(rejected.body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Origin", base)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-HopVista-Token", rejected.token)
		response, err = client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		var rejectedBody struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if err := json.NewDecoder(response.Body).Decode(&rejectedBody); err != nil {
			response.Body.Close()
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != rejected.status || response.Header.Get("Access-Control-Allow-Origin") != "" || rejectedBody.Error.Code != rejected.code {
			t.Fatalf("%s rejection: status=%d body=%+v", rejected.name, response.StatusCode, rejectedBody)
		}
	}
	request, err := http.NewRequest(http.MethodPost, base+"/api/v1/runs", bytes.NewBufferString(`{"target_url":"`+target.URL+`/","max_hops":4}`))
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
		t.Fatalf("run start: %d %+v", response.StatusCode, started)
	}
	until := time.Now().Add(12 * time.Second)
	for time.Now().Before(until) {
		response = get(started.StatusURL)
		var snapshot result.Run
		if err := json.NewDecoder(response.Body).Decode(&snapshot); err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if !snapshot.State.Terminal() {
			time.Sleep(25 * time.Millisecond)
			continue
		}
		if snapshot.State != result.RunCompleted || snapshot.SchemaVersion != 1 || len(snapshot.Families) != 2 || len(snapshot.Findings) != 1 || len(snapshot.Families[0].Exchanges) != 1 || len(snapshot.Families[0].Paths) != 1 {
			t.Fatalf("terminal result: %+v", snapshot)
		}
		x, path := snapshot.Families[0].Exchanges[0], snapshot.Families[0].Paths[0]
		if x.ResponseStatus == nil || *x.ResponseStatus != 503 || x.HTTP.Outcome != result.Completed || x.ActualPeerIP != "127.0.0.1" || path.Route.Outcome != result.Completed || path.Target.Replied != 5 || path.Trace.Outcome != result.DestinationReplied || snapshot.Families[1].Outcome != result.Unavailable {
			t.Fatalf("packaged measurements: exchange=%+v path=%+v ipv6=%+v", x, path, snapshot.Families[1])
		}
		compareTracert(t, path)
		comparePinnedCurl(t, client, base, session.Token, started.RunID, target.URL)
		if drainIP := os.Getenv("HOPVISTA_PACKAGED_DRAIN_IP"); drainIP != "" {
			comparePackagedStopping(t, client, base, session.Token, drainIP)
		}
		if httpsURL := os.Getenv("HOPVISTA_PACKAGED_HTTPS_URL"); httpsURL != "" {
			comparePinnedHTTPS(t, client, base, session.Token, httpsURL)
		}
		return
	}
	t.Fatal("packaged diagnostic run did not finish")
}

func comparePinnedHTTPS(t *testing.T, client *http.Client, base, token, rawURL string) {
	t.Helper()
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil {
		t.Fatalf("HOPVISTA_PACKAGED_HTTPS_URL must be an absolute HTTPS URL without credentials: %q", rawURL)
	}
	port := u.Port()
	if port == "" {
		port = "443"
	}
	body, err := json.Marshal(struct {
		TargetURL string `json:"target_url"`
		MaxHops   int    `json:"max_hops"`
	}{rawURL, 4})
	if err != nil {
		t.Fatal(err)
	}
	request, err := http.NewRequest(http.MethodPost, base+"/api/v1/runs", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Origin", base)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-HopVista-Token", token)
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	var started struct {
		StatusURL string `json:"status_url"`
	}
	if err := json.NewDecoder(response.Body).Decode(&started); err != nil {
		response.Body.Close()
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusAccepted || started.StatusURL == "" {
		t.Fatalf("packaged HTTPS start: %d %+v", response.StatusCode, started)
	}
	var snapshot result.Run
	until := time.Now().Add(25 * time.Second)
	for time.Now().Before(until) {
		response, err = client.Get(base + started.StatusURL)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.NewDecoder(response.Body).Decode(&snapshot); err != nil {
			response.Body.Close()
			t.Fatal(err)
		}
		response.Body.Close()
		if snapshot.State.Terminal() {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	if snapshot.State != result.RunCompleted || len(snapshot.Families) == 0 || len(snapshot.Families[0].Exchanges) != 1 {
		t.Fatalf("packaged HTTPS run: %+v", snapshot)
	}
	x := snapshot.Families[0].Exchanges[0]
	if x.Hostname != u.Hostname() || x.SelectedIP == "" || x.ActualPeerIP != x.SelectedIP || x.TCP.Outcome != result.Completed || x.TLS.Outcome != result.Completed || x.HTTP.Outcome != result.Completed || x.ResponseStatus == nil || x.TCP.DurationMS == nil || x.TLS.DurationMS == nil || x.HTTP.DurationMS == nil || x.ConnectionReused == nil || *x.ConnectionReused {
		t.Fatalf("packaged pinned HTTPS exchange: %+v", x)
	}
	resolve := u.Hostname() + ":" + port + ":" + x.SelectedIP
	curlPath, err := exec.LookPath("curl.exe")
	if err != nil {
		t.Fatal("independent curl.exe is unavailable: ", err)
	}
	curl := exec.Command(curlPath, "--silent", "--show-error", "--max-time", "8", "--ipv4", "--noproxy", "*", "--resolve", resolve, "--output", "NUL", "--write-out", "%{http_code}|%{remote_ip}|%{ssl_verify_result}|%{time_starttransfer}", rawURL)
	output, err := curl.CombinedOutput()
	if err != nil {
		t.Fatalf("pinned HTTPS curl failed: %v %s", err, output)
	}
	parts := strings.Split(strings.TrimSpace(string(output)), "|")
	if len(parts) != 4 || parts[0] != strconv.Itoa(*x.ResponseStatus) || parts[1] != x.SelectedIP || parts[2] != "0" {
		t.Fatalf("pinned HTTPS curl disagreed with packaged exchange: curl=%s app=%+v", output, x)
	}
	curlFirstByteSeconds, err := strconv.ParseFloat(parts[3], 64)
	if err != nil {
		t.Fatalf("invalid curl first-byte timing: %s", output)
	}
	appPhasesMS := *x.TCP.DurationMS + *x.TLS.DurationMS + *x.HTTP.DurationMS
	if curlFirstByteSeconds <= 0 || curlFirstByteSeconds > 8 || appPhasesMS <= 0 || math.Abs(appPhasesMS-curlFirstByteSeconds*1000) > 2000 {
		t.Fatalf("pinned HTTPS first-byte timing outside broad comparison range: curl=%s app_ms=%.1f", output, appPhasesMS)
	}
}

func comparePackagedStopping(t *testing.T, client *http.Client, base, token, rawIP string) {
	t.Helper()
	ip, err := netip.ParseAddr(rawIP)
	if err != nil || !ip.Is4() || ip.IsUnspecified() || ip.IsMulticast() {
		t.Fatalf("invalid packaged drain IPv4 address %q", rawIP)
	}
	post := func(path, body string) *http.Response {
		t.Helper()
		request, err := http.NewRequest(http.MethodPost, base+path, strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Origin", base)
		request.Header.Set("X-HopVista-Token", token)
		if body != "" {
			request.Header.Set("Content-Type", "application/json")
		}
		response, err := client.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		return response
	}
	body := `{"target_url":"http://` + rawIP + `:1/","max_hops":1}`
	response := post("/api/v1/runs", body)
	var started struct {
		RunID     string `json:"run_id"`
		StatusURL string `json:"status_url"`
		CancelURL string `json:"cancel_url"`
	}
	if err := json.NewDecoder(response.Body).Decode(&started); err != nil {
		response.Body.Close()
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusAccepted || started.RunID == "" {
		t.Fatalf("packaged drain run start: %d %+v", response.StatusCode, started)
	}
	var snapshot result.Run
	until := time.Now().Add(20 * time.Second)
	for time.Now().Before(until) {
		response, err = client.Get(base + started.StatusURL)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.NewDecoder(response.Body).Decode(&snapshot); err != nil {
			response.Body.Close()
			t.Fatal(err)
		}
		response.Body.Close()
		if snapshot.State.Terminal() {
			t.Fatalf("candidate finished before a pending target echo could be cancelled: %+v", snapshot)
		}
		if len(snapshot.Families) > 0 && len(snapshot.Families[0].Paths) == 1 {
			target := snapshot.Families[0].Paths[0].Target
			if target.Outcome == result.Pending && len(target.Samples) == 1 && target.Samples[0].Outcome == result.Timeout {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	if time.Now().After(until) {
		t.Fatal("packaged run did not publish a pending no-reply target series")
	}
	// The next target echo is scheduled immediately after the first timeout.
	time.Sleep(50 * time.Millisecond)
	response = post(started.CancelURL, "")
	var stopping result.Run
	if err := json.NewDecoder(response.Body).Decode(&stopping); err != nil {
		response.Body.Close()
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK || stopping.State != result.Stopping || stopping.StopRequestedAt == nil || stopping.EndedAt != nil || len(stopping.Findings) != 0 {
		t.Fatalf("packaged run did not remain stopping: %+v", stopping)
	}
	response = post("/api/v1/runs", body)
	var busy struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(response.Body).Decode(&busy); err != nil {
		response.Body.Close()
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusConflict || busy.Error.Code != "busy" {
		t.Fatalf("stopping run released busy slot: %d %+v", response.StatusCode, busy)
	}
	until = time.Now().Add(5 * time.Second)
	for time.Now().Before(until) {
		response, err = client.Get(base + started.StatusURL)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.NewDecoder(response.Body).Decode(&snapshot); err != nil {
			response.Body.Close()
			t.Fatal(err)
		}
		response.Body.Close()
		if snapshot.State.Terminal() {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if snapshot.State != result.RunCancelled || snapshot.EndedAt == nil || snapshot.StopRequestedAt == nil || snapshot.EndedAt.Before(*snapshot.StopRequestedAt) || len(snapshot.Families) == 0 || len(snapshot.Families[0].Paths) != 1 || len(snapshot.Findings) != 0 {
		t.Fatalf("packaged cancellation did not retain drained evidence: %+v", snapshot)
	}
	target := snapshot.Families[0].Paths[0].Target
	if len(target.Samples) != 2 || target.Attempted != 2 || target.Samples[0].Outcome != result.Timeout || target.Samples[1].Outcome != result.Timeout || target.Samples[1].StartedAt == nil || target.Samples[1].EndedAt == nil || target.Samples[1].StartedAt.After(*snapshot.StopRequestedAt) || target.Samples[1].EndedAt.Before(*snapshot.StopRequestedAt) {
		t.Fatalf("packaged cancellation lacked an in-flight echo or started another: %+v", target)
	}
}

func compareTracert(t *testing.T, path result.Path) {
	t.Helper()
	if path.SelectedIP != "127.0.0.1" || path.Trace.SelectedIP != path.SelectedIP || len(path.Trace.Hops) != 1 || path.Trace.Hops[0].TTL != 1 || path.Trace.Hops[0].ReplyAddress != path.SelectedIP {
		t.Fatalf("HopVista loopback trace is not a one-hop destination reply: %+v", path.Trace)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tracert := filepath.Join(os.Getenv("SystemRoot"), "System32", "tracert.exe")
	output, err := exec.CommandContext(ctx, tracert, "-d", "-4", "-h", "4", "-w", "1000", path.SelectedIP).CombinedOutput()
	if err != nil {
		t.Fatalf("Windows tracert failed: %v %s", err, output)
	}
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "1" && fields[len(fields)-1] == path.SelectedIP {
			return
		}
	}
	t.Fatalf("Windows tracert did not show the same one-hop loopback destination: %s", output)
}

func comparePinnedCurl(t *testing.T, client *http.Client, base, token, oldID, literalURL string) {
	t.Helper()
	hostURL := strings.Replace(literalURL, "127.0.0.1", "localhost", 1) + "/"
	request, err := http.NewRequest(http.MethodPost, base+"/api/v1/runs", bytes.NewBufferString(`{"target_url":"`+hostURL+`","max_hops":4}`))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Origin", base)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-HopVista-Token", token)
	response, err := client.Do(request)
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
	if response.StatusCode != http.StatusAccepted || started.RunID == oldID {
		t.Fatalf("restart: %d %+v", response.StatusCode, started)
	}
	response, err = client.Get(base + "/api/v1/runs/" + oldID)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("restart retained old run: %d", response.StatusCode)
	}
	until := time.Now().Add(12 * time.Second)
	var completed result.Run
	for time.Now().Before(until) {
		response, err = client.Get(base + started.StatusURL)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.NewDecoder(response.Body).Decode(&completed); err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if completed.State.Terminal() {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	if completed.State != result.RunCompleted || len(completed.Families) == 0 || len(completed.Families[0].Exchanges) != 1 {
		t.Fatalf("pinned hostname run: %+v", completed)
	}
	x := completed.Families[0].Exchanges[0]
	if x.ResponseStatus == nil || *x.ResponseStatus != 200 || x.Hostname != "localhost" || x.SelectedIP != "127.0.0.1" || x.ActualPeerIP != "127.0.0.1" || x.HTTP.Outcome != result.Completed {
		t.Fatalf("pinned hostname exchange: %+v", x)
	}
	if len(completed.Families[0].DNS) != 1 || completed.Families[0].DNS[0].SelectedIP != "127.0.0.1" {
		t.Fatalf("hostname selection: %+v", completed.Families[0].DNS)
	}
	if len(completed.Families[0].Paths) != 1 || completed.Families[0].Paths[0].SelectedIP != "127.0.0.1" {
		t.Fatalf("pinned path: %+v", completed.Families[0].Paths)
	}
	_, port, err := net.SplitHostPort(strings.TrimPrefix(literalURL, "http://"))
	if err != nil {
		t.Fatal(err)
	}
	curlPath, err := exec.LookPath("curl.exe")
	if err != nil {
		t.Fatal("independent curl.exe is unavailable: ", err)
	}
	curl := exec.Command(curlPath, "--silent", "--show-error", "--max-time", "5", "--ipv4", "--noproxy", "*", "--resolve", "localhost:"+port+":127.0.0.1", "--output", "NUL", "--write-out", "%{http_code}|%{remote_ip}", hostURL)
	output, err := curl.CombinedOutput()
	if err != nil {
		t.Fatalf("pinned curl failed: %v %s", err, output)
	}
	if strings.TrimSpace(string(output)) != "200|127.0.0.1" {
		t.Fatalf("pinned curl disagreed with executable: %s", output)
	}
}

func systemOnlyPATH(original []string) []string {
	filtered := make([]string, 0, len(original)+1)
	for _, entry := range original {
		if !strings.EqualFold(strings.SplitN(entry, "=", 2)[0], "PATH") {
			filtered = append(filtered, entry)
		}
	}
	root := os.Getenv("SystemRoot")
	if root == "" {
		root = `C:\Windows`
	}
	return append(filtered, "PATH="+filepath.Join(root, "System32"))
}
