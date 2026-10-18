package run_test

import (
	"context"
	"errors"
	"net/netip"
	"sync"
	"testing"
	"time"

	"hopvista/internal/probe"
	"hopvista/internal/probe/probefake"
	"hopvista/internal/result"
	"hopvista/internal/run"
)

type executorFunc func(context.Context, run.Input, func([]result.Family)) ([]result.Family, []result.Finding, error)

func (f executorFunc) Run(ctx context.Context, in run.Input, publish func([]result.Family)) ([]result.Family, []result.Finding, error) {
	return f(ctx, in, publish)
}

func family(outcome result.Outcome) result.Family {
	return result.Family{Family: "ipv4", Outcome: outcome, DNS: []result.DNS{}, Exchanges: []result.Exchange{}, Paths: []result.Path{}}
}

func waitState(t *testing.T, m *run.Manager, id string, want result.State) result.Run {
	t.Helper()
	until := time.After(2 * time.Second)
	for {
		got, err := m.Get(id)
		if err != nil {
			t.Fatal(err)
		}
		if got.State == want {
			return got
		}
		select {
		case <-until:
			t.Fatalf("state stayed %s, wanted %s", got.State, want)
		case <-time.After(time.Millisecond):
		}
	}
}

func TestPendingFakeRouteKeepsBusySlotUntilCleanup(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	t.Cleanup(func() { once.Do(func() { close(release) }) })
	now := func() time.Time { return time.Now().UTC() }
	adapter := probefake.New(now, []probefake.RouteStep{{Observation: result.Route{Measurement: result.Measurement{Outcome: result.Completed}}, Finish: func() { close(entered); <-release }}}, nil)
	executor := executorFunc(func(ctx context.Context, _ run.Input, publish func([]result.Family)) ([]result.Family, []result.Finding, error) {
		partial := []result.Family{family(result.Completed)}
		partial[0].DNS = []result.DNS{{Measurement: result.Measurement{ID: "dns-1", Outcome: result.Completed, Method: "fake_resolver"}, Hostname: "example.com", Addresses: []string{"192.0.2.10"}, SelectedIP: "192.0.2.10"}}
		publish(partial)
		ip := netip.MustParseAddr("192.0.2.10")
		route := adapter.Route(ctx, probe.RouteRequest{ID: "route-1", SelectedIP: ip, DeadlineAt: time.Now().Add(time.Minute)})
		partial[0].Paths = []result.Path{{ID: "path-1", SelectedIP: ip.String(), Route: route}}
		return partial, []result.Finding{{Type: "test", Family: "ipv4", Observation: "Finished.", Limit: "Fake evidence only.", EvidenceRefs: []string{"route-1"}}}, nil
	})
	m := run.NewManager(executor)
	first, err := m.Start(run.Input{TargetURL: "https://example.com/", MaxHops: 30})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("route did not enter")
	}
	if first.State != result.Running || first.SchemaVersion != 1 || first.MaxHops != 30 || first.ConnectionMode != "direct" {
		t.Fatalf("start: %+v", first)
	}
	if _, err := m.Start(run.Input{TargetURL: "https://example.net/", MaxHops: 64}); !errors.Is(err, run.ErrBusy) {
		t.Fatalf("expected busy, got %v", err)
	}
	stopping, err := m.Cancel(first.RunID)
	if err != nil {
		t.Fatal(err)
	}
	if stopping.State != result.Stopping || stopping.StopRequestedAt == nil || stopping.EndedAt != nil || len(stopping.Findings) != 0 || len(stopping.Families) != 1 {
		t.Fatalf("stopping snapshot: %+v", stopping)
	}
	stopping.Families[0].DNS[0].SelectedIP = "mutated"
	again, err := m.Get(first.RunID)
	if err != nil || again.Families[0].DNS[0].SelectedIP != "192.0.2.10" {
		t.Fatal("snapshot shares internal storage")
	}
	if _, err := m.Start(run.Input{TargetURL: "https://example.net/", MaxHops: 64}); !errors.Is(err, run.ErrBusy) {
		t.Fatal("slot released during stopping")
	}
	if len(adapter.Calls()) != 1 {
		t.Fatal("fake route was not started")
	}
	once.Do(func() { close(release) })
	terminal := waitState(t, m, first.RunID, result.RunCancelled)
	if terminal.EndedAt == nil || terminal.StopRequestedAt == nil || terminal.EndedAt.Before(*terminal.StopRequestedAt) || len(terminal.Findings) != 0 || len(terminal.Families[0].Paths) != 1 {
		t.Fatalf("terminal: %+v", terminal)
	}
	if repeat, err := m.Cancel(first.RunID); err != nil || repeat.State != result.RunCancelled {
		t.Fatal("terminal cancellation was not idempotent")
	}
	second, err := m.Start(run.Input{TargetURL: "https://example.net/", MaxHops: 64})
	if err != nil || second.RunID == first.RunID || second.MaxHops != 64 {
		t.Fatalf("restart: %+v %v", second, err)
	}
	if _, err := m.Get(first.RunID); !errors.Is(err, run.ErrNotFound) {
		t.Fatal("restart did not evict old result")
	}
	_, _ = m.Cancel(second.RunID)
}

