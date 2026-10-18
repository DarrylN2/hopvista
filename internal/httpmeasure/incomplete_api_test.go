package httpmeasure_test

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http/httptrace"
	"net/netip"
	"strings"
	"testing"
	"time"

	"hopvista/internal/engine"
	"hopvista/internal/httpmeasure"
	"hopvista/internal/result"
	"hopvista/internal/run"
)

func TestIncompleteMeasurementsThroughAPI(t *testing.T) {
	for _, tc := range []struct {
		name, code, phase, phrase string
		outcome                   result.Outcome
	}{
		{"dnsfailure", "dns_lookup_failed", "DNS", "lookup failed", result.Error},
		{"dnstimeout", "dns_timeout", "DNS", "lookup", result.Timeout},
		{"noipv4", "ipv4_unavailable", "DNS", "No usable IPv4", result.Unavailable},
		{"tcpfailure", "tcp_connect_failed", "TCP", "TCP", result.Error},
		{"tcptimeout", "tcp_connect_timeout", "TCP", "TCP", result.Timeout},
		{"tlsfailure", "tls_handshake_failed", "TLS", "TLS", result.Error},
		{"tlstimeout", "tls_handshake_timeout", "TLS", "TLS", result.Timeout},
		{"headertimeout", "response_header_timeout", "HTTP", "response", result.Timeout},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newTLSJourney(t, redirect("/failed"), noContent)
			if tc.phase == "HTTP" {
				f.responses[1] = ""
			}
			defer func() {
				for range f.dials {
					select {
					case <-f.complete:
					case <-time.After(6 * time.Second):
						t.Error("TLS fixture did not finish")
					}
				}
			}()
			resolver := lookupFunc(f.resolve)
			if tc.phase == "DNS" {
				resolver = func(context.Context, string) ([]netip.Addr, error) {
					switch tc.name {
					case "dnsfailure":
						return nil, errors.New("controlled lookup failure")
					case "dnstimeout":
						return nil, &net.DNSError{IsTimeout: true}
					default:
						return []netip.Addr{netip.MustParseAddr("2001:db8::1"), netip.MustParseAddr("::ffff:192.0.2.1"), netip.MustParseAddr("0.0.0.0")}, nil
					}
				}
			}
			attempt := 0
			dial := func(ctx context.Context, network, address string) (net.Conn, error) {
				attempt++
				if network != "tcp4" || address != "8.8.8.8:443" {
					t.Errorf("unexpected network work: %s %s", network, address)
				}
				if attempt == 1 {
					return f.dial(ctx, network, address)
				}
				if tc.phase == "HTTP" {
					return f.dial(ctx, network, address)
				}
				trace := httptrace.ContextClientTrace(ctx)
				trace.ConnectStart(network, address)
				if tc.phase == "TCP" {
					var err error = errors.New("controlled connection refusal")
					if tc.name == "tcptimeout" {
						err = &net.DNSError{IsTimeout: true}
					}
					trace.ConnectDone(network, address, err)
					return nil, err
				}
				client, server := net.Pipe()
				trace.ConnectDone(network, address, nil)
				go func() {
					defer server.Close()
					if tc.name == "tlstimeout" {
						_, _ = io.Copy(io.Discard, server)
					}
				}()
				return fixturePeer{client, &net.TCPAddr{IP: net.ParseIP("8.8.8.8"), Port: 443}}, nil
			}
			clock := &journeyClock{}
			adapter := journeyAdapter(clock, "8.8.8.8")
			e := &engine.Engine{HTTP: httpmeasure.ControlledTLSClient(f.roots, dial), Resolver: resolver, Probe: adapter, Clock: clock}
			snapshot := runEngineThroughAPI(t, run.NewManager(e), "https://start.example/start", nil)
			family := snapshot.Families[0]
			if snapshot.State != result.RunCompleted || family.Outcome != tc.outcome || family.Error == nil || family.Error.Code != tc.code || !strings.Contains(family.Error.Message, tc.phrase) {
				t.Fatalf("stop not explained: %+v", family)
			}
			var stopped result.Measurement
			if tc.phase == "DNS" {
				if attempt != 0 || len(adapter.Calls()) != 0 || len(family.Exchanges) != 0 || len(family.DNS) != 1 {
					t.Fatalf("fabricated IPv4 work: %+v", family)
				}
				stopped = family.DNS[0].Measurement
				if tc.name != "noipv4" && (stopped.Outcome != tc.outcome || stopped.DurationMS != nil || stopped.Error == nil || stopped.Error.Code != tc.code || stopped.StartedAt == nil || stopped.EndedAt == nil) {
					t.Fatalf("failed lookup evidence: %+v", stopped)
				}
				if tc.name == "noipv4" && (stopped.Outcome != result.Completed || stopped.DurationMS == nil || family.DNS[0].SelectedIP != "") {
					t.Fatalf("completed no-A lookup: %+v", family.DNS)
				}
			} else {
				if attempt != 2 || len(family.Exchanges) != 2 || len(family.DNS) != 1 {
					t.Fatalf("lost partial evidence: %+v", family)
				}
				earlier, x := family.Exchanges[0], family.Exchanges[1]
				if earlier.HTTP.Outcome != result.Completed || earlier.ResponseStatus == nil || *earlier.ResponseStatus != 302 || earlier.TLS.DurationMS == nil || x.DNS.Outcome != result.Skipped || x.DNS.DurationMS != nil || x.DNSRef != "dns-1" {
					t.Fatalf("lost completed exchange/pin: %+v", family)
				}
				switch tc.phase {
				case "TCP":
					stopped = x.TCP
				case "TLS":
					stopped = x.TLS
				default:
					stopped = x.HTTP
				}
				if stopped.Outcome != tc.outcome || stopped.DurationMS != nil || stopped.StartedAt == nil || stopped.Error == nil || stopped.Error.Code != tc.code || x.ResponseStatus != nil {
					t.Fatalf("stopped phase: %+v exchange=%+v", stopped, x)
				}
				if tc.phase != "HTTP" && (x.HTTP.Outcome != result.Unavailable || x.HTTP.StartedAt != nil || x.HTTP.DurationMS != nil) {
					t.Fatalf("unattempted HTTP: %+v", x.HTTP)
				}
				if tc.phase == "TLS" && (x.ActualPeerIP != "8.8.8.8" || x.TCP.Outcome != result.Completed || x.TCP.DurationMS == nil) {
					t.Fatalf("lost TCP: %+v", x)
				}
			}
			if stopped.BoundMS == nil || *stopped.BoundMS <= 0 || *stopped.BoundMS > 3000 {
				t.Fatalf("missing bound: %+v", stopped)
			}
			if len(snapshot.Findings) != 1 || len(snapshot.Findings[0].EvidenceRefs) != 1 || snapshot.Findings[0].EvidenceRefs[0] != stopped.ID || snapshot.Findings[0].Limit == "" {
				t.Fatalf("finding evidence: %+v", snapshot.Findings)
			}
			writeCapture(t, "run-"+tc.name, snapshot)
		})
	}
}
