package httpmeasure

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/netip"
	"sync"
	"testing"
	"time"

	"hopvista/internal/result"
	"hopvista/internal/target"
)

type resolverFunc func(context.Context, string) ([]netip.Addr, error)

func (f resolverFunc) LookupIPv4(ctx context.Context, host string) ([]netip.Addr, error) {
	return f(ctx, host)
}

type remoteConn struct {
	net.Conn
	remote net.Addr
}

func (c remoteConn) RemoteAddr() net.Addr { return c.remote }

type pipeFixture struct {
	mu        sync.Mutex
	dials     []string
	hosts     []string
	responses []string
	delay     time.Duration
	peer      string
	holdAt    int
	entered   chan struct{}
	release   chan struct{}
}

func (f *pipeFixture) dial(ctx context.Context, network, address string) (net.Conn, error) {
	f.mu.Lock()
	index := len(f.dials)
	f.dials = append(f.dials, network+" "+address)
	if index >= len(f.responses) {
		f.mu.Unlock()
		return nil, errors.New("unexpected connection")
	}
	response := f.responses[index]
	peer := f.peer
	delay := f.delay
	holdAt, entered, release := f.holdAt, f.entered, f.release
	f.mu.Unlock()
	if trace := httptrace.ContextClientTrace(ctx); trace != nil {
		if trace.ConnectStart != nil {
			trace.ConnectStart(network, address)
		}
		if trace.ConnectDone != nil {
			trace.ConnectDone(network, address, nil)
		}
	}
	client, server := net.Pipe()
	go func() {
		defer server.Close()
		req, err := http.ReadRequest(bufio.NewReader(server))
		if err != nil {
			return
		}
		f.mu.Lock()
		f.hosts = append(f.hosts, req.Host)
		f.mu.Unlock()
		if entered != nil && index == holdAt {
			close(entered)
			<-release
		}
		if delay > 0 {
			time.Sleep(delay)
		}
		_, _ = io.WriteString(server, response)
	}()
	if peer == "" {
		peer, _, _ = net.SplitHostPort(address)
	}
	remote := &net.TCPAddr{IP: net.ParseIP(peer), Port: 80}
	return remoteConn{Conn: client, remote: remote}, nil
}
func (f *pipeFixture) snapshot() ([]string, []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.dials...), append([]string(nil), f.hosts...)
}

func TestPinnedHTTPAndSameHostRedirectUseFreshConnections(t *testing.T) {
	f := &pipeFixture{peer: "8.8.8.8", responses: []string{
		"HTTP/1.1 302 Found\r\nLocation: /next\r\nContent-Length: 0\r\n\r\n",
		"HTTP/1.1 503 Service Unavailable\r\nContent-Length: 0\r\n\r\n",
	}}
	lookups := 0
	selector := target.NewSelector(resolverFunc(func(_ context.Context, host string) ([]netip.Addr, error) {
		lookups++
		if host != "example.com" {
			t.Errorf("host %s", host)
		}
		return []netip.Addr{netip.MustParseAddr("8.8.8.8"), netip.MustParseAddr("1.1.1.1")}, nil
	}), nil)
	c := NewClient()
	c.dial = f.dial
	j := c.Measure(context.Background(), "http://example.com/start", selector)
	if j.Outcome != result.Completed || j.StopReason != "" || j.FinalURL != "http://example.com/next" || j.FinalHost != "example.com" || len(j.Exchanges) != 2 || lookups != 1 {
		t.Fatalf("journey: %+v; lookups %d", j, lookups)
	}
	if len(j.DNS) != 1 || len(j.DNS[0].Addresses) != 2 || j.DNS[0].SelectedIP != "8.8.8.8" {
		t.Fatalf("DNS: %+v", j.DNS)
	}
	dials, hosts := f.snapshot()
	if len(dials) != 2 || len(hosts) != 2 {
		t.Fatalf("dials %v hosts %v", dials, hosts)
	}
	for i, x := range j.Exchanges {
		if dials[i] != "tcp4 8.8.8.8:80" || hosts[i] != "example.com" || x.SelectedIP != "8.8.8.8" || x.ActualPeerIP != "8.8.8.8" || x.ConnectionReused == nil || *x.ConnectionReused || x.TCP.Outcome != result.Completed || x.TCP.DurationMS == nil || x.HTTP.Outcome != result.Completed || x.HTTP.DurationMS == nil || x.TLS.Outcome != result.Skipped || x.TLS.DurationMS != nil {
			t.Fatalf("exchange %d: %+v", i, x)
		}
		if x.DNSRef != "dns-1" {
			t.Fatal("reused pin lost DNS evidence reference")
		}
	}
	if *j.Exchanges[0].ResponseStatus != 302 || *j.Exchanges[1].ResponseStatus != 503 || j.Exchanges[1].DNS.Outcome != result.Skipped {
		t.Fatalf("statuses: %+v", j.Exchanges)
	}
}

