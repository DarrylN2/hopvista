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
