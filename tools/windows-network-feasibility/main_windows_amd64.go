package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/netip"
	"os"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

type drainObservation struct {
	Case                 string          `json:"case"`
	StopRequestedAt      time.Time       `json:"stop_requested_at"`
	QuiescentAt          time.Time       `json:"quiescent_at"`
	DrainMS              float64         `json:"drain_ms"`
	PendingAtStop        bool            `json:"pending_at_stop"`
	CallsStarted         int             `json:"calls_started"`
	FollowingCallStarted bool            `json:"following_call_started"`
	ContextError         string          `json:"context_error"`
	DeadlineAt           *time.Time      `json:"deadline_at,omitempty"`
	Echo                 echoObservation `json:"echo"`
}

func drain(label string, ip netip.Addr, cutoff bool) drainObservation {
	var ctx context.Context
	var cancel context.CancelFunc
	if cutoff {
		ctx, cancel = context.WithTimeout(context.Background(), 150*time.Millisecond)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	started := make(chan struct{})
	done := make(chan echoObservation, 1)
	go func() {
		done <- echo(ctx, label, ip, 64, 750*time.Millisecond, false, 512, func() { close(started) })
	}()
	var o drainObservation
	o.Case = label
	if end, ok := ctx.Deadline(); ok {
		utc := end.UTC()
		o.DeadlineAt = &utc
	}
	select {
	case <-started:
		o.CallsStarted = 1
		if !cutoff {
			time.Sleep(25 * time.Millisecond)
			cancel()
		}
		<-ctx.Done()
		o.StopRequestedAt = time.Now().UTC()
		o.Echo = <-done
	case o.Echo = <-done:
		// A fast return may win this select even if the send boundary fired.
		if !o.Echo.StartedAt.IsZero() {
			if !cutoff {
				time.Sleep(25 * time.Millisecond)
				cancel()
			}
			<-ctx.Done()
		} else {
			cancel()
		}
		o.StopRequestedAt = time.Now().UTC()
	}
	if !o.Echo.StartedAt.IsZero() {
		o.CallsStarted = 1
	}
	o.QuiescentAt = time.Now().UTC()
	o.PendingAtStop = !o.Echo.EndedAt.IsZero() && o.Echo.EndedAt.After(o.StopRequestedAt)
	o.DrainMS = milliseconds(o.QuiescentAt.Sub(o.StopRequestedAt))
	if ctx.Err() != nil {
		o.ContextError = ctx.Err().Error()
	}
	// This is the entire disposable scheduling check, not an app coordinator.
	if ctx.Err() == nil {
		o.FollowingCallStarted = true
		_ = echo(ctx, "unexpected-following-call", ip, 64, 750*time.Millisecond, false, 512, nil)
	}
	return o
}

func environment() map[string]any {
	var version [284]byte // RTL_OSVERSIONINFOEXW
	binary.LittleEndian.PutUint32(version[:4], uint32(len(version)))
	code, _, _ := syscall.NewLazyDLL("ntdll.dll").NewProc("RtlGetVersion").Call(uintptr(unsafe.Pointer(&version[0])))
	var elevated uint32
	var returned uint32
	token, err := syscall.OpenCurrentProcessToken()
	if err != nil {
		panic(err)
	}
	defer token.Close()
	err = syscall.GetTokenInformation(token, 20 /*TokenElevation*/, (*byte)(unsafe.Pointer(&elevated)), 4, &returned)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		"go_version": runtime.Version(), "os": runtime.GOOS, "architecture": runtime.GOARCH,
		"windows_version":     fmt.Sprintf("%d.%d.%d", binary.LittleEndian.Uint32(version[4:8]), binary.LittleEndian.Uint32(version[8:12]), binary.LittleEndian.Uint32(version[12:16])),
		"version_return_code": code, "elevated": elevated != 0,
		"route_row_size": unsafe.Sizeof(routeRow{}), "route_next_hop_offset": unsafe.Offsetof(routeRow{}.Next),
		"sockaddr_size": unsafe.Sizeof(sockaddr{}), "options_size": 8, "reply_buffer_size": 512,
		"ipv4_reply_status_offset": 4, "ipv4_reply_rtt_offset": 8,
		"ipv6_reply_address_offset": 6, "ipv6_reply_status_offset": 28, "ipv6_reply_rtt_offset": 32,
		"ipv6_reply_scope_offset": 22,
		"native_mode":             "synchronous; Event/APC/context NULL; no raw sockets",
	}
}

func parseDestination(raw string, v6 bool) netip.Addr {
	ip, err := netip.ParseAddr(raw)
	if err != nil {
		panic(err)
	}
	ip = ip.Unmap()
	if ip.Is6() != v6 || ip.IsUnspecified() || ip.IsMulticast() || ip == netip.MustParseAddr("255.255.255.255") {
		panic("expected unicast literal in the requested family")
	}
	return ip
}

