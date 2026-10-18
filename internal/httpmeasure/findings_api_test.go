package httpmeasure_test

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"testing"
	"time"

	"hopvista/internal/engine"
	"hopvista/internal/httpmeasure"
	"hopvista/internal/probe/probefake"
	"hopvista/internal/result"
	"hopvista/internal/run"
	"hopvista/internal/target"
)

// Controlled recorded timings exercise Go findings through the real API/engine.
// They are not measurements of the pipe fixture's runtime.
type recordedJourney httpmeasure.Journey

func (j recordedJourney) Measure(context.Context, string, *target.Selector) httpmeasure.Journey {
	return httpmeasure.Journey(j)
}

func TestDominantWaitThroughAPI(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		dns, tcp, tls, http     *float64
		dnsOutcome, httpOutcome result.Outcome
		want, ref               string
	}{
		{"tie", number(200), number(200), number(10), number(20), result.Completed, result.Completed, "http_response_observed", "http"},
		{"missingcomparables", nil, nil, nil, number(200), result.Completed, result.Completed, "http_response_observed", "http"},
		{"missingduration", number(20), nil, nil, number(200), result.Completed, result.Completed, "dominant_measured_wait", "http"},
		{"measuredzero", nil, number(100), nil, number(0), result.Skipped, result.Completed, "dominant_measured_wait", "tcp"},
		{"unavailableexcluded", number(9000), number(100), nil, number(40), result.Unavailable, result.Completed, "dominant_measured_wait", "tcp"},
		{"failurebeforeranking", number(20), number(1000), number(10), nil, result.Completed, result.Timeout, "http_phase_incomplete", "http"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status := 200
			x := result.Exchange{ID: "exchange-2", RedirectIndex: 1, RequestedURL: "https://start.example/final", SelectedIP: "8.8.8.8", SelectedScope: "public", DNSRef: "dns-1", ResponseStatus: &status,
				DNS: result.Measurement{ID: "dns", Outcome: tc.dnsOutcome, DurationMS: tc.dns}, TCP: result.Measurement{ID: "tcp", Outcome: result.Completed, DurationMS: tc.tcp}, TLS: result.Measurement{ID: "tls", Outcome: result.Skipped, DurationMS: tc.tls}, HTTP: result.Measurement{ID: "http", Outcome: tc.httpOutcome, DurationMS: tc.http}}
			if tc.tls != nil {
				x.TLS.Outcome = result.Completed
			}
			outcome := result.Completed
			if tc.httpOutcome != result.Completed {
				x.ResponseStatus = nil
				outcome = tc.httpOutcome
			}
			earlierStatus := 302
			earlier := result.Exchange{ID: "exchange-1", RequestedURL: "https://start.example/start", Hostname: "start.example", Port: 443, SelectedIP: "8.8.8.8", SelectedScope: "public", DNSRef: "dns-1", ResponseStatus: &earlierStatus,
				DNS:  result.Measurement{ID: "earlier-dns", Outcome: result.Completed, DurationMS: number(20)},
				TCP:  result.Measurement{ID: "earlier-tcp", Outcome: result.Completed, DurationMS: number(10)},
				TLS:  result.Measurement{ID: "earlier-tls", Outcome: result.Completed, DurationMS: number(10)},
				HTTP: result.Measurement{ID: "earlier-http", Outcome: result.Completed, DurationMS: number(50000)}}
			j := recordedJourney{Outcome: outcome, SelectedIP: "8.8.8.8", DNS: []result.DNS{{Measurement: result.Measurement{ID: "dns-1", Outcome: result.Completed}, Hostname: "start.example"}}, Exchanges: []result.Exchange{earlier, x}}
			clock := &journeyClock{}
			e := &engine.Engine{HTTP: j, Probe: journeyAdapter(clock, "8.8.8.8"), Clock: clock}
			snapshot := runEngineThroughAPI(t, run.NewManager(e), "https://start.example/start", nil)
			if len(snapshot.Findings) != 1 {
				t.Fatalf("findings: %+v", snapshot.Findings)
			}
			f := snapshot.Findings[0]
			if f.Type != tc.want || f.ExchangeRef != "exchange-2" || len(f.EvidenceRefs) == 0 || f.EvidenceRefs[0] != tc.ref {
				t.Fatalf("finding: %+v", f)
			}
			for _, ref := range f.EvidenceRefs {
				if ref != "dns" && ref != "tcp" && ref != "tls" && ref != "http" {
					t.Fatalf("cross-exchange/unresolvable evidence: %s", ref)
				}
			}
			if tc.name == "measuredzero" && (snapshot.Families[0].Exchanges[1].HTTP.DurationMS == nil || *snapshot.Families[0].Exchanges[1].HTTP.DurationMS != 0) {
				t.Fatal("measured zero lost in JSON")
			}
			writeCapture(t, "run-"+tc.name, snapshot)
		})
	}
}

func number(v float64) *float64 { return &v }

func TestCompletedErrorResponsesThroughAPI(t *testing.T) {
	for _, status := range []int{404, 503} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			f := newTLSJourney(t, fmt.Sprintf("HTTP/1.1 %d Controlled\r\nContent-Length: 0\r\n\r\n", status))
			snapshot := runThroughAPI(t, f, "https://start.example/start", "8.8.8.8")
			family := snapshot.Families[0]
			x := family.Exchanges[0]
			if family.Outcome != result.Completed || family.Error != nil || x.ResponseStatus == nil || *x.ResponseStatus != status || x.HTTP.Outcome != result.Completed || x.HTTP.DurationMS == nil || x.TCP.DurationMS == nil || x.TLS.DurationMS == nil || snapshot.Findings[0].Type != "http_response_observed" || !strings.Contains(snapshot.Findings[0].Observation, fmt.Sprint(status)) {
				t.Fatalf("completed error response: %+v", snapshot)
			}
			writeCapture(t, fmt.Sprintf("run-status%d", status), snapshot)
		})
	}
}

