package target_test

import (
	"context"
	"errors"
	"net/netip"
	"reflect"
	"sync"
	"testing"
	"time"

	"hopvista/internal/result"
	"hopvista/internal/target"
)

type resolverReply struct {
	addresses []netip.Addr
	err       error
}
type fakeResolver struct {
	mu      sync.Mutex
	replies map[string][]resolverReply
	calls   []string
	onCall  func(context.Context)
}

func (f *fakeResolver) LookupIPv4(ctx context.Context, host string) ([]netip.Addr, error) {
	f.mu.Lock()
	f.calls = append(f.calls, host)
	replies := f.replies[host]
	if len(replies) > 0 {
		f.replies[host] = replies[1:]
	}
	f.mu.Unlock()
	if f.onCall != nil {
		f.onCall(ctx)
	}
	if len(replies) == 0 {
		return nil, errors.New("unscripted DNS")
	}
	return replies[0].addresses, replies[0].err
}
func (f *fakeResolver) callCount() int { f.mu.Lock(); defer f.mu.Unlock(); return len(f.calls) }
func ips(values ...string) []netip.Addr {
	out := make([]netip.Addr, 0, len(values))
	for _, v := range values {
		out = append(out, netip.MustParseAddr(v))
	}
	return out
}
func mustParse(t *testing.T, raw string) target.URL {
	t.Helper()
	u, p := target.Parse(raw)
	if p != nil {
		t.Fatalf("parse %q: %v", raw, p)
	}
	return u
}

func TestURLPolicy(t *testing.T) {
	for _, tc := range []struct{ raw, code string }{
		{"https://example.com/path?x=1#fragment", ""},
		{"http://127.0.0.1:8080/path", ""},
		{"http://10.0.0.1/", ""},
		{"http://169.254.1.2/", ""},
		{"http://192.0.2.1/", ""},
		{"https://example.com:443/", ""},
		{"example.com", "invalid_target_url"},
		{"ftp://example.com/", "invalid_target_url"},
		{"https://user:pass@example.com/", "invalid_target_url"},
		{"https://example.com:0/", "invalid_target_url"},
		{"https://example.com:65536/", "invalid_target_url"},
		{"https://example.com:/", "invalid_target_url"},
		{"http://[::1]/", "ipv6_not_supported_in_v1"},
		{"http://[::ffff:192.0.2.1]/", "ipv6_not_supported_in_v1"},
		{"http://0.0.0.0/", "invalid_target_url"},
		{"http://0.1.2.3/", "invalid_target_url"},
		{"http://224.0.0.1/", "invalid_target_url"},
		{"http://240.0.0.1/", "invalid_target_url"},
		{"http://255.255.255.255/", "invalid_target_url"},
	} {
		t.Run(tc.raw, func(t *testing.T) {
			u, p := target.Parse(tc.raw)
			if tc.code == "" {
				if p != nil {
					t.Fatal(p)
				}
				if u.String() == "" || u.Port() < 1 {
					t.Fatalf("bad parsed URL: %+v", u)
				}
				if tc.raw == "https://example.com/path?x=1#fragment" && u.String() != "https://example.com/path?x=1" {
					t.Fatal("fragment was retained")
				}
				return
			}
			if p == nil || p.Code != tc.code {
				t.Fatalf("got %+v, want %s", p, tc.code)
			}
		})
	}
}

func TestIANAScopeClassification(t *testing.T) {
	for _, tc := range []struct {
		ip     string
		scope  target.Scope
		usable bool
	}{
		{"8.8.8.8", target.Public, true},
		{"10.0.0.1", target.Private, true},
		{"172.16.0.1", target.Private, true},
		{"192.168.1.1", target.Private, true},
		{"127.0.0.2", target.Loopback, true},
		{"169.254.1.1", target.LinkLocal, true},
		{"100.64.0.1", target.SpecialUse, true},
		{"192.0.0.9", target.SpecialUse, true},
		{"192.0.2.10", target.SpecialUse, true},
		{"192.31.196.1", target.SpecialUse, true},
		{"192.88.99.1", target.SpecialUse, true},
		{"198.18.0.1", target.SpecialUse, true},
		{"198.51.100.1", target.SpecialUse, true},
		{"203.0.113.1", target.SpecialUse, true},
		{"224.0.0.1", target.SpecialUse, false},
		{"240.0.0.1", target.SpecialUse, false},
		{"0.0.0.0", target.SpecialUse, false},
	} {
		ip := netip.MustParseAddr(tc.ip)
		if got := target.Classify(ip); got != tc.scope {
			t.Errorf("%s classified %s, want %s", tc.ip, got, tc.scope)
		}
		if got := target.UsableIPv4(ip); got != tc.usable {
			t.Errorf("%s usable %v, want %v", tc.ip, got, tc.usable)
		}
	}
}

