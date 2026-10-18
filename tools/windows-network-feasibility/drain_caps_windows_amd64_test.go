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

type drainCapObservation struct {
	Destination     string          `json:"destination"`
	CapMS           int64           `json:"cap_ms"`
	StopRequestedAt time.Time       `json:"stop_requested_at"`
	QuiescentAt     time.Time       `json:"quiescent_at"`
	PendingAtStop   bool            `json:"pending_at_stop"`
	DrainMS         float64         `json:"drain_ms"`
	WorkAtStop      uint64          `json:"work_iterations_at_stop,omitempty"`
	WorkAtQuiescent uint64          `json:"work_iterations_at_quiescent,omitempty"`
	Echo            echoObservation `json:"echo"`
}

type drainLoadObservation struct {
	OriginalProcs int       `json:"original_gomaxprocs"`
	TestProcs     int       `json:"test_gomaxprocs"`
	Workers       int       `json:"load_workers"`
	StartedAt     time.Time `json:"load_started_at"`
	StoppedAt     time.Time `json:"load_stopped_at"`
	Iterations    uint64    `json:"work_iterations"`
}

// Opt-in finite sampling of cancellation drain after a full-cap native send.
// Context cancellation stops later scheduling; it does not cancel this call.
func TestLiveIPv4DrainCaps(t *testing.T) {
	if os.Getenv("HOPVISTA_DRAIN_CAPS") != "1" {
		t.Skip("enable with run.ps1 -DrainCapsCheck")
	}
	if environment()["elevated"] == true {
		t.Fatal("requires an unelevated process")
	}
	var load *drainLoadObservation
	var work atomic.Uint64
	var stopLoad func()
	output := ".local/drain-caps.json"
	if os.Getenv("HOPVISTA_DRAIN_LOAD") == "1" {
		output = ".local/loaded-drain-caps.json"
		load = &drainLoadObservation{OriginalProcs: runtime.GOMAXPROCS(1), TestProcs: 1, Workers: 2, StartedAt: time.Now().UTC()}
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
	var observations []drainCapObservation
	var failures []string
	for _, raw := range []string{"192.0.2.1", "198.51.100.1", "203.0.113.1"} {
		ip := netip.MustParseAddr(raw)
		for _, cap := range []time.Duration{750 * time.Millisecond, 900 * time.Millisecond, time.Second} {
			ctx, cancel := context.WithCancel(context.Background())
			started := make(chan struct{})
			done := make(chan echoObservation, 1)
			go func() {
				done <- echo(ctx, "cancel-pending-v4", ip, 64, cap, false, 512, func() { close(started) })
			}()
			var o echoObservation
			finished := false
			select {
			case <-started:
				time.Sleep(25 * time.Millisecond)
			case o = <-done:
				finished = true
				failures = append(failures, raw+": native send did not remain pending for stop request")
			}
			stopped := time.Now()
			workAtStop := work.Load()
			cancel()
			if !finished {
				o = <-done
			}
			quiescent := time.Now()
			workAtQuiescent := work.Load()
			item := drainCapObservation{
				Destination: raw, CapMS: cap.Milliseconds(), StopRequestedAt: stopped.UTC(), QuiescentAt: quiescent.UTC(),
				PendingAtStop: !o.EndedAt.IsZero() && o.EndedAt.After(stopped), DrainMS: milliseconds(quiescent.Sub(stopped)),
				WorkAtStop: workAtStop, WorkAtQuiescent: workAtQuiescent, Echo: o,
			}
			observations = append(observations, item)
			if load != nil && workAtQuiescent <= workAtStop {
				failures = append(failures, raw+": load workers made no progress during pending drain")
			}
			t.Logf("target=%s cap=%dms pending=%v drain=%.2fms native=%.2fms outcome=%s close_ok=%v", raw, item.CapMS, item.PendingAtStop, item.DrainMS, o.ElapsedMS, o.Outcome, o.CloseOK != nil && *o.CloseOK)
			if !item.PendingAtStop || o.TimeoutMS != uint32(cap.Milliseconds()) || o.Outcome != "timeout" || o.NativeError != 11010 || o.CloseOK == nil || !*o.CloseOK {
				failures = append(failures, raw+": pending no-reply drain prerequisite or cleanup failed")
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
		CapturedAt   time.Time             `json:"captured_at"`
		Environment  map[string]any        `json:"environment"`
		Observations []drainCapObservation `json:"observations"`
		Gate         string                `json:"gate"`
		Load         *drainLoadObservation `json:"scheduler_load,omitempty"`
	}{time.Now().UTC(), environment(), observations, "incomplete; finite local drain samples do not establish a worst-case bound", load}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(output, data, 0600); err != nil {
		t.Fatal(err)
	}
	for _, failure := range failures {
		t.Error(failure)
	}
}
