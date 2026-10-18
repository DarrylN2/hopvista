# Milestone 16: production adapter start-margin sample

On 2026-10-15, an opt-in test exercised the production Windows IPv4 echo adapter against candidate no-reply address `192.0.2.1` on the current machine and network. For each configured cap (750, 900, and 1,000 ms), a budget of cap + 249 ms was refused as unstarted `deadline` evidence with no send timestamp, address, or RTT. Three further calls per cap used the full native cap with a fixture budget of cap + the empirical 250 ms margin + 50 ms. All nine started, returned a no-reply timeout, and completed adapter cleanup before their fixture deadlines. The largest observed positive check-to-return overrun above the configured cap was under 1 ms in this sample.

Run from the repository root where ICMP calls are permitted:

```powershell
$env:GOCACHE = Join-Path (Get-Location) '.tools/go-build'
$env:GOTOOLCHAIN = 'local'
$env:HOPVISTA_LIVE_DRAIN_IP = '192.0.2.1'
.\.tools\go\bin\go.exe test ./internal/probe/windowsipv4 -run TestLiveStartMarginSampling -count=1 -v
```

The `+50 ms` fixture headroom is additional to the product's 250 ms start margin; it gives this sampling test room for setup and scheduler variation. These are unloaded local observations, not a worst-case completion guarantee or broader validation of the margin. The earlier failed loaded boundary experiment remains failed evidence. A genuinely stalled `GetBestRoute2`, another Windows account/build/network, and broader packaged-executable timing remain unverified. Release acceptance remains open.
