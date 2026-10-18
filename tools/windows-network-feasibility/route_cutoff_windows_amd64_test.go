package main

import (
	"context"
	"encoding/json"
	"net/netip"
	"os"
	"testing"
	"time"
)

type routeCutoffObservation struct {
	CapMS           int64           `json:"cap_ms"`
	BudgetMS        int64           `json:"budget_ms"`
	InducedDelayMS  int64           `json:"induced_route_phase_delay_ms"`
	RouteReturnCode uint32          `json:"route_return_code"`
	RouteElapsedMS  float64         `json:"route_elapsed_ms"`
	DeadlineAt      time.Time       `json:"deadline_at"`
	ReturnedAt      time.Time       `json:"returned_at"`
	PastDeadline    bool            `json:"past_deadline_at_return"`
	Echo            echoObservation `json:"echo"`
}

type expiredRouteObservation struct {
	CapMS                int64           `json:"cap_ms"`
	BudgetMS             int64           `json:"budget_ms"`
	PostDeadlineDelayMS  int64           `json:"post_deadline_delay_ms"`
	RouteReturnCode      uint32          `json:"route_return_code"`
	RouteElapsedMS       float64         `json:"route_elapsed_ms"`
	DeadlineAt           time.Time       `json:"deadline_at"`
	StopObservedAt       time.Time       `json:"stop_observed_at"`
	RoutePhaseReturnedAt time.Time       `json:"route_phase_returned_at"`
	QuiescentAt          time.Time       `json:"quiescent_at"`
	StopToQuiescentMS    float64         `json:"stop_to_quiescent_ms"`
	Echo                 echoObservation `json:"echo"`
}

// Tests route-then-probe ordering with an induced delay. This does not bound
// GetBestRoute2 itself, which has no cancellation or timeout argument.
func TestLiveRouteThenProbeCutoff(t *testing.T) {
	if os.Getenv("HOPVISTA_ROUTE_CUTOFF") != "1" {
		t.Skip("enable with run.ps1 -RouteCutoffCheck")
	}
	if environment()["elevated"] == true {
		t.Fatal("requires an unelevated process")
	}
	ip := netip.MustParseAddr("192.0.2.1")
	var observations []routeCutoffObservation
	for _, cap := range []time.Duration{750 * time.Millisecond, 900 * time.Millisecond, time.Second} {
		budget := cap + 80*time.Millisecond
		ctx, cancel := context.WithTimeout(context.Background(), budget)
		deadline, _ := ctx.Deadline()
		r := route(ip)
		if r.ReturnCode != 0 || r.Error != "" {
			cancel()
			t.Fatalf("IPv4 route prerequisite unavailable: code=%d error=%s", r.ReturnCode, r.Error)
		}
		time.Sleep(100 * time.Millisecond)
		o := guardedProbeCap(ctx, "after-delayed-route", cap)
		returned := time.Now()
		cancel()
		item := routeCutoffObservation{cap.Milliseconds(), budget.Milliseconds(), 100, r.ReturnCode, r.ElapsedMS, deadline.UTC(), returned.UTC(), !returned.Before(deadline), o}
		observations = append(observations, item)
		if o.Outcome != "not_started" || !o.StartedAt.IsZero() || o.CloseOK != nil || o.TimeoutMS != 0 {
			t.Errorf("cap=%dms: probe started after route phase consumed start allowance: %+v", cap.Milliseconds(), o)
		}
		t.Logf("cap=%dms route=%.2fms refused=%v sent=%v", cap.Milliseconds(), r.ElapsedMS, o.Outcome == "not_started", !o.StartedAt.IsZero())
	}
	// The route phase deliberately stays active past the cutoff. The real route
	// lookup runs first; the induced pause models a slow route phase, not a
	// measured GetBestRoute2 stall or a cancellation mechanism for that API.
	const cap = 750 * time.Millisecond
	const budget = 50 * time.Millisecond
	const postDeadlineDelay = 30 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), budget)
	deadline, _ := ctx.Deadline()
	stopObserved := make(chan time.Time, 1)
	go func() {
		<-ctx.Done()
		stopObserved <- time.Now()
	}()
	r := route(ip)
	if r.ReturnCode != 0 || r.Error != "" {
		cancel()
		t.Fatalf("IPv4 route prerequisite unavailable: code=%d error=%s", r.ReturnCode, r.Error)
	}
	if delay := time.Until(deadline.Add(postDeadlineDelay)); delay > 0 {
		time.Sleep(delay)
	}
	routePhaseReturned := time.Now()
	stopAt := <-stopObserved
	o := guardedProbeCap(ctx, "after-expired-route-phase", cap)
	quiescentAt := time.Now()
	cancel()
	expired := expiredRouteObservation{
		CapMS: cap.Milliseconds(), BudgetMS: budget.Milliseconds(), PostDeadlineDelayMS: postDeadlineDelay.Milliseconds(),
		RouteReturnCode: r.ReturnCode, RouteElapsedMS: r.ElapsedMS, DeadlineAt: deadline.UTC(),
		StopObservedAt: stopAt.UTC(), RoutePhaseReturnedAt: routePhaseReturned.UTC(),
		QuiescentAt: quiescentAt.UTC(), StopToQuiescentMS: float64(quiescentAt.Sub(stopAt).Microseconds()) / 1000,
		Echo: o,
	}
	if routePhaseReturned.Before(deadline) || routePhaseReturned.Before(stopAt) {
		t.Errorf("route phase did not outlast the cutoff: %+v", expired)
	}
	if o.Outcome != "not_started" || !o.StartedAt.IsZero() || o.CloseOK != nil || o.TimeoutMS != 0 {
		t.Errorf("probe started after route phase crossed cutoff: %+v", o)
	}
	t.Logf("expired route phase: stop-to-quiescent=%.2fms echo=%s", expired.StopToQuiescentMS, o.Outcome)
	if err := os.MkdirAll(".local", 0700); err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(struct {
		CapturedAt   time.Time                `json:"captured_at"`
		Environment  map[string]any           `json:"environment"`
		Destination  string                   `json:"destination"`
		Observations []routeCutoffObservation `json:"observations"`
		ExpiredRoute expiredRouteObservation  `json:"expired_route_phase"`
		Gate         string                   `json:"gate"`
	}{time.Now().UTC(), environment(), ip.String(), observations, expired, "incomplete; induced route-phase delay tests refusal, not a route latency bound"}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".local/route-cutoff.json", data, 0600); err != nil {
		t.Fatal(err)
	}
}
