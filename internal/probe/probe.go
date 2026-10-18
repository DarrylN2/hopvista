// Package probe is the IPv4 route/echo seam. It contains no Windows API types.
package probe

import (
	"context"
	"net/netip"
	"time"

	"hopvista/internal/result"
)

// InitialStartMargin is empirical headroom accepted for development on the
// reviewed Windows conditions, not a universal scheduling or completion bound.
const InitialStartMargin = 250 * time.Millisecond

type RouteRequest struct {
	ID         string
	SelectedIP netip.Addr
	DeadlineAt time.Time
}

type EchoRequest struct {
	ID         string
	SelectedIP netip.Addr
	TTL        int           // 1..255; trace uses 1..64, ordinary echo uses 64
	Cap        time.Duration // full native cap, positive integral ms, at most 1 s
	DeadlineAt time.Time
}

// Adapter performs synchronous calls. Context cancellation stops NEW work; it
// cannot be assumed to interrupt a started route or echo. Return only after
// completion AND cleanup. The caller must retain its run slot until return.
//
// Route checks cancellation/cutoff before native entry; it has no promised
// native time bound. Recheck after Route returns before scheduling Echo.
// Echo rechecks after setup, immediately before send, and starts only if its
// full Cap + InitialStartMargin fits. Never shrink Cap to remaining time.
// Record an unstarted refusal with no StartedAt, reply address or RTT.
// The clock check and native entry are not atomic under arbitrary scheduling.
// Started calls preserve actual observations even if cancellation arrives;
// the coordinator decides how to retain them in a stopping run.
type Adapter interface {
	Route(context.Context, RouteRequest) result.Route
	Echo(context.Context, EchoRequest) result.Echo
}

// StartOutcome centralizes the start policy for both fake and native adapters.
// A zero return permits work. Invalid request configuration is an error.
// A context deadline, when present, can only shorten the run cutoff.
func StartOutcome(ctx context.Context, now, cutoff time.Time, cap time.Duration) result.Outcome {
	if ctx.Err() == context.Canceled {
		return result.Cancelled
	}
	if ctx.Err() == context.DeadlineExceeded {
		return result.Deadline
	}
	if cutoff.IsZero() || cap < 0 || cap > time.Second || cap%time.Millisecond != 0 {
		return result.Error
	}
	if end, ok := ctx.Deadline(); ok && end.Before(cutoff) {
		cutoff = end
	}
	if !now.Before(cutoff) {
		return result.Deadline
	}
	if cap > 0 && cutoff.Sub(now) < cap+InitialStartMargin {
		return result.Deadline
	}
	return ""
}

func ValidIPv4(ip netip.Addr) bool {
	return ip.Is4() && !ip.IsUnspecified() && !ip.IsMulticast() && ip != netip.MustParseAddr("255.255.255.255")
}
