package engine

import (
	"context"
	"errors"
	"net/netip"
	"strings"
	"sync"
	"testing"
	"time"

	"hopvista/internal/httpmeasure"
	"hopvista/internal/probe/probefake"
	"hopvista/internal/result"
	"hopvista/internal/run"
	"hopvista/internal/target"
)

type journeyFunc func(context.Context, string, *target.Selector) httpmeasure.Journey

func (f journeyFunc) Measure(ctx context.Context, url string, selector *target.Selector) httpmeasure.Journey {
	return f(ctx, url, selector)
}

type pendingHTTP struct{ entered chan struct{} }

func (p pendingHTTP) Measure(ctx context.Context, url string, s *target.Selector) httpmeasure.Journey {
	return p.MeasureWithProgress(ctx, url, s, nil)
}
func (p pendingHTTP) MeasureWithProgress(ctx context.Context, _ string, _ *target.Selector, publish func(httpmeasure.Journey)) httpmeasure.Journey {
	dns := result.DNS{Measurement: result.Measurement{ID: "dns-1", Outcome: result.Completed, Method: "fake_resolver"}, Hostname: "example.com", Addresses: []string{"192.0.2.1"}, SelectedIP: "192.0.2.1"}
	j := httpmeasure.Journey{Outcome: result.Pending, DNS: []result.DNS{dns}, Exchanges: []result.Exchange{}, SelectedIP: "192.0.2.1"}
	if publish != nil {
		publish(j)
	}
	close(p.entered)
	<-ctx.Done()
	j.Outcome = result.Cancelled
	j.StopReason = "cancelled"
	return j
}

type testClock struct{ now time.Time }

func (c *testClock) Now() time.Time { return c.now }
func (c *testClock) Wait(ctx context.Context, d time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if d > 0 {
		c.now = c.now.Add(d)
	}
	return nil
}

func TestEnginePublishesHTTPThenFinalHostPath(t *testing.T) {
	clock := &testClock{now: time.Now().UTC()}
	ip := netip.MustParseAddr("127.0.0.1")
	onLink := true
	echoes := make([]probefake.EchoStep, 6)
	for i := range echoes {
		rtt := float64(i)
		echoes[i] = probefake.EchoStep{Observation: result.Echo{Measurement: result.Measurement{Outcome: result.Completed}, Status: result.EchoReply, ReplyAddress: ip.String(), RTTMS: &rtt}}
	}
	a := probefake.New(clock.Now, []probefake.RouteStep{{Observation: result.Route{Measurement: result.Measurement{Outcome: result.Completed}, OnLink: &onLink}}}, echoes)
	status := 503
	http := journeyFunc(func(_ context.Context, url string, _ *target.Selector) httpmeasure.Journey {
		return httpmeasure.Journey{Outcome: result.Completed, DNS: []result.DNS{{Measurement: result.Measurement{ID: "dns-1", Outcome: result.Skipped}, Hostname: "127.0.0.1", SelectedIP: ip.String()}}, Exchanges: []result.Exchange{{ID: "exchange-1", RequestedURL: url, SelectedIP: ip.String(), HTTP: result.Measurement{ID: "exchange-1-http", Outcome: result.Completed}, ResponseStatus: &status}}, FinalURL: url, FinalHost: "127.0.0.1", SelectedIP: ip.String()}
	})
	e := &Engine{HTTP: http, Probe: a, Clock: clock}
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(30*time.Second))
	defer cancel()
	var stages []int
	families, findings, err := e.Run(ctx, run.Input{TargetURL: "http://127.0.0.1/", MaxHops: 16}, func(f []result.Family) { stages = append(stages, len(f[0].Paths)) })
	if err != nil || len(stages) < 3 || stages[0] != 0 || stages[len(stages)-1] != 1 {
		t.Fatalf("publish stages=%v err=%v", stages, err)
	}
	if len(families) != 2 || families[1].Outcome != result.Unavailable || families[1].Error.Code != "ipv6_not_supported_in_v1" {
		t.Fatalf("family split: %+v", families)
	}
	p := families[0].Paths[0]
	if p.Trace.TargetRole != "final_host" || p.Trace.Outcome != result.DestinationReplied || p.Target.Attempted != 5 || p.Target.Replied != 5 {
		t.Fatalf("path: %+v", p)
	}
	if len(findings) != 1 || findings[0].EvidenceRefs[0] != "exchange-1-http" || findings[0].ExchangeRef != "exchange-1" {
		t.Fatalf("finding: %+v", findings)
	}
}

