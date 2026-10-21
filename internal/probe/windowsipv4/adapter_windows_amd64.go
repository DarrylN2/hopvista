//go:build windows && amd64

// Package windowsipv4 implements the synchronous IPv4 IP Helper adapter.
// Native handles and status codes stay inside this package.
package windowsipv4

import (
	"context"
	"encoding/binary"
	"net/netip"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"hopvista/internal/probe"
	"hopvista/internal/result"
)

var dll = syscall.NewLazyDLL("iphlpapi.dll")
var getBestRoute2 = dll.NewProc("GetBestRoute2")
var icmpCreateFile = dll.NewProc("IcmpCreateFile")
var icmpSendEcho2 = dll.NewProc("IcmpSendEcho2")
var icmpCloseHandle = dll.NewProc("IcmpCloseHandle")

// observeEcho is a package-local test seam. Production constructs Adapter{}
// and retains the schema-1 returned-and-closed timestamp in Measurement.
type Adapter struct {
	observeEcho func(nativeEchoTrace)
}

type nativeEchoTrace struct {
	StartedAt    time.Time
	ReturnedAt   time.Time
	ClosedAt     time.Time
	HandleClosed bool
}

var _ probe.Adapter = Adapter{}

// These layouts match the reviewed standalone harness. The byte arrays avoid
// placing Go pointers in native buffers and the uint64 supplies ABI alignment.
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

func socketAddress(ip netip.Addr) sockaddr {
	var out sockaddr
	binary.LittleEndian.PutUint16(out[:2], 2)
	b := ip.As4()
	copy(out[4:8], b[:])
	return out
}

func readAddress(b sockaddr) (netip.Addr, bool) {
	if binary.LittleEndian.Uint16(b[:2]) != 2 {
		return netip.Addr{}, false
	}
	return netip.AddrFrom4([4]byte(b[4:8])), true
}

func (Adapter) Route(ctx context.Context, req probe.RouteRequest) result.Route {
	out := result.Route{Measurement: result.Measurement{ID: req.ID, Method: "get_best_route2_ipv4"}, SelectedIP: req.SelectedIP.String()}
	if req.ID == "" || !probe.ValidIPv4(req.SelectedIP) || req.DeadlineAt.IsZero() {
		out.Outcome = result.Error
		out.Error = invalid()
		return out
	}
	if stop := probe.StartOutcome(ctx, time.Now(), req.DeadlineAt, 0); stop != "" {
		out.Outcome = stop
		return out
	}
	dest := socketAddress(req.SelectedIP)
	var row routeRow
	var source sockaddr
	started := time.Now().UTC()
	out.StartedAt = &started
	code, _, _ := getBestRoute2.Call(0, 0, 0, uintptr(unsafe.Pointer(&dest)), 0, uintptr(unsafe.Pointer(&row)), uintptr(unsafe.Pointer(&source)))
	ended := time.Now().UTC()
	out.EndedAt = &ended
	runtime.KeepAlive(dest)
	runtime.KeepAlive(row)
	runtime.KeepAlive(source)
	if code != 0 {
		out.Outcome = result.Error
		out.Error = nativeError("route_lookup_failed")
		return out
	}
	src, srcOK := readAddress(source)
	next, nextOK := readAddress(row.Next)
	if !srcOK || !nextOK {
		out.Outcome = result.Error
		out.Error = nativeError("route_address_invalid")
		return out
	}
	onLink := next.IsUnspecified()
	out.Outcome = result.Completed
	out.SourceIP = src.String()
	out.InterfaceIndex = &row.Index
	out.OnLink = &onLink
	if !onLink {
		out.NextHop = next.String()
	}
	duration := ended.Sub(started).Seconds() * 1000
	out.DurationMS = &duration
	return out
}

