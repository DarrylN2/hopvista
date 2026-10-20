package httpmeasure_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"hopvista/internal/engine"
	"hopvista/internal/httpmeasure"
	"hopvista/internal/pathmeasure"
	"hopvista/internal/probe/probefake"
	"hopvista/internal/result"
	"hopvista/internal/run"
)

// Scripts feed normalized observations at the probe boundary. Only TLS/HTTP
// are real protocol work; none of these cases probes the named public IPs.
func pathReply(status result.EchoStatus, rtt float64, address string) probefake.EchoStep {
	e := result.Echo{Measurement: result.Measurement{Outcome: result.Completed}, Status: status, ReplyAddress: address, RTTMS: &rtt}
	if status == result.NoReply {
		e.Outcome = result.Timeout
		e.ReplyAddress = ""
		e.RTTMS = nil
	}
	return probefake.EchoStep{Observation: e}
}

func TestDifficultPathsThroughAPI(t *testing.T) {
	for _, tc := range []struct {
		name              string
		hops              int // zero omits max_hops, testing the API default
		onLink, attempted bool
		trace             result.TraceOutcome
	}{
		{"highsilent", 4, false, false, result.DestinationReplied},
		{"onlink", 4, true, false, result.DestinationReplied},
		{"limit1", 1, true, false, result.HopLimitReached},
		{"default30", 0, true, false, result.HopLimitReached},
		{"limit64", 64, true, false, result.HopLimitReached},
		{"unreachable", 4, false, false, result.UnreachableResponse},
		{"attempted", 4, false, true, result.DestinationReplied},
		{"deadline", 4, true, false, result.TraceDeadline},
		{"cancelled", 4, true, false, result.TraceCancelled},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newTLSJourney(t, noContent)
			if tc.attempted {
				f.responses = []string{redirect("https://private.example/blocked")}
			}
			clock := &journeyClock{}
			index := uint32(7)
			route := result.Route{Measurement: result.Measurement{Outcome: result.Completed}, OnLink: &tc.onLink, SourceIP: "192.0.2.10", InterfaceIndex: &index}
			var script []probefake.EchoStep
			if !tc.onLink {
				route.NextHop = "192.0.2.254"
				for _, rtt := range []float64{1, 2, 1} {
					script = append(script, pathReply(result.EchoReply, rtt, route.NextHop))
				}
			}
			for i, rtt := range []float64{0, -1, 14, 20, 18} {
				if rtt < 0 || tc.trace == result.HopLimitReached || (tc.trace == result.UnreachableResponse && i > 1) {
					script = append(script, pathReply(result.NoReply, 0, ""))
				} else {
					script = append(script, pathReply(result.EchoReply, rtt, "8.8.8.8"))
				}
			}
			limit := tc.hops
			if limit == 0 {
				limit = 30
			}
			if tc.trace == result.HopLimitReached {
				for i := 0; i < limit; i++ {
					script = append(script, pathReply(result.NoReply, 0, ""))
				}
			} else if tc.trace == result.UnreachableResponse {
				script = append(script, pathReply(result.TimeExceeded, 5, route.NextHop), pathReply(result.UnreachableReply, 6, "192.0.2.253"))
			} else {
				script = append(script, pathReply(result.NoReply, 0, ""), pathReply(result.TimeExceeded, 450, "192.0.2.2"), pathReply(result.TimeExceeded, 12, "192.0.2.3"), pathReply(result.EchoReply, 14, "8.8.8.8"))
			}
			entered, release := make(chan struct{}), make(chan struct{})
			var once sync.Once
			t.Cleanup(func() { once.Do(func() { close(release) }) })
			if tc.trace == result.TraceCancelled {
				script[0].Setup = func() { close(entered); <-release }
			}
			adapter := probefake.New(clock.Now, []probefake.RouteStep{{Observation: route}}, script)
			e := &engine.Engine{HTTP: httpmeasure.ControlledTLSClient(f.roots, f.dial), Resolver: lookupFunc(f.resolve), Probe: adapter, Clock: clock}
			manager := run.NewManager(e)
			if tc.trace == result.TraceDeadline {
				manager = run.NewManagerWithWindow(e, 500*time.Millisecond)
			}
			var hops *int
			if tc.hops != 0 {
				hops = &tc.hops
			}
			cancelled := false
			snapshot := runEngineThroughAPIWithHops(t, manager, "https://start.example/path", hops, func(s result.Run, cancel func(), _ func()) {
				if tc.trace != result.TraceCancelled || cancelled {
					return
				}
				select {
				case <-entered:
					cancelled = true
					cancel()
					once.Do(func() { close(release) })
				default:
				}
			})
			select {
			case <-f.complete:
			case <-time.After(6 * time.Second):
				t.Fatal("TLS fixture did not finish")
			}
			family := snapshot.Families[0]
			if snapshot.MaxHops != limit || len(family.Exchanges) != 1 || len(family.Paths) != 1 {
				t.Fatalf("API path contract: %+v", snapshot)
			}
			x, p := family.Exchanges[0], family.Paths[0]
			if x.HTTP.Outcome != result.Completed || x.HTTP.DurationMS == nil || x.ResponseStatus == nil || p.SelectedIP != "8.8.8.8" || p.Trace.SelectedIP != "8.8.8.8" || p.Route.SourceIP != "192.0.2.10" || p.Route.InterfaceIndex == nil || *p.Route.InterfaceIndex != 7 || p.Trace.Outcome != tc.trace || p.Trace.MaxHops != limit {
				t.Fatalf("observations lost: exchange=%+v path=%+v", x, p)
			}
			role := "final_host"
			if tc.attempted {
				role = "attempted_host"
				if family.FinalHost != "" || family.FinalURL != "" {
					t.Fatal("policy stop invented a final host")
				}
			} else if family.FinalHost != "start.example" || family.FinalURL != "https://start.example/path" {
				t.Fatal("completed HTTP identity lost")
			}
			if p.Trace.TargetRole != role {
				t.Fatalf("path identity: %+v", p.Trace)
			}
			if tc.onLink {
				if p.Route.NextHop != "" || p.Gateway.Outcome != result.Skipped || p.Gateway.Attempted != 0 || len(p.Gateway.Samples) != 0 {
					t.Fatalf("invented gateway: %+v", p)
				}
			} else if p.Gateway.Attempted != 3 || p.Gateway.Replied != 3 || len(p.Gateway.Samples) != 3 || p.Gateway.Samples[0].SelectedIP != route.NextHop || *p.Gateway.Samples[0].BoundMS != 900 {
				t.Fatalf("gateway evidence: %+v", p.Gateway)
			}
			switch tc.trace {
			case result.DestinationReplied:
				if p.Trace.HighestAttemptedTTL != 4 || len(p.Trace.Hops) != 4 || p.Trace.Hops[0].Status != result.NoReply || *p.Trace.Hops[1].RTTMS != 450 || *p.Trace.Hops[2].RTTMS != 12 || p.Trace.Hops[3].ReplyAddress != "8.8.8.8" {
					t.Fatalf("high/silent hop sequence: %+v", p.Trace)
				}
				if p.Target.Attempted != 5 || p.Target.Replied != 4 || *p.Target.Samples[0].RTTMS != 0 || p.Target.Variation == nil || p.Target.Variation.ValueMS == nil || *p.Target.Variation.ValueMS != 6 || p.Target.Variation.Formula != pathmeasure.VariationFormula {
					t.Fatalf("target variation: %+v", p.Target)
				}
			case result.HopLimitReached:
				if p.Trace.HighestAttemptedTTL != limit || len(p.Trace.Hops) != limit || p.Target.Attempted != 5 || p.Target.Replied != 0 || p.Target.Variation.Outcome != result.Unavailable || p.Target.Variation.ValueMS != nil {
					t.Fatalf("HTTP/ICMP disagreement: %+v", p)
				}
			case result.UnreachableResponse:
				if p.Trace.HighestAttemptedTTL != 2 || len(p.Trace.Hops) != 2 || p.Trace.Hops[1].Status != result.UnreachableReply || p.Trace.Hops[1].ReplyAddress != "192.0.2.253" || p.Target.Replied != 1 || p.Target.Variation.Outcome != result.Unavailable || p.Target.Variation.ValueMS != nil {
					t.Fatalf("unreachable evidence: %+v", p)
				}
			case result.TraceCancelled, result.TraceDeadline:
				if len(p.Trace.Hops) != 0 || p.Trace.HighestAttemptedTTL != 0 || p.Target.Attempted != 0 || len(p.Target.Samples) != 1 || p.Target.Samples[0].Status != result.NotSent || p.Target.Samples[0].StartedAt != nil || p.Target.Samples[0].EndedAt != nil || p.Target.Samples[0].RTTMS != nil || p.Target.Samples[0].ReplyAddress != "" {
					t.Fatalf("unsent probe invented evidence: %+v", p)
				}
				want := result.Deadline
				if tc.trace == result.TraceCancelled {
					want = result.Cancelled
				}
				if p.Target.Samples[0].Outcome != want || family.Outcome != want || p.Target.Samples[0].BoundMS == nil || *p.Target.Samples[0].BoundMS != 1000 {
					t.Fatalf("stop outcomes: %+v", p)
				}
			}
			if tc.trace == result.TraceCancelled {
				if snapshot.State != result.RunCancelled || len(snapshot.Findings) != 0 {
					t.Fatal("cancelled run invented finding")
				}
			} else {
				if snapshot.State != result.RunCompleted || len(snapshot.Findings) != 1 || snapshot.Findings[0].ExchangeRef != x.ID {
					t.Fatalf("Go finding: %+v", snapshot.Findings)
				}
				finding := snapshot.Findings[0]
				// Real HTTP phase timings may meet the dominant-wait rule on a
				// loaded machine. Either valid HTTP finding must exclude ICMP.
				validType := finding.Type == "http_response_observed" || finding.Type == "dominant_measured_wait"
				if tc.attempted {
					validType = finding.Type == "redirect_stopped"
				}
				if !validType || len(finding.EvidenceRefs) == 0 || finding.Limit == "" {
					t.Fatalf("path became website diagnosis: %+v", finding)
				}
				for _, ref := range finding.EvidenceRefs {
					if ref != x.DNS.ID && ref != x.TCP.ID && ref != x.TLS.ID && ref != x.HTTP.ID {
						t.Fatalf("ICMP entered website finding: %+v", finding)
					}
				}
			}
			for i, h := range p.Trace.Hops {
				if h.TTL != i+1 || h.SelectedIP != "8.8.8.8" || h.BoundMS == nil || *h.BoundMS != 750 {
					t.Fatalf("hop order/identity/bound: %+v", h)
				}
				if h.Status == result.NoReply && (h.RTTMS != nil || h.ReplyAddress != "" || h.DurationMS != nil || h.Outcome != result.Timeout) {
					t.Fatalf("silent hop fabricated evidence: %+v", h)
				}
			}
			for _, call := range adapter.Calls() {
				if call.Route != nil && call.Route.SelectedIP.String() != "8.8.8.8" {
					t.Fatal("route used another selection")
				}
				if call.Echo != nil && call.Echo.SelectedIP.String() != "8.8.8.8" && call.Echo.SelectedIP.String() != route.NextHop {
					t.Fatal("echo used another selection")
				}
			}
			checkPathReferences(t, snapshot)
			writeCapture(t, "run-"+tc.name, snapshot)
		})
	}
}

func checkPathReferences(t *testing.T, snapshot result.Run) {
	t.Helper()
	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	var refs []string
	var visit func(any)
	visit = func(value any) {
		switch v := value.(type) {
		case []any:
			for _, item := range v {
				visit(item)
			}
		case map[string]any:
			if id, ok := v["id"].(string); ok {
				if id == "" || ids[id] {
					t.Fatalf("empty/duplicate evidence ID: %q", id)
				}
				ids[id] = true
			}
			for key, item := range v {
				if key == "dns_ref" || key == "exchange_ref" {
					if ref, ok := item.(string); ok && ref != "" {
						refs = append(refs, ref)
					}
				}
				if key == "evidence_refs" {
					for _, ref := range item.([]any) {
						refs = append(refs, fmt.Sprint(ref))
					}
				}
				visit(item)
			}
		}
	}
	visit(value)
	for _, ref := range refs {
		if !ids[ref] {
			t.Fatalf("unresolvable evidence reference %s", ref)
		}
	}
	if !strings.Contains(string(data), "\"schema_version\":1") {
		t.Fatal("schema version changed")
	}
}
