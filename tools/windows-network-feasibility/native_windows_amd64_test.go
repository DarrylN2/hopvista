package main

import (
	"context"
	"encoding/binary"
	"net/netip"
	"testing"
	"time"
	"unsafe"
)

func TestWindowsLayoutsAndScopes(t *testing.T) {
	if unsafe.Sizeof(routeRow{}) != 104 || unsafe.Offsetof(routeRow{}.Next) != 44 {
		t.Fatal("route ABI layout changed")
	}
	for _, raw := range []string{"127.0.0.1", "::1", "fe80::1234%7", "::ffff:192.0.2.1"} {
		ip := netip.MustParseAddr(raw)
		b, err := socketAddress(ip)
		if err != nil {
			t.Fatal(err)
		}
		got, err := readSockaddr(b[:])
		if err != nil || got != ip.Unmap() {
			t.Fatalf("%s -> %s: %v", ip, got, err)
		}
	}
	if _, err := socketAddress(netip.MustParseAddr("fe80::1")); err == nil {
		t.Fatal("accepted missing scope")
	}
	b, err := socketAddress(netip.MustParseAddr("fe80::1%12"))
	if err != nil || binary.LittleEndian.Uint32(b[24:]) != 12 {
		t.Fatal("scope not encoded at sockaddr offset 24")
	}
}

func TestTimeoutDoesNotExceedRemainingBudget(t *testing.T) {
	now := time.Now()
	ctx, cancel := context.WithDeadline(context.Background(), now.Add(150*time.Millisecond))
	defer cancel()
	n, err := nativeTimeout(ctx, 750*time.Millisecond, now.Add(500*time.Microsecond))
	if err != nil || n != 149 {
		t.Fatalf("timeout=%d err=%v", n, err)
	}
	if _, err := nativeTimeout(ctx, time.Second, now.Add(150*time.Millisecond)); err == nil {
		t.Fatal("started at cutoff")
	}
	cancel()
	if _, err := nativeTimeout(ctx, time.Second, now); err == nil {
		t.Fatal("started after cancellation")
	}
}

func TestFullBudgetGuardAtSendBoundary(t *testing.T) {
	now := time.Now()
	ctx, cancel := context.WithDeadline(context.Background(), now.Add(800*time.Millisecond))
	defer cancel()
	if !fullBudgetAvailable(ctx, 750*time.Millisecond, 50*time.Millisecond, now) {
		t.Fatal("refused exact full budget")
	}
	if fullBudgetAvailable(ctx, 750*time.Millisecond, 50*time.Millisecond, now.Add(time.Nanosecond)) {
		t.Fatal("accepted shortened full budget")
	}
}

func TestGuardedTimeoutNeverTrimsFullCap(t *testing.T) {
	now := time.Now()
	const cap = 750 * time.Millisecond
	const margin = 250 * time.Millisecond
	ctx, cancel := context.WithDeadline(context.Background(), now.Add(cap+margin))
	defer cancel()
	if got, err := guardedNativeTimeout(ctx, cap, margin, now); err != nil || got != 750 {
		t.Fatalf("exact budget: timeout=%d err=%v", got, err)
	}
	// Simulate a scheduling pause after an earlier guard: the native timeout
	// must be refused, never shortened to the remaining budget.
	if got, err := guardedNativeTimeout(ctx, cap, margin, now.Add(100*time.Millisecond)); err == nil || got != 0 {
		t.Fatalf("post-guard pause: timeout=%d err=%v", got, err)
	}
	if got, err := guardedNativeTimeout(context.Background(), cap+500*time.Microsecond, margin, now); err == nil || got != 0 {
		t.Fatalf("fractional cap: timeout=%d err=%v", got, err)
	}
	cancel()
	if got, err := guardedNativeTimeout(ctx, cap, margin, now); err == nil || got != 0 {
		t.Fatalf("cancelled: timeout=%d err=%v", got, err)
	}
}

func TestNativeFailureIsNotAnUnreachableReply(t *testing.T) {
	for _, test := range []struct {
		code  uint32
		reply bool
		want  string
	}{
		{0, true, "echo_reply"}, {11013, true, "time_exceeded_reply"}, {11041, true, "time_exceeded_reply"},
		{11003, true, "unreachable_reply"}, {11003, false, "api_error"}, {11010, false, "timeout"}, {5, false, "api_error"},
	} {
		if got := outcome(test.code, test.reply); got != test.want {
			t.Fatalf("%+v got %s", test, got)
		}
	}
}

func TestDecodeIPv4UnreachableReply(t *testing.T) {
	// ICMP_ECHO_REPLY32 scalar fields: responder at 0, status at 4,
	// round-trip time at 8, and reply data size at 12.
	buffer := make([]byte, 512)
	copy(buffer[:4], []byte{192, 0, 2, 254})
	binary.LittleEndian.PutUint32(buffer[4:8], 11003) // IP_DEST_HOST_UNREACHABLE
	binary.LittleEndian.PutUint32(buffer[8:12], 7)
	binary.LittleEndian.PutUint16(buffer[12:14], 8)
	o := echoObservation{ReplyCount: 1}
	decodeEchoReply(&o, netip.MustParseAddr("192.0.2.1"), buffer)
	if o.Outcome != "unreachable_reply" || o.Status == nil || *o.Status != 11003 || o.ReplyFrom != "192.0.2.254" || o.RTT == nil || *o.RTT != 7 || o.DataSize == nil || *o.DataSize != 8 || o.NativeError != 0 {
		t.Fatalf("unreachable reply decoded incorrectly: %+v", o)
	}
	if got := outcome(11003, false); got != "api_error" {
		t.Fatalf("native error misclassified as reply: %s", got)
	}
}

func TestCancelledEchoDoesNotCreateNativeHandle(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got := echo(ctx, "cancel-before-start", netip.MustParseAddr("127.0.0.1"), 64, time.Second, false, 512, func() { t.Fatal("entered native send") })
	if got.Outcome != "not_started" || got.CloseOK != nil || !got.StartedAt.IsZero() {
		t.Fatalf("unexpected native work after cancellation: %+v", got)
	}
}
