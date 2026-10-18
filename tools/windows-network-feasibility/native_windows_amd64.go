package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"net/netip"
	"runtime"
	"strconv"
	"syscall"
	"time"
	"unsafe"
)

var (
	ipHelper  = syscall.NewLazyDLL("iphlpapi.dll")
	create4   = ipHelper.NewProc("IcmpCreateFile")
	create6   = ipHelper.NewProc("Icmp6CreateFile")
	closeICMP = ipHelper.NewProc("IcmpCloseHandle")
	send4     = ipHelper.NewProc("IcmpSendEcho2")
	send6     = ipHelper.NewProc("Icmp6SendEcho2")
	bestRoute = ipHelper.NewProc("GetBestRoute2")
)

// Byte layouts are Windows ABI layouts, not Go structs containing Go pointers.
// uint64 backing storage supplies alignment for NET_LUID and native buffers.
// All calls are synchronous (Event/APC/context = NULL). Buffers remain alive
// until the native call returns; no cancellation goroutine closes the handle.
type sockaddr [28]byte
type routeRow struct {
	Luid                                       uint64
	Index                                      uint32
	Prefix                                     [32]byte
	Next                                       sockaddr
	SitePrefix                                 byte
	_                                          [3]byte
	Valid, Preferred, Metric, Protocol         uint32
	Loopback, Autoconfigure, Publish, Immortal byte
	Age, Origin                                uint32
}

func socketAddress(ip netip.Addr) (sockaddr, error) {
	var out sockaddr
	ip = ip.Unmap()
	if ip.Is4() {
		binary.LittleEndian.PutUint16(out[:2], 2)
		b := ip.As4()
		copy(out[4:8], b[:])
		return out, nil
	}
	if !ip.Is6() {
		return out, fmt.Errorf("invalid address")
	}
	binary.LittleEndian.PutUint16(out[:2], 23)
	b := ip.As16()
	copy(out[8:24], b[:])
	if ip.IsLinkLocalUnicast() && ip.Zone() == "" {
		return out, fmt.Errorf("IPv6 link-local destination needs a zone")
	}
	if zone := ip.Zone(); zone != "" {
		index, err := strconv.ParseUint(zone, 10, 32)
		if err != nil {
			iface, lookupErr := net.InterfaceByName(zone)
			if lookupErr != nil {
				return out, lookupErr
			}
			index = uint64(iface.Index)
		}
		if index == 0 {
			return out, fmt.Errorf("zone must identify an interface")
		}
		binary.LittleEndian.PutUint32(out[24:28], uint32(index))
	}
	return out, nil
}

func readSockaddr(b []byte) (netip.Addr, error) {
	if len(b) < 28 {
		return netip.Addr{}, fmt.Errorf("short SOCKADDR_INET")
	}
	switch binary.LittleEndian.Uint16(b[:2]) {
	case 2:
		return netip.AddrFrom4([4]byte(b[4:8])), nil
	case 23:
		ip := netip.AddrFrom16([16]byte(b[8:24]))
		if scope := binary.LittleEndian.Uint32(b[24:28]); scope != 0 {
			ip = ip.WithZone(strconv.FormatUint(uint64(scope), 10))
		}
		return ip, nil
	default:
		return netip.Addr{}, fmt.Errorf("unexpected sockaddr family")
	}
}

type routeObservation struct {
	Destination string  `json:"destination"`
	Family      string  `json:"family"`
	Scope       string  `json:"scope"`
	APICalled   bool    `json:"api_called"`
	Source      string  `json:"source,omitempty"`
	Interface   uint32  `json:"interface_index,omitempty"`
	NextHop     string  `json:"next_hop,omitempty"`
	OnLink      bool    `json:"on_link"`
	ReturnCode  uint32  `json:"return_code"`
	ElapsedMS   float64 `json:"elapsed_ms"`
	Error       string  `json:"error,omitempty"`
}

