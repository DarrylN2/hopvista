// Package engine coordinates one IPv4 diagnostic run from pinned HTTP evidence
// through route and ICMP path sampling. The run manager owns cutoff and storage.
package engine

import (
	"context"
	"fmt"
	"net/netip"

	"hopvista/internal/httpmeasure"
	"hopvista/internal/pathmeasure"
	"hopvista/internal/probe"
	"hopvista/internal/result"
	"hopvista/internal/run"
	"hopvista/internal/target"
)

type HTTPMeasurer interface {
	Measure(context.Context, string, *target.Selector) httpmeasure.Journey
}

type progressingHTTP interface {
	MeasureWithProgress(context.Context, string, *target.Selector, func(httpmeasure.Journey)) httpmeasure.Journey
}

type Engine struct {
	HTTP     HTTPMeasurer
	Resolver target.Resolver
	Probe    probe.Adapter
	Clock    pathmeasure.Clock
}

var _ run.Executor = (*Engine)(nil)

func New(prober probe.Adapter) *Engine {
	return &Engine{HTTP: httpmeasure.NewClient(), Probe: prober, Clock: pathmeasure.RealClock{}}
}

func (e *Engine) Run(ctx context.Context, input run.Input, publish func([]result.Family)) ([]result.Family, []result.Finding, error) {
	if e == nil || e.HTTP == nil || e.Probe == nil || publish == nil {
		return nil, nil, fmt.Errorf("measurement engine is incomplete")
	}
	clock := e.Clock
	if clock == nil {
		clock = pathmeasure.RealClock{}
	}
	_, deadlineOK := ctx.Deadline()
	if !deadlineOK {
		return nil, nil, fmt.Errorf("measurement deadline is required")
	}
	selector := target.NewSelector(e.Resolver, clock.Now)
	ipv4 := result.Family{Family: "ipv4", Outcome: result.Pending, DNS: []result.DNS{}, Exchanges: []result.Exchange{}, Paths: []result.Path{}}
	families := []result.Family{ipv4, unsupportedIPv6()}
	var journey httpmeasure.Journey
	if progressive, ok := e.HTTP.(progressingHTTP); ok {
		journey = progressive.MeasureWithProgress(ctx, input.TargetURL, selector, func(progress httpmeasure.Journey) {
			ipv4 = familyFromJourney(progress)
			families[0] = ipv4
			publish(families)
		})
	} else {
		journey = e.HTTP.Measure(ctx, input.TargetURL, selector)
	}
	ipv4 = familyFromJourney(journey)
	families[0] = ipv4
	publish(families)
	if journey.SelectedIP != "" && ctx.Err() == nil {
		ip, err := netip.ParseAddr(journey.SelectedIP)
		if err != nil || !probe.ValidIPv4(ip) {
			return families, nil, fmt.Errorf("HTTP journey returned an invalid selected IPv4 address")
		}
		deadline, _ := ctx.Deadline()
		role := "attempted_host"
		if journey.FinalURL != "" {
			role = "final_host"
		}
		path := pathmeasure.MeasureWithProgress(ctx, e.Probe, clock, pathmeasure.Request{ID: "path-1", SelectedIP: ip, DeadlineAt: deadline, MaxHops: input.MaxHops, TargetRole: role}, func(progress result.Path) {
			ipv4.Paths = []result.Path{progress}
			families[0] = ipv4
			publish(families)
		})
		ipv4.Paths = []result.Path{path}
		if path.Trace.Outcome == result.TraceDeadline {
			ipv4.Outcome = result.Deadline
		}
		if path.Trace.Outcome == result.TraceCancelled {
			ipv4.Outcome = result.Cancelled
		}
		families[0] = ipv4
		publish(families)
	}
	return families, findings(ipv4), nil
}

func familyFromJourney(j httpmeasure.Journey) result.Family {
	f := result.Family{Family: "ipv4", Outcome: j.Outcome, DNS: j.DNS, Exchanges: j.Exchanges, Paths: []result.Path{}, FinalURL: j.FinalURL, FinalHost: j.FinalHost}
	if j.StopReason != "" {
		f.Error = &result.ErrorData{Code: j.StopReason, Message: "The IPv4 measurement journey stopped before a final response."}
	}
	return f
}

func unsupportedIPv6() result.Family {
	return result.Family{Family: "ipv6", Outcome: result.Unavailable, Error: &result.ErrorData{Code: "ipv6_not_supported_in_v1", Message: "IPv6 measurements are deferred in this version."}, DNS: []result.DNS{}, Exchanges: []result.Exchange{}, Paths: []result.Path{}}
}
