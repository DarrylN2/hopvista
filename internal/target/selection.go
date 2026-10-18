package target

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"strconv"
	"time"

	"hopvista/internal/result"
)

const lookupBound = 2 * time.Second

type Resolver interface {
	LookupIPv4(context.Context, string) ([]netip.Addr, error)
}

type systemResolver struct{}

func (systemResolver) LookupIPv4(ctx context.Context, host string) ([]netip.Addr, error) {
	return net.DefaultResolver.LookupNetIP(ctx, "ip4", host)
}

type Selection struct {
	url     URL
	ip      netip.Addr
	scope   Scope
	dnsRef  string
	answers []netip.Addr // resolver-returned A addresses, including unusable ones
}

func (s Selection) URL() URL              { return s.url }
func (s Selection) IP() netip.Addr        { return s.ip }
func (s Selection) Scope() Scope          { return s.scope }
func (s Selection) DNSRef() string        { return s.dnsRef }
func (s Selection) Answers() []netip.Addr { return append([]netip.Addr(nil), s.answers...) }

type cachedSelection struct {
	selection Selection
	problem   *Problem
}

// Selector belongs to one diagnostic run and is used sequentially by its
// coordinator. Host selections (including no-A outcomes) are cached once;
// a redirect to the same host reuses the original pinned IP.
type Selector struct {
	resolver Resolver
	now      func() time.Time
	byHost   map[string]cachedSelection
	nextID   int
}

func NewSelector(resolver Resolver, now func() time.Time) *Selector {
	if resolver == nil {
		resolver = systemResolver{}
	}
	if now == nil {
		now = time.Now
	}
	return &Selector{resolver: resolver, now: now, byHost: make(map[string]cachedSelection)}
}

// Select returns a DNS record only for a new host. A literal IP receives a
// skipped DNS record. Cached hosts return nil and retain the prior DNSRef.
func (s *Selector) Select(ctx context.Context, u URL) (Selection, *result.DNS, *Problem) {
	if cached, ok := s.byHost[u.hostKey]; ok {
		selected := cached.selection
		selected.url = u
		return selected, nil, copyProblem(cached.problem)
	}
	s.nextID++
	id := "dns-" + strconv.Itoa(s.nextID)
	selected := Selection{url: u, dnsRef: id, answers: []netip.Addr{}}
	dns := &result.DNS{Measurement: result.Measurement{ID: id, Method: "local_resolver_ipv4", Outcome: result.Unavailable}, Hostname: u.hostname, Addresses: []string{}}
	if u.literal.IsValid() {
		selected.ip, selected.scope = u.literal, Classify(u.literal)
		dns.Method, dns.Outcome = "literal_ip", result.Skipped
		dns.SelectedIP, dns.SelectedScope = selected.ip.String(), string(selected.scope)
		s.byHost[u.hostKey] = cachedSelection{selection: selected}
		return selected, dns, nil
	}
	if ctx.Err() != nil {
		problem := stoppedProblem(ctx)
		dns.Outcome, dns.Error = stoppedOutcome(ctx), &result.ErrorData{Code: problem.Code, Message: problem.Message}
		return selected, dns, problem
	}
	started := s.now().UTC()
	dns.StartedAt = &started
	lookupCtx, cancel := context.WithTimeout(ctx, lookupBound)
	defer cancel()
	bound := lookupBound
	if deadline, ok := lookupCtx.Deadline(); ok && deadline.Sub(started) < bound {
		bound = deadline.Sub(started)
	}
	if bound < 0 {
		bound = 0
	}
	boundMS := bound.Seconds() * 1000
	dns.BoundMS = &boundMS
	answers, err := s.resolver.LookupIPv4(lookupCtx, u.hostname)
	ended := s.now().UTC()
	dns.EndedAt = &ended
	if lookupCtx.Err() != nil {
		if ctx.Err() == nil && lookupCtx.Err() == context.DeadlineExceeded {
			problem := &Problem{Code: "dns_timeout", Message: "No IPv4 lookup answer arrived within two seconds."}
			dns.Outcome, dns.Error = result.Timeout, &result.ErrorData{Code: problem.Code, Message: problem.Message}
			s.byHost[u.hostKey] = cachedSelection{selection: selected, problem: copyProblem(problem)}
			return selected, dns, problem
		}
		problem := stoppedProblem(lookupCtx)
		dns.Outcome, dns.Error = stoppedOutcome(lookupCtx), &result.ErrorData{Code: problem.Code, Message: problem.Message}
		return selected, dns, problem
	}
	if err != nil {
		problem := &Problem{Code: "dns_lookup_failed", Message: "The local IPv4 lookup failed."}
		if errors.Is(err, context.DeadlineExceeded) {
			problem = &Problem{Code: "dns_timeout", Message: "No IPv4 lookup answer arrived within two seconds."}
			dns.Outcome = result.Timeout
		}
		if dns.Outcome != result.Timeout {
			dns.Outcome = result.Error
		}
		dns.Error = &result.ErrorData{Code: problem.Code, Message: problem.Message}
		s.byHost[u.hostKey] = cachedSelection{selection: selected, problem: copyProblem(problem)}
		return selected, dns, problem
	}
	duration := ended.Sub(started).Seconds() * 1000
	if duration < 0 {
		duration = 0
	}
	dns.DurationMS = &duration
	dns.Outcome = result.Completed
	for _, ip := range answers {
		if !ip.Is4() {
			continue
		}
		dns.Addresses = append(dns.Addresses, ip.String())
		selected.answers = append(selected.answers, ip)
		if !selected.ip.IsValid() && UsableIPv4(ip) {
			selected.ip, selected.scope = ip, Classify(ip)
		}
	}
	if !selected.ip.IsValid() {
		problem := &Problem{Code: "ipv4_unavailable", Message: "No usable IPv4 A address was found for this hostname."}
		s.byHost[u.hostKey] = cachedSelection{selection: selected, problem: copyProblem(problem)}
		return selected, dns, problem
	}
	dns.SelectedIP, dns.SelectedScope = selected.ip.String(), string(selected.scope)
	s.byHost[u.hostKey] = cachedSelection{selection: selected}
	return selected, dns, nil
}

func stoppedOutcome(ctx context.Context) result.Outcome {
	if ctx.Err() == context.DeadlineExceeded {
		return result.Deadline
	}
	return result.Cancelled
}
func stoppedProblem(ctx context.Context) *Problem {
	if ctx.Err() == context.DeadlineExceeded {
		return &Problem{Code: "deadline", Message: "The IPv4 lookup reached the measurement cutoff."}
	}
	return &Problem{Code: "cancelled", Message: "The diagnostic run was cancelled before IPv4 selection."}
}
func copyProblem(p *Problem) *Problem {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}