func route(ip netip.Addr) routeObservation {
	o := routeObservation{Destination: ip.String(), Family: "ipv4", Scope: scope(ip)}
	if ip.Is6() {
		o.Family = "ipv6"
	}
	dest, err := socketAddress(ip)
	if err != nil {
		o.Error = err.Error()
		return o
	}
	var row routeRow
	var source sockaddr
	start := time.Now()
	o.APICalled = true
	code, _, _ := bestRoute.Call(0, 0, 0, uintptr(unsafe.Pointer(&dest)), 0, uintptr(unsafe.Pointer(&row)), uintptr(unsafe.Pointer(&source)))
	o.ElapsedMS = milliseconds(time.Since(start))
	runtime.KeepAlive(dest)
	runtime.KeepAlive(row)
	runtime.KeepAlive(source)
	o.ReturnCode = uint32(code)
	if code != 0 {
		o.Error = syscall.Errno(code).Error()
		return o
	}
	src, err := readSockaddr(source[:])
	if err != nil {
		o.Error = err.Error()
		return o
	}
	next, err := readSockaddr(row.Next[:])
	if err != nil {
		o.Error = err.Error()
		return o
	}
	o.Source, o.Interface, o.NextHop, o.OnLink = src.String(), row.Index, next.String(), next.IsUnspecified()
	return o
}

type echoObservation struct {
	Case        string    `json:"case"`
	Destination string    `json:"destination"`
	Family      string    `json:"family"`
	Scope       string    `json:"scope"`
	API         string    `json:"api"`
	HopLimit    byte      `json:"ttl_or_hop_limit"`
	TimeoutMS   uint32    `json:"timeout_ms"`
	RequestSize int       `json:"request_size"`
	ReplySize   int       `json:"reply_buffer_size"`
	StartedAt   time.Time `json:"started_at"`
	EndedAt     time.Time `json:"ended_at"`
	ElapsedMS   float64   `json:"elapsed_ms"`
	ReplyCount  uint32    `json:"reply_count"`
	Status      *uint32   `json:"reply_status,omitempty"`
	NativeError uint32    `json:"native_error"`
	Outcome     string    `json:"outcome"`
	ReplyFrom   string    `json:"reply_from,omitempty"`
	RTT         *uint32   `json:"rtt_ms,omitempty"`
	DataSize    *uint16   `json:"reply_data_size,omitempty"`
	ReplyScope  *uint32   `json:"reply_scope_id,omitempty"`
	CloseOK     *bool     `json:"close_ok,omitempty"`
	Error       string    `json:"error,omitempty"`
}

func outcome(code uint32, reply bool) string {
	switch code {
	case 0:
		if reply {
			return "echo_reply"
		}
	case 11010:
		return "timeout"
	case 11013, 11014, 11041:
		if reply {
			return "time_exceeded_reply"
		}
	case 11002, 11003, 11004, 11005, 11040, 11045:
		if reply {
			return "unreachable_reply"
		}
	}
	if reply {
		return "other_icmp_reply"
	}
	return "api_error"
}

func nativeTimeout(ctx context.Context, cap time.Duration, now time.Time) (uint32, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if end, ok := ctx.Deadline(); ok && end.Sub(now) < cap {
		cap = end.Sub(now)
	}
	// Floor so the native timeout never exceeds remaining budget.
	if cap < time.Millisecond {
		return 0, context.DeadlineExceeded
	}
	return uint32(cap / time.Millisecond), nil
}

func echo(ctx context.Context, label string, ip netip.Addr, ttl byte, cap time.Duration, invalidHandle bool, replySize int, began func()) (o echoObservation) {
	return echoInternal(ctx, label, ip, ttl, cap, invalidHandle, replySize, began, 0, nil)
}

func fullBudgetAvailable(ctx context.Context, cap, margin time.Duration, now time.Time) bool {
	if ctx.Err() != nil {
		return false
	}
	end, ok := ctx.Deadline()
	return !ok || end.Sub(now) >= cap+margin
}

