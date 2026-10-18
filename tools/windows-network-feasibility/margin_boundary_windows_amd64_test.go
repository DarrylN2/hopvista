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

const candidateStartMargin = 250 * time.Millisecond

type marginBoundaryObservation struct {
	Destination       string          `json:"destination"`
	CapMS             int64           `json:"cap_ms"`
	BudgetMS          int64           `json:"budget_ms"`
	Attempt           int             `json:"attempt"`
	DeadlineAt        time.Time       `json:"deadline_at"`
	ReturnedAt        time.Time       `json:"returned_at"`
	RemainingAtSendMS float64         `json:"remaining_at_send_ms,omitempty"`
	CheckToCleanupMS  float64         `json:"check_to_cleanup_ms"`
	PastDeadline      bool            `json:"past_deadline_at_return"`
	Echo              echoObservation `json:"echo"`
	WorkBefore        uint64          `json:"work_iterations_before,omitempty"`
	WorkAfter         uint64          `json:"work_iterations_after,omitempty"`
}

type boundaryLoadObservation struct {
	OriginalProcs int       `json:"original_gomaxprocs"`
	TestProcs     int       `json:"test_gomaxprocs"`
	Workers       int       `json:"load_workers"`
	StartedAt     time.Time `json:"load_started_at"`
	StoppedAt     time.Time `json:"load_stopped_at"`
	Iterations    uint64    `json:"work_iterations"`
}

// Candidate-margin experiment only. This does not select a product policy or
// establish a worst-case native completion bound.
func TestLiveCandidateMarginBoundary(t *testing.T) {
	if os.Getenv("HOPVISTA_MARGIN_BOUNDARY") != "1" {
		t.Skip("enable with run.ps1 -MarginBoundaryCheck")
	}
	if environment()["elevated"] == true {
		t.Fatal("requires an unelevated process")
	}
	var load *boundaryLoadObservation
	var work atomic.Uint64
	var stopLoad func()
	output := ".local/margin-boundary.json"
	if os.Getenv("HOPVISTA_MARGIN_LOAD") == "1" {
		output = ".local/loaded-margin-boundary.json"
		load = &boundaryLoadObservation{OriginalProcs: runtime.GOMAXPROCS(1), TestProcs: 1, Workers: 2, StartedAt: time.Now().UTC()}
		stop, ready, done := make(chan struct{}), make(chan struct{}, 2), make(chan struct{}, 2)
		for worker := 0; worker < load.Workers; worker++ {
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
		// Run this before writing the capture, and also on an early test failure.
		defer func() { runtime.GOMAXPROCS(load.OriginalProcs) }()
		stopLoad = func() {
			if !load.StoppedAt.IsZero() {
				return
			}
			close(stop)
			<-done
			<-done
			load.StoppedAt = time.Now().UTC()
			load.Iterations = work.Load()
		}
		defer stopLoad()
	}
	var observations []marginBoundaryObservation
	for _, cap := range []time.Duration{750 * time.Millisecond, 900 * time.Millisecond, time.Second} {
		// One millisecond below the proposed threshold must refuse before handle creation.
		budget := cap + candidateStartMargin - time.Millisecond
		ctx, cancel := context.WithTimeout(context.Background(), budget)
		deadline, _ := ctx.Deadline()
		checked := time.Now()
		refused := candidateMarginProbe(ctx, "margin-boundary-refusal", netip.MustParseAddr("192.0.2.1"), cap)
		returned := time.Now()
		cancel()
		observations = append(observations, marginBoundaryObservation{
			Destination: "192.0.2.1", CapMS: cap.Milliseconds(), BudgetMS: budget.Milliseconds(),
			DeadlineAt: deadline.UTC(), ReturnedAt: returned.UTC(), CheckToCleanupMS: milliseconds(returned.Sub(checked)),
			PastDeadline: !returned.Before(deadline), Echo: refused,
		})
		if refused.Outcome != "not_started" || !refused.StartedAt.IsZero() || refused.CloseOK != nil || refused.TimeoutMS != 0 {
			t.Errorf("cap=%dms: candidate boundary did not refuse before handle creation: %+v", cap.Milliseconds(), refused)
		}
	}
	for _, raw := range []string{"192.0.2.1", "198.51.100.1", "203.0.113.1"} {
		ip := netip.MustParseAddr(raw)
		for _, cap := range []time.Duration{750 * time.Millisecond, 900 * time.Millisecond, time.Second} {
			for attempt := 1; attempt <= 3; attempt++ {
				// The extra 20 ms leaves a small setup window before the exact guard.
				budget := cap + candidateStartMargin + 20*time.Millisecond
				ctx, cancel := context.WithTimeout(context.Background(), budget)
				deadline, _ := ctx.Deadline()
				checked := time.Now()
				workBefore := work.Load()
				o := candidateMarginProbe(ctx, "margin-boundary-full-cap", ip, cap)
				returned := time.Now()
				workAfter := work.Load()
				cancel()
				item := marginBoundaryObservation{
					Destination: raw, CapMS: cap.Milliseconds(), BudgetMS: budget.Milliseconds(), Attempt: attempt,
					DeadlineAt: deadline.UTC(), ReturnedAt: returned.UTC(), CheckToCleanupMS: milliseconds(returned.Sub(checked)),
					PastDeadline: !returned.Before(deadline), Echo: o,
					WorkBefore: workBefore, WorkAfter: workAfter,
				}
				if !o.StartedAt.IsZero() {
					item.RemainingAtSendMS = milliseconds(deadline.Sub(o.StartedAt))
				}
				observations = append(observations, item)
				if load != nil && workAfter <= workBefore {
					t.Errorf("target=%s cap=%dms attempt=%d: no load progress during probe", raw, cap.Milliseconds(), attempt)
				}
				if o.TimeoutMS != uint32(cap.Milliseconds()) || o.Outcome != "timeout" || o.NativeError != 11010 || o.CloseOK == nil || !*o.CloseOK || item.PastDeadline {
					t.Errorf("target=%s cap=%dms attempt=%d: candidate full-cap boundary failed: %+v", raw, cap.Milliseconds(), attempt, item)
				}
			}
		}
	}
	if stopLoad != nil {
		stopLoad()
	}
	if err := os.MkdirAll(".local", 0700); err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(struct {
		CapturedAt        time.Time                   `json:"captured_at"`
		Environment       map[string]any              `json:"environment"`
		CandidateMarginMS int64                       `json:"candidate_margin_ms"`
		Observations      []marginBoundaryObservation `json:"observations"`
		Gate              string                      `json:"gate"`
		Load              *boundaryLoadObservation    `json:"scheduler_load,omitempty"`
	}{time.Now().UTC(), environment(), candidateStartMargin.Milliseconds(), observations, "incomplete; candidate margin tested only in local fixture, not a worst-case bound", load}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(output, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func candidateMarginProbe(ctx context.Context, label string, ip netip.Addr, cap time.Duration) echoObservation {
	if !fullBudgetAvailable(ctx, cap, candidateStartMargin, time.Now()) {
		return echoObservation{Case: label, Destination: ip.String(), Family: "ipv4", Scope: "documentation", API: "IcmpSendEcho2 (not called)", HopLimit: 64, Outcome: "not_started", Error: "insufficient time for full native probe budget and candidate margin"}
	}
	return echoInternal(ctx, label, ip, 64, cap, false, 512, nil, candidateStartMargin, nil)
}