func TestFirstUsableAAndCacheAgainstChangedAnswers(t *testing.T) {
	clock := time.Date(2026, 10, 14, 20, 0, 0, 0, time.UTC)
	f := &fakeResolver{replies: map[string][]resolverReply{"Example.COM": {
		{addresses: ips("::1", "0.0.0.0", "10.0.0.5", "8.8.8.8")},
		{addresses: ips("1.1.1.1")},
	}}}
	f.onCall = func(context.Context) { clock = clock.Add(12 * time.Millisecond) }
	s := target.NewSelector(f, func() time.Time { return clock })
	u := mustParse(t, "https://Example.COM:8443/path")
	selected, dns, p := s.Select(context.Background(), u)
	if p != nil || dns == nil || dns.Outcome != result.Completed || selected.IP().String() != "10.0.0.5" || selected.Scope() != target.Private || selected.DNSRef() != dns.ID || u.Port() != 8443 {
		t.Fatalf("selection: %+v %+v %v", selected, dns, p)
	}
	if !reflect.DeepEqual(dns.Addresses, []string{"0.0.0.0", "10.0.0.5", "8.8.8.8"}) || dns.DurationMS == nil || *dns.DurationMS != 12 || dns.BoundMS == nil || *dns.BoundMS <= 0 || *dns.BoundMS > 2000 {
		t.Fatalf("DNS evidence: %+v", dns)
	}
	selected.Answers()[1] = netip.MustParseAddr("9.9.9.9")
	selected2, dns2, p := s.Select(context.Background(), mustParse(t, "https://example.com/next"))
	if p != nil || dns2 != nil || selected2.IP().String() != "10.0.0.5" || selected2.Answers()[1].String() != "10.0.0.5" || selected2.DNSRef() != dns.ID || f.callCount() != 1 {
		t.Fatalf("rebound/changed answer displaced pin: %+v %+v %v", selected2, dns2, p)
	}
}