func main() {
	remote4 := flag.String("ipv4", "1.1.1.1", "literal remote IPv4 destination")
	remote6 := flag.String("ipv6", "2606:4700:4700::1111", "literal remote IPv6 destination")
	silent4 := flag.String("no-reply-ipv4", "192.0.2.1", "candidate no-reply IPv4 destination; silence must be observed")
	silent6 := flag.String("no-reply-ipv6", "2001:db8::1", "candidate no-reply IPv6 destination; requires a route")
	hops := flag.Int("hops", 4, "TTL/Hop Limit steps (1-64), continue after timeout")
	flag.Parse()
	if *hops < 1 || *hops > 64 {
		fmt.Fprintln(os.Stderr, "hops must be 1-64")
		os.Exit(2)
	}
	v4, v6 := parseDestination(*remote4, false), parseDestination(*remote6, true)
	n4, n6 := parseDestination(*silent4, false), parseDestination(*silent6, true)
	env := environment()
	if env["elevated"] == true {
		fmt.Fprintln(os.Stderr, "Run this feasibility gate unelevated.")
		os.Exit(2)
	}
	var routes []routeObservation
	var echoes []echoObservation
	ctx := context.Background()
	probe := func(label string, ip netip.Addr, ttl byte) {
		echoes = append(echoes, echo(ctx, label, ip, ttl, 750*time.Millisecond, false, 512, nil))
	}
	for _, ip := range []netip.Addr{netip.MustParseAddr("127.0.0.1"), netip.IPv6Loopback(), v4, v6, n4, n6} {
		routes = append(routes, route(ip))
	}
	probe("localhost-v4", netip.MustParseAddr("127.0.0.1"), 64)
	probe("localhost-v6", netip.IPv6Loopback(), 64)
	probe("localhost-v6-hop1", netip.IPv6Loopback(), 1)
	// Exercise a real zone against this machine's own link-local address,
	// when present. This is local scope interop, not remote IPv6 evidence.
	interfaces, interfaceErr := net.Interfaces()
	if interfaceErr != nil {
		env["scoped_local_ipv6_error"] = interfaceErr.Error()
	}
	env["scoped_local_ipv6_available"] = false
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		addresses, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, raw := range addresses {
			prefix, err := netip.ParsePrefix(raw.String())
			if err != nil || !prefix.Addr().Is6() || !prefix.Addr().IsLinkLocalUnicast() {
				continue
			}
			ip := prefix.Addr().WithZone(fmt.Sprint(iface.Index))
			routes = append(routes, route(ip))
			probe("scoped-local-v6", ip, 1)
			env["scoped_local_ipv6_available"] = true
			break
		}
		if env["scoped_local_ipv6_available"] == true {
			break
		}
	}
	for _, ip := range []netip.Addr{v4, v6} {
		r := route(ip)
		if r.Error == "" && !r.OnLink {
			gateway := netip.MustParseAddr(r.NextHop)
			if gateway.Is6() && gateway.IsLinkLocalUnicast() && gateway.Zone() == "" {
				gateway = gateway.WithZone(fmt.Sprint(r.Interface))
			}
			routes = append(routes, route(gateway))
			probe("selected-next-hop", gateway, 64)
		}
		for ttl := 1; ttl <= *hops; ttl++ {
			probe("remote-low-limit", ip, byte(ttl))
			last := echoes[len(echoes)-1]
			if last.Outcome == "echo_reply" || last.Outcome == "unreachable_reply" || last.Outcome == "api_error" {
				break
			}
		}
		probe("remote-echo", ip, 64)
	}
	// Consecutive unanswered TTLs demonstrate continuation when this path is silent.
	for ttl := 1; ttl <= 4; ttl++ {
		probe("no-reply-continuation", n4, byte(ttl))
	}
	probe("no-reply-v4", n4, 64)
	probe("no-reply-v6", n6, 64)
	echoes = append(echoes, echo(ctx, "invalid-handle", netip.MustParseAddr("127.0.0.1"), 64, 750*time.Millisecond, true, 512, nil))
	echoes = append(echoes, echo(ctx, "short-buffer", netip.MustParseAddr("127.0.0.1"), 64, 750*time.Millisecond, false, 8, nil))
	// A local scope preflight makes no native call when a required zone is absent.
	routes = append(routes, route(netip.MustParseAddr("fe80::1")))
	drains := []drainObservation{drain("cancel-v4", n4, false), drain("cutoff-v4", n4, true), drain("cancel-v6", n6, false), drain("cutoff-v6", n6, true)}
	result := struct {
		CapturedAt  time.Time          `json:"captured_at"`
		Environment map[string]any     `json:"environment"`
		Routes      []routeObservation `json:"routes"`
		Echoes      []echoObservation  `json:"echoes"`
		Drains      []drainObservation `json:"drains"`
		Gate        string             `json:"gate"`
	}{time.Now().UTC(), env, routes, echoes, drains, "requires evidence review; exit 0 means capture completed, not feasibility passed"}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
