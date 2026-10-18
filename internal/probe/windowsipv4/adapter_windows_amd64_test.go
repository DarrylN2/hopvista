//go:build windows && amd64

package windowsipv4

import (
	"context"
	"encoding/binary"
	"net/netip"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"hopvista/internal/probe"
	"hopvista/internal/result"
)

func TestRouteLayoutAndIPv4Address(t *testing.T) {
	if unsafe.Sizeof(routeRow{}) != 104 || unsafe.Offsetof(routeRow{}.Next) != 44 {
		t.Fatal("route ABI layout changed")
	}
	ip := netip.MustParseAddr("192.0.2.10")
	got, ok := readAddress(socketAddress(ip))
	if !ok || got != ip {
		t.Fatalf("address: %v %v", got, ok)
	}
}

func TestReplyClassification(t *testing.T) {
	for _, tc := range []struct {
		native  uint32
		want    result.EchoStatus
		outcome result.Outcome
	}{
		{0, result.EchoReply, result.Completed}, {11013, result.TimeExceeded, result.Completed},
		{11041, result.TimeExceeded, result.Completed}, {11003, result.UnreachableReply, result.Completed},
		{5, result.CallError, result.Error},
	} {
		var b [512]byte
		copy(b[:4], []byte{192, 0, 2, 254})
		binary.LittleEndian.PutUint32(b[4:8], tc.native)
		binary.LittleEndian.PutUint32(b[8:12], 7)
		got := decodeReply(result.Echo{}, b[:])
		if got.Status != tc.want || got.Outcome != tc.outcome {
			t.Fatalf("status %d: %+v", tc.native, got)
		}
		if tc.outcome == result.Completed && (got.ReplyAddress != "192.0.2.254" || got.RTTMS == nil || *got.RTTMS != 7) {
			t.Fatalf("reply evidence: %+v", got)
		}
		if tc.outcome != result.Completed && (got.ReplyAddress != "" || got.RTTMS != nil) {
			t.Fatalf("invented evidence: %+v", got)
		}
	}
}

func TestRefusalDoesNotEnterNative(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	deadline := time.Now().Add(3 * time.Second)
	route := (Adapter{}).Route(ctx, probe.RouteRequest{ID: "r", SelectedIP: netip.MustParseAddr("127.0.0.1"), DeadlineAt: deadline})
	echo := (Adapter{}).Echo(ctx, probe.EchoRequest{ID: "e", SelectedIP: netip.MustParseAddr("127.0.0.1"), DeadlineAt: deadline, TTL: 64, Cap: time.Second})
	if route.Outcome != result.Cancelled || route.StartedAt != nil || echo.Outcome != result.Cancelled || echo.StartedAt != nil || echo.Status != result.NotSent {
		t.Fatalf("unexpected native start: %+v %+v", route, echo)
	}
	short := (Adapter{}).Echo(context.Background(), probe.EchoRequest{ID: "short", SelectedIP: netip.MustParseAddr("127.0.0.1"), DeadlineAt: time.Now().Add(time.Second), TTL: 64, Cap: time.Second})
	if short.Outcome != result.Deadline || short.StartedAt != nil {
		t.Fatalf("full-cap refusal: %+v", short)
	}
}

func TestLiveLoopback(t *testing.T) {
	if os.Getenv("HOPVISTA_LIVE_ADAPTER") != "1" {
		t.Skip("set HOPVISTA_LIVE_ADAPTER=1 for opt-in Windows check")
	}
	ip := netip.MustParseAddr("127.0.0.1")
	deadline := time.Now().Add(5 * time.Second)
	a := Adapter{}
	r := a.Route(context.Background(), probe.RouteRequest{ID: "route", SelectedIP: ip, DeadlineAt: deadline})
	if r.Outcome != result.Completed || r.OnLink == nil || !*r.OnLink || r.SourceIP == "" || r.NextHop != "" {
		t.Fatalf("loopback route: %+v", r)
	}
	e := a.Echo(context.Background(), probe.EchoRequest{ID: "echo", SelectedIP: ip, TTL: 64, Cap: time.Second, DeadlineAt: deadline})
	if e.Outcome != result.Completed || e.Status != result.EchoReply || e.ReplyAddress != ip.String() || e.RTTMS == nil || e.StartedAt == nil || e.EndedAt == nil {
		t.Fatalf("loopback echo: %+v", e)
	}
}

