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

type minuteProbeObservation struct {
	ScheduledAt     time.Time       `json:"scheduled_at"`
	RouteReturnCode uint32          `json:"route_return_code"`
	RouteElapsedMS  float64         `json:"route_elapsed_ms"`
	CheckedAt       time.Time       `json:"checked_at"`
	ReturnedAt      time.Time       `json:"returned_at"`
	Echo            echoObservation `json:"echo"`
}

type minuteRoutePhaseObservation struct {
	ScheduledAt      time.Time       `json:"scheduled_at"`
	StartedAt        time.Time       `json:"started_at"`
	RouteReturnCode  uint32          `json:"route_return_code"`
	RouteElapsedMS   float64         `json:"route_elapsed_ms"`
	PhaseReturnedAt  time.Time       `json:"phase_returned_at"`
	PostPhaseCheckAt time.Time       `json:"post_phase_check_at"`
	Echo             echoObservation `json:"echo"`
}

// A one-minute harness fixture, not application run orchestration. The busy
// Go workers create process-local scheduler contention during the final five
// seconds. The route-phase delay after GetBestRoute2 is artificial.
func TestLiveMinuteCutoffUnderSchedulerLoad(t *testing.T) {
	if os.Getenv("HOPVISTA_MINUTE_CUTOFF") != "1" {
		t.Skip("enable with run.ps1 -MinuteCutoffCheck")
	}
	if environment()["elevated"] == true {
		t.Fatal("requires an unelevated process")
	}
	const runLength = 60 * time.Second
	const cap = time.Second
	ip := netip.MustParseAddr("192.0.2.1")
	started := time.Now()
	deadline := started.Add(runLength)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	stopObserved := make(chan time.Time, 1)
	go func() {
		<-ctx.Done()
		stopObserved <- time.Now()
	}()

	// Constrain this test process to one Go scheduler processor and run two
	// CPU-bound workers in the final five seconds. No system-wide setting changes.
	originalProcs := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(originalProcs)
	loadStop := make(chan struct{})
	loadDone := make(chan struct{}, 2)
	var workIterations atomic.Uint64
	loadScheduled := started.Add(55 * time.Second)
	routeScheduled := started.Add(59*time.Second + 500*time.Millisecond)
	probeScheduled := started.Add(58*time.Second + 700*time.Millisecond)
	loadStarted := time.Time{}
	sleepUntil(loadScheduled)
	loadStarted = time.Now()
	for worker := 0; worker < 2; worker++ {
		go func(seed uint64) {
			x := seed
			for {
				select {
				case <-loadStop:
					loadDone <- struct{}{}
					return
				default:
				}
				for i := 0; i < 10000; i++ {
					x = x*1664525 + 1013904223
				}
				workIterations.Add(10000)
			}
		}(uint64(worker + 1))
	}

	routePhaseDone := make(chan minuteRoutePhaseObservation, 1)
	go func() {
		sleepUntil(routeScheduled)
		item := minuteRoutePhaseObservation{ScheduledAt: routeScheduled.UTC(), StartedAt: time.Now().UTC()}
		if ctx.Err() == nil {
			r := route(ip)
			item.RouteReturnCode, item.RouteElapsedMS = r.ReturnCode, r.ElapsedMS
		}
		// Hold this already-started phase past the real 60-second cutoff.
		sleepUntil(deadline.Add(30 * time.Millisecond))
		item.PhaseReturnedAt = time.Now().UTC()
		item.PostPhaseCheckAt = time.Now().UTC()
		item.Echo = candidateMarginProbe(ctx, "after-minute-route-phase", ip, cap)
		routePhaseDone <- item
	}()

	sleepUntil(probeScheduled)
	firstRoute := routeObservation{}
	if ctx.Err() == nil {
		firstRoute = route(ip)
	} else {
		firstRoute.Error = ctx.Err().Error()
	}
	probe := minuteProbeObservation{ScheduledAt: probeScheduled.UTC(), RouteReturnCode: firstRoute.ReturnCode, RouteElapsedMS: firstRoute.ElapsedMS, CheckedAt: time.Now().UTC()}
	if firstRoute.ReturnCode == 0 && firstRoute.Error == "" {
		probe.Echo = candidateMarginProbe(ctx, "minute-loaded-full-cap", ip, cap)
	}
	probe.ReturnedAt = time.Now().UTC()
	// A subsequent attempt must be refused; this is a guard check, not new
	// native measurement work.
	sleepUntil(started.Add(59 * time.Second))
	lateCheckedAt := time.Now()
	late := candidateMarginProbe(ctx, "minute-late-refusal", ip, cap)
	lateReturnedAt := time.Now()
	phase := <-routePhaseDone
	stopAt := <-stopObserved
	close(loadStop)
	<-loadDone
	<-loadDone
	loadStopped := time.Now()
	quiescentAt := time.Now()

	if firstRoute.ReturnCode != 0 || firstRoute.Error != "" {
		t.Errorf("pre-probe route unavailable: code=%d error=%s", firstRoute.ReturnCode, firstRoute.Error)
	}
	if probe.Echo.TimeoutMS != uint32(cap.Milliseconds()) || probe.Echo.Outcome != "timeout" || probe.Echo.NativeError != 11010 || probe.Echo.CloseOK == nil || !*probe.Echo.CloseOK {
		t.Errorf("loaded full-cap probe did not complete as expected: %+v", probe.Echo)
	}
	if late.Outcome != "not_started" || !late.StartedAt.IsZero() || late.CloseOK != nil {
		t.Errorf("late probe entered native work: %+v", late)
	}
	if phase.StartedAt.After(deadline) || phase.RouteReturnCode != 0 || phase.PhaseReturnedAt.Before(deadline) || phase.Echo.Outcome != "not_started" || !phase.Echo.StartedAt.IsZero() || phase.Echo.CloseOK != nil {
		t.Errorf("route phase did not stop cleanly across cutoff: %+v", phase)
	}
	if stopAt.After(quiescentAt) || loadStarted.After(probe.CheckedAt) || workIterations.Load() == 0 {
		t.Error("cutoff or scheduler-load ordering was not observed")
	}
	if err := os.MkdirAll(".local", 0700); err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(struct {
		CapturedAt         time.Time                   `json:"captured_at"`
		Environment        map[string]any              `json:"environment"`
		CandidateMarginMS  int64                       `json:"candidate_margin_ms"`
		RunStartedAt       time.Time                   `json:"run_started_at"`
		DeadlineAt         time.Time                   `json:"deadline_at"`
		StopObservedAt     time.Time                   `json:"stop_observed_at"`
		QuiescentAt        time.Time                   `json:"quiescent_at"`
		StopToQuiescentMS  float64                     `json:"stop_to_quiescent_ms"`
		OriginalGOMAXPROCS int                         `json:"original_gomaxprocs"`
		TestGOMAXPROCS     int                         `json:"test_gomaxprocs"`
		LoadWorkers        int                         `json:"load_workers"`
		LoadStartedAt      time.Time                   `json:"load_started_at"`
		LoadStoppedAt      time.Time                   `json:"load_stopped_at"`
		WorkIterations     uint64                      `json:"work_iterations"`
		Probe              minuteProbeObservation      `json:"probe"`
		LateCheckedAt      time.Time                   `json:"late_checked_at"`
		LateReturnedAt     time.Time                   `json:"late_returned_at"`
		LateProbe          echoObservation             `json:"late_probe"`
		RoutePhase         minuteRoutePhaseObservation `json:"route_phase"`
		Gate               string                      `json:"gate"`
	}{time.Now().UTC(), environment(), candidateStartMargin.Milliseconds(), started.UTC(), deadline.UTC(), stopAt.UTC(), quiescentAt.UTC(), milliseconds(quiescentAt.Sub(stopAt)), originalProcs, 1, 2, loadStarted.UTC(), loadStopped.UTC(), workIterations.Load(), probe, lateCheckedAt.UTC(), lateReturnedAt.UTC(), late, phase, "incomplete; one loaded 60-second local fixture cannot establish a production or worst-case bound"}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".local/minute-cutoff.json", data, 0600); err != nil {
		t.Fatal(err)
	}
	t.Logf("60-second fixture: probe=%s late=%s route-phase=%s stop-to-quiescent=%.2fms iterations=%d", probe.Echo.Outcome, late.Outcome, phase.Echo.Outcome, milliseconds(quiescentAt.Sub(stopAt)), workIterations.Load())
}

func sleepUntil(at time.Time) {
	if remaining := time.Until(at); remaining > 0 {
		time.Sleep(remaining)
	}
}
