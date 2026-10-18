package engine

import (
	"fmt"
	"sort"

	"hopvista/internal/result"
)

// findings chooses one conservative observation. Only measurements from the
// same exchange are compared; ICMP RTTs never become website phase timings.
func findings(f result.Family) []result.Finding {
	if len(f.Exchanges) > 0 {
		x := f.Exchanges[len(f.Exchanges)-1]
		if x.HTTP.Outcome != result.Completed {
			phase, m := stoppedPhase(x)
			return []result.Finding{{Type: "http_phase_incomplete", Family: "ipv4", ExchangeRef: x.ID, Observation: fmt.Sprintf("This diagnostic run stopped during the %s phase for %s over IPv4.", phase, x.RequestedURL), Limit: "This direct attempt does not establish the website's general availability or the cause of the incomplete phase.", EvidenceRefs: []string{m.ID}}}
		}
		if x.ResponseStatus == nil {
			return []result.Finding{}
		}
		if x.RedirectStopReason != "" {
			return []result.Finding{{Type: "redirect_stopped", Family: "ipv4", ExchangeRef: x.ID, Observation: fmt.Sprintf("This diagnostic run received HTTP status %d from %s, then stopped at the redirect (%s).", status(x), x.RequestedURL, x.RedirectStopReason), Limit: "This policy stop does not establish whether the redirected website is reachable.", EvidenceRefs: []string{x.HTTP.ID}}}
		}
		if status(x) >= 400 {
			return []result.Finding{{Type: "http_response_observed", Family: "ipv4", ExchangeRef: x.ID, Observation: fmt.Sprintf("This diagnostic run received HTTP status %d from %s over IPv4.", status(x), x.RequestedURL), Limit: "This completed direct response does not establish what a browser experienced or why the server returned that status.", EvidenceRefs: []string{x.HTTP.ID}}}
		}
		if best, next, ok := dominantPhase(x); ok {
			return []result.Finding{{Type: "dominant_measured_wait", Family: "ipv4", ExchangeRef: x.ID, Observation: fmt.Sprintf("The %s phase was the longest measured wait in this direct IPv4 exchange with %s (%.0f ms).", best.name, x.RequestedURL, best.ms), Limit: "This comparison does not identify a server, network, gateway, or hop cause, or describe browser page load time.", EvidenceRefs: []string{best.id, next.id}}}
		}
		return []result.Finding{{Type: "http_response_observed", Family: "ipv4", ExchangeRef: x.ID, Observation: fmt.Sprintf("This diagnostic run received HTTP status %d from %s over IPv4; no single measured phase met the dominant-wait rule.", status(x), x.RequestedURL), Limit: "This direct exchange does not establish browser timing or a cause for the observed waits.", EvidenceRefs: []string{x.HTTP.ID}}}
	}
	if len(f.DNS) > 0 {
		d := f.DNS[len(f.DNS)-1]
		observation := fmt.Sprintf("This diagnostic run did not select a usable IPv4 address for %s.", d.Hostname)
		if d.Outcome == result.Timeout {
			observation = fmt.Sprintf("The IPv4 lookup for %s did not answer within its bound in this diagnostic run.", d.Hostname)
		}
		return []result.Finding{{Type: "ipv4_selection_incomplete", Family: "ipv4", Observation: observation, Limit: "This local resolver observation does not establish the hostname's availability elsewhere.", EvidenceRefs: []string{d.ID}}}
	}
	return []result.Finding{}
}

func status(x result.Exchange) int {
	if x.ResponseStatus == nil {
		return 0
	}
	return *x.ResponseStatus
}

func stoppedPhase(x result.Exchange) (string, result.Measurement) {
	if failedPhase(x.TCP.Outcome) {
		return "TCP connection", x.TCP
	}
	if failedPhase(x.TLS.Outcome) {
		return "TLS handshake", x.TLS
	}
	return "HTTP first response byte", x.HTTP
}

func failedPhase(o result.Outcome) bool {
	return o == result.Timeout || o == result.Error || o == result.Cancelled || o == result.Deadline
}

type phase struct {
	name, id string
	ms       float64
}

func dominantPhase(x result.Exchange) (phase, phase, bool) {
	var phases []phase
	for _, item := range []struct {
		name        string
		measurement result.Measurement
	}{
		{"DNS lookup", x.DNS}, {"TCP connection", x.TCP}, {"TLS handshake", x.TLS}, {"HTTP first response byte", x.HTTP},
	} {
		m := item.measurement
		if m.Outcome == result.Completed && m.DurationMS != nil && *m.DurationMS >= 0 {
			phases = append(phases, phase{name: item.name, id: m.ID, ms: *m.DurationMS})
		}
	}
	if len(phases) < 2 {
		return phase{}, phase{}, false
	}
	sort.SliceStable(phases, func(i, j int) bool { return phases[i].ms > phases[j].ms })
	if phases[0].ms < 100 || phases[0].ms < 2*phases[1].ms {
		return phase{}, phase{}, false
	}
	return phases[0], phases[1], true
}
