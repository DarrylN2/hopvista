package main

import (
	"context"
	"encoding/json"
	"net/netip"
	"os"
	"testing"
	"time"
)

// Opt-in live regression loop. Default unit tests never send network probes.
// The 50 ms allowance is an investigation threshold, not a new product promise.
func TestLiveNativeTimeout(t *testing.T) {
	if os.Getenv("HOPVISTA_LIVE_TIMEOUT") != "1" {
		t.Skip("enable with run.ps1 -TimeoutCheck")
	}
	if environment()["elevated"] == true {
		t.Fatal("requires an unelevated process")
	}
	var observations []echoObservation
	for attempt := 0; attempt < 3; attempt++ {
		o := echo(context.Background(), "150ms-native-timeout", netip.MustParseAddr("192.0.2.1"), 64, 150*time.Millisecond, false, 512, nil)
		observations = append(observations, o)
		t.Logf("attempt=%d timeout=%dms elapsed=%.2fms replies=%d native_error=%d close_ok=%v", attempt+1, o.TimeoutMS, o.ElapsedMS, o.ReplyCount, o.NativeError, o.CloseOK != nil && *o.CloseOK)
		if o.Outcome != "timeout" || o.NativeError != 11010 {
			t.Errorf("no-reply prerequisite unavailable: outcome=%s error=%d", o.Outcome, o.NativeError)
		} else if o.ElapsedMS > float64(o.TimeoutMS)+50 {
			t.Errorf("native timeout exceeded investigation allowance: requested=%dms actual=%.2fms", o.TimeoutMS, o.ElapsedMS)
		}
		if o.CloseOK == nil || !*o.CloseOK {
			t.Error("native handle was not closed successfully")
		}
	}
	if err := os.MkdirAll(".local", 0700); err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(observations, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".local/timeout-observations.json", data, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestLiveTimeoutSweep(t *testing.T) {
	if os.Getenv("HOPVISTA_TIMEOUT_SWEEP") != "1" {
		t.Skip("enable with run.ps1 -TimeoutSweep")
	}
	if environment()["elevated"] == true {
		t.Fatal("requires an unelevated process")
	}
	var observations []echoObservation
	for _, timeout := range []time.Duration{50, 150, 300, 500, 750, 1000, 1500} {
		o := echo(context.Background(), "timeout-sweep", netip.MustParseAddr("192.0.2.1"), 64, timeout*time.Millisecond, false, 512, nil)
		observations = append(observations, o)
		t.Logf("timeout=%dms elapsed=%.2fms replies=%d native_error=%d", o.TimeoutMS, o.ElapsedMS, o.ReplyCount, o.NativeError)
		if o.CloseOK == nil || !*o.CloseOK {
			t.Error("native handle cleanup failed")
		}
	}
	data, err := json.MarshalIndent(observations, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".local/timeout-sweep.json", data, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestLiveAsyncTimeout(t *testing.T) {
	if os.Getenv("HOPVISTA_ASYNC_TIMEOUT") != "1" {
		t.Skip("enable with run.ps1 -AsyncTimeoutCheck")
	}
	if environment()["elevated"] == true {
		t.Fatal("requires an unelevated process")
	}
	var observations []asyncObservation
	for _, destination := range []string{"127.0.0.1", "::1", "192.0.2.1"} {
		o := asyncEcho(netip.MustParseAddr(destination), 150)
		observations = append(observations, o)
		t.Logf("destination=%s submitted=%.2fms pending=%v completed=%.2fms event=%v replies=%d outcome=%s error=%d buffers_unpinned=%v", destination, o.SubmissionMS, o.Pending, o.Echo.ElapsedMS, o.EventSignaled, o.Echo.ReplyCount, o.Echo.Outcome, o.Echo.NativeError, o.BuffersUnpinned)
		if !o.EventSignaled || !o.BuffersUnpinned || !o.EventClosed || o.Echo.CloseOK == nil || !*o.Echo.CloseOK {
			t.Error("asynchronous completion/cleanup not verified")
		}
		if destination == "192.0.2.1" && (o.Echo.Outcome != "timeout" || o.Echo.ElapsedMS > 200) {
			t.Errorf("async path did not meet 150ms + 50ms investigation allowance: outcome=%s elapsed=%.2fms", o.Echo.Outcome, o.Echo.ElapsedMS)
		}
		if o.Echo.Outcome == "completion_unverified" {
			break
		}
	}
	data, err := json.MarshalIndent(observations, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".local/async-timeout-observations.json", data, 0600); err != nil {
		t.Fatal(err)
	}
}

// Candidate policy experiment only: no production adapter or coordinator.
func guardedProbe(ctx context.Context, label string) echoObservation {
	return guardedProbeCap(ctx, label, 750*time.Millisecond)
}

func guardedProbeCap(ctx context.Context, label string, cap time.Duration) echoObservation {
	// Reserve the entire cap plus this experiment's 50 ms allowance.
	// Refuse before handle creation. This does not redefine the Windows timeout.
	refused := echoObservation{Case: label, Destination: "192.0.2.1", Family: "ipv4", Scope: "documentation", API: "IcmpSendEcho2 (not called)", HopLimit: 64, Outcome: "not_started"}
	if err := ctx.Err(); err != nil {
		refused.Error = err.Error()
		return refused
	}
	if end, ok := ctx.Deadline(); ok && time.Until(end) < cap+50*time.Millisecond {
		refused.Error = "insufficient time for full native probe budget and investigation allowance"
		return refused
	}
	return echoInternal(ctx, label, netip.MustParseAddr("192.0.2.1"), 64, cap, false, 512, nil, 50*time.Millisecond, nil)
}

// A pause after the first guard must not turn the native call into a short
// timeout. This deliberately exercises the post-setup check with a real handle.
func TestLiveGuardAfterSetupDelay(t *testing.T) {
	if os.Getenv("HOPVISTA_GUARD_DELAY") != "1" {
		t.Skip("enable with run.ps1 -GuardDelayCheck")
	}
	if environment()["elevated"] == true {
		t.Fatal("requires an unelevated process")
	}
	const cap = 750 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), cap+50*time.Millisecond)
	defer cancel()
	o := echoInternal(ctx, "post-setup-delay", netip.MustParseAddr("192.0.2.1"), 64, cap, false, 512, func() { t.Error("entered native send") }, 50*time.Millisecond, func() { time.Sleep(80 * time.Millisecond) })
	if o.Outcome != "not_started" || !o.StartedAt.IsZero() || o.TimeoutMS != 0 || o.CloseOK == nil || !*o.CloseOK {
		t.Fatalf("post-setup guard did not refuse and close handle: %+v", o)
	}
	data, err := json.MarshalIndent(struct {
		CapturedAt  time.Time       `json:"captured_at"`
		Environment map[string]any  `json:"environment"`
		CapMS       int64           `json:"cap_ms"`
		BudgetMS    int64           `json:"budget_ms"`
		DelayMS     int64           `json:"induced_setup_delay_ms"`
		Echo        echoObservation `json:"echo"`
		Gate        string          `json:"gate"`
	}{time.Now().UTC(), environment(), cap.Milliseconds(), (cap + 50*time.Millisecond).Milliseconds(), 80, o, "incomplete; this only checks refusal after one induced setup delay"}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(".local", 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".local/guard-delay.json", data, 0600); err != nil {
		t.Fatal(err)
	}
	t.Logf("post-setup refusal: outcome=%s sent=%v close_ok=%v", o.Outcome, !o.StartedAt.IsZero(), *o.CloseOK)
}

type cutoffCapObservation struct {
	CapMS                int64           `json:"cap_ms"`
	BudgetMS             int64           `json:"budget_ms"`
	DeadlineAt           time.Time       `json:"deadline_at"`
	ReturnedAt           time.Time       `json:"returned_at"`
	RemainingAtStartMS   float64         `json:"remaining_at_start_ms,omitempty"`
	PastDeadlineAtReturn bool            `json:"past_deadline_at_return"`
	Echo                 echoObservation `json:"echo"`
}

// Harness-only check of all approved per-probe caps. A returned observation
// includes synchronous native completion and the deferred handle close.
func TestLiveCutoffCaps(t *testing.T) {
	if os.Getenv("HOPVISTA_CUTOFF_CAPS") != "1" {
		t.Skip("enable with run.ps1 -CutoffCapsCheck")
	}
	if environment()["elevated"] == true {
		t.Fatal("requires an unelevated process")
	}
	var observations []cutoffCapObservation
	for _, cap := range []time.Duration{750 * time.Millisecond, 900 * time.Millisecond, time.Second} {
		for _, budget := range []time.Duration{150 * time.Millisecond, cap + 49*time.Millisecond, cap + 150*time.Millisecond} {
			attempts := 1
			if budget == cap+150*time.Millisecond {
				attempts = 3
			}
			for attempt := 0; attempt < attempts; attempt++ {
				ctx, cancel := context.WithTimeout(context.Background(), budget)
				deadline, _ := ctx.Deadline()
				o := guardedProbeCap(ctx, "cutoff-cap", cap)
				returned := time.Now()
				cancel()
				item := cutoffCapObservation{CapMS: cap.Milliseconds(), BudgetMS: budget.Milliseconds(), DeadlineAt: deadline.UTC(), ReturnedAt: returned.UTC(), PastDeadlineAtReturn: !returned.Before(deadline), Echo: o}
				if !o.StartedAt.IsZero() {
					item.RemainingAtStartMS = milliseconds(deadline.Sub(o.StartedAt))
				}
				observations = append(observations, item)
				t.Logf("cap=%dms budget=%dms attempt=%d started=%v native_timeout=%dms elapsed=%.2fms outcome=%s deadline_passed=%v close_ok=%v", item.CapMS, item.BudgetMS, attempt+1, !o.StartedAt.IsZero(), o.TimeoutMS, o.ElapsedMS, o.Outcome, item.PastDeadlineAtReturn, o.CloseOK != nil && *o.CloseOK)
				if budget < cap+50*time.Millisecond {
					if o.Outcome != "not_started" || !o.StartedAt.IsZero() || o.CloseOK != nil {
						t.Error("short remaining budget entered native work")
					}
					continue
				}
				if o.TimeoutMS != uint32(cap.Milliseconds()) || o.Outcome != "timeout" || o.NativeError != 11010 {
					t.Error("full-cap no-reply prerequisite or native timeout not observed")
				}
				if o.CloseOK == nil || !*o.CloseOK {
					t.Error("native handle cleanup not verified")
				}
				if item.PastDeadlineAtReturn || o.ElapsedMS > float64(cap.Milliseconds()+50) {
					t.Error("native completion exceeded this experiment's budget or allowance")
				}
			}
		}
	}
	if err := os.MkdirAll(".local", 0700); err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(struct {
		CapturedAt             time.Time              `json:"captured_at"`
		Environment            map[string]any         `json:"environment"`
		Target                 string                 `json:"target"`
		InvestigationAllowance int64                  `json:"investigation_allowance_ms"`
		Observations           []cutoffCapObservation `json:"observations"`
		Gate                   string                 `json:"gate"`
	}{time.Now().UTC(), environment(), "192.0.2.1", 50, observations, "incomplete; cap samples do not validate the near-cutoff contract or worst-case bound"}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".local/cutoff-cap-observations.json", data, 0600); err != nil {
		t.Fatal(err)
	}
}

type marginObservation struct {
	Destination         string          `json:"destination"`
	CapMS               int64           `json:"cap_ms"`
	BudgetMS            int64           `json:"budget_ms"`
	Attempt             int             `json:"attempt"`
	PolicyCheckAt       time.Time       `json:"policy_check_at"`
	DeadlineAt          time.Time       `json:"deadline_at"`
	ReturnedAt          time.Time       `json:"returned_at"`
	CheckToSendMS       float64         `json:"check_to_send_ms"`
	CheckToCleanupMS    float64         `json:"check_to_cleanup_ms"`
	NativeToCleanupMS   float64         `json:"native_to_cleanup_ms"`
	PastFixtureDeadline bool            `json:"past_fixture_deadline"`
	Echo                echoObservation `json:"echo"`
}

// Opt-in local timing characterization, not a proof of a worst-case margin.
// Full-cap calls use three documentation-address candidates, five samples each.
func TestLiveMarginStudy(t *testing.T) {
	if os.Getenv("HOPVISTA_MARGIN_STUDY") != "1" {
		t.Skip("enable with run.ps1 -MarginStudy")
	}
	if environment()["elevated"] == true {
		t.Fatal("requires an unelevated process")
	}
	const allowance = 150 * time.Millisecond // Diagnostic observation window, not approved margin.
	var observations []marginObservation
	for _, raw := range []string{"192.0.2.1", "198.51.100.1", "203.0.113.1"} {
		ip := netip.MustParseAddr(raw)
		for _, cap := range []time.Duration{750 * time.Millisecond, 900 * time.Millisecond, time.Second} {
			for attempt := 1; attempt <= 5; attempt++ {
				budget := cap + allowance
				ctx, cancel := context.WithTimeout(context.Background(), budget)
				deadline, _ := ctx.Deadline()
				checked := time.Now()
				o := echo(ctx, "margin-study", ip, 64, cap, false, 512, nil)
				returned := time.Now()
				cancel()
				item := marginObservation{Destination: raw, CapMS: cap.Milliseconds(), BudgetMS: budget.Milliseconds(), Attempt: attempt, PolicyCheckAt: checked.UTC(), DeadlineAt: deadline.UTC(), ReturnedAt: returned.UTC(), CheckToCleanupMS: milliseconds(returned.Sub(checked)), PastFixtureDeadline: !returned.Before(deadline), Echo: o}
				if !o.StartedAt.IsZero() {
					item.CheckToSendMS = milliseconds(o.StartedAt.Sub(checked))
					item.NativeToCleanupMS = milliseconds(returned.Sub(o.EndedAt))
				}
				observations = append(observations, item)
				t.Logf("target=%s cap=%dms attempt=%d native=%.2fms check_to_cleanup=%.2fms outcome=%s error=%d close_ok=%v", raw, item.CapMS, attempt, o.ElapsedMS, item.CheckToCleanupMS, o.Outcome, o.NativeError, o.CloseOK != nil && *o.CloseOK)
				if o.TimeoutMS != uint32(cap.Milliseconds()) || o.Outcome != "timeout" || o.NativeError != 11010 {
					t.Error("no-reply prerequisite or full native cap not observed")
				}
				if o.CloseOK == nil || !*o.CloseOK || item.PastFixtureDeadline {
					t.Error("completion/cleanup did not fit the diagnostic observation window")
				}
			}
		}
	}
	if err := os.MkdirAll(".local", 0700); err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(struct {
		CapturedAt             time.Time           `json:"captured_at"`
		Environment            map[string]any      `json:"environment"`
		ObservationAllowanceMS int64               `json:"observation_allowance_ms"`
		Observations           []marginObservation `json:"observations"`
		Gate                   string              `json:"gate"`
	}{time.Now().UTC(), environment(), allowance.Milliseconds(), observations, "incomplete; local samples do not validate a production margin or worst-case bound"}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".local/margin-study.json", data, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestGuardRefusesShortRemainingBudget(t *testing.T) {
	for _, remaining := range []time.Duration{time.Nanosecond, 150 * time.Millisecond, 799 * time.Millisecond} {
		ctx, cancel := context.WithTimeout(context.Background(), remaining)
		o := guardedProbe(ctx, "short-budget")
		cancel()
		if o.Outcome != "not_started" || !o.StartedAt.IsZero() || o.CloseOK != nil {
			t.Fatalf("unsafe native work with %s remaining: %+v", remaining, o)
		}
	}
}

func TestLiveGuardedCutoff(t *testing.T) {
	if os.Getenv("HOPVISTA_GUARDED_CUTOFF") != "1" {
		t.Skip("enable with run.ps1 -GuardedCutoffCheck")
	}
	if environment()["elevated"] == true {
		t.Fatal("requires an unelevated process")
	}
	var observations []echoObservation
	short, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	o := guardedProbe(short, "guarded-short-remaining-budget")
	cancel()
	observations = append(observations, o)
	t.Logf("short budget: outcome=%s timeout=%dms elapsed=%.2fms started=%v", o.Outcome, o.TimeoutMS, o.ElapsedMS, !o.StartedAt.IsZero())
	if o.Outcome != "not_started" || !o.StartedAt.IsZero() || o.CloseOK != nil {
		t.Error("started native work without a full probe budget")
	}
	for attempt := 0; attempt < 3; attempt++ {
		ctx, stop := context.WithTimeout(context.Background(), 900*time.Millisecond)
		o = guardedProbe(ctx, "guarded-full-budget")
		stop()
		observations = append(observations, o)
		t.Logf("full budget: timeout=%dms elapsed=%.2fms outcome=%s", o.TimeoutMS, o.ElapsedMS, o.Outcome)
		if o.TimeoutMS != 750 || o.Outcome != "timeout" || o.ElapsedMS > 800 || o.CloseOK == nil || !*o.CloseOK {
			t.Error("full-budget native completion/cleanup not verified within investigation allowance")
		}
	}
	data, err := json.MarshalIndent(observations, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".local/guarded-cutoff-observations.json", data, 0600); err != nil {
		t.Fatal(err)
	}
}
