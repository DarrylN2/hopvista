package probefake_test

import (
	"context"
	"net/netip"
	"sync"
	"testing"
	"time"

	"hopvista/internal/probe"
	"hopvista/internal/probe/probefake"
	"hopvista/internal/result"
)

type clock struct {
	mu sync.Mutex
	at time.Time
}

func newClock() *clock                   { return &clock{at: time.Date(2026, 10, 14, 20, 0, 0, 0, time.UTC)} }
func (c *clock) now() time.Time          { c.mu.Lock(); defer c.mu.Unlock(); return c.at }
func (c *clock) advance(d time.Duration) { c.mu.Lock(); defer c.mu.Unlock(); c.at = c.at.Add(d) }
func request(c *clock) probe.EchoRequest {
	return probe.EchoRequest{ID: "echo-1", SelectedIP: netip.MustParseAddr("192.0.2.10"), TTL: 64, Cap: time.Second, DeadlineAt: c.now().Add(time.Minute)}
}

func TestNormalizedEchoObservations(t *testing.T) {
	for _, tc := range []struct {
		status  result.EchoStatus
		outcome result.Outcome
		reply   bool
	}{
		{result.EchoReply, result.Completed, true}, {result.TimeExceeded, result.Completed, true},
		{result.UnreachableReply, result.Completed, true}, {result.NoReply, result.Timeout, false},
		{result.CallError, result.Error, false},
	} {
		t.Run(string(tc.status), func(t *testing.T) {
			c := newClock()
			raw := result.Echo{Measurement: result.Measurement{Outcome: tc.outcome}, Status: tc.status}
			if tc.reply {
				rtt := 0.0
				raw.RTTMS = &rtt
				raw.ReplyAddress = "192.0.2.10"
			}
			if tc.outcome == result.Error {
				raw.Error = &result.ErrorData{Code: "adapter_failure", Message: "scripted API failure"}
			}
			var a probe.Adapter = probefake.New(c.now, nil, []probefake.EchoStep{{Observation: raw, Finish: func() { c.advance(time.Second) }}})
			r := request(c)
			got := a.Echo(context.Background(), r)
			if got.Outcome != tc.outcome || got.Status != tc.status || got.ID != r.ID || got.SelectedIP != r.SelectedIP.String() || got.TTL != 64 {
				t.Fatalf("%+v", got)
			}
			if got.StartedAt == nil || got.EndedAt == nil || *got.BoundMS != 1000 {
				t.Fatalf("missing full-cap metadata: %+v", got)
			}
			if tc.reply && (got.RTTMS == nil || *got.RTTMS != 0 || got.DurationMS == nil || *got.DurationMS != 1000) {
				t.Fatal("measured zero RTT must remain distinct from absent RTT")
			}
			if !tc.reply && (got.RTTMS != nil || got.ReplyAddress != "" || got.DurationMS != nil) {
				t.Fatal("failed measurement invented timing or reply")
			}
		})
	}
}

func TestRefusalBeforeSend(t *testing.T) {
	for _, mode := range []string{"budget", "cancel_before", "cancel_setup", "delay_setup"} {
		t.Run(mode, func(t *testing.T) {
			c := newClock()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			r := request(c)
			want := result.Deadline
			setup := func() {}
			switch mode {
			case "budget":
				r.DeadlineAt = c.now().Add(1249 * time.Millisecond)
			case "cancel_before":
				cancel()
				want = result.Cancelled
			case "cancel_setup":
				setup = cancel
				want = result.Cancelled
			case "delay_setup":
				setup = func() { c.advance(time.Minute) }
			}
			a := probefake.New(c.now, nil, []probefake.EchoStep{{Setup: setup, Finish: func() { t.Error("refused send started") }}})
			got := a.Echo(ctx, r)
			if got.Outcome != want || got.Status != result.NotSent || got.StartedAt != nil || got.RTTMS != nil || got.DurationMS != nil || len(a.Calls()) != 0 {
				t.Fatalf("refusal: %+v", got)
			}
		})
	}
}

func TestDelayedRouteCannotAuthorizeLateEcho(t *testing.T) {
	c := newClock()
	onLink := true
	index := uint32(4)
	a := probefake.New(c.now, []probefake.RouteStep{{Observation: result.Route{Measurement: result.Measurement{Outcome: result.Completed}, SourceIP: "192.0.2.1", InterfaceIndex: &index, OnLink: &onLink}, Finish: func() { c.advance(time.Minute) }}}, nil)
	r := request(c)
	route := a.Route(context.Background(), probe.RouteRequest{ID: "route-1", SelectedIP: r.SelectedIP, DeadlineAt: r.DeadlineAt})
	if route.Outcome != result.Completed || route.OnLink == nil || !*route.OnLink || route.NextHop != "" || *route.DurationMS != 60000 {
		t.Fatalf("%+v", route)
	}
	echo := a.Echo(context.Background(), r)
	if echo.Outcome != result.Deadline || echo.StartedAt != nil || len(a.Calls()) != 1 {
		t.Fatalf("late echo: %+v", echo)
	}
}

