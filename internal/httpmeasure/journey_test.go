package httpmeasure_test

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/netip"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"

	"hopvista/internal/httpmeasure"
	"hopvista/internal/result"
	"hopvista/internal/target"
)

type lookupFunc func(context.Context, string) ([]netip.Addr, error)

func (f lookupFunc) LookupIPv4(ctx context.Context, host string) ([]netip.Addr, error) {
	return f(ctx, host)
}

type fixturePeer struct {
	net.Conn
	peer net.Addr
}

func (c fixturePeer) RemoteAddr() net.Addr { return c.peer }

type requestRecord struct {
	index             int
	Authority         string    `json:"authority"`
	Path              string    `json:"path"`
	SNI               string    `json:"sni"`
	CertificateSHA256 string    `json:"certificate_sha256"`
	Dial              string    `json:"dial"`
	FirstByteAt       time.Time `json:"first_byte_at"`
	HeadersAt         time.Time `json:"headers_at"`
}

// Only DNS and the socket are simulated. Each dial creates a fresh TLS server
// over net.Pipe with a claimed RemoteAddr, never a public network connection.
type tlsJourney struct {
	mu              sync.Mutex
	cert            tls.Certificate
	roots           *x509.CertPool
	identity        string
	responses       []string
	records         []requestRecord
	dials           []string
	lookups         []string
	events          []string
	complete        chan struct{}
	headerGap       time.Duration
	requestReceived chan struct{}
}

func newTLSJourney(t *testing.T, responses ...string) *tlsJourney {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "HopVista controlled fixture"}, DNSNames: []string{"start.example", "next.example"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	leaf, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(leaf)
	hash := sha256.Sum256(der)
	return &tlsJourney{cert: tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}, roots: roots, identity: hex.EncodeToString(hash[:]), responses: responses, complete: make(chan struct{}, 8)}
}

func (f *tlsJourney) resolve(_ context.Context, host string) ([]netip.Addr, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lookups = append(f.lookups, host)
	f.events = append(f.events, "resolve "+host)
	switch host {
	case "start.example":
		return []netip.Addr{netip.MustParseAddr("8.8.8.8"), netip.MustParseAddr("1.0.0.1")}, nil
	case "next.example", "wrong.example":
		return []netip.Addr{netip.MustParseAddr("1.1.1.1")}, nil
	case "private.example":
		return []netip.Addr{netip.MustParseAddr("10.0.0.5")}, nil
	default:
		return nil, fmt.Errorf("unexpected DNS lookup: %s", host)
	}
}

func (f *tlsJourney) dial(ctx context.Context, network, address string) (net.Conn, error) {
	f.mu.Lock()
	index := len(f.dials)
	f.dials = append(f.dials, network+" "+address)
	f.events = append(f.events, "dial "+network+" "+address)
	if index >= len(f.responses) {
		f.mu.Unlock()
		return nil, fmt.Errorf("unexpected dial %s", address)
	}
	response := f.responses[index]
	f.mu.Unlock()
	client, server := net.Pipe()
	_ = server.SetDeadline(time.Now().Add(5 * time.Second))
	trace := httptrace.ContextClientTrace(ctx)
	if trace != nil && trace.ConnectStart != nil {
		trace.ConnectStart(network, address)
	}
	if trace != nil && trace.ConnectDone != nil {
		trace.ConnectDone(network, address, nil)
	}
	go func() {
		defer func() { server.Close(); f.complete <- struct{}{} }()
		// TLS 1.2 avoids TLS 1.3 ticket/alert writes deadlocking on net.Pipe's
		// unbuffered full-duplex socket. Real-socket TLS uses production defaults.
		conn := tls.Server(server, &tls.Config{Certificates: []tls.Certificate{f.cert}, MinVersion: tls.VersionTLS12, MaxVersion: tls.VersionTLS12})
		if conn.Handshake() != nil {
			return
		}
		req, err := http.ReadRequest(bufio.NewReader(conn))
		if err != nil {
			return
		}
		if response == "" {
			if f.requestReceived != nil {
				close(f.requestReceived)
			}
			// Hold all response bytes until the client times out or cancels.
			_, _ = io.Copy(io.Discard, conn)
			return
		}
		record := requestRecord{index: index, Authority: req.Host, Path: req.URL.RequestURI(), SNI: conn.ConnectionState().ServerName, CertificateSHA256: f.identity, Dial: network + " " + address}
		record.FirstByteAt = time.Now().UTC()
		_, _ = io.WriteString(conn, response[:1])
		if f.headerGap > 0 {
			time.Sleep(f.headerGap)
		}
		_, _ = io.WriteString(conn, response[1:])
		record.HeadersAt = time.Now().UTC()
		f.mu.Lock()
		f.records = append(f.records, record)
		f.mu.Unlock()
	}()
	host, port, _ := net.SplitHostPort(address)
	portNum := 443
	fmt.Sscan(port, &portNum)
	return fixturePeer{client, &net.TCPAddr{IP: net.ParseIP(host), Port: portNum}}, nil
}

func redirect(location string) string {
	return "HTTP/1.1 302 Found\r\nLocation: " + location + "\r\nContent-Length: 0\r\n\r\n"
}

const noContent = "HTTP/1.1 204 No Content\r\nContent-Length: 0\r\n\r\n"

