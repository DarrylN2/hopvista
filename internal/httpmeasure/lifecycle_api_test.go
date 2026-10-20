package httpmeasure_test

import (
	"sync"
	"testing"
	"time"

	"hopvista/internal/engine"
	"hopvista/internal/httpmeasure"
	"hopvista/internal/probe/probefake"
	"hopvista/internal/result"
	"hopvista/internal/run"
)

// TLS/HTTP and the engine/API are real. The route Finish hook models a
// synchronous call owning resources until explicitly released; it is not an
// observed Windows GetBestRoute2 stall. Only the deadline case shortens the
// manager window, through its existing test-only constructor.
func TestHeldRouteStopThroughAPI(t *testing.T) {
	for _, cancelled := range []bool{true, false} {
		name := "cutoff"
		if cancelled {
			name = "cancelroute"
		}
		t.Run(name, func(t *testing.T) {
			f := newTLSJourney(t, noContent)
			entered, release := make(chan struct{}), make(chan struct{})
			var once sync.Once
			t.Cleanup(func() { once.Do(func() { close(release) }) })
			onLink := true
			adapter := probefake.New(time.Now, []probefake.RouteStep{{
				Observation: result.Route{Measurement: result.Measurement{Outcome: result.Completed}, OnLink: &onLink},
				Finish:      func() { close(entered); <-release },
			}}, nil)
			e := &engine.Engine{HTTP: httpmeasure.ControlledTLSClient(f.roots, f.dial), Resolver: lookupFunc(f.resolve), Probe: adapter}
			manager := run.NewManager(e)
			if !cancelled {
				manager = run.NewManagerWithWindow(e, 500*time.Millisecond)
			}
			didCancel, sawStopping := false, false
			snapshot := runEngineThroughAPI(t, manager, "https://start.example/retained", func(s result.Run, cancel func(), assertBusy func()) {
				select {
				case <-entered:
				default:
					return
				}
				if s.State == result.Running && cancelled && !didCancel {
					didCancel = true
					cancel()
					return
				}
				if s.State != result.Stopping {
					return
				}
				if s.StopRequestedAt == nil || s.EndedAt != nil || len(s.Findings) != 0 || s.MaxHops != 1 || len(s.Families[0].Exchanges) != 1 || s.Families[0].Exchanges[0].HTTP.Outcome != result.Completed {
					t.Fatalf("lost retained evidence/ownership: %+v", s)
				}
				if len(adapter.Calls()) != 1 {
					t.Fatal("late route authorized new echo")
				}
				writeCapture(t, "run-"+name+"stopping", s)
				// Repeated cancel is idempotent; after deadline it must not
				// replace the cutoff outcome. Check busy before releasing ownership.
				cancel()
				assertBusy()
				if sawStopping {
					once.Do(func() { close(release) })
				}
				sawStopping = true
			})
			want, outcome := result.RunCompleted, result.Deadline
			if cancelled {
				want, outcome = result.RunCancelled, result.Cancelled
			}
			if !sawStopping || snapshot.State != want || snapshot.StopRequestedAt == nil || snapshot.EndedAt == nil || len(snapshot.Families[0].Paths) != 1 {
				t.Fatalf("lifecycle: %+v", snapshot)
			}
			p := snapshot.Families[0].Paths[0]
			if p.Route.Outcome != result.Completed || p.Route.StartedAt == nil || p.Route.EndedAt == nil || p.Route.StartedAt.After(*snapshot.StopRequestedAt) || p.Route.EndedAt.Before(*snapshot.StopRequestedAt) || p.Route.EndedAt.After(*snapshot.EndedAt) || p.Target.Outcome != outcome || len(p.Target.Samples) != 0 || len(p.Trace.Hops) != 0 || snapshot.Families[0].Outcome != outcome || len(adapter.Calls()) != 1 {
				t.Fatalf("late route or invented send: %+v", p)
			}
			writeCapture(t, "run-"+name, snapshot)
		})
	}
}