// A guarded call either uses its full cap or does not start. Evaluate the
// guard and timeout from one clock reading so a pause cannot silently trim it.
func guardedNativeTimeout(ctx context.Context, cap, margin time.Duration, now time.Time) (uint32, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if cap < time.Millisecond || cap%time.Millisecond != 0 || cap/time.Millisecond > time.Duration(^uint32(0)) {
		return 0, fmt.Errorf("full native timeout must fit an integral number of milliseconds")
	}
	if !fullBudgetAvailable(ctx, cap, margin, now) {
		return 0, context.DeadlineExceeded
	}
	return uint32(cap / time.Millisecond), nil
}

// startMargin is used only by the guarded feasibility experiment. beforeSend
// lets the live test simulate a scheduling delay after handle creation.
func echoInternal(ctx context.Context, label string, ip netip.Addr, ttl byte, cap time.Duration, invalidHandle bool, replySize int, began func(), startMargin time.Duration, beforeSend func()) (o echoObservation) {
	ip = ip.Unmap()
	o = echoObservation{Case: label, Destination: ip.String(), Family: "ipv4", Scope: scope(ip), API: "IcmpSendEcho2", HopLimit: ttl, RequestSize: 8, ReplySize: replySize}
	if ip.Is6() {
		o.Family, o.API = "ipv6", "Icmp6SendEcho2"
	}
	if err := ctx.Err(); err != nil {
		o.Outcome, o.Error = "not_started", err.Error()
		return
	}
	dest, err := socketAddress(ip)
	if err != nil {
		o.Outcome, o.Error = "unsupported_scope", err.Error()
		return
	}
	create := create4
	if ip.Is6() {
		create = create6
	}
	handle, _, callErr := create.Call()
	if handle == ^uintptr(0) || handle == 0 {
		o.Outcome, o.NativeError, o.Error = "api_error", errno(callErr), callErr.Error()
		return
	}
	defer func() {
		closed, _, closeErr := closeICMP.Call(handle)
		ok := closed != 0
		o.CloseOK = &ok
		if !ok {
			o.Error += " close: " + closeErr.Error()
		}
	}()
	var options [8]byte // IP_OPTION_INFORMATION32: TTL/TOS/flags/size + NULL pointer32.
	options[0] = ttl
	request := [8]byte{'h', 'v', '-', 'c', 'h', 'e', 'c', 'k'}
	buffer := make([]uint64, (replySize+7)/8)
	bytes := unsafe.Slice((*byte)(unsafe.Pointer(&buffer[0])), len(buffer)*8)
	var source sockaddr
	if ip.Is6() {
		// GetBestRoute2 supplies source and interface scope when a route exists.
		r := route(ip)
		if r.Error == "" {
			source, _ = socketAddress(netip.MustParseAddr(r.Source))
		} else {
			source, _ = socketAddress(netip.IPv6Unspecified())
		}
	}
	if err := ctx.Err(); err != nil {
		o.Outcome, o.Error = "cancelled", err.Error()
		return
	}
	if beforeSend != nil {
		beforeSend()
	}
	if startMargin > 0 {
		o.TimeoutMS, err = guardedNativeTimeout(ctx, cap, startMargin, time.Now())
	} else {
		o.TimeoutMS, err = nativeTimeout(ctx, cap, time.Now())
	}
	if err != nil {
		o.Outcome, o.Error = "not_started", err.Error()
		return
	}
	callHandle := handle
	if invalidHandle {
		callHandle = ^uintptr(0)
	}
	if startMargin > 0 && !fullBudgetAvailable(ctx, cap, startMargin, time.Now()) {
		o.Outcome, o.Error, o.TimeoutMS = "not_started", "insufficient time for full native probe budget and candidate margin", 0
		return
	}
	start := time.Now()
	o.StartedAt = start.UTC()
	if began != nil {
		began()
	}
	var count uintptr
	if ip.Is4() {
		b := ip.As4()
		count, _, callErr = send4.Call(callHandle, 0, 0, 0, uintptr(binary.LittleEndian.Uint32(b[:])), uintptr(unsafe.Pointer(&request[0])), uintptr(len(request)), uintptr(unsafe.Pointer(&options[0])), uintptr(unsafe.Pointer(&buffer[0])), uintptr(replySize), uintptr(o.TimeoutMS))
	} else {
		count, _, callErr = send6.Call(callHandle, 0, 0, 0, uintptr(unsafe.Pointer(&source)), uintptr(unsafe.Pointer(&dest)), uintptr(unsafe.Pointer(&request[0])), uintptr(len(request)), uintptr(unsafe.Pointer(&options[0])), uintptr(unsafe.Pointer(&buffer[0])), uintptr(replySize), uintptr(o.TimeoutMS))
	}
	o.EndedAt = time.Now().UTC()
	o.ElapsedMS = milliseconds(time.Since(start))
	runtime.KeepAlive(request)
	runtime.KeepAlive(options)
	runtime.KeepAlive(buffer)
	runtime.KeepAlive(source)
	runtime.KeepAlive(dest)
	o.ReplyCount = uint32(count)
	if count == 0 {
		o.NativeError = errno(callErr)
		o.Outcome = outcome(o.NativeError, false)
		o.Error = callErr.Error()
		return
	}
	decodeEchoReply(&o, ip, bytes)
	return
}

