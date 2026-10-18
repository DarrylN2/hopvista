package httpmeasure_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"hopvista/internal/engine"
	"hopvista/internal/httpmeasure"
	"hopvista/internal/localapi"
	"hopvista/internal/probe/probefake"
	"hopvista/internal/result"
	"hopvista/internal/run"
)

// Skip only path sampling waits; DNS/TCP/TLS/HTTP use their actual clocks.
type journeyClock struct{}

func (c *journeyClock) Now() time.Time { return time.Now() }
func (c *journeyClock) Wait(ctx context.Context, d time.Duration) error {
	return ctx.Err()
}

func runThroughAPI(t *testing.T, f *tlsJourney, url, pathIP string) result.Run {
	t.Helper()
	clock := &journeyClock{}
	adapter := journeyAdapter(clock, pathIP)
	e := &engine.Engine{HTTP: httpmeasure.ControlledTLSClient(f.roots, f.dial), Resolver: lookupFunc(f.resolve), Probe: adapter, Clock: clock}
	snapshot := runEngineThroughAPI(t, run.NewManager(e), url, nil)
	if snapshot.State != result.RunCompleted {
		t.Fatalf("lifecycle: %+v", snapshot)
	}
	for range f.dials {
		select {
		case <-f.complete:
		case <-time.After(6 * time.Second):
			t.Fatal("TLS server did not finish")
		}
	}
	sort.Slice(f.records, func(i, j int) bool { return f.records[i].index < f.records[j].index })
	for _, c := range adapter.Calls() {
		if c.Route != nil && c.Route.SelectedIP.String() != pathIP {
			t.Fatalf("wrong route target: %+v", c)
		}
		if c.Echo != nil && c.Echo.SelectedIP.String() != pathIP {
			t.Fatalf("wrong echo target: %+v", c)
		}
	}
	return snapshot
}

func journeyAdapter(clock *journeyClock, pathIP string) *probefake.Adapter {
	onLink, rtt := true, 1.0
	echoes := make([]probefake.EchoStep, 6)
	for i := range echoes {
		echoes[i].Observation = result.Echo{Measurement: result.Measurement{Outcome: result.Completed}, Status: result.EchoReply, ReplyAddress: pathIP, RTTMS: &rtt}
	}
	return probefake.New(clock.Now, []probefake.RouteStep{{Observation: result.Route{Measurement: result.Measurement{Outcome: result.Completed}, OnLink: &onLink}}}, echoes)

}

// Exercise the session/start/status boundary with the real manager and engine.
// The optional observer can cancel a pending run through the same API.
func runEngineThroughAPI(t *testing.T, manager *run.Manager, url string, observe func(result.Run, func())) result.Run {
	hops := 1
	return runEngineThroughAPIWithHops(t, manager, url, &hops, observe)
}