func TestEngineLabelsIncompleteRedirectAsAttemptedHost(t *testing.T) {
	clock := &testClock{now: time.Now().UTC()}
	ip := netip.MustParseAddr("127.0.0.1")
	onLink := true
	echoes := make([]probefake.EchoStep, 6)
	for i := range echoes {
		echoes[i] = probefake.EchoStep{Observation: result.Echo{Measurement: result.Measurement{Outcome: result.Timeout}, Status: result.NoReply}}
	}
	a := probefake.New(clock.Now, []probefake.RouteStep{{Observation: result.Route{Measurement: result.Measurement{Outcome: result.Completed}, OnLink: &onLink}}}, echoes)
	http := journeyFunc(func(_ context.Context, url string, _ *target.Selector) httpmeasure.Journey {
		return httpmeasure.Journey{Outcome: result.Error, StopReason: "tcp_connect_failed", Exchanges: []result.Exchange{{ID: "exchange-1", RequestedURL: url, SelectedIP: ip.String(), TCP: result.Measurement{ID: "exchange-1-tcp", Outcome: result.Error}, HTTP: result.Measurement{ID: "exchange-1-http", Outcome: result.Unavailable}}}, SelectedIP: ip.String()}
	})
	e := &Engine{HTTP: http, Probe: a, Clock: clock}
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(30*time.Second))
	defer cancel()
	families, findings, err := e.Run(ctx, run.Input{TargetURL: "http://127.0.0.1/", MaxHops: 1}, func([]result.Family) {})
	if err != nil || families[0].Paths[0].Trace.TargetRole != "attempted_host" || families[0].Paths[0].Trace.Outcome != result.HopLimitReached || families[0].Outcome != result.Error || findings[0].EvidenceRefs[0] != "exchange-1-tcp" {
		t.Fatalf("attempted host: %+v %+v %v", families, findings, err)
	}
}

func TestEngineNoSelectedIPDoesNotProbe(t *testing.T) {
	clock := &testClock{now: time.Now().UTC()}
	a := probefake.New(clock.Now, nil, nil)
	http := journeyFunc(func(_ context.Context, _ string, _ *target.Selector) httpmeasure.Journey {
		return httpmeasure.Journey{Outcome: result.Unavailable, StopReason: "ipv4_unavailable", DNS: []result.DNS{{Measurement: result.Measurement{ID: "dns-1", Outcome: result.Completed}, Hostname: "example.com"}}}
	})
	e := &Engine{HTTP: http, Probe: a, Clock: clock}
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(30*time.Second))
	defer cancel()
	families, findings, err := e.Run(ctx, run.Input{TargetURL: "https://example.com", MaxHops: 16}, func([]result.Family) {})
	if err != nil || len(families[0].Paths) != 0 || len(a.Calls()) != 0 || len(findings) != 1 || findings[0].EvidenceRefs[0] != "dns-1" {
		t.Fatalf("no selection: %+v %+v %v", families, findings, err)
	}
}

