package main

// Experimental event-path check only. It always waits for native completion
// before releasing storage; observing a context deadline is not completion.
import (
	"encoding/binary"
	"fmt"
	"net/netip"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

type asyncObservation struct {
	Echo             echoObservation `json:"echo"`
	SubmissionReturn uint32          `json:"submission_return"`
	SubmissionError  uint32          `json:"submission_error"`
	SubmissionMS     float64         `json:"submission_ms"`
	Pending          bool            `json:"pending"`
	WaitReturn       uint32          `json:"wait_return"`
	EventSignaled    bool            `json:"event_signaled"`
	BuffersUnpinned  bool            `json:"buffers_unpinned"`
	GCDuringPending  bool            `json:"gc_during_pending"`
	EventClosed      bool            `json:"event_closed"`
}

// A failed completion check must not unpin buffers that Windows may still own.
// This diagnostic-only retention ends with the process; no subsequent probe runs.
var unresolvedAsyncPins []*runtime.Pinner

func asyncEcho(ip netip.Addr, timeout uint32) (result asyncObservation) {
	ip = ip.Unmap()
	o := &result.Echo
	*o = echoObservation{Case: "async-event", Destination: ip.String(), Family: "ipv4", Scope: scope(ip), API: "IcmpSendEcho2(event)", HopLimit: 64, TimeoutMS: timeout, RequestSize: 8, ReplySize: 512}
	create, send, parse := create4, send4, ipHelper.NewProc("IcmpParseReplies")
	if ip.Is6() {
		o.Family, o.API = "ipv6", "Icmp6SendEcho2(event)"
		create, send, parse = create6, send6, ipHelper.NewProc("Icmp6ParseReplies")
	}
	kernel := syscall.NewLazyDLL("kernel32.dll")
	handle, _, err := create.Call()
	if handle == 0 || handle == ^uintptr(0) {
		o.Outcome, o.NativeError, o.Error = "api_error", errno(err), err.Error()
		return
	}
	event, _, err := kernel.NewProc("CreateEventW").Call(0, 1, 0, 0)
	if event == 0 {
		closeICMP.Call(handle)
		o.Outcome, o.Error = "api_error", err.Error()
		return
	}
	pins := new(runtime.Pinner)
	cleanupAllowed := true
	defer func() {
		if !cleanupAllowed {
			unresolvedAsyncPins = append(unresolvedAsyncPins, pins)
			o.Error += "; unresolved native operation: all native buffers/handles retained until diagnostic process exit"
			return
		}
		closed, _, closeErr := closeICMP.Call(handle)
		ok := closed != 0
		o.CloseOK = &ok
		if !ok {
			o.Error += "; close: " + closeErr.Error()
		}
		pins.Unpin()
		result.BuffersUnpinned = true
		closed, _, closeErr = kernel.NewProc("CloseHandle").Call(event)
		result.EventClosed = closed != 0
		if closed == 0 {
			o.Error += "; event close: " + closeErr.Error()
		}
	}()
	request, options, reply := new([8]byte), new([8]byte), new([64]uint64)
	copy(request[:], []byte("hv-check"))
	options[0] = 64
	pins.Pin(&request[0])
	pins.Pin(&options[0])
	pins.Pin(&reply[0])
	var source, destination *sockaddr
	if ip.Is6() {
		source, destination = new(sockaddr), new(sockaddr)
		src, _ := socketAddress(netip.IPv6Unspecified())
		dst, addrErr := socketAddress(ip)
		if addrErr != nil {
			o.Outcome, o.Error = "unsupported_scope", addrErr.Error()
			return
		}
		*source, *destination = src, dst
		pins.Pin(source)
		pins.Pin(destination)
	}
	start := time.Now()
	o.StartedAt = start.UTC()
	var count uintptr
	if ip.Is4() {
		b := ip.As4()
		count, _, err = send.Call(handle, event, 0, 0, uintptr(binary.LittleEndian.Uint32(b[:])), uintptr(unsafe.Pointer(&request[0])), 8, uintptr(unsafe.Pointer(&options[0])), uintptr(unsafe.Pointer(&reply[0])), 512, uintptr(timeout))
	} else {
		count, _, err = send.Call(handle, event, 0, 0, uintptr(unsafe.Pointer(source)), uintptr(unsafe.Pointer(destination)), uintptr(unsafe.Pointer(&request[0])), 8, uintptr(unsafe.Pointer(&options[0])), uintptr(unsafe.Pointer(&reply[0])), 512, uintptr(timeout))
	}
	result.SubmissionMS = milliseconds(time.Since(start))
	result.SubmissionReturn, result.SubmissionError = uint32(count), errno(err)
	result.Pending = (count == 0 && result.SubmissionError == 997) || count == 997
	if count == 0 && !result.Pending {
		o.Outcome, o.NativeError, o.Error = outcome(errno(err), false), errno(err), err.Error()
		return
	}
	if result.Pending {
		runtime.GC()
		result.GCDuringPending = true
	}
	wait, _, waitErr := kernel.NewProc("WaitForSingleObject").Call(event, 2000)
	result.WaitReturn, result.EventSignaled = uint32(wait), wait == 0
	o.EndedAt, o.ElapsedMS = time.Now().UTC(), milliseconds(time.Since(start))
	if wait != 0 {
		cleanupAllowed = false
		o.Outcome, o.Error = "completion_unverified", fmt.Sprintf("event not signaled: wait=%d error=%v", wait, waitErr)
		return
	}
	count, _, err = parse.Call(uintptr(unsafe.Pointer(&reply[0])), 512)
	o.ReplyCount = uint32(count)
	if count == 0 {
		o.NativeError, o.Outcome, o.Error = errno(err), outcome(errno(err), false), err.Error()
		return
	}
	decodeEchoReply(o, ip, unsafe.Slice((*byte)(unsafe.Pointer(&reply[0])), 512))
	return
}
