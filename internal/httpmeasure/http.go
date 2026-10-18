// Package httpmeasure measures direct, pinned IPv4 HTTP exchanges. It never
// follows a redirect inside net/http; each next host passes target policy first.
package httpmeasure

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"hopvista/internal/result"
	"hopvista/internal/target"
)

type Journey struct {
	Outcome    result.Outcome
	DNS        []result.DNS
	Exchanges  []result.Exchange
	StopReason string
	FinalURL   string
	FinalHost  string
	SelectedIP string
}

type Client struct {
	// Roots is only for controlled trust fixtures. Nil uses the platform roots.
	Roots  *x509.CertPool
	dial   func(context.Context, string, string) (net.Conn, error)
	bounds phaseBounds
}

type phaseBounds struct{ connect, tls, header time.Duration }

var productionBounds = phaseBounds{connect: 2 * time.Second, tls: 2 * time.Second, header: 3 * time.Second}
var errPeerMismatch = errors.New("actual peer differs from selected IPv4 address")

func NewClient() *Client { return &Client{} }

// Measure performs at most four sequential HTTP exchanges. A returned
// StopReason is a partial outcome, never a claim about general reachability.
func (c *Client) Measure(ctx context.Context, rawURL string, selector *target.Selector) Journey {
	return c.MeasureWithProgress(ctx, rawURL, selector, nil)
}

// MeasureWithProgress synchronously publishes detached snapshots after DNS
// selection and each completed exchange. A pending exchange is not presented
// as completed evidence; the previous observations stay visible until return.
func (c *Client) MeasureWithProgress(ctx context.Context, rawURL string, selector *target.Selector, publish func(Journey)) Journey {
	j := Journey{DNS: []result.DNS{}, Exchanges: []result.Exchange{}, Outcome: result.Pending}
	emit := func() {
		if publish == nil {
			return
		}
		snapshot := j
		snapshot.DNS = append([]result.DNS{}, j.DNS...)
		snapshot.Exchanges = append([]result.Exchange{}, j.Exchanges...)
		publish(snapshot)
	}
	if selector == nil {
		j.StopReason = "selector_unavailable"
		j.Outcome = result.Unavailable
		return j
	}
	u, problem := target.Parse(rawURL)
	if problem != nil {
		j.StopReason = problem.Code
		j.Outcome = selectionOutcome(problem.Code)
		return j
	}
	selected, dns, problem := selector.Select(ctx, u)
	if dns != nil {
		j.DNS = append(j.DNS, *dns)
	}
	if selected.IP().IsValid() {
		j.SelectedIP = selected.IP().String()
	}
	emit()
	if problem != nil {
		j.StopReason = problem.Code
		j.Outcome = selectionOutcome(problem.Code)
		return j
	}
	for index := 0; index <= target.MaxRedirects; index++ {
		if ctx.Err() != nil {
			j.StopReason = contextCode(ctx)
			j.Outcome = contextOutcome(ctx)
			return j
		}
		x := c.exchange(ctx, index, selected, dns)
		j.Exchanges = append(j.Exchanges, x)
		j.SelectedIP = selected.IP().String()
		emit()
		if x.HTTP.Outcome != result.Completed {
			j.StopReason = phaseStopReason(x)
			j.Outcome = x.HTTP.Outcome
			if j.Outcome == result.Unavailable {
				if x.TCP.Outcome == result.Error || x.TCP.Outcome == result.Timeout || x.TCP.Outcome == result.Cancelled || x.TCP.Outcome == result.Deadline {
					j.Outcome = x.TCP.Outcome
				}
				if x.TLS.Outcome == result.Error || x.TLS.Outcome == result.Timeout || x.TLS.Outcome == result.Cancelled || x.TLS.Outcome == result.Deadline {
					j.Outcome = x.TLS.Outcome
				}
			}
			return j
		}
		if x.ResponseStatus == nil || !redirectStatus(*x.ResponseStatus) {
			j.Outcome, j.FinalURL, j.FinalHost = result.Completed, selected.URL().String(), selected.URL().Hostname()
			return j
		}
		decision := selector.NextRedirect(ctx, selected, x.Location, index)
		if decision.DNS != nil {
			j.DNS = append(j.DNS, *decision.DNS)
			emit()
		}
		if decision.Next == nil {
			j.StopReason = decision.StopReason
			j.Exchanges[len(j.Exchanges)-1].RedirectStopReason = decision.StopReason
			j.Outcome = redirectStopOutcome(decision.StopReason)
			return j
		}
		selected = *decision.Next
		dns = decision.DNS
	}
	panic("redirect loop exceeded checked limit")
}