func TestHTTPResponseWithSilentICMPKeepsObservationsSeparate(t *testing.T) {
	clock := &testClock{now: time.Now().UTC()}
	ip := netip.MustParseAddr("127.0.0.1")
	onLink := true
	echoes := make([]probefake.EchoStep, 7)
	for i := range echoes {
		echoes[i] = probefake.EchoStep{Observation: result.Echo{Measurement: result.Measurement{Outcome: result.Timeout}, Status: result.NoReply}}
	}
	a := probefake.New(clock.Now, []probefake.RouteStep{{Observation: result.Route{Measurement: result.Measurement{Outcome: result.Completed}, OnLink: &onLink}}}, echoes)
	http := journeyFunc(func(_ context.Context, url string, _ *target.Selector) httpmeasure.Journey {
		s := 200
		return httpmeasure.Journey{Outcome: result.Completed, FinalURL: url, SelectedIP: ip.String(), Exchanges: []result.Exchange{{ID: "exchange-1", RequestedURL: url, HTTP: result.Measurement{ID: "exchange-1-http", Outcome: result.Completed}, ResponseStatus: &s}}}
	})
	e := &Engine{HTTP: http, Probe: a, Clock: clock}
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(30*time.Second))
	defer cancel()
	families, findings, err := e.Run(ctx, run.Input{TargetURL: "http://127.0.0.1/", MaxHops: 2}, func([]result.Family) {})
	if err != nil || families[0].Outcome != result.Completed || families[0].Paths[0].Target.Replied != 0 || families[0].Paths[0].Trace.Outcome != result.HopLimitReached || len(findings) != 1 || findings[0].Type != "http_response_observed" || findings[0].EvidenceRefs[0] != "exchange-1-http" || !strings.Contains(findings[0].Observation, "status 200") || strings.Contains(strings.ToLower(findings[0].Observation), "unreachable") {
		t.Fatalf("HTTP/ICMP disagreement: %+v %+v %v", families, findings, err)
	}
}

func TestCancelledPendingRouteRetainsManagerSlot(t *testing.T) {
	ip := netip.MustParseAddr("127.0.0.1")
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	t.Cleanup(func() { once.Do(func() { close(release) }) })
	a := probefake.New(time.Now, []probefake.RouteStep{{Observation: result.Route{Measurement: result.Measurement{Outcome: result.Completed}}, Finish: func() { close(entered); <-release }}}, nil)
	http := journeyFunc(func(_ context.Context, url string, _ *target.Selector) httpmeasure.Journey {
		status := 200
		return httpmeasure.Journey{Outcome: result.Completed, Exchanges: []result.Exchange{{ID: "exchange-1", RequestedURL: url, HTTP: result.Measurement{ID: "exchange-1-http", Outcome: result.Completed}, ResponseStatus: &status}}, FinalURL: url, SelectedIP: ip.String()}
	})
	m := run.NewManager(&Engine{HTTP: http, Probe: a})
	started, err := m.Start(run.Input{TargetURL: "http://127.0.0.1/", MaxHops: 4})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("route did not start")
	}
	stopping, err := m.Cancel(started.RunID)
	if err != nil || stopping.State != result.Stopping || stopping.EndedAt != nil {
		t.Fatalf("stopping: %+v %v", stopping, err)
	}
	if _, err := m.Start(run.Input{TargetURL: "http://127.0.0.1/", MaxHops: 4}); !errors.Is(err, run.ErrBusy) {
		t.Fatalf("slot released before native-style cleanup: %v", err)
	}
	once.Do(func() { close(release) })
	until := time.Now().Add(2 * time.Second)
	for time.Now().Before(until) {
		snapshot, err := m.Get(started.RunID)
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.State != result.RunCancelled {
			time.Sleep(time.Millisecond)
			continue
		}
		if len(snapshot.Families) != 2 || len(snapshot.Families[0].Paths) != 1 || snapshot.Families[0].Paths[0].Trace.Outcome != result.TraceCancelled || len(a.Calls()) != 1 || len(snapshot.Findings) != 0 {
			t.Fatalf("cancelled evidence: %+v", snapshot)
		}
		return
	}
	t.Fatal("cancelled run did not quiesce")
}