func decodeEchoReply(o *echoObservation, ip netip.Addr, bytes []byte) {
	// Parse only the first synchronous reply. Never dereference pointer32 data.
	statusOffset, rttOffset := 4, 8
	if ip.Is6() {
		statusOffset, rttOffset = 28, 32
		addr := netip.AddrFrom16([16]byte(bytes[6:22])) // packed IPV6_ADDRESS_EX
		// IPV6_ADDRESS_EX is packed: 2-byte port + 4-byte flow +
		// 16-byte address + 4-byte scope = 26 bytes, then outer padding.
		// Microsoft documents this reply field in network byte order.
		// Live local replies here returned zero; nonzero reply scope is unverified.
		zone := binary.BigEndian.Uint32(bytes[22:26])
		o.ReplyScope = &zone
		if zone != 0 {
			addr = addr.WithZone(strconv.FormatUint(uint64(zone), 10))
		}
		o.ReplyFrom = addr.String()
	} else {
		o.ReplyFrom = netip.AddrFrom4([4]byte(bytes[:4])).String()
		size := binary.LittleEndian.Uint16(bytes[12:14])
		o.DataSize = &size
	}
	status := binary.LittleEndian.Uint32(bytes[statusOffset : statusOffset+4])
	o.Status = &status
	o.Outcome = outcome(status, true)
	if status == 0 || o.Outcome == "time_exceeded_reply" || o.Outcome == "unreachable_reply" {
		rtt := binary.LittleEndian.Uint32(bytes[rttOffset : rttOffset+4])
		o.RTT = &rtt
	}
}

func errno(err error) uint32 {
	if n, ok := err.(syscall.Errno); ok {
		return uint32(n)
	}
	return 0
}
func milliseconds(d time.Duration) float64 { return float64(d) / float64(time.Millisecond) }

// This is a harness label, not the future production IANA address policy.
func scope(ip netip.Addr) string {
	ip = ip.Unmap()
	switch {
	case ip.IsLoopback():
		return "loopback"
	case ip.IsLinkLocalUnicast():
		return "link_local"
	case ip.IsPrivate():
		return "private_or_unique_local"
	case netip.MustParsePrefix("192.0.2.0/24").Contains(ip), netip.MustParsePrefix("2001:db8::/32").Contains(ip):
		return "documentation"
	case ip == netip.MustParseAddr("1.1.1.1"), ip == netip.MustParseAddr("2606:4700:4700::1111"):
		return "public"
	default:
		return "other_unicast_review_address_policy"
	}
}
