package run_test

import (
	"context"
	"errors"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"hopvista/internal/result"
	"hopvista/internal/run"
)

// This opt-in wall-clock check uses the production 60-second window. The held
// executor models a pending synchronous route; it is not native stall evidence.
func TestLiveMinuteCutoffUnderProcessLoad(t *testing.T) {
	if os.Getenv("HOPVISTA_LIVE_MINUTE_CUTOFF") != "1" {
		t.Skip("set HOPVISTA_LIVE_MINUTE_CUTOFF=1 for the real-minute cutoff check")
	}
	entered := make(chan struct{})
	contextStopped := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }) })
	m := run.NewManager(executorFunc(func(ctx context.Context, _ run.Input, _ func([]result.Family)) ([]result.Family, []result.Finding, error) {
		close(entered)
		<-ctx.Done()
		close(contextStopped)
		<-release
		return []result.Family{family(result.Deadline)}, nil, nil
	}))
	started, err := m.Start(run.Input{TargetURL: "http://example.com/", MaxHops: 64})
	if err != nil {
		t.Fatal(err)
	}
	<-entered
	if got := started.DeadlineAt.Sub(started.StartedAt); got != run.MeasurementWindow {
		t.Fatalf("production cutoff window = %s", got)
	}
	loadAt := started.DeadlineAt.Add(-5 * time.Second)
	if wait := time.Until(loadAt); wait > 0 {
		time.Sleep(wait)
	}
	before, err := m.Get(started.RunID)
	if err != nil || before.State != result.Running || before.StopRequestedAt != nil {
		t.Fatalf("run stopped before the final five seconds: %+v %v", before, err)
	}
	previousProcs := runtime.GOMAXPROCS(1)
	workerStop := make(chan struct{})
	var workers sync.WaitGroup
	var progress [2]atomic.Uint64
	for i := 0; i < 2; i++ {
		index := i
		workers.Add(1)
		go func() {
			defer workers.Done()
			for {
				select {
				case <-workerStop:
					return
				default:
				}
				for j := 0; j < 1000; j++ {
				}
				progress[index].Add(1)
			}
		}()
	}
	defer func() {
		close(workerStop)
		workers.Wait()
		runtime.GOMAXPROCS(previousProcs)
	}()
	<-contextStopped
	stopping := waitState(t, m, started.RunID, result.Stopping)
	if progress[0].Load() == 0 || progress[1].Load() == 0 || stopping.StopRequestedAt == nil || stopping.EndedAt != nil || len(stopping.Findings) != 0 {
		t.Fatalf("loaded cutoff did not retain a pending run: %+v worker progress=%d,%d", stopping, progress[0].Load(), progress[1].Load())
	}
	if lag := stopping.StopRequestedAt.Sub(started.DeadlineAt); lag < 0 || lag > 2*time.Second {
		t.Fatalf("stop observed outside the reviewed local window: %s", lag)
	}
	if _, err := m.Start(run.Input{TargetURL: "http://example.net/", MaxHops: 1}); !errors.Is(err, run.ErrBusy) {
		t.Fatalf("pending executor released the busy slot: %v", err)
	}
	releaseOnce.Do(func() { close(release) })
	terminal := waitState(t, m, started.RunID, result.RunCompleted)
	if terminal.EndedAt == nil || terminal.StopRequestedAt == nil || terminal.EndedAt.Before(*terminal.StopRequestedAt) || terminal.Families[0].Outcome != result.Deadline {
		t.Fatalf("terminal cutoff result: %+v", terminal)
	}
	t.Logf("real 60-second cutoff observed %s after deadline; held executor quiesced %s after stop", stopping.StopRequestedAt.Sub(started.DeadlineAt), terminal.EndedAt.Sub(*terminal.StopRequestedAt))
}
