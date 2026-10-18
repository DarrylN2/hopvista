# Milestone 19: production-duration cutoff with a held executor

On 2026-10-15, an opt-in run-manager test used the actual production 60-second measurement window. A scripted executor remained pending across the deadline to model a synchronous route call that does not return on context cancellation. Two CPU-bound Go workers ran during the final five seconds with `GOMAXPROCS=1`, and both made progress. The manager observed `stopping` about 15.95 ms after the nominal deadline, retained its busy slot and withheld findings while the executor was held, then recorded terminal quiescence after the test released it. In this one run, the terminal timestamp was about 20.81 ms after the observed stop.

Run from the repository root:

```powershell
$env:GOCACHE = Join-Path (Get-Location) '.tools/go-build'
$env:GOTOOLCHAIN = 'local'
$env:HOPVISTA_LIVE_MINUTE_CUTOFF = '1'
.\.tools\go\bin\go.exe test ./internal/run -run TestLiveMinuteCutoffUnderProcessLoad -count=1 -v
```

The held executor is simulated. This check demonstrates wall-clock cutoff and lifecycle ordering at the production duration on this machine; it does not show a genuinely stalled native `GetBestRoute2`, set a native drain bound, or verify another Windows account/build. The 250 ms native start margin remains empirical. Release acceptance remains open.
