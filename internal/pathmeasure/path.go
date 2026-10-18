// Package pathmeasure runs the bounded IPv4 route, echo, and TTL plan for one
// selected host. Its only network dependency is the normalized probe adapter.
package pathmeasure

import (
	"context"
	"fmt"
	"net/netip"
	"sort"
	"time"

	"hopvista/internal/probe"
	"hopvista/internal/result"
)

const VariationFormula = "median(abs(rtt_ms[i] - rtt_ms[i-1])) over adjacent successful attempts in send order"

type Clock interface {
	Now() time.Time
	Wait(context.Context, time.Duration) error
}

type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }
func (RealClock) Wait(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

type Request struct {
	ID         string
	SelectedIP netip.Addr
	DeadlineAt time.Time
	MaxHops    int
	TargetRole string // final_host or attempted_host
}

// Measure preserves each reply and refusal. It never starts a new native call
// after cancellation/cutoff; a started synchronous adapter call owns the slot
// until it returns and closes its handle.
func Measure(ctx context.Context, adapter probe.Adapter, clock Clock, req Request) result.Path {
	return MeasureWithProgress(ctx, adapter, clock, req, nil)
}

// MeasureWithProgress calls publish synchronously with detached path snapshots
// after route, each echo sample, and each trace hop. Findings remain the run
// manager's responsibility and are never emitted from here.
func MeasureWithProgress(ctx context.Context, adapter probe.Adapter, clock Clock, req Request, publish func(result.Path)) result.Path {
	if clock == nil {
		clock = RealClock{}
	}
	path := result.Path{ID: req.ID, SelectedIP: req.SelectedIP.String()}
	path.Gateway = result.EchoSeries{ID: req.ID + "-gateway", Outcome: result.Skipped, Samples: []result.Echo{}}
	path.Target = result.EchoSeries{ID: req.ID + "-target", Outcome: result.Skipped, Samples: []result.Echo{}}
	path.Trace = result.Trace{ID: req.ID + "-trace", SelectedIP: req.SelectedIP.String(), Family: "ipv4", TargetRole: req.TargetRole, MaxHops: req.MaxHops, Outcome: result.TraceUnavailable, Hops: []result.Echo{}}
	emit := func() {
		if publish == nil {
			return
		}
		copy := path
		copy.Gateway.Samples = append([]result.Echo{}, path.Gateway.Samples...)
		copy.Target.Samples = append([]result.Echo{}, path.Target.Samples...)
		copy.Trace.Hops = append([]result.Echo{}, path.Trace.Hops...)
		publish(copy)
	}
	if req.ID == "" || !probe.ValidIPv4(req.SelectedIP) || req.DeadlineAt.IsZero() || req.MaxHops < 1 || req.MaxHops > 64 || (req.TargetRole != "final_host" && req.TargetRole != "attempted_host") {
		path.Route = result.Route{Measurement: result.Measurement{ID: req.ID + "-route", Outcome: result.Error, Method: "route_selection", Error: &result.ErrorData{Code: "invalid_path_request", Message: "Path request is incomplete."}}, SelectedIP: req.SelectedIP.String()}
		return path
	}
	path.Route = adapter.Route(ctx, probe.RouteRequest{ID: req.ID + "-route", SelectedIP: req.SelectedIP, DeadlineAt: req.DeadlineAt})
	emit()
	if s := pathStop(path.Route.Outcome, stop(ctx, clock, req.DeadlineAt)); s != "" {
		path.Gateway.Outcome, path.Target.Outcome, path.Trace.Outcome = s, s, traceStop(s)
		emit()
		return path
	}
	if path.Route.Outcome == result.Completed && path.Route.OnLink != nil && !*path.Route.OnLink && path.Route.NextHop != "" {
		if gateway, err := netip.ParseAddr(path.Route.NextHop); err == nil && probe.ValidIPv4(gateway) {
			path.Gateway = series(ctx, adapter, clock, path.Gateway.ID, gateway, 3, 250*time.Millisecond, 900*time.Millisecond, req.DeadlineAt, func(s result.EchoSeries) { path.Gateway = s; emit() })
		}
	}
	if s := pathStop(path.Gateway.Outcome, stop(ctx, clock, req.DeadlineAt)); s != "" {
		path.Target.Outcome, path.Trace.Outcome = s, traceStop(s)
		emit()
		return path
	}
	path.Target = series(ctx, adapter, clock, path.Target.ID, req.SelectedIP, 5, 300*time.Millisecond, time.Second, req.DeadlineAt, func(s result.EchoSeries) { path.Target = s; emit() })
	path.Target.Variation = variation(path.Target.ID+"-variation", path.Target.Samples)
	emit()
	if s := pathStop(path.Target.Outcome, stop(ctx, clock, req.DeadlineAt)); s != "" {
		path.Trace.Outcome = traceStop(s)
		emit()
		return path
	}
	path.Trace = traceWithProgress(ctx, adapter, clock, path.Trace, req.SelectedIP, req.DeadlineAt, func(tr result.Trace) { path.Trace = tr; emit() })
	return path
}

func stop(ctx context.Context, clock Clock, deadline time.Time) result.Outcome {
	return probe.StartOutcome(ctx, clock.Now(), deadline, 0)
}
func stopped(o result.Outcome) bool { return o == result.Cancelled || o == result.Deadline }
func pathStop(observed, current result.Outcome) result.Outcome {
	if stopped(observed) {
		return observed
	}
	if stopped(current) {
		return current
	}
	return ""
}

func series(ctx context.Context, adapter probe.Adapter, clock Clock, id string, ip netip.Addr, count int, spacing, cap time.Duration, deadline time.Time, publish func(result.EchoSeries)) result.EchoSeries {
	out := result.EchoSeries{ID: id, Outcome: result.Pending, Samples: []result.Echo{}}
	first := clock.Now()
	for i := 0; i < count; i++ {
		if i > 0 {
			wait := first.Add(time.Duration(i) * spacing).Sub(clock.Now())
			if err := clock.Wait(ctx, wait); err != nil {
				out.Outcome = stop(ctx, clock, deadline)
				if out.Outcome == "" {
					out.Outcome = result.Cancelled
				}
				break
			}
		}
		if s := stop(ctx, clock, deadline); s != "" {
			out.Outcome = s
			break
		}
		e := adapter.Echo(ctx, probe.EchoRequest{ID: fmt.Sprintf("%s-%d", id, i+1), SelectedIP: ip, TTL: 64, Cap: cap, DeadlineAt: deadline})
		out.Samples = append(out.Samples, e)
		if e.StartedAt != nil {
			out.Attempted++
		}
		if e.Status == result.EchoReply && e.Outcome == result.Completed {
			out.Replied++
		}
		if publish != nil {
			publish(out)
		}
		if stopped(e.Outcome) {
			out.Outcome = e.Outcome
			break
		}
	}
	if out.Outcome == result.Pending {
		out.Outcome = result.Completed
	}
	if publish != nil {
		publish(out)
	}
	return out
}

func variation(id string, samples []result.Echo) *result.Variation {
	v := &result.Variation{ID: id, Outcome: result.Unavailable, Formula: VariationFormula, EvidenceRefs: []string{}}
	var successful []result.Echo
	for _, e := range samples {
		if e.Status == result.EchoReply && e.Outcome == result.Completed && e.RTTMS != nil && e.StartedAt != nil {
			successful = append(successful, e)
		}
	}
	if len(successful) < 3 {
		return v
	}
	sort.SliceStable(successful, func(i, j int) bool { return successful[i].StartedAt.Before(*successful[j].StartedAt) })
	deltas := make([]float64, 0, len(successful)-1)
	for i := 1; i < len(successful); i++ {
		d := *successful[i].RTTMS - *successful[i-1].RTTMS
		if d < 0 {
			d = -d
		}
		deltas = append(deltas, d)
		v.EvidenceRefs = append(v.EvidenceRefs, successful[i-1].ID, successful[i].ID)
	}
	sort.Float64s(deltas)
	mid := len(deltas) / 2
	value := deltas[mid]
	if len(deltas)%2 == 0 {
		value = (deltas[mid-1] + deltas[mid]) / 2
	}
	v.Outcome = result.Completed
	v.ValueMS = &value
	return v
}

func trace(ctx context.Context, adapter probe.Adapter, clock Clock, out result.Trace, ip netip.Addr, deadline time.Time) result.Trace {
	return traceWithProgress(ctx, adapter, clock, out, ip, deadline, nil)
}

func traceWithProgress(ctx context.Context, adapter probe.Adapter, clock Clock, out result.Trace, ip netip.Addr, deadline time.Time, publish func(result.Trace)) result.Trace {
	out.Outcome = result.TracePending
	if publish != nil {
		publish(out)
	}
	for ttl := 1; ttl <= out.MaxHops; ttl++ {
		if s := stop(ctx, clock, deadline); s != "" {
			out.Outcome = traceStop(s)
			if publish != nil {
				publish(out)
			}
			return out
		}
		e := adapter.Echo(ctx, probe.EchoRequest{ID: fmt.Sprintf("%s-ttl-%d", out.ID, ttl), SelectedIP: ip, TTL: ttl, Cap: 750 * time.Millisecond, DeadlineAt: deadline})
		out.Hops = append(out.Hops, e)
		if e.StartedAt != nil {
			out.HighestAttemptedTTL = ttl
		}
		if stopped(e.Outcome) {
			out.Outcome = traceStop(e.Outcome)
			if publish != nil {
				publish(out)
			}
			return out
		}
		if e.Outcome == result.Error {
			out.Outcome = result.TraceError
			if publish != nil {
				publish(out)
			}
			return out
		}
		if e.Status == result.UnreachableReply {
			out.Outcome = result.UnreachableResponse
			if publish != nil {
				publish(out)
			}
			return out
		}
		if e.Status == result.EchoReply && e.ReplyAddress == ip.String() {
			out.Outcome = result.DestinationReplied
			if publish != nil {
				publish(out)
			}
			return out
		}
		if publish != nil {
			publish(out)
		}
	}
	out.Outcome = result.HopLimitReached
	if publish != nil {
		publish(out)
	}
	return out
}

func traceStop(o result.Outcome) result.TraceOutcome {
	if o == result.Cancelled {
		return result.TraceCancelled
	}
	return result.TraceDeadline
}
