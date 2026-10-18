# Milestone 14: production native echo cancellation sample

On 2026-10-15, an opt-in test exercised the production Windows IPv4 adapter against candidate no-reply address `192.0.2.1` with a full 1,000 ms native echo cap. The first call remained pending when cancellation was requested 25 ms after launch. An attempted second echo on the cancelled context returned `cancelled` without a native send timestamp, reply address, or RTT. The started call returned a native timeout/no-reply observation after cancellation; adapter return occurred after its handle-close path. The observed stop-to-return interval was about 640 ms on this current machine and network.

Run from the repository root where ICMP calls are permitted:

```powershell
$env:GOCACHE = Join-Path (Get-Location) '.tools/go-build'
$env:GOTOOLCHAIN = 'local'
$env:HOPVISTA_LIVE_DRAIN_IP = '192.0.2.1'
.\.tools\go\bin\go.exe test ./internal/probe/windowsipv4 -run TestLivePendingEchoCancellation -count=1 -v
```

The fixture requires an actually pending no-reply call; a fast reply or error fails rather than counting as drain evidence. This is one production-adapter sample, not a native latest-return bound or proof that a Go context interrupts `IcmpSendEcho2`. It does not exercise a genuinely stalled `GetBestRoute2`, another Windows account/build, wider network conditions, or the packaged app's stopping display. The 250 ms start margin remains empirical and release acceptance remains open.
