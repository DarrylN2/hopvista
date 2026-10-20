//go:build windows && amd64

package appsmoke

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"testing"
	"time"

	"hopvista/internal/result"
)

// No alternate executable or shortened manager window: all route/echo calls
// below are the production Windows adapter. The candidate must remain silent
// far enough into a 64-hop trace to exhaust the real measurement budget.
func TestPackagedMinuteCutoff(t *testing.T) {
	runPackagedMinuteCutoff(t, false)
}

// The process pause is a harness-only scheduling fixture: real native work is
// in flight, but its long wall-clock duration is externally induced. It does
// not demonstrate a naturally stalled GetBestRoute2 or echo call.
func TestPackagedSuspendedMinuteCutoff(t *testing.T) {
	runPackagedMinuteCutoff(t, true)
}

func runPackagedMinuteCutoff(t *testing.T, pauseNative bool) {
	rawIP := os.Getenv("HOPVISTA_PACKAGED_CUTOFF_IP")
	if rawIP == "" {
		t.Skip("set HOPVISTA_PACKAGED_CUTOFF_IP to a no-reply IPv4 candidate")
	}
	ip, err := netip.ParseAddr(rawIP)
	if err != nil || !ip.Is4() || ip.IsUnspecified() || ip.IsMulticast() {
		t.Fatalf("invalid cutoff candidate %q", rawIP)
	}
	p := startPackagedProcess(t, copyPackagedExecutable(t))
	if pauseNative {
		ownDisposablePackagedProcess(t, p)
	}
	client := &http.Client{Timeout: 5 * time.Second, Transport: &http.Transport{Proxy: nil}}
	defer client.CloseIdleConnections()
	request := func(method, path, token string, body []byte) (int, []byte) {
		t.Helper()
		req, err := http.NewRequest(method, p.base+path, bytes.NewReader(body))
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
		var data json.RawMessage
		if err := json.NewDecoder(response.Body).Decode(&data); err != nil {
			t.Fatal(err)
		}
		return response.StatusCode, data
	}
	code, data := request(http.MethodGet, "/api/v1/session", "", nil)
	var session struct {
		Token string `json:"mutation_token"`
	}
	if err := json.Unmarshal(data, &session); err != nil || code != 200 || session.Token == "" {
		t.Fatal("session unavailable")
	}
	body, err := json.Marshal(map[string]any{"target_url": "http://" + rawIP + ":1/", "max_hops": 64})
	if err != nil {
		t.Fatal(err)
	}
	code, data = request(http.MethodPost, "/api/v1/runs", session.Token, body)
	var started struct {
		StatusURL string `json:"status_url"`
	}
	if err := json.Unmarshal(data, &started); err != nil || code != http.StatusAccepted || started.StatusURL == "" {
		t.Fatalf("start: %d %s", code, data)
	}
	var snapshot result.Run
	var previous result.Run
	paused := false
	var pausedEchoID string
	until := time.Now().Add(75 * time.Second) // harness observation limit, not a native drain promise
	for time.Now().Before(until) {
		code, data = request(http.MethodGet, started.StatusURL, "", nil)
		if err := json.Unmarshal(data, &snapshot); err != nil || code != 200 {
			t.Fatalf("snapshot: %d %s", code, data)
		}
		if snapshot.MaxHops != 64 || snapshot.DeadlineAt.Sub(snapshot.StartedAt) != 60*time.Second {
			t.Fatalf("changed settings/window: %+v", snapshot)
		}
		if !snapshot.State.Terminal() && len(snapshot.Findings) != 0 {
			t.Fatal("final findings published before quiescence")
		}
		if previous.State == result.Stopping && snapshot.State == result.Running {
			t.Fatal("stopping regressed to running")
		}
		if snapshot.State == result.Stopping {
			if snapshot.StopRequestedAt == nil || snapshot.EndedAt != nil {
				t.Fatal("invalid stopping timestamps")
			}
			writeCutoffCapture(t, "minutestopping", snapshot)
		}
		if pauseNative && !paused && len(snapshot.Families) > 0 && len(snapshot.Families[0].Paths) == 1 {
			path := snapshot.Families[0].Paths[0]
			if path.Target.Outcome == result.Pending && len(path.Target.Samples) == 1 && path.Target.Samples[0].Outcome == result.Timeout {
				// The second target echo follows the first timeout immediately.
				// The known no-reply fixture keeps it native-pending beyond 50 ms.
				time.Sleep(50 * time.Millisecond)
				pausedEchoID = path.Target.ID + "-2"
				pausePackagedUntil(t, p, snapshot.DeadlineAt.Add(500*time.Millisecond))
				paused = true
			}
		}
		if snapshot.State.Terminal() {
			break
		}
		previous = snapshot
		time.Sleep(20 * time.Millisecond)
	}
	// Save even a failed candidate's terminal/last snapshot for diagnosis.
	capture := "minute"
	if pauseNative {
		capture = "pausedminute"
	}
	writeCutoffCapture(t, capture, snapshot)
	if pauseNative && !paused {
		t.Fatal("native no-reply fixture never reached the pause trigger")
	}
	if !snapshot.State.Terminal() || snapshot.State != result.RunCompleted || snapshot.EndedAt == nil || len(snapshot.Families) == 0 || len(snapshot.Families[0].Paths) != 1 {
		t.Fatalf("no quiescent result: %+v", snapshot)
	}
	path := snapshot.Families[0].Paths[0]
	if pauseNative {
		found := false
		for _, e := range path.Target.Samples {
			if e.ID == pausedEchoID && e.StartedAt != nil && e.EndedAt != nil && e.StartedAt.Before(snapshot.DeadlineAt) && e.EndedAt.After(snapshot.DeadlineAt) {
				found = true
			}
		}
		if !found {
			t.Fatalf("process pause did not span a real native echo: id=%s samples=%+v", pausedEchoID, path.Target.Samples)
		}
	}
	if path.Trace.Outcome != result.TraceDeadline || snapshot.Families[0].Outcome != result.Deadline {
		t.Fatalf("candidate did not exhaust the window: trace=%s family=%s elapsed=%s", path.Trace.Outcome, snapshot.Families[0].Outcome, snapshot.EndedAt.Sub(snapshot.StartedAt))
	}
	if snapshot.StopRequestedAt == nil || snapshot.StopRequestedAt.Before(snapshot.DeadlineAt) || snapshot.EndedAt.Before(snapshot.DeadlineAt) {
		t.Fatal("candidate exhausted admission budget before the wall-clock cutoff; this does not exercise real-minute stopping")
	}
	if pauseNative {
		if len(path.Target.Samples) != 2 || path.Target.Outcome != result.Deadline || len(path.Trace.Hops) != 0 {
			t.Fatal("paused target echo did not retain partial evidence or stop scheduling")
		}
	} else if len(path.Trace.Hops) < 2 || len(path.Target.Samples) == 0 {
		t.Fatal("completed trace evidence missing")
	}
	stop := snapshot.DeadlineAt
	if snapshot.StopRequestedAt != nil && snapshot.StopRequestedAt.Before(stop) {
		stop = *snapshot.StopRequestedAt
	}
	check := func(m result.Measurement) {
		t.Helper()
		if m.StartedAt != nil && (m.StartedAt.After(stop) || m.EndedAt == nil || m.EndedAt.After(*snapshot.EndedAt)) {
			t.Fatalf("native start/cleanup outside ownership: %+v", m)
		}
	}
	check(path.Route.Measurement)
	for _, samples := range [][]result.Echo{path.Gateway.Samples, path.Target.Samples, path.Trace.Hops} {
		for _, e := range samples {
			check(e.Measurement)
			if e.StartedAt == nil && (e.RTTMS != nil || e.ReplyAddress != "" || e.EndedAt != nil) {
				t.Fatalf("unstarted probe invented evidence: %+v", e)
			}
			if e.StartedAt != nil {
				t.Logf("native echo id=%s cap_ms=%v start=%s returned-and-closed=%s outcome=%s", e.ID, *e.BoundMS, e.StartedAt.Format(time.RFC3339Nano), e.EndedAt.Format(time.RFC3339Nano), e.Outcome)
			}
		}
	}
	if len(path.Trace.Hops) > 0 {
		last := path.Trace.Hops[len(path.Trace.Hops)-1]
		if last.StartedAt == nil && (last.Outcome != result.Deadline || last.Status != result.NotSent || last.RTTMS != nil) {
			t.Fatalf("invalid budget refusal: %+v", last)
		}
	}
	t.Logf("cutoff=%s stop-observed=%v quiescence=%s elapsed=%s", snapshot.DeadlineAt.Format(time.RFC3339Nano), snapshot.StopRequestedAt, snapshot.EndedAt.Format(time.RFC3339Nano), snapshot.EndedAt.Sub(snapshot.StartedAt))
	for i := 0; i < 3; i++ {
		_, stable := request(http.MethodGet, started.StatusURL, "", nil)
		if !bytes.Equal(data, stable) {
			t.Fatal("terminal snapshot changed after quiescence")
		}
	}
	// Starting another run is allowed only after the observed terminal result.
	code, data = request(http.MethodPost, "/api/v1/runs", session.Token, []byte(`{"target_url":"http://127.0.0.1:1/","max_hops":1}`))
	if code != http.StatusAccepted {
		t.Fatalf("quiescent slot not released: %d %s", code, data)
	}
}

func writeCutoffCapture(t *testing.T, name string, snapshot result.Run) {
	t.Helper()
	if dir := os.Getenv("HOPVISTA_CUTOFF_CAPTURE"); dir != "" {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
		data, err := json.MarshalIndent(snapshot, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "run-"+name+".json"), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