func (f *tlsJourney) measure(t *testing.T, url string) httpmeasure.Journey {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	j := httpmeasure.ControlledTLSClient(f.roots, f.dial).Measure(ctx, url, target.NewSelector(lookupFunc(f.resolve), nil))
	for range f.dials {
		select {
		case <-f.complete:
		case <-ctx.Done():
			t.Fatal("TLS fixture did not finish")
		}
	}
	sort.Slice(f.records, func(i, j int) bool { return f.records[i].index < f.records[j].index })
	return j
}

func writeCapture(t *testing.T, name string, value any) {
	t.Helper()
	dir := os.Getenv("HOPVISTA_JOURNEY_CAPTURE")
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".json"), append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestHTTPSRelativeSameHostAndNewHostJourney(t *testing.T) {
	f := newTLSJourney(t, redirect("/relative?step=2#ignored"), redirect("https://start.example:8443/absolute"), redirect("https://next.example/final"), noContent)
	f.headerGap = 40 * time.Millisecond
	j := f.measure(t, "https://start.example/start")
	if j.Outcome != result.Completed || j.FinalURL != "https://next.example/final" || j.FinalHost != "next.example" || j.SelectedIP != "1.1.1.1" || len(j.Exchanges) != 4 || len(j.DNS) != 2 {
		t.Fatalf("journey: %+v", j)
	}
	if !reflect.DeepEqual(f.lookups, []string{"start.example", "next.example"}) {
		t.Fatalf("unchecked/repeated DNS: %v", f.lookups)
	}
	wantAuthorities := []string{"start.example", "start.example", "start.example:8443", "next.example"}
	wantPaths := []string{"/start", "/relative?step=2", "/absolute", "/final"}
	wantDials := []string{"tcp4 8.8.8.8:443", "tcp4 8.8.8.8:443", "tcp4 8.8.8.8:8443", "tcp4 1.1.1.1:443"}
	if !reflect.DeepEqual(f.dials, wantDials) || len(f.records) != 4 {
		t.Fatalf("dials=%v records=%+v", f.dials, f.records)
	}
	wantEvents := []string{"resolve start.example", "dial tcp4 8.8.8.8:443", "dial tcp4 8.8.8.8:443", "dial tcp4 8.8.8.8:8443", "resolve next.example", "dial tcp4 1.1.1.1:443"}
	if !reflect.DeepEqual(f.events, wantEvents) {
		t.Fatalf("new-host resolution must precede its pinned dial: %v", f.events)
	}
	for i, x := range j.Exchanges {
		r := f.records[i]
		if r.Authority != wantAuthorities[i] || r.Path != wantPaths[i] || r.SNI != x.Hostname || r.CertificateSHA256 != f.identity || x.SelectedIP != x.ActualPeerIP || x.ConnectionReused == nil || *x.ConnectionReused || x.RedirectIndex != i || x.SelectedScope != "public" {
			t.Fatalf("exchange %d identity: %+v %+v", i, x, r)
		}
		if x.ResponseStatus == nil || (i < 3 && *x.ResponseStatus != 302) || (i == 3 && *x.ResponseStatus != 204) {
			t.Fatalf("status: %+v", x)
		}
		for _, phase := range []result.Measurement{x.TCP, x.TLS, x.HTTP} {
			if phase.Outcome != result.Completed || phase.Method == "" || phase.BoundMS == nil || phase.StartedAt == nil || phase.EndedAt == nil || phase.DurationMS == nil || phase.EndedAt.Before(*phase.StartedAt) {
				t.Fatalf("phase: %+v", phase)
			}
			if i > 0 && phase.StartedAt.Before(*j.Exchanges[i-1].HTTP.EndedAt) {
				t.Fatalf("phase belongs to earlier exchange: %+v", phase)
			}
		}
		if x.TCP.Method != "http_transport_connect" || *x.TCP.BoundMS != 2000 || x.TLS.Method != "http_transport_tls" || *x.TLS.BoundMS != 2000 || x.HTTP.Method != "request_written_to_first_response_header_byte" || *x.HTTP.BoundMS != 3000 {
			t.Fatalf("phase method/bound: %+v", x)
		}
		if x.HTTP.EndedAt.Before(r.FirstByteAt) || !x.HTTP.EndedAt.Before(r.HeadersAt) {
			t.Fatalf("first byte was measured at complete headers: %+v %+v", x.HTTP, r)
		}
		if (i == 1 || i == 2) && (x.DNS.Outcome != result.Skipped || x.DNS.Method != "reused_selection" || x.DNS.DurationMS != nil || x.DNS.StartedAt != nil || x.DNS.BoundMS != nil || x.DNSRef != "dns-1") {
			t.Fatalf("same-host selection: %+v", x)
		}
		if i == 0 || i == 3 {
			if x.DNS.Outcome != result.Completed || x.DNS.DurationMS == nil || x.DNS.StartedAt == nil || x.DNS.EndedAt == nil || x.DNS.BoundMS == nil {
				t.Fatalf("new host DNS: %+v", x)
			}
		}
	}
	writeCapture(t, "https-journey", struct {
		Journey  httpmeasure.Journey
		Requests []requestRecord
	}{j, f.records})
}