func TestUnavailableAndCancelledSelection(t *testing.T) {
	f := &fakeResolver{replies: map[string][]resolverReply{
		"only-v6.example":  {{addresses: ips("::1", "2001:db8::1")}},
		"unusable.example": {{addresses: ips("0.0.0.0", "224.0.0.1")}},
		"broken.example":   {{err: errors.New("private resolver detail")}},
		"timeout.example":  {{err: context.DeadlineExceeded}},
	}}
	s := target.NewSelector(f, nil)
	for _, tc := range []struct {
		host, code string
		outcome    result.Outcome
	}{
		{"only-v6.example", "ipv4_unavailable", result.Completed},
		{"unusable.example", "ipv4_unavailable", result.Completed},
		{"broken.example", "dns_lookup_failed", result.Error},
		{"timeout.example", "dns_timeout", result.Timeout},
	} {
		selection, dns, p := s.Select(context.Background(), mustParse(t, "https://"+tc.host+"/"))
		if p == nil || p.Code != tc.code || dns == nil || dns.Outcome != tc.outcome || selection.IP().IsValid() {
			t.Fatalf("%s: %+v %+v %v", tc.host, selection, dns, p)
		}
		_, second, p2 := s.Select(context.Background(), mustParse(t, "https://"+tc.host+"/other"))
		if second != nil || p2 == nil || p2.Code != p.Code {
			t.Fatal("unavailable host was looked up twice")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, dns, p := s.Select(ctx, mustParse(t, "https://cancel.example/"))
	if p == nil || p.Code != "cancelled" || dns.Outcome != result.Cancelled || dns.StartedAt != nil || f.callCount() != 4 {
		t.Fatalf("cancelled lookup started: %+v %v", dns, p)
	}
}

func TestRedirectPolicyBeforeNewConnection(t *testing.T) {
	f := &fakeResolver{replies: map[string][]resolverReply{
		"public.example":    {{addresses: ips("8.8.8.8")}},
		"public2.example":   {{addresses: ips("1.1.1.1")}},
		"private.example":   {{addresses: ips("10.0.0.5")}},
		"loopback.example":  {{addresses: ips("127.0.0.1")}},
		"linklocal.example": {{addresses: ips("169.254.2.3")}},
		"special.example":   {{addresses: ips("100.64.0.1")}},
		"v6.example":        {{addresses: ips("::1")}},
	}}
	s := target.NewSelector(f, nil)
	initial, _, p := s.Select(context.Background(), mustParse(t, "https://public.example/start"))
	if p != nil {
		t.Fatal(p)
	}
	for _, tc := range []struct {
		location, reason, selected string
		calls                      int
	}{
		{"/next", "", "8.8.8.8", 1},
		{"https://PUBLIC.example:8443/other", "", "8.8.8.8", 1},
		{"https://public2.example/", "", "1.1.1.1", 2},
		{"http://private.example/", "redirect_requires_new_run", "10.0.0.5", 3},
		{"//private.example/other", "redirect_requires_new_run", "10.0.0.5", 3},
		{"http://loopback.example/", "redirect_requires_new_run", "127.0.0.1", 4},
		{"http://linklocal.example/", "redirect_requires_new_run", "169.254.2.3", 5},
		{"http://special.example/", "redirect_requires_new_run", "100.64.0.1", 6},
		{"http://v6.example/", "ipv4_unavailable", "", 7},
		{"http://[::1]/", "ipv6_not_supported_in_v1", "", 7},
		{"file:///etc/passwd", "invalid_redirect", "", 7},
		{"https://user:password@public2.example/", "invalid_redirect", "", 7},
		{"", "invalid_redirect", "", 7},
	} {
		d := s.NextRedirect(context.Background(), initial, tc.location, 0)
		if d.StopReason != tc.reason || d.SelectedIP != tc.selected || f.callCount() != tc.calls {
			t.Fatalf("%s: %+v; calls %d", tc.location, d, f.callCount())
		}
		if tc.reason == "" && (d.Next == nil || d.Next.IP().String() != tc.selected) {
			t.Fatalf("allowed redirect missing approved pin: %+v", d)
		}
		if tc.reason != "" && d.Next != nil {
			t.Fatalf("stopped redirect exposed a dial selection: %+v", d)
		}
	}
	if d := s.NextRedirect(context.Background(), initial, "https://new.example/", 3); d.StopReason != "redirect_limit" || d.Next != nil || f.callCount() != 7 {
		t.Fatalf("limit triggered network work: %+v", d)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if d := s.NextRedirect(ctx, initial, "https://unseen.example/", 0); d.StopReason != "cancelled" || d.Next != nil || f.callCount() != 7 {
		t.Fatalf("cancel triggered lookup: %+v", d)
	}
}

func TestExplicitPrivateHostAndSameHostRedirect(t *testing.T) {
	f := &fakeResolver{replies: map[string][]resolverReply{"device.local": {{addresses: ips("192.168.1.25")}, {addresses: ips("8.8.8.8")}}}}
	s := target.NewSelector(f, nil)
	initial, _, p := s.Select(context.Background(), mustParse(t, "http://device.local/"))
	if p != nil || initial.Scope() != target.Private {
		t.Fatalf("explicit private hostname: %+v %v", initial, p)
	}
	d := s.NextRedirect(context.Background(), initial, "/admin", 0)
	if d.StopReason != "" || d.Next == nil || d.Next.IP().String() != "192.168.1.25" || d.DNS != nil || f.callCount() != 1 {
		t.Fatalf("same-host pin changed: %+v", d)
	}
	literal, _, p := s.Select(context.Background(), mustParse(t, "http://127.0.0.1/"))
	if p != nil || literal.Scope() != target.Loopback || f.callCount() != 1 {
		t.Fatalf("explicit loopback failed: %+v %v", literal, p)
	}
}