func TestProgressRetainsDNSAndFirstRedirectDuringPendingExchange(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	defer close(release)
	f := &pipeFixture{peer: "8.8.8.8", responses: []string{
		"HTTP/1.1 302 Found\r\nLocation: /next\r\nContent-Length: 0\r\n\r\n",
		"HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n",
	}, holdAt: 1, entered: entered, release: release}
	selector := target.NewSelector(resolverFunc(func(_ context.Context, _ string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("8.8.8.8")}, nil
	}), nil)
	c := NewClient()
	c.dial = f.dial
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	updates := make(chan Journey, 8)
	done := make(chan Journey, 1)
	go func() {
		done <- c.MeasureWithProgress(ctx, "http://example.com/start", selector, func(j Journey) { updates <- j })
	}()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("second exchange did not start")
	}
	var seen []Journey
	for {
		select {
		case snapshot := <-updates:
			seen = append(seen, snapshot)
		default:
			goto collected
		}
	}
collected:
	if len(seen) < 2 || len(seen[0].DNS) != 1 || len(seen[0].Exchanges) != 0 || seen[0].SelectedIP != "8.8.8.8" || len(seen[1].Exchanges) != 1 || seen[1].Exchanges[0].ResponseStatus == nil || *seen[1].Exchanges[0].ResponseStatus != 302 {
		t.Fatalf("partial HTTP progress: %+v", seen)
	}
	cancel()
	select {
	case j := <-done:
		if j.Outcome != result.Cancelled || len(j.Exchanges) != 2 || len(seen[0].Exchanges) != 0 || len(seen[1].Exchanges) != 1 {
			t.Fatalf("cancelled journey or mutated snapshots: %+v %+v", j, seen)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("cancelled HTTP exchange did not return")
	}
}

func TestSensitiveRedirectStopsBeforeDial(t *testing.T) {
	f := &pipeFixture{peer: "8.8.8.8", responses: []string{"HTTP/1.1 302 Found\r\nLocation: http://private.example/\r\nContent-Length: 0\r\n\r\n"}}
	lookups := map[string]int{}
	selector := target.NewSelector(resolverFunc(func(_ context.Context, host string) ([]netip.Addr, error) {
		lookups[host]++
		if host == "private.example" {
			return []netip.Addr{netip.MustParseAddr("10.0.0.5")}, nil
		}
		return []netip.Addr{netip.MustParseAddr("8.8.8.8")}, nil
	}), nil)
	c := NewClient()
	c.dial = f.dial
	j := c.Measure(context.Background(), "http://public.example/", selector)
	dials, _ := f.snapshot()
	if j.StopReason != "redirect_requires_new_run" || j.Outcome != result.Skipped || len(j.Exchanges) != 1 || len(dials) != 1 || lookups["private.example"] != 1 || len(j.DNS) != 2 || j.DNS[1].SelectedIP != "10.0.0.5" || j.Exchanges[0].RedirectStopReason != "redirect_requires_new_run" {
		t.Fatalf("unsafe redirect: %+v dials %v lookups %v", j, dials, lookups)
	}
}

func TestPublicNewHostRedirectGetsNewPin(t *testing.T) {
	f := &pipeFixture{responses: []string{
		"HTTP/1.1 302 Found\r\nLocation: http://new.example/next\r\nContent-Length: 0\r\n\r\n",
		"HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n",
	}}
	selector := target.NewSelector(resolverFunc(func(_ context.Context, host string) ([]netip.Addr, error) {
		if host == "start.example" {
			return []netip.Addr{netip.MustParseAddr("8.8.8.8")}, nil
		}
		if host == "new.example" {
			return []netip.Addr{netip.MustParseAddr("1.1.1.1")}, nil
		}
		return nil, errors.New("unexpected hostname")
	}), nil)
	c := NewClient()
	c.dial = f.dial
	j := c.Measure(context.Background(), "http://start.example/", selector)
	dials, hosts := f.snapshot()
	if j.Outcome != result.Completed || j.FinalHost != "new.example" || len(j.DNS) != 2 || len(j.Exchanges) != 2 || len(dials) != 2 || dials[0] != "tcp4 8.8.8.8:80" || dials[1] != "tcp4 1.1.1.1:80" || hosts[0] != "start.example" || hosts[1] != "new.example" || j.Exchanges[1].DNS.Outcome != result.Completed {
		t.Fatalf("new host was not repinned: %+v dials %v hosts %v", j, dials, hosts)
	}
}

func TestRedirectLimitStopsAtFourExchanges(t *testing.T) {
	f := &pipeFixture{responses: []string{
		"HTTP/1.1 302 Found\r\nLocation: /one\r\nContent-Length: 0\r\n\r\n",
		"HTTP/1.1 302 Found\r\nLocation: /two\r\nContent-Length: 0\r\n\r\n",
		"HTTP/1.1 302 Found\r\nLocation: /three\r\nContent-Length: 0\r\n\r\n",
		"HTTP/1.1 302 Found\r\nLocation: /four\r\nContent-Length: 0\r\n\r\n",
	}}
	selector := target.NewSelector(resolverFunc(func(context.Context, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("8.8.8.8")}, nil
	}), nil)
	c := NewClient()
	c.dial = f.dial
	j := c.Measure(context.Background(), "http://example.com/", selector)
	dials, _ := f.snapshot()
	if j.StopReason != "redirect_limit" || j.Outcome != result.Skipped || len(j.Exchanges) != 4 || len(dials) != 4 || j.Exchanges[3].RedirectStopReason != "redirect_limit" {
		t.Fatalf("redirect limit: %+v dials %v", j, dials)
	}
}