func (a Adapter) Echo(ctx context.Context, req probe.EchoRequest) (out result.Echo) {
	bound := req.Cap.Seconds() * 1000
	out = result.Echo{Measurement: result.Measurement{ID: req.ID, Method: "icmp_send_echo2_ipv4", BoundMS: &bound}, SelectedIP: req.SelectedIP.String(), TTL: req.TTL, Status: result.NotSent}
	if req.ID == "" || !probe.ValidIPv4(req.SelectedIP) || req.TTL < 1 || req.TTL > 255 || req.Cap <= 0 || req.Cap > time.Second || req.Cap%time.Millisecond != 0 || req.DeadlineAt.IsZero() {
		out.Outcome = result.Error
		out.Error = invalid()
		return
	}
	if stop := probe.StartOutcome(ctx, time.Now(), req.DeadlineAt, req.Cap); stop != "" {
		out.Outcome = stop
		return
	}
	handle, _, _ := icmpCreateFile.Call()
	if handle == 0 || handle == ^uintptr(0) {
		out.Outcome = result.Error
		out.Status = result.CallError
		out.Error = nativeError("icmp_handle_open_failed")
		return
	}
	var returnedAt time.Time
	defer func() {
		closed, _, _ := icmpCloseHandle.Call(handle)
		closedAt := time.Now().UTC()
		if closed == 0 {
			out.Outcome = result.Error
			out.Status = result.CallError
			out.Error = nativeError("icmp_handle_close_failed")
			out.ReplyAddress = ""
			out.RTTMS = nil
			out.DurationMS = nil
		}
		if out.StartedAt != nil {
			out.EndedAt = &closedAt
			if out.Outcome == result.Completed {
				d := closedAt.Sub(*out.StartedAt).Seconds() * 1000
				out.DurationMS = &d
			}
			if a.observeEcho != nil && !returnedAt.IsZero() {
				a.observeEcho(nativeEchoTrace{StartedAt: *out.StartedAt, ReturnedAt: returnedAt, ClosedAt: closedAt, HandleClosed: closed != 0})
			}
		}
	}()
	if stop := probe.StartOutcome(ctx, time.Now(), req.DeadlineAt, req.Cap); stop != "" {
		out.Outcome = stop
		return
	}
	var options [8]byte // IP_OPTION_INFORMATION32: TTL followed by zero options and pointer32.
	options[0] = byte(req.TTL)
	request := [8]byte{'h', 'v', '-', 'c', 'h', 'e', 'c', 'k'}
	var reply [64]uint64 // 512 aligned bytes, including ICMP_ECHO_REPLY32.
	if stop := probe.StartOutcome(ctx, time.Now(), req.DeadlineAt, req.Cap); stop != "" {
		out.Outcome = stop
		return
	}
	started := time.Now().UTC()
	out.StartedAt = &started
	ip := req.SelectedIP.As4()
	count, _, callErr := icmpSendEcho2.Call(handle, 0, 0, 0, uintptr(binary.LittleEndian.Uint32(ip[:])), uintptr(unsafe.Pointer(&request[0])), uintptr(len(request)), uintptr(unsafe.Pointer(&options[0])), uintptr(unsafe.Pointer(&reply[0])), uintptr(unsafe.Sizeof(reply)), uintptr(req.Cap/time.Millisecond))
	returnedAt = time.Now().UTC()
	runtime.KeepAlive(request)
	runtime.KeepAlive(options)
	runtime.KeepAlive(reply)
	if count == 0 {
		if errno(callErr) == 11010 {
			out.Outcome = result.Timeout
			out.Status = result.NoReply
		} else {
			out.Outcome = result.Error
			out.Status = result.CallError
			out.Error = nativeError("icmp_send_failed")
		}
		return
	}
	return decodeReply(out, unsafe.Slice((*byte)(unsafe.Pointer(&reply[0])), int(unsafe.Sizeof(reply))))
}

func decodeReply(out result.Echo, b []byte) result.Echo {
	if len(b) < 16 {
		out.Outcome = result.Error
		out.Status = result.CallError
		out.Error = nativeError("icmp_reply_invalid")
		return out
	}
	status := binary.LittleEndian.Uint32(b[4:8])
	addr := netip.AddrFrom4([4]byte(b[:4]))
	switch status {
	case 0:
		out.Status = result.EchoReply
	case 11013, 11014, 11041:
		out.Status = result.TimeExceeded
	case 11002, 11003, 11004, 11005, 11040, 11045:
		out.Status = result.UnreachableReply
	default:
		out.Outcome = result.Error
		out.Status = result.CallError
		out.Error = nativeError("icmp_reply_status_unknown")
		return out
	}
	out.Outcome = result.Completed
	out.ReplyAddress = addr.String()
	rtt := float64(binary.LittleEndian.Uint32(b[8:12]))
	out.RTTMS = &rtt
	return out
}

func errno(err error) uint32 {
	if n, ok := err.(syscall.Errno); ok {
		return uint32(n)
	}
	return 0
}
func invalid() *result.ErrorData {
	return &result.ErrorData{Code: "invalid_probe_request", Message: "A valid IPv4 request and full-cap budget are required."}
}
func nativeError(code string) *result.ErrorData {
	return &result.ErrorData{Code: code, Message: "Windows IP Helper did not complete this measurement."}
}
