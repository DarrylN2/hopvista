//go:build windows && amd64

package pathmeasure_test

import (
	"context"
	"encoding/json"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"hopvista/internal/pathmeasure"
	"hopvista/internal/probe/windowsipv4"
	"hopvista/internal/result"
)

// Run explicitly with HOPVISTA_REMOTE_TRACE_IP set to a reachable public IPv4
// address. Tracert uses three probes per TTL; HopVista uses one, so only stable
// semantic observations are compared here, not exact RTTs or every hop.
func TestLiveRemoteTracertComparison(t *testing.T) {
	rawIP := os.Getenv("HOPVISTA_REMOTE_TRACE_IP")
	if rawIP == "" {
		t.Skip("set HOPVISTA_REMOTE_TRACE_IP for an opt-in remote comparison")
	}
	ip, err := netip.ParseAddr(rawIP)
	if err != nil || !ip.Is4() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsMulticast() || ip.IsUnspecified() {
		t.Fatalf("expected a public IPv4 literal, got %q", rawIP)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	path := pathmeasure.Measure(ctx, windowsipv4.Adapter{}, pathmeasure.RealClock{}, pathmeasure.Request{
		ID: "remote-comparison", SelectedIP: ip, DeadlineAt: time.Now().Add(20 * time.Second), MaxHops: 4, TargetRole: "final_host",
	})
	if path.Route.Outcome != result.Completed || path.Trace.Family != "ipv4" || path.Trace.SelectedIP != rawIP || len(path.Trace.Hops) == 0 {
		t.Fatalf("HopVista remote path did not complete: %+v", path)
	}
	compareWindowsRoute(t, ctx, rawIP, path.Route)
	tracert := filepath.Join(os.Getenv("SystemRoot"), "System32", "tracert.exe")
	cmd := exec.CommandContext(ctx, tracert, "-d", "-4", "-h", "4", "-w", "750", rawIP)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Windows tracert failed: %v %s", err, output)
	}
	firstTracertIP := ""
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != strconv.Itoa(1) {
			continue
		}
		if parsed, err := netip.ParseAddr(fields[len(fields)-1]); err == nil && parsed.Is4() {
			firstTracertIP = parsed.String()
		}
	}
	if firstTracertIP == "" || path.Trace.Hops[0].Status != result.TimeExceeded || path.Trace.Hops[0].ReplyAddress != firstTracertIP {
		t.Fatalf("first remote hop disagreed: HopVista=%+v tracert=%s", path.Trace.Hops[0], output)
	}
	t.Logf("same destination %s; first-hop addresses matched; HopVista outcome %s; Windows tracert used three probes per TTL", rawIP, path.Trace.Outcome)
}

func compareWindowsRoute(t *testing.T, ctx context.Context, ip string, route result.Route) {
	t.Helper()
	powershell := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	command := "Find-NetRoute -RemoteIPAddress " + ip + " | Select-Object InterfaceIndex,IPAddress,NextHop | ConvertTo-Json -Compress"
	output, err := exec.CommandContext(ctx, powershell, "-NoProfile", "-NonInteractive", "-Command", command).CombinedOutput()
	if err != nil {
		t.Fatalf("Windows route query failed: %v %s", err, output)
	}
	type row struct {
		InterfaceIndex uint32
		IPAddress      string
		NextHop        string
	}
	var rows []row
	if strings.HasPrefix(strings.TrimSpace(string(output)), "[") {
		if err := json.Unmarshal(output, &rows); err != nil {
			t.Fatalf("Windows route JSON: %v %s", err, output)
		}
	} else {
		var only row
		if err := json.Unmarshal(output, &only); err != nil {
			t.Fatalf("Windows route JSON: %v %s", err, output)
		}
		rows = []row{only}
	}
	var source, nextHop string
	var interfaceIndex uint32
	for _, row := range rows {
		if row.IPAddress != "" {
			source = row.IPAddress
			interfaceIndex = row.InterfaceIndex
		}
		if row.NextHop != "" {
			nextHop = row.NextHop
			if interfaceIndex != 0 && row.InterfaceIndex != interfaceIndex {
				t.Fatalf("Windows source and next-hop interfaces differ: %s", output)
			}
		}
	}
	if source == "" || route.InterfaceIndex == nil || route.OnLink == nil || route.SourceIP != source || *route.InterfaceIndex != interfaceIndex || route.NextHop != nextHop || *route.OnLink != (nextHop == "") {
		t.Fatalf("HopVista route disagreed with Windows route: HopVista=%+v Windows=%s", route, output)
	}
}