func TestInterruptedHTTPThroughAPI(t *testing.T) {
	for _, mode := range []string{"cancelled", "deadline"} {
		t.Run(mode, func(t *testing.T) {
			f := newTLSJourney(t, redirect("/held"), "")
			f.requestReceived = make(chan struct{})
			clock := &journeyClock{}
			e := &engine.Engine{HTTP: httpmeasure.ControlledTLSClient(f.roots, f.dial), Resolver: lookupFunc(f.resolve), Probe: probefake.New(clock.Now, nil, nil), Clock: clock}
			window := 5 * time.Second
			if mode == "deadline" {
				window = 300 * time.Millisecond
			}
			cancelled := false
			snapshot := runEngineThroughAPI(t, run.NewManagerWithWindow(e, window), "https://start.example/start", func(s result.Run, cancel func()) {
				if mode != "cancelled" || cancelled || s.State != result.Running || len(s.Families) == 0 || len(s.Families[0].Exchanges) != 1 {
					return
				}
				select {
				case <-f.requestReceived:
					writeCapture(t, "run-running", s)
					cancelled = true
					cancel()
				default:
				}
			})
			for range f.dials {
				select {
				case <-f.complete:
				case <-time.After(6 * time.Second):
					t.Fatal("fixture did not finish")
				}
			}
			family := snapshot.Families[0]
			want, state := result.Deadline, result.RunCompleted
			if mode == "cancelled" {
				want, state = result.Cancelled, result.RunCancelled
			}
			if snapshot.State != state || family.Outcome != want || len(family.Exchanges) != 2 || family.Exchanges[0].HTTP.Outcome != result.Completed {
				t.Fatalf("interrupted result: %+v", snapshot)
			}
			x := family.Exchanges[1]
			if family.Error == nil || family.Error.Code != mode || x.HTTP.Error == nil || x.HTTP.Error.Code != mode {
				t.Fatalf("interruption mislabelled: family=%+v HTTP=%+v", family.Error, x.HTTP)
			}
			if x.TCP.Outcome != result.Completed || x.TLS.Outcome != result.Completed || x.HTTP.Outcome != want || x.HTTP.DurationMS != nil || x.HTTP.StartedAt == nil || x.ResponseStatus != nil || len(family.Paths) != 0 {
				t.Fatalf("partial evidence: %+v", x)
			}
			if mode == "cancelled" && len(snapshot.Findings) != 0 {
				t.Fatal("cancelled run invented finding")
			}
			writeCapture(t, "run-"+mode, snapshot)
		})
	}
}

func TestFirstByteRetainedWhenHeadersTimeoutThroughAPI(t *testing.T) {
	f := newTLSJourney(t, noContent)
	f.headerGap = 3200 * time.Millisecond
	snapshot := runThroughAPI(t, f, "https://start.example/start", "8.8.8.8")
	family := snapshot.Families[0]
	x := family.Exchanges[0]
	if family.Outcome != result.Timeout || family.Error == nil || family.Error.Code != "response_header_timeout" || x.HTTP.Outcome != result.Completed || x.HTTP.DurationMS == nil || x.HTTP.Error == nil || x.ResponseStatus != nil || len(snapshot.Findings) != 1 || snapshot.Findings[0].Type != "http_phase_incomplete" || snapshot.Findings[0].EvidenceRefs[0] != x.HTTP.ID || !strings.Contains(snapshot.Findings[0].Observation, "first response-header byte was measured") {
		t.Fatalf("first byte mistaken for full response: %+v", snapshot)
	}
	writeCapture(t, "run-partialheaders", snapshot)
}

func TestRedirectLookupFailureThroughAPI(t *testing.T) {
	f := newTLSJourney(t, redirect("https://missing.example/next"))
	clock := &journeyClock{}
	e := &engine.Engine{HTTP: httpmeasure.ControlledTLSClient(f.roots, f.dial), Resolver: lookupFunc(func(ctx context.Context, host string) ([]netip.Addr, error) {
		if host == "missing.example" {
			return nil, errors.New("controlled DNS failure")
		}
		return f.resolve(ctx, host)
	}), Probe: journeyAdapter(clock, "8.8.8.8"), Clock: clock}
	snapshot := runEngineThroughAPI(t, run.NewManager(e), "https://start.example/start", nil)
	<-f.complete
	family := snapshot.Families[0]
	if len(family.DNS) != 2 || len(family.Exchanges) != 1 || len(f.dials) != 1 || family.Exchanges[0].HTTP.Outcome != result.Completed || family.DNS[0].DurationMS == nil || family.DNS[1].Outcome != result.Error || len(snapshot.Findings) != 1 || snapshot.Findings[0].Type != "ipv4_selection_incomplete" || snapshot.Findings[0].EvidenceRefs[0] != "dns-2" || !strings.Contains(snapshot.Findings[0].Observation, "missing.example") {
		t.Fatalf("redirect DNS failure hidden: %+v", snapshot)
	}
	writeCapture(t, "run-redirectdns", snapshot)
}
