package pathmeasure

import (
	"context"
	"fmt"
	"net/netip"
	"sync"
	"testing"
	"time"

	"hopvista/internal/probe/probefake"
	"hopvista/internal/result"
)

type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *fakeClock) Now() time.Time { c.mu.Lock(); defer c.mu.Unlock(); return c.now }
func (c *fakeClock) Wait(ctx context.Context, d time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.mu.Lock()
	if d > 0 {
		c.now = c.now.Add(d)
	}
	c.mu.Unlock()
	return nil
}
func (c *fakeClock) Advance(d time.Duration) { c.mu.Lock(); c.now = c.now.Add(d); c.mu.Unlock() }

func echo(status result.EchoStatus, rtt float64, address string) probefake.EchoStep {
	o := result.Echo{Measurement: result.Measurement{Outcome: result.Completed}, Status: status, ReplyAddress: address}
	if status == result.NoReply {
		o.Outcome = result.Timeout
	} else {
		o.RTTMS = &rtt
	}
	return probefake.EchoStep{Observation: o}
}

func TestOnLinkSamplingAndSilentHop(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)}
	ip := netip.MustParseAddr("192.0.2.1")
	onLink := true
	routes := []probefake.RouteStep{{Observation: result.Route{Measurement: result.Measurement{Outcome: result.Completed}, OnLink: &onLink, SourceIP: "192.0.2.2"}}}
	echoes := []probefake.EchoStep{echo(result.EchoReply, 10, ip.String()), echo(result.NoReply, 0, ""), echo(result.EchoReply, 14, ip.String()), echo(result.EchoReply, 20, ip.String()), echo(result.EchoReply, 18, ip.String()), echo(result.TimeExceeded, 4, "192.0.2.254"), echo(result.NoReply, 0, ""), echo(result.EchoReply, 10, ip.String())}
	a := probefake.New(clock.Now, routes, echoes)
	var progress []result.Path
	p := MeasureWithProgress(context.Background(), a, clock, Request{ID: "p", SelectedIP: ip, DeadlineAt: clock.Now().Add(30 * time.Second), MaxHops: 4, TargetRole: "final_host"}, func(snapshot result.Path) { progress = append(progress, snapshot) })
	if p.Gateway.Outcome != result.Skipped || len(p.Gateway.Samples) != 0 {
		t.Fatalf("on-link gateway: %+v", p.Gateway)
	}
	if p.Target.Attempted != 5 || p.Target.Replied != 4 || p.Target.Variation == nil || p.Target.Variation.Outcome != result.Completed || *p.Target.Variation.ValueMS != 4 {
		t.Fatalf("target sampling/variation: %+v", p.Target)
	}
	if p.Trace.Outcome != result.DestinationReplied || p.Trace.HighestAttemptedTTL != 3 || len(p.Trace.Hops) != 3 || p.Trace.Hops[1].Status != result.NoReply {
		t.Fatalf("trace: %+v", p.Trace)
	}
	var firstSample *result.Path
	for i := range progress {
		if len(progress[i].Target.Samples) == 1 {
			firstSample = &progress[i]
			break
		}
	}
	if firstSample == nil || firstSample.Target.Outcome != result.Pending || len(firstSample.Target.Samples) != 1 || firstSample.Target.Samples[0].ID != "p-target-1" || firstSample.Gateway.Samples == nil || firstSample.Trace.Hops == nil {
		t.Fatalf("progress snapshot was mutated or used null arrays: %+v", firstSample)
	}
	calls := a.Calls()
	if len(calls) != 9 {
		t.Fatalf("calls=%d", len(calls))
	}
	for i := 1; i <= 5; i++ {
		if calls[i].Echo.SelectedIP != ip || calls[i].Echo.TTL != 64 || calls[i].Echo.Cap != time.Second {
			t.Fatalf("target call %d: %+v", i, calls[i])
		}
	}
}

func TestSeparateGatewayAndUnreachableTrace(t *testing.T) {
	clock := &fakeClock{now: time.Now().UTC()}
	ip := netip.MustParseAddr("192.0.2.1")
	gateway := netip.MustParseAddr("192.0.2.254")
	onLink := false
	routes := []probefake.RouteStep{{Observation: result.Route{Measurement: result.Measurement{Outcome: result.Completed}, OnLink: &onLink, NextHop: gateway.String()}}}
	var echoes []probefake.EchoStep
	for i := 0; i < 8; i++ {
		echoes = append(echoes, echo(result.NoReply, 0, ""))
	}
	echoes = append(echoes, echo(result.TimeExceeded, 2, gateway.String()), echo(result.UnreachableReply, 3, gateway.String()))
	a := probefake.New(clock.Now, routes, echoes)
	p := Measure(context.Background(), a, clock, Request{ID: "p", SelectedIP: ip, DeadlineAt: clock.Now().Add(30 * time.Second), MaxHops: 16, TargetRole: "attempted_host"})
	if p.Gateway.Attempted != 3 || p.Gateway.Replied != 0 || p.Trace.TargetRole != "attempted_host" || p.Trace.Outcome != result.UnreachableResponse || p.Trace.HighestAttemptedTTL != 2 {
		t.Fatalf("path: %+v", p)
	}
	for _, call := range a.Calls()[1:4] {
		if call.Echo.SelectedIP != gateway || call.Echo.Cap != 900*time.Millisecond {
			t.Fatalf("gateway call: %+v", call)
		}
	}
}

