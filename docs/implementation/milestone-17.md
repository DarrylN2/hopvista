# Milestone 17: loaded production adapter start-margin sample

On 2026-10-15, an opt-in production Windows IPv4 adapter test repeated the full-cap start policy under process-local load. Two CPU-bound Go workers were active with `GOMAXPROCS=1`; both made progress during each accepted native call. The test used candidate no-reply address `192.0.2.1`, the 750, 900, and 1,000 ms caps, and fixture budgets of cap + the empirical 250 ms margin + 50 ms. Across at most five attempts per cap, it required three started full-cap calls. One 750 ms attempt safely refused before send; the next three started and returned before their fixture deadlines. Three calls at each of the 900 and 1,000 ms caps also started and returned before their fixture deadlines. The test joined both workers and restored `GOMAXPROCS` afterward.

Run from the repository root where ICMP calls are permitted:

```powershell
$env:GOCACHE = Join-Path (Get-Location) '.tools/go-build'
$env:GOTOOLCHAIN = 'local'
$env:HOPVISTA_LIVE_DRAIN_IP = '192.0.2.1'
.\.tools\go\bin\go.exe test ./internal/probe/windowsipv4 -run TestLiveLoadedStartMarginSampling -count=1 -v
```

The safe refusal remains part of the result; it is not counted as a started call. The earlier failed loaded boundary matrix in the standalone harness also remains failed evidence. This pass samples one process-load condition on the current machine, not system-wide stress or a worst-case scheduling/native return bound. Another Windows account/build/network, a genuinely stalled `GetBestRoute2`, and broader packaged-executable timing remain unverified. Release acceptance remains open.
