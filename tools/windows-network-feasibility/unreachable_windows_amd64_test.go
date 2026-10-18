package main

import (
	"context"
	"encoding/json"
	"net/netip"
	"os"
	"testing"
	"time"
)

// Opt-in live check using a separately started controlled IPv4 responder.
func TestLiveIPv4UnreachableReply(t *testing.T) {
	if os.Getenv("HOPVISTA_LIVE_UNREACHABLE") != "1" {
		t.Skip("requires the controlled WSL unreachable fixture")
	}
	if environment()["elevated"] == true {
		t.Fatal("requires an unelevated process")
	}
	ip, err := netip.ParseAddr(os.Getenv("HOPVISTA_UNREACHABLE_TARGET"))
	if err != nil || !ip.Is4() {
		t.Fatal("set HOPVISTA_UNREACHABLE_TARGET to the fixture's IPv4 address")
	}
	o := echo(context.Background(), "controlled-unreachable-v4", ip, 64, 750*time.Millisecond, false, 512, nil)
	if err := os.MkdirAll(".local", 0700); err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(struct {
		CapturedAt  time.Time       `json:"captured_at"`
		Environment map[string]any  `json:"environment"`
		Echo        echoObservation `json:"echo"`
		Gate        string          `json:"gate"`
	}{time.Now().UTC(), environment(), o, "live unreachable reply must be reviewed with fixture output"}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".local/unreachable-fixture.json", data, 0600); err != nil {
		t.Fatal(err)
	}
	t.Logf("reply_count=%d status=%v native_error=%d outcome=%s close_ok=%v", o.ReplyCount, o.Status, o.NativeError, o.Outcome, o.CloseOK != nil && *o.CloseOK)
	if o.ReplyCount == 0 || o.Status == nil || o.Outcome != "unreachable_reply" || o.NativeError != 0 || o.CloseOK == nil || !*o.CloseOK {
		t.Fatalf("controlled IPv4 unreachable reply not decoded: %+v", o)
	}
}