func TestDeadlineAfterPendingRouteStartsNoEcho(t *testing.T) {
	clock := &fakeClock{now: time.Now().UTC()}
	ip := netip.MustParseAddr("192.0.2.1")
	a := probefake.New(clock.Now, []probefake.RouteStep{{Observation: result.Route{Measurement: result.Measurement{Outcome: result.Completed}}, Finish: func() { clock.Advance(2 * time.Second) }}}, nil)
	p := Measure(context.Background(), a, clock, Request{ID: "p", SelectedIP: ip, DeadlineAt: clock.Now().Add(time.Second), MaxHops: 16, TargetRole: "final_host"})
	if len(a.Calls()) != 1 || p.Target.Attempted != 0 || p.Target.Outcome != result.Deadline || p.Trace.Outcome != result.TraceDeadline || len(p.Trace.Hops) != 0 {
		t.Fatalf("echo scheduled after route cutoff: %+v", p)
	}
}

func TestVariationNeedsThreeSuccessfulReplies(t *testing.T) {
	now := time.Now().UTC()
	a, b := 0.0, 12.0
	samples := []result.Echo{
		{Measurement: result.Measurement{ID: "first", Outcome: result.Completed, StartedAt: &now}, Status: result.EchoReply, RTTMS: &a},
		{Measurement: result.Measurement{ID: "miss", Outcome: result.Timeout}, Status: result.NoReply},
		{Measurement: result.Measurement{ID: "second", Outcome: result.Completed, StartedAt: &now}, Status: result.EchoReply, RTTMS: &b},
	}
	v := variation("v", samples)
	if v.Outcome != result.Unavailable || v.ValueMS != nil || len(v.EvidenceRefs) != 0 {
		t.Fatalf("insufficient evidence: %+v", v)
	}
}

func TestTraceHopLimitIsNotDestination(t *testing.T) {
	clock := &fakeClock{now: time.Now().UTC()}
	ip := netip.MustParseAddr("192.0.2.1")
	a := probefake.New(clock.Now, nil, []probefake.EchoStep{echo(result.NoReply, 0, ""), echo(result.TimeExceeded, 3, "192.0.2.254")})
	out := trace(context.Background(), a, clock, result.Trace{ID: "trace", SelectedIP: ip.String(), Family: "ipv4", MaxHops: 2}, ip, clock.Now().Add(10*time.Second))
	if out.Outcome != result.HopLimitReached || out.HighestAttemptedTTL != 2 || len(out.Hops) != 2 {
		t.Fatalf("trace outcome: %+v", out)
	}
}

func TestTraceConfiguredHopLimits(t *testing.T) {
	for _, limit := range []int{1, 30, 64} {
		t.Run(fmt.Sprintf("%d", limit), func(t *testing.T) {
			clock := &fakeClock{now: time.Now().UTC()}
			ip := netip.MustParseAddr("192.0.2.1")
			echoes := make([]probefake.EchoStep, limit)
			for i := range echoes {
				echoes[i] = echo(result.NoReply, 0, "")
			}
			a := probefake.New(clock.Now, nil, echoes)
			out := trace(context.Background(), a, clock, result.Trace{ID: "trace", SelectedIP: ip.String(), Family: "ipv4", MaxHops: limit}, ip, clock.Now().Add(time.Minute))
			if out.Outcome != result.HopLimitReached || out.HighestAttemptedTTL != limit || len(out.Hops) != limit {
				t.Fatalf("limit %d: %+v", limit, out)
			}
			for i, call := range a.Calls() {
				if call.Echo.TTL != i+1 {
					t.Fatalf("hop %d TTL=%d", i+1, call.Echo.TTL)
				}
			}
		})
	}
}

func TestPathAcceptsSixtyFourHops(t *testing.T) {
	clock := &fakeClock{now: time.Now().UTC()}
	ip := netip.MustParseAddr("192.0.2.1")
	onLink := true
	routes := []probefake.RouteStep{{Observation: result.Route{Measurement: result.Measurement{Outcome: result.Completed}, OnLink: &onLink}}}
	echoes := make([]probefake.EchoStep, 5+64)
	for i := range echoes {
		echoes[i] = echo(result.NoReply, 0, "")
	}
	a := probefake.New(clock.Now, routes, echoes)
	p := Measure(context.Background(), a, clock, Request{ID: "p", SelectedIP: ip, DeadlineAt: clock.Now().Add(time.Minute), MaxHops: 64, TargetRole: "final_host"})
	if p.Route.Outcome != result.Completed || p.Trace.Outcome != result.HopLimitReached || p.Trace.HighestAttemptedTTL != 64 || len(p.Trace.Hops) != 64 {
		t.Fatalf("64-hop path: %+v", p)
	}
}