func TestTraceProgressPersistsThroughCancellation(t *testing.T) {
	ip := netip.MustParseAddr("127.0.0.1")
	onLink := true
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	t.Cleanup(func() { once.Do(func() { close(release) }) })
	echoes := make([]probefake.EchoStep, 7)
	for i := 0; i < 5; i++ {
		echoes[i] = probefake.EchoStep{Observation: result.Echo{Measurement: result.Measurement{Outcome: result.Timeout}, Status: result.NoReply}}
	}
	rtt := 3.0
	echoes[5] = probefake.EchoStep{Observation: result.Echo{Measurement: result.Measurement{Outcome: result.Completed}, Status: result.TimeExceeded, ReplyAddress: "127.0.0.2", RTTMS: &rtt}}
	echoes[6] = probefake.EchoStep{Observation: result.Echo{Measurement: result.Measurement{Outcome: result.Timeout}, Status: result.NoReply}, Finish: func() { close(entered); <-release }}
	a := probefake.New(time.Now, []probefake.RouteStep{{Observation: result.Route{Measurement: result.Measurement{Outcome: result.Completed}, OnLink: &onLink}}}, echoes)
	http := journeyFunc(func(_ context.Context, url string, _ *target.Selector) httpmeasure.Journey {
		s := 200
		return httpmeasure.Journey{Outcome: result.Completed, SelectedIP: ip.String(), FinalURL: url, Exchanges: []result.Exchange{{ID: "x", RequestedURL: url, HTTP: result.Measurement{ID: "http", Outcome: result.Completed}, ResponseStatus: &s}}}
	})
	m := run.NewManager(&Engine{HTTP: http, Probe: a})
	started, err := m.Start(run.Input{TargetURL: "http://127.0.0.1/", MaxHops: 4})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(4 * time.Second):
		t.Fatal("second trace hop did not start")
	}
	active, err := m.Get(started.RunID)
	if err != nil || active.State != result.Running || len(active.Families[0].Paths) != 1 {
		t.Fatalf("active snapshot: %+v %v", active, err)
	}
	p := active.Families[0].Paths[0]
	if p.Target.Attempted != 5 || p.Trace.Outcome != result.TracePending || len(p.Trace.Hops) != 1 || p.Trace.Hops[0].Status != result.TimeExceeded || len(active.Findings) != 0 {
		t.Fatalf("partial trace: %+v", active)
	}
	stopping, err := m.Cancel(started.RunID)
	if err != nil || stopping.State != result.Stopping || stopping.EndedAt != nil || len(stopping.Families[0].Paths[0].Trace.Hops) != 1 {
		t.Fatalf("stopping snapshot: %+v %v", stopping, err)
	}
	once.Do(func() { close(release) })
	until := time.Now().Add(2 * time.Second)
	for time.Now().Before(until) {
		terminal, err := m.Get(started.RunID)
		if err != nil {
			t.Fatal(err)
		}
		if !terminal.State.Terminal() {
			time.Sleep(time.Millisecond)
			continue
		}
		if terminal.State != result.RunCancelled || terminal.Families[0].Paths[0].Trace.Outcome != result.TraceCancelled || len(terminal.Families[0].Paths[0].Trace.Hops) != 2 || len(terminal.Findings) != 0 || len(a.Calls()) != 8 {
			t.Fatalf("terminal partial trace: %+v", terminal)
		}
		return
	}
	t.Fatal("cancelled trace did not quiesce")
}

func TestHTTPProgressVisibleBeforeCancellation(t *testing.T) {
	entered := make(chan struct{})
	a := probefake.New(time.Now, nil, nil)
	m := run.NewManager(&Engine{HTTP: pendingHTTP{entered: entered}, Probe: a})
	started, err := m.Start(run.Input{TargetURL: "http://example.com/", MaxHops: 16})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("HTTP selection did not publish")
	}
	active, err := m.Get(started.RunID)
	if err != nil || active.State != result.Running || len(active.Families) != 2 || len(active.Families[0].DNS) != 1 || active.Families[0].Outcome != result.Pending || len(active.Families[0].Exchanges) != 0 || len(active.Findings) != 0 {
		t.Fatalf("HTTP progress: %+v %v", active, err)
	}
	stopping, err := m.Cancel(started.RunID)
	if err != nil || stopping.State != result.Stopping || len(stopping.Families[0].DNS) != 1 {
		t.Fatalf("cancel snapshot: %+v %v", stopping, err)
	}
	until := time.Now().Add(2 * time.Second)
	for time.Now().Before(until) {
		terminal, err := m.Get(started.RunID)
		if err != nil {
			t.Fatal(err)
		}
		if !terminal.State.Terminal() {
			time.Sleep(time.Millisecond)
			continue
		}
		if terminal.State != result.RunCancelled || len(terminal.Families[0].DNS) != 1 || terminal.Families[0].Outcome != result.Cancelled || len(terminal.Findings) != 0 || len(a.Calls()) != 0 {
			t.Fatalf("terminal HTTP evidence: %+v", terminal)
		}
		return
	}
	t.Fatal("HTTP cancellation did not finish")
}