func TestInvalidRedirectStopsWithoutAnotherDial(t *testing.T) {
	f := &pipeFixture{peer: "8.8.8.8", responses: []string{"HTTP/1.1 302 Found\r\nLocation: file:///etc/passwd\r\nContent-Length: 0\r\n\r\n"}}
	lookups := 0
	selector := target.NewSelector(resolverFunc(func(context.Context, string) ([]netip.Addr, error) {
		lookups++
		return []netip.Addr{netip.MustParseAddr("8.8.8.8")}, nil
	}), nil)
	c := NewClient()
	c.dial = f.dial
	j := c.Measure(context.Background(), "http://example.com/", selector)
	dials, _ := f.snapshot()
	if j.Outcome != result.Skipped || j.StopReason != "invalid_redirect" || len(j.Exchanges) != 1 || j.Exchanges[0].ResponseStatus == nil || *j.Exchanges[0].ResponseStatus != 302 || j.Exchanges[0].RedirectStopReason != "invalid_redirect" || len(dials) != 1 || lookups != 1 {
		t.Fatalf("invalid redirect was followed or misclassified: %+v dials=%v lookups=%d", j, dials, lookups)
	}
}

func TestConnectFailureHasNoNumericTiming(t *testing.T) {
	selector := target.NewSelector(resolverFunc(func(context.Context, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("8.8.8.8")}, nil
	}), nil)
	c := NewClient()
	c.dial = func(ctx context.Context, network, address string) (net.Conn, error) {
		err := errors.New("controlled connect refusal")
		if trace := httptrace.ContextClientTrace(ctx); trace != nil {
			trace.ConnectStart(network, address)
			trace.ConnectDone(network, address, err)
		}
		return nil, err
	}
	j := c.Measure(context.Background(), "http://example.com/", selector)
	if j.StopReason != "tcp_connect_failed" || j.Outcome != result.Error || len(j.Exchanges) != 1 || j.Exchanges[0].TCP.DurationMS != nil || j.Exchanges[0].TCP.Error == nil || j.Exchanges[0].TCP.Error.Code != "tcp_connect_failed" || j.Exchanges[0].HTTP.DurationMS != nil || j.Exchanges[0].ResponseStatus != nil {
		t.Fatalf("connect failure: %+v", j)
	}
}

func TestPinnedDialRejectsUnexpectedPeerBeforeRequest(t *testing.T) {
	f := &pipeFixture{peer: "1.1.1.1", responses: []string{"HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n"}}
	selector := target.NewSelector(resolverFunc(func(context.Context, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("8.8.8.8")}, nil
	}), nil)
	c := NewClient()
	c.dial = f.dial
	j := c.Measure(context.Background(), "http://example.com/", selector)
	_, hosts := f.snapshot()
	if j.StopReason != "peer_mismatch" || len(hosts) != 0 || len(j.Exchanges) != 1 || j.Exchanges[0].HTTP.DurationMS != nil {
		t.Fatalf("peer mismatch was not stopped before HTTP: %+v hosts %v", j, hosts)
	}
}

func TestHeaderTimeoutHasNoFabricatedFirstByte(t *testing.T) {
	f := &pipeFixture{peer: "8.8.8.8", delay: 100 * time.Millisecond, responses: []string{"HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n"}}
	selector := target.NewSelector(resolverFunc(func(context.Context, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("8.8.8.8")}, nil
	}), nil)
	c := NewClient()
	c.dial = f.dial
	c.bounds = phaseBounds{connect: time.Second, tls: time.Second, header: 25 * time.Millisecond}
	j := c.Measure(context.Background(), "http://example.com/", selector)
	if j.StopReason != "response_header_timeout" || j.Outcome != result.Timeout || len(j.Exchanges) != 1 || j.Exchanges[0].HTTP.DurationMS != nil || j.Exchanges[0].HTTP.BoundMS == nil || *j.Exchanges[0].HTTP.BoundMS != 25 || j.Exchanges[0].TCP.Outcome != result.Completed {
		t.Fatalf("header timeout: %+v", j)
	}
}

func TestCancelledBeforeExchangeMakesNoConnection(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	lookups := 0
	selector := target.NewSelector(resolverFunc(func(context.Context, string) ([]netip.Addr, error) { lookups++; return nil, nil }), nil)
	c := NewClient()
	j := c.Measure(ctx, "http://example.com/", selector)
	if j.Outcome != result.Cancelled || j.StopReason != "cancelled" || lookups != 0 || len(j.Exchanges) != 0 {
		t.Fatalf("cancelled: %+v lookups %d", j, lookups)
	}
}