func TestCutoffRemainsStoppingUntilExecutorReturns(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	t.Cleanup(func() { once.Do(func() { close(release) }) })
	m := run.NewManagerWithWindow(executorFunc(func(ctx context.Context, _ run.Input, _ func([]result.Family)) ([]result.Family, []result.Finding, error) {
		close(entered)
		<-release
		return []result.Family{family(result.Deadline)}, nil, nil
	}), 20*time.Millisecond)
	first, err := m.Start(run.Input{TargetURL: "http://example.com/", MaxHops: 1})
	if err != nil {
		t.Fatal(err)
	}
	<-entered
	stopping := waitState(t, m, first.RunID, result.Stopping)
	if stopping.StopRequestedAt == nil || stopping.EndedAt != nil {
		t.Fatal("cutoff did not retain pending run")
	}
	if _, err := m.Start(run.Input{TargetURL: "http://example.com/", MaxHops: 30}); !errors.Is(err, run.ErrBusy) {
		t.Fatal("cutoff released active slot")
	}
	once.Do(func() { close(release) })
	terminal := waitState(t, m, first.RunID, result.RunCompleted)
	if terminal.EndedAt == nil || terminal.StopRequestedAt == nil || terminal.Families[0].Outcome != result.Deadline {
		t.Fatalf("cutoff result: %+v", terminal)
	}
}

func TestCompletedFindingsOnlyAfterQuiescenceAndFailureSemantics(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	t.Cleanup(func() { once.Do(func() { close(release) }) })
	m := run.NewManager(executorFunc(func(_ context.Context, _ run.Input, _ func([]result.Family)) ([]result.Family, []result.Finding, error) {
		close(entered)
		<-release
		return []result.Family{family(result.Completed)}, []result.Finding{{Type: "observed", Family: "ipv4", Observation: "A response arrived.", Limit: "Only this run.", EvidenceRefs: []string{"http-1"}}}, nil
	}))
	r, err := m.Start(run.Input{TargetURL: "http://example.com/", MaxHops: 30})
	if err != nil {
		t.Fatal(err)
	}
	<-entered
	current, _ := m.Get(r.RunID)
	if len(current.Findings) != 0 || current.EndedAt != nil {
		t.Fatal("finding leaked before quiescence")
	}
	once.Do(func() { close(release) })
	ended := waitState(t, m, r.RunID, result.RunCompleted)
	if len(ended.Findings) != 1 || ended.EndedAt == nil {
		t.Fatal("terminal finding missing")
	}
	fail := run.NewManager(executorFunc(func(_ context.Context, _ run.Input, _ func([]result.Family)) ([]result.Family, []result.Finding, error) {
		return nil, nil, errors.New("internal test detail")
	}))
	f, err := fail.Start(run.Input{TargetURL: "http://example.com/", MaxHops: 30})
	if err != nil {
		t.Fatal(err)
	}
	failed := waitState(t, fail, f.RunID, result.Failed)
	if failed.Error == nil || failed.Error.Code != "orchestration_failed" || failed.Error.Message == "internal test detail" {
		t.Fatal("internal error leaked")
	}
}

func TestPlaceholderHonestOutcome(t *testing.T) {
	m := run.NewManager(run.Placeholder{})
	r, err := m.Start(run.Input{TargetURL: "https://example.com/", MaxHops: 30})
	if err != nil {
		t.Fatal(err)
	}
	ended := waitState(t, m, r.RunID, result.RunCompleted)
	if len(ended.Families) != 1 || ended.Families[0].Outcome != result.Unavailable || ended.Families[0].Error.Code != "measurement_not_connected" || len(ended.Findings) != 0 {
		t.Fatalf("placeholder invented evidence: %+v", ended)
	}
}
