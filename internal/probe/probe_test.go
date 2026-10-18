package probe_test

import (
	"context"
	"testing"
	"time"

	"hopvista/internal/probe"
	"hopvista/internal/result"
)

func TestFullCapStartPolicy(t *testing.T) {
	now := time.Date(2026, 10, 14, 20, 0, 0, 0, time.UTC)
	for _, cap := range []time.Duration{750 * time.Millisecond, 900 * time.Millisecond, time.Second} {
		for _, delta := range []time.Duration{-time.Nanosecond, 0, time.Nanosecond} {
			got := probe.StartOutcome(context.Background(), now, now.Add(cap+probe.InitialStartMargin+delta), cap)
			want := result.Outcome("")
			if delta < 0 {
				want = result.Deadline
			}
			if got != want {
				t.Fatalf("cap %v delta %v: got %q, want %q", cap, delta, got, want)
			}
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got := probe.StartOutcome(ctx, now, now.Add(time.Minute), time.Second); got != result.Cancelled {
		t.Fatal(got)
	}
	for _, cap := range []time.Duration{-1, time.Second + time.Millisecond, time.Microsecond} {
		if got := probe.StartOutcome(context.Background(), now, now.Add(time.Minute), cap); got != result.Error {
			t.Fatalf("invalid cap %v: %s", cap, got)
		}
	}
	if got := probe.StartOutcome(context.Background(), now, now, 0); got != result.Deadline {
		t.Fatal(got)
	}
	if got := probe.StartOutcome(context.Background(), now, time.Time{}, 0); got != result.Error {
		t.Fatal(got)
	}
}

func TestContextDeadlineCannotExtendRunCutoff(t *testing.T) {
	now := time.Now()
	ctx, cancel := context.WithDeadline(context.Background(), now.Add(time.Second))
	defer cancel()
	if got := probe.StartOutcome(ctx, now, now.Add(time.Minute), time.Second); got != result.Deadline {
		t.Fatal(got)
	}
	ctx2, cancel2 := context.WithDeadline(context.Background(), now.Add(time.Minute))
	defer cancel2()
	if got := probe.StartOutcome(ctx2, now, now.Add(time.Second), time.Second); got != result.Deadline {
		t.Fatal(got)
	}
}