func selectionOutcome(code string) result.Outcome {
	switch code {
	case "cancelled":
		return result.Cancelled
	case "deadline":
		return result.Deadline
	case "dns_timeout":
		return result.Timeout
	case "ipv4_unavailable", "ipv6_not_supported_in_v1":
		return result.Unavailable
	default:
		return result.Error
	}
}
func redirectStopOutcome(code string) result.Outcome {
	switch code {
	case "redirect_requires_new_run", "redirect_limit", "invalid_redirect":
		return result.Skipped
	default:
		return selectionOutcome(code)
	}
}
func contextOutcome(ctx context.Context) result.Outcome {
	if ctx.Err() == context.DeadlineExceeded {
		return result.Deadline
	}
	return result.Cancelled
}
func contextCode(ctx context.Context) string {
	if ctx.Err() == context.DeadlineExceeded {
		return "deadline"
	}
	return "cancelled"
}
func redirectStatus(status int) bool {
	return status == 301 || status == 302 || status == 303 || status == 307 || status == 308
}

func (c *Client) exchange(ctx context.Context, index int, selected target.Selection, dns *result.DNS) result.Exchange {
	u := selected.URL()
	id := "exchange-" + strconv.Itoa(index+1)
	x := result.Exchange{ID: id, RedirectIndex: index, RequestedURL: u.String(), Hostname: u.Hostname(), Port: u.Port(), SelectedIP: selected.IP().String(), SelectedScope: string(selected.Scope()), DNSRef: selected.DNSRef()}
	x.DNS = result.Measurement{ID: id + "-dns", Outcome: result.Skipped, Method: "reused_selection"}
	if dns != nil {
		x.DNS = dns.Measurement
		x.DNS.ID = id + "-dns"
	}
	x.TCP = result.Measurement{ID: id + "-tcp", Outcome: result.Unavailable, Method: "http_transport_connect"}
	x.TLS = result.Measurement{ID: id + "-tls", Outcome: result.Unavailable, Method: "http_transport_tls"}
	x.HTTP = result.Measurement{ID: id + "-http", Outcome: result.Unavailable, Method: "request_written_to_first_response_header_byte"}
	bounds := c.bounds
	if bounds.connect == 0 {
		bounds = productionBounds
	}
	connectBound, tlsBound, headerBound := ms(bounds.connect), ms(bounds.tls), ms(bounds.header)
	x.TCP.BoundMS = &connectBound
	x.TLS.BoundMS = &tlsBound
	x.HTTP.BoundMS = &headerBound
	if strings.HasPrefix(u.String(), "http:") {
		x.TLS.Outcome = result.Skipped
		x.TLS.Method = "http_not_tls"
		x.TLS.BoundMS = nil
	}
	if ctx.Err() != nil {
		x.TCP.Outcome = contextOutcome(ctx)
		x.HTTP.Outcome = contextOutcome(ctx)
		return x
	}

	recorder := &traceRecorder{}
	trace := &httptrace.ClientTrace{
		ConnectStart:         func(_, _ string) { recorder.setConnectStart() },
		ConnectDone:          func(_, _ string, err error) { recorder.setConnectDone(err) },
		TLSHandshakeStart:    func() { recorder.setTLSStart() },
		TLSHandshakeDone:     func(_ tls.ConnectionState, err error) { recorder.setTLSDone(err) },
		WroteRequest:         func(info httptrace.WroteRequestInfo) { recorder.setWritten(info.Err) },
		GotFirstResponseByte: func() { recorder.setFirstByte() },
		GotConn:              func(info httptrace.GotConnInfo) { recorder.setGotConn(info) },
	}
	request, err := http.NewRequestWithContext(httptrace.WithClientTrace(ctx, trace), http.MethodGet, u.String(), nil)
	if err != nil {
		x.HTTP.Outcome = result.Error
		x.HTTP.Error = &result.ErrorData{Code: "invalid_http_request", Message: "The HTTP request could not be prepared."}
		return x
	}
	tlsConfig := &tls.Config{RootCAs: c.Roots, ServerName: u.Hostname(), MinVersion: tls.VersionTLS12}
	transport := &http.Transport{
		Proxy: nil, DisableKeepAlives: true, ForceAttemptHTTP2: false,
		TLSClientConfig: tlsConfig, TLSHandshakeTimeout: bounds.tls, ResponseHeaderTimeout: bounds.header,
		DialContext: func(dialCtx context.Context, network, authority string) (net.Conn, error) {
			if network != "tcp" {
				return nil, errors.New("unexpected dial network")
			}
			host, port, err := net.SplitHostPort(authority)
			if err != nil || !strings.EqualFold(host, u.Hostname()) || port != strconv.Itoa(u.Port()) {
				return nil, errors.New("unexpected dial authority")
			}
			address := net.JoinHostPort(selected.IP().String(), port)
			dial := c.dial
			if dial == nil {
				d := &net.Dialer{Timeout: bounds.connect}
				dial = d.DialContext
			}
			conn, err := dial(dialCtx, "tcp4", address)
			if err != nil {
				return nil, err
			}
			peer, ok := peerIPv4(conn.RemoteAddr())
			if !ok || peer != selected.IP() {
				conn.Close()
				return nil, errPeerMismatch
			}
			recorder.setPeer(peer.String())
			return conn, nil
		},
	}
	defer transport.CloseIdleConnections()
	response, roundErr := transport.RoundTrip(request)
	if response != nil {
		defer response.Body.Close()
	}
	rec := recorder.snapshot()
	x.ActualPeerIP = rec.peer
	if rec.gotConn {
		reused := rec.reused
		x.ConnectionReused = &reused
	}
	populatePhase(&x.TCP, rec.connectStart, rec.connectEnd, rec.connectErr, roundErr, ctx)
	if x.TLS.Outcome != result.Skipped {
		populatePhase(&x.TLS, rec.tlsStart, rec.tlsEnd, rec.tlsErr, roundErr, ctx)
	}
	populatePhase(&x.HTTP, rec.written, rec.firstByte, rec.writeErr, roundErr, ctx)
	if rec.reused {
		x.TCP.Outcome = result.Unavailable
		x.TCP.DurationMS = nil
		x.TLS.Outcome = result.Unavailable
		x.TLS.DurationMS = nil
		x.HTTP.Outcome = result.Error
		x.HTTP.DurationMS = nil
		x.HTTP.Error = &result.ErrorData{Code: "connection_reused", Message: "The measured exchange reused a connection."}
		return x
	}
	if roundErr != nil {
		if x.HTTP.Outcome == result.Unavailable {
			x.HTTP.Outcome = errorOutcome(roundErr, ctx)
		}
		x.HTTP.Error = &result.ErrorData{Code: roundErrorCode(roundErr, rec), Message: "The HTTP exchange did not return response headers."}
		return x
	}
	if response == nil {
		x.HTTP.Outcome = result.Error
		x.HTTP.Error = &result.ErrorData{Code: "no_http_response", Message: "No HTTP response was returned."}
		return x
	}
	status := response.StatusCode
	x.ResponseStatus = &status
	x.Location = response.Header.Get("Location")
	return x
}

