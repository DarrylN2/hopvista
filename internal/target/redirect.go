package target

import (
	"context"

	"hopvista/internal/result"
)

const MaxRedirects = 3

// RedirectDecision has Next only after all URL, IPv4 selection and new-host
// scope checks pass. The future HTTP transport must dial only Next.IP.
type RedirectDecision struct {
	Next          *Selection
	AttemptedURL  string
	SelectedIP    string
	SelectedScope Scope
	DNS           *result.DNS
	StopReason    string
}

// NextRedirect permits at most three follows (four exchanges total). For a
// new host, it resolves once and allows the follow only for a public address.
// Same-host redirects reuse the run's prior selection, including across ports.
func (s *Selector) NextRedirect(ctx context.Context, from Selection, location string, followed int) RedirectDecision {
	if !from.ip.IsValid() || from.url.value == nil {
		return RedirectDecision{StopReason: "invalid_redirect"}
	}
	if followed >= MaxRedirects {
		return RedirectDecision{StopReason: "redirect_limit"}
	}
	if followed < 0 {
		return RedirectDecision{StopReason: "invalid_redirect"}
	}
	nextURL, problem := resolveRedirect(from.url, location)
	if problem != nil {
		return RedirectDecision{StopReason: problem.Code}
	}
	decision := RedirectDecision{AttemptedURL: nextURL.String()}
	if ctx.Err() != nil {
		decision.StopReason = stoppedProblem(ctx).Code
		return decision
	}
	selected, dns, problem := s.Select(ctx, nextURL)
	decision.DNS = dns
	if problem != nil {
		decision.StopReason = problem.Code
		return decision
	}
	decision.SelectedIP, decision.SelectedScope = selected.ip.String(), selected.scope
	if nextURL.hostKey != from.url.hostKey && selected.scope != Public {
		decision.StopReason = "redirect_requires_new_run"
		return decision
	}
	decision.Next = &selected
	return decision
}