// This models an indefinitely pending synchronous call until explicitly released.
// It proves fake ownership semantics, not that a real GetBestRoute2 was stalled.
func TestStartedWorkRetainsOwnershipThroughCancellationAndCleanup(t *testing.T) {
	for _, route := range []bool{true, false} {
		t.Run(map[bool]string{true: "route", false: "echo"}[route], func(t *testing.T) {
			c := newClock()
			entered := make(chan struct{})
			release := make(chan struct{})
			done := make(chan struct{})
			var once sync.Once
			defer once.Do(func() { close(release) })
			finish := func() { close(entered); <-release; c.advance(2 * time.Second) }
			a := probefake.New(c.now, []probefake.RouteStep{{Observation: result.Route{Measurement: result.Measurement{Outcome: result.Completed}}, Finish: finish}}, []probefake.EchoStep{{Observation: result.Echo{Measurement: result.Measurement{Outcome: result.Timeout}, Status: result.NoReply}, Finish: finish}})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			r := request(c)
			go func() {
				defer close(done)
				if route {
					a.Route(ctx, probe.RouteRequest{ID: "route-1", SelectedIP: r.SelectedIP, DeadlineAt: r.DeadlineAt})
				} else {
					a.Echo(ctx, r)
				}
			}()
			select {
			case <-entered:
			case <-time.After(2 * time.Second):
				t.Fatal("fake did not enter")
			}
			cancel()
			select {
			case <-done:
				t.Fatal("returned before cleanup release")
			default:
			}
			if len(a.Calls()) != 1 {
				t.Fatal("started work lost")
			}
			got := a.Echo(ctx, r)
			if got.Outcome != result.Cancelled || len(a.Calls()) != 1 {
				t.Fatal("new work started after cancellation")
			}
			once.Do(func() { close(release) })
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Fatal("fake failed to drain")
			}
		})
	}
}

func TestInvalidRequestsAndScriptExhaustionDoNotStartWork(t *testing.T) {
	c := newClock()
	a := probefake.New(c.now, nil, nil)
	for _, ip := range []string{"::1", "::ffff:192.0.2.1", "0.0.0.0", "224.0.0.1", "255.255.255.255"} {
		r := request(c)
		r.SelectedIP = netip.MustParseAddr(ip)
		if got := a.Echo(context.Background(), r); got.Error == nil || got.Error.Code != "invalid_probe_request" {
			t.Fatalf("%s: %+v", ip, got)
		}
		if got := a.Route(context.Background(), probe.RouteRequest{ID: "route-1", SelectedIP: r.SelectedIP, DeadlineAt: r.DeadlineAt}); got.Error == nil {
			t.Fatalf("%s: %+v", ip, got)
		}
	}
	if got := a.Echo(context.Background(), request(c)); got.Error == nil || got.Error.Code != "fake_script_exhausted" {
		t.Fatalf("%+v", got)
	}
	for _, cap := range []time.Duration{0, -time.Millisecond, time.Microsecond, time.Second + time.Millisecond} {
		r := request(c)
		r.Cap = cap
		if got := a.Echo(context.Background(), r); got.Error == nil || got.Error.Code != "invalid_probe_request" {
			t.Fatalf("invalid cap %v: %+v", cap, got)
		}
	}
	if len(a.Calls()) != 0 {
		t.Fatal("invalid work started")
	}
}

func TestRouteFailuresAndCancellationStayNormalized(t *testing.T) {
	c := newClock()
	request := probe.RouteRequest{ID: "route-1", SelectedIP: netip.MustParseAddr("192.0.2.10"), DeadlineAt: c.now().Add(time.Minute)}
	a := probefake.New(c.now, []probefake.RouteStep{{Observation: result.Route{Measurement: result.Measurement{Outcome: result.Error, Error: &result.ErrorData{Code: "route_unavailable", Message: "No route was selected."}}}}}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got := a.Route(ctx, request); got.Outcome != result.Cancelled || got.StartedAt != nil || len(a.Calls()) != 0 {
		t.Fatalf("cancelled route: %+v", got)
	}
	got := a.Route(context.Background(), request)
	if got.Outcome != result.Error || got.Error == nil || got.Error.Code != "route_unavailable" || got.DurationMS != nil || got.BoundMS != nil || got.SourceIP != "" || got.OnLink != nil {
		t.Fatalf("route failure: %+v", got)
	}
	if len(a.Calls()) != 1 {
		t.Fatal("route script was consumed by cancellation")
	}
	// Calls returns detached request copies, protecting later assertions.
	calls := a.Calls()
	calls[0].Route.ID = "mutated"
	if a.Calls()[0].Route.ID != "route-1" {
		t.Fatal("calls exposed internal request storage")
	}
}
