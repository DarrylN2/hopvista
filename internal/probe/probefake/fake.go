// Package probefake provides scripted, network-free observations through the
// same interface as the future Windows adapter. Hooks model setup and pending
// synchronous calls; an in-flight Finish hook deliberately ignores cancellation.
package probefake

import (
	"context"
	"sync"
	"time"

	"hopvista/internal/probe"
	"hopvista/internal/result"
)

type RouteStep struct {
	Observation result.Route
	Finish      func() // may block until the test explicitly releases it
}

type EchoStep struct {
	Observation result.Echo
	Setup       func() // runs before the final send guard
	Finish      func() // runs after send; resources remain owned until it returns
}

type Call struct {
	Route *probe.RouteRequest
	Echo  *probe.EchoRequest
}

type Adapter struct {
	mu     sync.Mutex
	now    func() time.Time
	routes []RouteStep
	echoes []EchoStep
	calls  []Call
}

var _ probe.Adapter = (*Adapter)(nil)

// now must be concurrency-safe. Configure immutable scripts before use. Hooks
// are called without holding the fake's mutex so tests can inspect pending work.
func New(now func() time.Time, routes []RouteStep, echoes []EchoStep) *Adapter {
	return &Adapter{now: now, routes: append([]RouteStep(nil), routes...), echoes: append([]EchoStep(nil), echoes...)}
}

// Calls reports only started work, including work still pending cleanup.
func (a *Adapter) Calls() []Call {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]Call, len(a.calls))
	for i, c := range a.calls {
		if c.Route != nil {
			r := *c.Route
			out[i].Route = &r
		}
		if c.Echo != nil {
			e := *c.Echo
			out[i].Echo = &e
		}
	}
	return out
}

func (a *Adapter) Route(ctx context.Context, r probe.RouteRequest) result.Route {
	o := result.Route{Measurement: result.Measurement{ID: r.ID, Method: "route_selection"}, SelectedIP: r.SelectedIP.String()}
	if r.ID == "" || !probe.ValidIPv4(r.SelectedIP) || r.DeadlineAt.IsZero() {
		o.Outcome = result.Error
		o.Error = invalidRequest()
		return o
	}
	if stop := probe.StartOutcome(ctx, a.now(), r.DeadlineAt, 0); stop != "" {
		o.Outcome = stop
		return o
	}
	a.mu.Lock()
	if len(a.routes) == 0 {
		a.mu.Unlock()
		o.Outcome = result.Error
		o.Error = exhausted()
		return o
	}
	s := a.routes[0]
	a.routes = a.routes[1:]
	a.calls = append(a.calls, Call{Route: &r})
	a.mu.Unlock()
	started := a.now().UTC()
	if s.Finish != nil {
		s.Finish()
	}
	o = s.Observation
	o.ID, o.Method, o.SelectedIP = r.ID, "route_selection", r.SelectedIP.String()
	ended := a.now().UTC()
	o.StartedAt, o.EndedAt = &started, &ended
	o.BoundMS, o.DurationMS = nil, nil // route has no native bound
	if o.Outcome == result.Completed {
		d := ended.Sub(started).Seconds() * 1000
		o.DurationMS = &d
	}
	return o
}

func (a *Adapter) Echo(ctx context.Context, r probe.EchoRequest) result.Echo {
	bound := r.Cap.Seconds() * 1000
	o := result.Echo{Measurement: result.Measurement{ID: r.ID, Method: "icmp_echo", BoundMS: &bound}, SelectedIP: r.SelectedIP.String(), TTL: r.TTL, Status: result.NotSent}
	if r.ID == "" || !probe.ValidIPv4(r.SelectedIP) || r.DeadlineAt.IsZero() || r.TTL < 1 || r.TTL > 255 || r.Cap <= 0 || r.Cap > time.Second || r.Cap%time.Millisecond != 0 {
		o.Outcome = result.Error
		o.Error = invalidRequest()
		return o
	}
	if stop := probe.StartOutcome(ctx, a.now(), r.DeadlineAt, r.Cap); stop != "" {
		o.Outcome = stop
		return o
	}
	a.mu.Lock()
	if len(a.echoes) == 0 {
		a.mu.Unlock()
		o.Outcome = result.Error
		o.Error = exhausted()
		return o
	}
	s := a.echoes[0]
	a.echoes = a.echoes[1:]
	a.mu.Unlock()
	if s.Setup != nil {
		s.Setup()
	}
	if stop := probe.StartOutcome(ctx, a.now(), r.DeadlineAt, r.Cap); stop != "" {
		o.Outcome = stop
		return o
	}
	started := a.now().UTC()
	a.mu.Lock()
	a.calls = append(a.calls, Call{Echo: &r})
	a.mu.Unlock()
	if s.Finish != nil {
		s.Finish()
	}
	o = s.Observation
	o.ID, o.Method, o.SelectedIP, o.TTL = r.ID, "icmp_echo", r.SelectedIP.String(), r.TTL
	ended := a.now().UTC()
	o.StartedAt, o.EndedAt, o.BoundMS = &started, &ended, &bound
	o.DurationMS = nil
	if o.Outcome == result.Completed {
		d := ended.Sub(started).Seconds() * 1000
		o.DurationMS = &d
	}
	return o
}

func invalidRequest() *result.ErrorData {
	return &result.ErrorData{Code: "invalid_probe_request", Message: "A valid IPv4 request and full-cap budget are required."}
}
func exhausted() *result.ErrorData {
	return &result.ErrorData{Code: "fake_script_exhausted", Message: "No scripted observation remains."}
}
