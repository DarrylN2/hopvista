package engine

import (
	"strings"
	"testing"

	"hopvista/internal/result"
)

func measured(id string, ms float64) result.Measurement {
	return result.Measurement{ID: id, Outcome: result.Completed, DurationMS: &ms}
}

func TestDominantWaitRuleEdges(t *testing.T) {
	for _, tc := range []struct {
		name          string
		first, second float64
		dominant      bool
	}{
		{"below_100", 99, 49, false}, {"at_100_and_2x", 100, 50, true},
		{"below_2x", 100, 50.1, false}, {"above_both", 201, 100, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status := 200
			x := result.Exchange{ID: "exchange-1", RequestedURL: "https://example.com/", DNS: measured("dns", tc.second), TCP: measured("tcp", tc.first), TLS: result.Measurement{ID: "tls", Outcome: result.Skipped}, HTTP: result.Measurement{ID: "http", Outcome: result.Completed}, ResponseStatus: &status}
			f := findings(result.Family{Exchanges: []result.Exchange{x}})
			if len(f) != 1 {
				t.Fatalf("findings: %+v", f)
			}
			if (f[0].Type == "dominant_measured_wait") != tc.dominant {
				t.Fatalf("dominant=%v: %+v", tc.dominant, f[0])
			}
			if tc.dominant && (len(f[0].EvidenceRefs) != 2 || f[0].EvidenceRefs[0] != "tcp" || f[0].EvidenceRefs[1] != "dns") {
				t.Fatalf("evidence: %+v", f[0])
			}
		})
	}
}

func TestFindingUsesStoppedPhaseAndCompletedErrorStatus(t *testing.T) {
	for _, tc := range []struct {
		name                      string
		x                         result.Exchange
		wantType, wantRef, phrase string
	}{
		{"tcp", result.Exchange{ID: "x", RequestedURL: "https://example.com/", TCP: result.Measurement{ID: "tcp", Outcome: result.Timeout}, HTTP: result.Measurement{ID: "http", Outcome: result.Unavailable}}, "http_phase_incomplete", "tcp", "TCP connection"},
		{"tls", result.Exchange{ID: "x", RequestedURL: "https://example.com/", TCP: result.Measurement{ID: "tcp", Outcome: result.Completed}, TLS: result.Measurement{ID: "tls", Outcome: result.Error}, HTTP: result.Measurement{ID: "http", Outcome: result.Unavailable}}, "http_phase_incomplete", "tls", "TLS handshake"},
		{"first_byte", result.Exchange{ID: "x", RequestedURL: "https://example.com/", TCP: result.Measurement{ID: "tcp", Outcome: result.Completed}, TLS: result.Measurement{ID: "tls", Outcome: result.Completed}, HTTP: result.Measurement{ID: "http", Outcome: result.Timeout}}, "http_phase_incomplete", "http", "HTTP first response byte"},
		{"service_unavailable", func() result.Exchange {
			s := 503
			return result.Exchange{ID: "x", RequestedURL: "https://example.com/", HTTP: result.Measurement{ID: "http", Outcome: result.Completed}, ResponseStatus: &s}
		}(), "http_response_observed", "http", "status 503"},
		{"redirect_limit", func() result.Exchange {
			s := 302
			return result.Exchange{ID: "x", RequestedURL: "https://example.com/", HTTP: result.Measurement{ID: "http", Outcome: result.Completed}, ResponseStatus: &s, RedirectStopReason: "redirect_limit"}
		}(), "redirect_stopped", "http", "redirect_limit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := findings(result.Family{Exchanges: []result.Exchange{tc.x}})
			if len(f) != 1 || f[0].Type != tc.wantType || len(f[0].EvidenceRefs) != 1 || f[0].EvidenceRefs[0] != tc.wantRef || !strings.Contains(f[0].Observation, tc.phrase) || f[0].ExchangeRef != "x" || f[0].Limit == "" {
				t.Fatalf("finding: %+v", f)
			}
		})
	}
}

func TestNoDominanceAcrossExchangesOrFromICMP(t *testing.T) {
	status := 200
	x := result.Exchange{ID: "x", RequestedURL: "http://example.com/", DNS: result.Measurement{ID: "dns", Outcome: result.Skipped}, TCP: measured("tcp", 99), TLS: result.Measurement{ID: "tls", Outcome: result.Skipped}, HTTP: result.Measurement{ID: "http", Outcome: result.Completed}, ResponseStatus: &status}
	f := findings(result.Family{Exchanges: []result.Exchange{x}, Paths: []result.Path{{Target: result.EchoSeries{Samples: []result.Echo{{RTTMS: func() *float64 { v := 500.0; return &v }()}}}}}})
	if len(f) != 1 || f[0].Type == "dominant_measured_wait" {
		t.Fatalf("invented comparison: %+v", f)
	}
}
