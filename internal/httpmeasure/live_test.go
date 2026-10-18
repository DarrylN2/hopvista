package httpmeasure

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"sync"
	"testing"
	"time"

	"hopvista/internal/result"
	"hopvista/internal/target"
)

func liveHTTP(t *testing.T) {
	t.Helper()
	if os.Getenv("HOPVISTA_LIVE_HTTP") != "1" {
		t.Skip("opt-in loopback HTTP fixture")
	}
}
func loopbackSelector() *target.Selector {
	return target.NewSelector(resolverFunc(func(context.Context, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("127.0.0.1")}, nil
	}), nil)
}

func TestLivePinnedHTTPAndHeadersOnly(t *testing.T) {
	liveHTTP(t)
	var mu sync.Mutex
	hosts := []string{}
	done := make(chan struct{})
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		hosts = append(hosts, r.Host)
		mu.Unlock()
		if r.URL.Path == "/headers" {
			w.WriteHeader(200)
			w.(http.Flusher).Flush()
			<-done
			return
		}
		w.WriteHeader(503)
	}))
	defer s.Close()
	defer close(done)
	_, port, err := net.SplitHostPort(s.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	c := NewClient()
	start := time.Now()
	j := c.Measure(context.Background(), "http://example.com:"+port+"/headers", loopbackSelector())
	if time.Since(start) > time.Second || j.Outcome != result.Completed || len(j.Exchanges) != 1 || j.Exchanges[0].TCP.Outcome != result.Completed || j.Exchanges[0].HTTP.Outcome != result.Completed || j.Exchanges[0].ActualPeerIP != "127.0.0.1" || *j.Exchanges[0].ResponseStatus != 200 {
		t.Fatalf("headers-only measurement: %+v", j)
	}
	j = c.Measure(context.Background(), "http://example.com:"+port+"/status", loopbackSelector())
	if j.Outcome != result.Completed || *j.Exchanges[0].ResponseStatus != 503 || j.Exchanges[0].ConnectionReused == nil || *j.Exchanges[0].ConnectionReused {
		t.Fatalf("503/new connection: %+v", j)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(hosts) != 2 || hosts[0] != "example.com:"+port || hosts[1] != "example.com:"+port {
		t.Fatalf("authority changed: %v", hosts)
	}
}

func TestLivePinnedTLSIdentityAndSNI(t *testing.T) {
	liveHTTP(t)
	var mu sync.Mutex
	snis := []string{}
	hosts := []string{}
	s := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		hosts = append(hosts, r.Host)
		mu.Unlock()
		w.WriteHeader(204)
	}))
	s.TLS = &tls.Config{GetConfigForClient: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
		mu.Lock()
		snis = append(snis, hello.ServerName)
		mu.Unlock()
		return nil, nil
	}}
	s.StartTLS()
	defer s.Close()
	cert := s.Certificate()
	if len(cert.DNSNames) == 0 {
		t.Fatal("TLS fixture has no DNS SAN")
	}
	if len(cert.IPAddresses) == 0 {
		t.Fatal("TLS fixture has no IP SAN for literal verification")
	}
	host := cert.DNSNames[0]
	_, port, err := net.SplitHostPort(s.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(cert)
	c := NewClient()
	c.Roots = roots
	j := c.Measure(context.Background(), "https://"+host+":"+port+"/", loopbackSelector())
	if j.Outcome != result.Completed || len(j.Exchanges) != 1 || j.Exchanges[0].TLS.Outcome != result.Completed || j.Exchanges[0].TLS.DurationMS == nil || j.Exchanges[0].TCP.Outcome != result.Completed || j.Exchanges[0].HTTP.Outcome != result.Completed || j.Exchanges[0].ActualPeerIP != "127.0.0.1" || *j.Exchanges[0].ResponseStatus != 204 {
		t.Fatalf("trusted TLS: %+v", j)
	}
	wrong := c.Measure(context.Background(), "https://wrong.example:"+port+"/", loopbackSelector())
	if wrong.StopReason != "tls_verification_failed" || wrong.Exchanges[0].TLS.Error == nil || wrong.Exchanges[0].TLS.Error.Code != "tls_verification_failed" || wrong.Exchanges[0].HTTP.DurationMS != nil || wrong.Exchanges[0].TLS.DurationMS != nil {
		t.Fatalf("wrong-name certificate accepted: %+v", wrong)
	}
	untrusted := NewClient().Measure(context.Background(), "https://"+host+":"+port+"/", loopbackSelector())
	if untrusted.StopReason != "tls_verification_failed" || untrusted.Exchanges[0].HTTP.DurationMS != nil {
		t.Fatalf("untrusted certificate accepted: %+v", untrusted)
	}
	mu.Lock()
	if len(hosts) != 1 || hosts[0] != host+":"+port || len(snis) < 3 || snis[0] != host || snis[1] != "wrong.example" || snis[2] != host {
		mu.Unlock()
		t.Fatalf("Host/SNI mismatch: hosts %v, SNI %v", hosts, snis)
	}
	mu.Unlock()
	if len(cert.IPAddresses) > 0 {
		ip := cert.IPAddresses[0].String()
		literal := c.Measure(context.Background(), "https://"+ip+":"+port+"/", target.NewSelector(nil, nil))
		if literal.Outcome != result.Completed {
			t.Fatalf("IP SAN did not verify: %+v", literal)
		}
		mu.Lock()
		lastSNI := snis[len(snis)-1]
		mu.Unlock()
		if lastSNI != "" {
			t.Fatalf("IP literal sent SNI %q", lastSNI)
		}
	}
}

func TestLivePinnedHTTPSRedirectJourney(t *testing.T) {
	liveHTTP(t)
	var mu sync.Mutex
	var authorities, snis []string
	s := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		authorities = append(authorities, r.Host)
		snis = append(snis, r.TLS.ServerName)
		mu.Unlock()
		if r.URL.Path == "/start" {
			w.Header().Set("Location", "/final")
			w.WriteHeader(http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	s.StartTLS()
	defer s.Close()
	host := s.Certificate().DNSNames[0]
	_, port, err := net.SplitHostPort(s.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(s.Certificate())
	c := NewClient()
	c.Roots = roots
	j := c.Measure(context.Background(), "https://"+host+":"+port+"/start", loopbackSelector())
	if j.Outcome != result.Completed || j.FinalHost != host || len(j.Exchanges) != 2 || len(j.DNS) != 1 {
		t.Fatalf("real TCP HTTPS journey: %+v", j)
	}
	for _, x := range j.Exchanges {
		if x.SelectedIP != "127.0.0.1" || x.ActualPeerIP != "127.0.0.1" || x.ConnectionReused == nil || *x.ConnectionReused || x.TCP.Outcome != result.Completed || x.TLS.Outcome != result.Completed || x.HTTP.Outcome != result.Completed {
			t.Fatalf("real TCP evidence: %+v", x)
		}
	}
	if j.Exchanges[1].DNS.Outcome != result.Skipped || j.Exchanges[1].DNS.DurationMS != nil || j.Exchanges[1].DNSRef != "dns-1" {
		t.Fatalf("same-host pin: %+v", j.Exchanges[1])
	}
	mu.Lock()
	defer mu.Unlock()
	if len(authorities) != 2 || len(snis) != 2 {
		t.Fatalf("requests: %v %v", authorities, snis)
	}
	for i := range authorities {
		if authorities[i] != host+":"+port || snis[i] != host {
			t.Fatalf("authority/SNI: %v %v", authorities, snis)
		}
	}
}