// Nil hops exercises the API's default rather than sending a configured limit.
func runEngineThroughAPIWithHops(t *testing.T, manager *run.Manager, url string, hops *int, observe func(result.Run, func())) result.Run {
	t.Helper()
	handler, err := localapi.NewHandler("127.0.0.1:8123", manager)
	if err != nil {
		t.Fatal(err)
	}
	call := func(method, path, body, token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, "http://127.0.0.1:8123"+path, strings.NewReader(body))
		if method == "POST" {
			req.Header.Set("Origin", "http://127.0.0.1:8123")
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-HopVista-Token", token)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		return w
	}
	session := call("GET", "/api/v1/session", "", "")
	var credentials struct {
		Token string `json:"mutation_token"`
	}
	if err := json.Unmarshal(session.Body.Bytes(), &credentials); err != nil {
		t.Fatal(err)
	}
	input := map[string]any{"target_url": url}
	if hops != nil {
		input["max_hops"] = *hops
	}
	body, _ := json.Marshal(input)
	start := call("POST", "/api/v1/runs", string(body), credentials.Token)
	if start.Code != http.StatusAccepted {
		t.Fatalf("start: %d %s", start.Code, start.Body)
	}
	var started struct {
		StatusURL string `json:"status_url"`
	}
	if err := json.Unmarshal(start.Body.Bytes(), &started); err != nil {
		t.Fatal(err)
	}
	until := time.Now().Add(10 * time.Second)
	for time.Now().Before(until) {
		response := call("GET", started.StatusURL, "", "")
		var snapshot result.Run
		if err := json.Unmarshal(response.Body.Bytes(), &snapshot); err != nil {
			t.Fatal(err)
		}
		if response.Code != http.StatusOK {
			t.Fatalf("status: %d %s", response.Code, response.Body)
		}
		if observe != nil {
			observe(snapshot, func() {
				response := call("POST", started.StatusURL+"/cancel", "", credentials.Token)
				if response.Code != http.StatusOK {
					t.Fatalf("cancel: %d %s", response.Code, response.Body)
				}
				var stopping result.Run
				if err := json.Unmarshal(response.Body.Bytes(), &stopping); err != nil {
					t.Fatal(err)
				}
				if stopping.State == result.Stopping {
					if len(stopping.Findings) != 0 {
						t.Fatal("finding published while stopping")
					}
					writeCapture(t, "run-stopping", stopping)
				}
			})
		}
		if snapshot.State.Terminal() {
			if snapshot.SchemaVersion != 1 || snapshot.ConnectionMode != "direct" || snapshot.TargetURL != url || len(snapshot.Families) != 2 {
				t.Fatalf("API contract: %+v", snapshot)
			}
			return snapshot
		}
		if len(snapshot.Findings) != 0 {
			t.Fatal("finding published before terminal result")
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("API run did not finish")
	return result.Run{}
}

func TestHTTPSRedirectStopReasonThroughAPI(t *testing.T) {
	f := newTLSJourney(t, redirect("/earlier"), redirect("https://private.example/blocked"))
	snapshot := runThroughAPI(t, f, "https://start.example/start", "8.8.8.8")
	family := snapshot.Families[0]
	if family.Error == nil || family.Error.Code != "redirect_requires_new_run" || !strings.Contains(family.Error.Message, "non-public") || !strings.Contains(family.Error.Message, "new diagnostic run") {
		t.Fatalf("stop reason is not explained for the current UI: %+v", family.Error)
	}
}

func TestHTTPSJourneysThroughAPI(t *testing.T) {
	for _, tc := range []struct {
		name                            string
		responses                       []string
		stop, explanation, pathIP, role string
		outcome                         result.Outcome
		exchanges, dials, lookups       int
	}{
		{"completed", []string{redirect("/relative"), redirect("https://start.example:8443/absolute"), redirect("https://next.example/final"), noContent}, "", "", "1.1.1.1", "final_host", result.Completed, 4, 4, 2},
		{"invalid", []string{redirect("/earlier"), redirect("file:///blocked")}, "invalid_redirect", "not an allowed HTTP or HTTPS URL", "8.8.8.8", "attempted_host", result.Skipped, 2, 2, 1},
		{"ipv6", []string{redirect("/earlier"), redirect("https://[::1]/blocked")}, "ipv6_not_supported_in_v1", "No IPv6 connection was attempted", "8.8.8.8", "attempted_host", result.Unavailable, 2, 2, 1},
		{"sensitive", []string{redirect("/earlier"), redirect("https://private.example/blocked")}, "redirect_requires_new_run", "non-public", "8.8.8.8", "attempted_host", result.Skipped, 2, 2, 2},
		{"limit", []string{redirect("/one"), redirect("/two"), redirect("/three"), redirect("https://unresolved.example/four")}, "redirect_limit", "three redirects (four exchanges)", "8.8.8.8", "attempted_host", result.Skipped, 4, 4, 1},
		{"certificate", []string{redirect("/earlier"), redirect("https://wrong.example/final"), noContent}, "tls_verification_failed", "certificate verification", "1.1.1.1", "attempted_host", result.Error, 3, 3, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newTLSJourney(t, tc.responses...)
			snapshot := runThroughAPI(t, f, "https://start.example/start", tc.pathIP)
			family := snapshot.Families[0]
			if family.Outcome != tc.outcome || len(family.Exchanges) != tc.exchanges || len(f.dials) != tc.dials || len(f.lookups) != tc.lookups || len(family.Paths) != 1 {
				t.Fatalf("result=%+v dials=%v lookups=%v", family, f.dials, f.lookups)
			}
			last := family.Exchanges[len(family.Exchanges)-1]
			path := family.Paths[0]
			if path.SelectedIP != tc.pathIP || path.Trace.SelectedIP != tc.pathIP || path.Trace.TargetRole != tc.role || path.Trace.Outcome != result.DestinationReplied || last.SelectedIP != tc.pathIP {
				t.Fatalf("incorrect trace identity: %+v last=%+v", path, last)
			}
			if tc.stop == "" {
				if family.Error != nil || family.FinalHost != "next.example" || family.FinalURL != "https://next.example/final" {
					t.Fatalf("final host: %+v", family)
				}
			} else {
				if family.FinalHost != "" || family.FinalURL != "" || family.Error == nil || family.Error.Code != tc.stop || !strings.Contains(family.Error.Message, tc.explanation) {
					t.Fatalf("stop: %+v", family)
				}
				if tc.name == "certificate" {
					if last.ActualPeerIP != "1.1.1.1" || last.TCP.Outcome != result.Completed || last.TLS.Outcome != result.Error || last.TLS.DurationMS != nil || last.TLS.StartedAt == nil || last.TLS.EndedAt == nil || last.TLS.Error == nil || last.TLS.Error.Code != tc.stop || last.HTTP.DurationMS != nil || last.ResponseStatus != nil || len(f.records) != 2 {
						t.Fatalf("certificate bypass or lost failed phase/peer: %+v", last)
					}
				} else if last.RedirectStopReason != tc.stop || last.HTTP.Outcome != result.Completed {
					t.Fatalf("stopped redirect lost response: %+v", last)
				}
			}
			ids := map[string]bool{}
			for _, dns := range family.DNS {
				ids[dns.ID] = true
			}
			for i, x := range family.Exchanges {
				if x.RedirectIndex != i || x.ID == "" || !ids[x.DNSRef] {
					t.Fatalf("exchange order/DNS reference: %+v", x)
				}
				ids[x.ID] = true
				for _, m := range []result.Measurement{x.DNS, x.TCP, x.TLS, x.HTTP} {
					if ids[m.ID] || m.ID == "" {
						t.Fatalf("duplicate/missing phase id: %+v", m)
					}
					ids[m.ID] = true
				}
				if i < 2 && (x.ResponseStatus == nil || *x.ResponseStatus != 302 || x.ActualPeerIP != "8.8.8.8" || x.SelectedScope != "public" || x.ConnectionReused == nil || *x.ConnectionReused) {
					t.Fatalf("earlier evidence lost: %+v", x)
				}
			}
			if len(snapshot.Findings) != 1 || snapshot.Findings[0].ExchangeRef != last.ID {
				t.Fatalf("finding exchange: %+v", snapshot.Findings)
			}
			for _, ref := range snapshot.Findings[0].EvidenceRefs {
				if !ids[ref] {
					t.Fatalf("unresolvable finding evidence %s", ref)
				}
			}
			writeCapture(t, "run-"+tc.name, snapshot)
			writeCapture(t, "requests-"+tc.name, f.records)
		})
	}
}
