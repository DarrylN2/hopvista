package main

import (
	"context"
	"encoding/json"
	"net/netip"
	"os"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

type loadedMarginSample struct {
	Destination      string          `json:"destination"`
	CapMS            int64           `json:"cap_ms"`
	BudgetMS         int64           `json:"budget_ms"`
	Attempt          int             `json:"attempt"`
	DeadlineAt       time.Time       `json:"deadline_at"`
	ReturnedAt       time.Time       `json:"returned_at"`
	CheckToCleanupMS float64         `json:"check_to_cleanup_ms"`
	PastDeadline     bool            `json:"past_deadline_at_return"`
	WorkBefore       uint64          `json:"work_iterations_before"`
	WorkAfter        uint64          `json:"work_iterations_after"`
	Echo             echoObservation `json:"echo"`
}

// This is a finite loaded sampling check. A safe post-setup refusal is
// recorded and retried so each cap/path gets three actual native calls.
// Neither a completed sample nor the sample maximum validates a worst case.
func TestLiveLoadedMarginSampling(t *testing.T) {
	if os.Getenv("HOPVISTA_LOADED_MARGIN_SAMPLING") != "1" {
		t.Skip("enable with run.ps1 -LoadedMarginSamplingCheck")
	}
	if environment()["elevated"] == true {
		t.Fatal("requires an unelevated process")
	}
	const callsPerCell = 3
	const attemptsPerCell = 5
	const setupHeadroom = 50 * time.Millisecond
	var work atomic.Uint64
	originalProcs := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(originalProcs)
	stop, ready, done := make(chan struct{}), make(chan struct{}, 2), make(chan struct{}, 2)
	for worker := 0; worker < 2; worker++ {
		go func(seed uint64) {
			x := seed
			first := true
			for {
				select {
				case <-stop:
					done <- struct{}{}
					return
				default:
				}
				for i := 0; i < 10000; i++ {
					x = x*1664525 + 1013904223
				}
				work.Add(10000)
				if first {
					ready <- struct{}{}
					first = false
				}
			}
		}(uint64(worker + 1))
	}
	<-ready
	<-ready
	startedLoadAt := time.Now()
	stoppedLoadAt := time.Time{}
	stopLoad := func() {
		if !stoppedLoadAt.IsZero() {
			return
		}
		close(stop)
		<-done
		<-done
		stoppedLoadAt = time.Now()
	}
	defer stopLoad()

	var samples []loadedMarginSample
	for _, raw := range []string{"192.0.2.1", "198.51.100.1", "203.0.113.1"} {
		ip := netip.MustParseAddr(raw)
		for _, cap := range []time.Duration{750 * time.Millisecond, 900 * time.Millisecond, time.Second} {
			started := 0
			for attempt := 1; attempt <= attemptsPerCell && started < callsPerCell; attempt++ {
				budget := cap + candidateStartMargin + setupHeadroom
				ctx, cancel := context.WithTimeout(context.Background(), budget)
				deadline, _ := ctx.Deadline()
				checked := time.Now()
				workBefore := work.Load()
				echo := candidateMarginProbe(ctx, "loaded-margin-sampling", ip, cap)
				returned := time.Now()
				workAfter := work.Load()
				cancel()
				sample := loadedMarginSample{
					Destination: raw, CapMS: cap.Milliseconds(), BudgetMS: budget.Milliseconds(), Attempt: attempt,
					DeadlineAt: deadline.UTC(), ReturnedAt: returned.UTC(), CheckToCleanupMS: milliseconds(returned.Sub(checked)),
					PastDeadline: !returned.Before(deadline), WorkBefore: workBefore, WorkAfter: workAfter, Echo: echo,
				}
				samples = append(samples, sample)
				if workAfter <= workBefore {
					t.Errorf("%s cap=%d attempt=%d: load workers made no progress", raw, cap.Milliseconds(), attempt)
				}
				switch echo.Outcome {
				case "not_started":
					if !echo.StartedAt.IsZero() || echo.TimeoutMS != 0 || (echo.CloseOK != nil && !*echo.CloseOK) {
						t.Errorf("%s cap=%d attempt=%d: invalid safe refusal: %+v", raw, cap.Milliseconds(), attempt, sample)
					}
				case "timeout":
					started++
					if echo.TimeoutMS != uint32(cap.Milliseconds()) || echo.NativeError != 11010 || echo.CloseOK == nil || !*echo.CloseOK || sample.PastDeadline {
						t.Errorf("%s cap=%d attempt=%d: full-cap call failed: %+v", raw, cap.Milliseconds(), attempt, sample)
					}
				default:
					t.Errorf("%s cap=%d attempt=%d: unexpected outcome: %+v", raw, cap.Milliseconds(), attempt, sample)
				}
			}
			if started != callsPerCell {
				t.Errorf("%s cap=%d: only %d of %d actual calls started in %d attempts", raw, cap.Milliseconds(), started, callsPerCell, attemptsPerCell)
			}
		}
	}
	stopLoad()
	if err := os.MkdirAll(".local", 0700); err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(struct {
		CapturedAt        time.Time            `json:"captured_at"`
		Environment       map[string]any       `json:"environment"`
		CandidateMarginMS int64                `json:"candidate_margin_ms"`
		SetupHeadroomMS   int64                `json:"fixture_setup_headroom_ms"`
		OriginalProcs     int                  `json:"original_gomaxprocs"`
		TestProcs         int                  `json:"test_gomaxprocs"`
		Workers           int                  `json:"load_workers"`
		LoadStartedAt     time.Time            `json:"load_started_at"`
		LoadStoppedAt     time.Time            `json:"load_stopped_at"`
		WorkIterations    uint64               `json:"work_iterations"`
		Samples           []loadedMarginSample `json:"samples"`
		Gate              string               `json:"gate"`
	}{time.Now().UTC(), environment(), candidateStartMargin.Milliseconds(), setupHeadroom.Milliseconds(), originalProcs, 1, 2, startedLoadAt.UTC(), stoppedLoadAt.UTC(), work.Load(), samples, "incomplete; finite loaded samples do not validate a production margin or worst-case bound"}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".local/loaded-margin-sampling.json", data, 0600); err != nil {
		t.Fatal(err)
	}
}
