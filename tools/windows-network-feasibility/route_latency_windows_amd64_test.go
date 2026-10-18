package main

import (
	"encoding/json"
	"net/netip"
	"os"
	"testing"
	"time"
)

type routeLatencyObservation struct {
	Destination   string    `json:"destination"`
	Family        string    `json:"family"`
	Attempt       int       `json:"attempt"`
	StartedAt     time.Time `json:"started_at"`
	ReturnedAt    time.Time `json:"returned_at"`
	APIElapsedMS  float64   `json:"api_elapsed_ms"`
	CallElapsedMS float64   `json:"call_elapsed_ms"`
	ReturnCode    uint32    `json:"return_code"`
}

type routeBatchObservation struct {
	Destination string  `json:"destination"`
	Calls       int     `json:"calls"`
	ElapsedMS   float64 `json:"elapsed_ms"`
	ReturnCode  uint32  `json:"return_code"`
}

// Opt-in route timing characterization. GetBestRoute2 has no cancellation or
// timeout parameter, so finite samples cannot establish a worst-case bound.
func TestLiveRouteLatency(t *testing.T) {
	if os.Getenv("HOPVISTA_ROUTE_LATENCY") != "1" {
		t.Skip("enable with run.ps1 -RouteLatencyStudy")
	}
	if environment()["elevated"] == true {
		t.Fatal("requires an unelevated process")
	}
	var observations []routeLatencyObservation
	var batches []routeBatchObservation
	for _, raw := range []string{"1.1.1.1", "192.0.2.1", "::1", "2606:4700:4700::1111"} {
		ip := netip.MustParseAddr(raw)
		for attempt := 1; attempt <= 25; attempt++ {
			started := time.Now()
			r := route(ip)
			returned := time.Now()
			observations = append(observations, routeLatencyObservation{
				Destination: raw, Family: r.Family, Attempt: attempt,
				StartedAt: started.UTC(), ReturnedAt: returned.UTC(),
				APIElapsedMS: r.ElapsedMS, CallElapsedMS: milliseconds(returned.Sub(started)),
				ReturnCode: r.ReturnCode,
			})
		}
		const batchCalls = 1000
		started := time.Now()
		var code uint32
		for attempt := 0; attempt < batchCalls; attempt++ {
			code = route(ip).ReturnCode
		}
		batches = append(batches, routeBatchObservation{raw, batchCalls, milliseconds(time.Since(started)), code})
	}
	if err := os.MkdirAll(".local", 0700); err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(struct {
		CapturedAt   time.Time                 `json:"captured_at"`
		Environment  map[string]any            `json:"environment"`
		Observations []routeLatencyObservation `json:"observations"`
		Batches      []routeBatchObservation   `json:"batches"`
		Gate         string                    `json:"gate"`
	}{time.Now().UTC(), environment(), observations, batches, "incomplete; route samples do not establish a worst-case bound"}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".local/route-latency.json", data, 0600); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{"1.1.1.1", "192.0.2.1", "::1", "2606:4700:4700::1111"} {
		var maxMS float64
		var code uint32
		for _, o := range observations {
			if o.Destination == raw {
				if o.CallElapsedMS > maxMS {
					maxMS = o.CallElapsedMS
				}
				code = o.ReturnCode
			}
		}
		t.Logf("target=%s samples=25 max_call_elapsed=%.2fms final_return_code=%d", raw, maxMS, code)
	}
	for _, b := range batches {
		t.Logf("target=%s batch_calls=%d batch_elapsed=%.2fms final_return_code=%d", b.Destination, b.Calls, b.ElapsedMS, b.ReturnCode)
	}
}