func peerIPv4(addr net.Addr) (netip.Addr, bool) {
	if tcp, ok := addr.(*net.TCPAddr); ok {
		ip, ok := netip.AddrFromSlice(tcp.IP)
		if ok && ip.Unmap().Is4() {
			return ip.Unmap(), true
		}
	}
	return netip.Addr{}, false
}
func ms(d time.Duration) float64 { return d.Seconds() * 1000 }
func populatePhase(m *result.Measurement, start, end time.Time, phaseErr, roundErr error, ctx context.Context) {
	if !start.IsZero() {
		v := start.UTC()
		m.StartedAt = &v
	}
	if !end.IsZero() {
		v := end.UTC()
		m.EndedAt = &v
	}
	if !start.IsZero() && !end.IsZero() && phaseErr == nil {
		v := end.Sub(start).Seconds() * 1000
		if v < 0 {
			v = 0
		}
		m.DurationMS = &v
		m.Outcome = result.Completed
		return
	}
	if phaseErr != nil {
		m.Outcome = errorOutcome(phaseErr, ctx)
		m.Error = &result.ErrorData{Code: phaseErrorCode(m.Method, phaseErr, ctx), Message: "The measured phase failed."}
		return
	}
	if !start.IsZero() && roundErr != nil {
		m.Outcome = errorOutcome(roundErr, ctx)
		return
	}
	if ctx.Err() != nil {
		m.Outcome = contextOutcome(ctx)
	}
}
func phaseErrorCode(method string, err error, ctx context.Context) string {
	if ctx.Err() != nil {
		return contextCode(ctx)
	}
	switch method {
	case "http_transport_connect":
		if errorOutcome(err, ctx) == result.Timeout {
			return "tcp_connect_timeout"
		}
		return "tcp_connect_failed"
	case "http_transport_tls":
		var unknown x509.UnknownAuthorityError
		var host x509.HostnameError
		var invalid x509.CertificateInvalidError
		if errors.As(err, &unknown) || errors.As(err, &host) || errors.As(err, &invalid) {
			return "tls_verification_failed"
		}
		if errorOutcome(err, ctx) == result.Timeout {
			return "tls_handshake_timeout"
		}
		return "tls_handshake_failed"
	case "request_written_to_first_response_header_byte":
		return "http_write_failed"
	default:
		return "phase_failed"
	}
}
func errorOutcome(err error, ctx context.Context) result.Outcome {
	if ctx.Err() != nil {
		return contextOutcome(ctx)
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return result.Timeout
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return result.Timeout
	}
	return result.Error
}
func roundErrorCode(err error, rec traceSnapshot) string {
	if errors.Is(err, errPeerMismatch) {
		return "peer_mismatch"
	}
	if rec.connectErr != nil {
		if strings.Contains(rec.connectErr.Error(), "actual peer differs") {
			return "peer_mismatch"
		}
		return "tcp_connect_failed"
	}
	if rec.tlsErr != nil {
		var unknown x509.UnknownAuthorityError
		var host x509.HostnameError
		var invalid x509.CertificateInvalidError
		if errors.As(rec.tlsErr, &unknown) || errors.As(rec.tlsErr, &host) || errors.As(rec.tlsErr, &invalid) {
			return "tls_verification_failed"
		}
		return "tls_handshake_failed"
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "response_header_timeout"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "deadline"
	}
	if errors.Is(err, context.Canceled) {
		return "cancelled"
	}
	return "http_exchange_failed"
}

func phaseStopReason(x result.Exchange) string {
	if x.HTTP.Error != nil {
		return x.HTTP.Error.Code
	}
	if x.TLS.Error != nil {
		return x.TLS.Error.Code
	}
	if x.TCP.Error != nil {
		return x.TCP.Error.Code
	}
	return fmt.Sprintf("http_%s", x.HTTP.Outcome)
}