// This opt-in check needs a path that stays pending through the 25 ms stop
// request and returns a native timeout. It records one tested condition only.
func TestLivePendingEchoCancellation(t *testing.T) {
	rawIP := os.Getenv("HOPVISTA_LIVE_DRAIN_IP")
	if rawIP == "" {
		t.Skip("set HOPVISTA_LIVE_DRAIN_IP to a candidate no-reply IPv4 address")
	}
	ip, err := netip.ParseAddr(rawIP)
	if err != nil || !probe.ValidIPv4(ip) {
		t.Fatalf("invalid IPv4 address %q", rawIP)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	deadline := time.Now().Add(5 * time.Second)
	request := probe.EchoRequest{ID: "pending", SelectedIP: ip, TTL: 64, Cap: time.Second, DeadlineAt: deadline}
	done := make(chan result.Echo, 1)
	go func() { done <- (Adapter{}).Echo(ctx, request) }()
	time.Sleep(25 * time.Millisecond)
	select {
	case early := <-done:
		t.Fatalf("candidate did not remain pending: %+v", early)
	default:
	}
	stopped := time.Now().UTC()
	cancel()
	refused := (Adapter{}).Echo(ctx, probe.EchoRequest{ID: "after-stop", SelectedIP: ip, TTL: 64, Cap: time.Second, DeadlineAt: deadline})
	if refused.Outcome != result.Cancelled || refused.StartedAt != nil || refused.Status != result.NotSent || refused.ReplyAddress != "" || refused.RTTMS != nil {
		t.Fatalf("new echo started after cancellation: %+v", refused)
	}
	first := <-done // retain the in-flight call until native return and handle cleanup
	if first.Outcome != result.Timeout || first.Status != result.NoReply || first.StartedAt == nil || first.EndedAt == nil || first.StartedAt.After(stopped) || first.EndedAt.Before(stopped) || first.ReplyAddress != "" || first.RTTMS != nil {
		t.Fatalf("pending native echo did not drain as observed: %+v", first)
	}
	t.Logf("pending native echo returned after cancellation and cleanup; stop-to-return %.1f ms", first.EndedAt.Sub(stopped).Seconds()*1000)
}

// This samples the production adapter's empirical full-cap start policy on
// one explicit no-reply path. It does not establish a worst-case bound.
func TestLiveStartMarginSampling(t *testing.T) {
	rawIP := os.Getenv("HOPVISTA_LIVE_DRAIN_IP")
	if rawIP == "" {
		t.Skip("set HOPVISTA_LIVE_DRAIN_IP to a candidate no-reply IPv4 address")
	}
	ip, err := netip.ParseAddr(rawIP)
	if err != nil || !probe.ValidIPv4(ip) {
		t.Fatalf("invalid IPv4 address %q", rawIP)
	}
	for _, cap := range []time.Duration{750 * time.Millisecond, 900 * time.Millisecond, time.Second} {
		t.Run(cap.String(), func(t *testing.T) {
			refused := (Adapter{}).Echo(context.Background(), probe.EchoRequest{
				ID: "short-budget", SelectedIP: ip, TTL: 64, Cap: cap,
				DeadlineAt: time.Now().Add(cap + probe.InitialStartMargin - time.Millisecond),
			})
			if refused.Outcome != result.Deadline || refused.StartedAt != nil || refused.Status != result.NotSent || refused.ReplyAddress != "" || refused.RTTMS != nil {
				t.Fatalf("full-cap refusal: %+v", refused)
			}
			var largestOverrun time.Duration
			for sample := 0; sample < 3; sample++ {
				checkedAt := time.Now().UTC()
				deadline := checkedAt.Add(cap + probe.InitialStartMargin + 50*time.Millisecond)
				echo := (Adapter{}).Echo(context.Background(), probe.EchoRequest{
					ID: "full-cap", SelectedIP: ip, TTL: 64, Cap: cap, DeadlineAt: deadline,
				})
				if echo.Outcome != result.Timeout || echo.Status != result.NoReply || echo.StartedAt == nil || echo.EndedAt == nil || echo.EndedAt.After(deadline) || echo.ReplyAddress != "" || echo.RTTMS != nil {
					t.Fatalf("full-cap sample %d: %+v deadline=%s", sample+1, echo, deadline)
				}
				if overrun := echo.EndedAt.Sub(checkedAt) - cap; overrun > largestOverrun {
					largestOverrun = overrun
				}
			}
			t.Logf("three full-cap no-reply calls returned and closed before their fixture deadlines; largest check-to-return overrun %s", largestOverrun)
		})
	}
}

// This repeats the production-adapter sample under process-local contention.
// Safe pre-send refusals are counted, not reclassified as completed calls.
func TestLiveLoadedStartMarginSampling(t *testing.T) {
	rawIP := os.Getenv("HOPVISTA_LIVE_DRAIN_IP")
	if rawIP == "" {
		t.Skip("set HOPVISTA_LIVE_DRAIN_IP to a candidate no-reply IPv4 address")
	}
	ip, err := netip.ParseAddr(rawIP)
	if err != nil || !probe.ValidIPv4(ip) {
		t.Fatalf("invalid IPv4 address %q", rawIP)
	}
	previousProcs := runtime.GOMAXPROCS(1)
	stopWorkers := make(chan struct{})
	var workers sync.WaitGroup
	var progress [2]atomic.Uint64
	for i := 0; i < 2; i++ {
		index := i
		workers.Add(1)
		go func() {
			defer workers.Done()
			var n uint64
			for {
				select {
				case <-stopWorkers:
					return
				default:
				}
				for j := 0; j < 1000; j++ {
					n++
				}
				progress[index].Add(n)
			}
		}()
	}
	defer func() {
		close(stopWorkers)
		workers.Wait()
		runtime.GOMAXPROCS(previousProcs)
	}()
	readyUntil := time.Now().Add(2 * time.Second)
	for (progress[0].Load() == 0 || progress[1].Load() == 0) && time.Now().Before(readyUntil) {
		time.Sleep(time.Millisecond)
	}
	if progress[0].Load() == 0 || progress[1].Load() == 0 {
		t.Fatal("CPU workers did not start")
	}
	for _, cap := range []time.Duration{750 * time.Millisecond, 900 * time.Millisecond, time.Second} {
		t.Run(cap.String(), func(t *testing.T) {
			started, refused := 0, 0
			for attempts := 0; attempts < 5 && started < 3; attempts++ {
				beforeProgress := [2]uint64{progress[0].Load(), progress[1].Load()}
				checkedAt := time.Now().UTC()
				deadline := checkedAt.Add(cap + probe.InitialStartMargin + 50*time.Millisecond)
				echo := (Adapter{}).Echo(context.Background(), probe.EchoRequest{
					ID: "loaded-full-cap", SelectedIP: ip, TTL: 64, Cap: cap, DeadlineAt: deadline,
				})
				if echo.Outcome == result.Deadline && echo.StartedAt == nil && echo.Status == result.NotSent && echo.ReplyAddress == "" && echo.RTTMS == nil {
					refused++
					continue
				}
				if echo.Outcome != result.Timeout || echo.Status != result.NoReply || echo.StartedAt == nil || echo.EndedAt == nil || echo.EndedAt.After(deadline) || echo.ReplyAddress != "" || echo.RTTMS != nil {
					t.Fatalf("loaded full-cap attempt %d: %+v deadline=%s", attempts+1, echo, deadline)
				}
				if progress[0].Load() <= beforeProgress[0] || progress[1].Load() <= beforeProgress[1] {
					t.Fatal("CPU workers made no progress during native call")
				}
				started++
			}
			if started != 3 {
				t.Fatalf("only %d full-cap calls started in five attempts; safe refusals=%d", started, refused)
			}
			t.Logf("three loaded full-cap calls returned before fixture deadlines; safe pre-send refusals=%d", refused)
		})
	}
}
