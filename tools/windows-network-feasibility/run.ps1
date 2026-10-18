param(
    [string]$Go = '',
    [string]$IPv4 = '1.1.1.1',
    [string]$IPv6 = '2606:4700:4700::1111',
    [string]$NoReplyIPv4 = '192.0.2.1',
    [string]$NoReplyIPv6 = '2001:db8::1',
    [ValidateRange(1, 64)][int]$Hops = 4,
    [switch]$TimeoutCheck,
    [switch]$TimeoutSweep,
    [switch]$NativeTimeoutComparison,
    [switch]$AsyncTimeoutCheck,
    [switch]$GuardedCutoffCheck,
    [switch]$GuardDelayCheck,
    [switch]$CutoffCapsCheck,
    [switch]$MarginStudy,
    [switch]$MarginBoundaryCheck,
    [switch]$LoadedMarginBoundaryCheck,
    [switch]$LoadedMarginSamplingCheck,
    [switch]$MinuteCutoffCheck,
    [switch]$RouteLatencyStudy,
    [switch]$RouteCutoffCheck,
    [switch]$DrainCapsCheck,
    [switch]$LoadedDrainCapsCheck,
    [switch]$UnreachableCheck,
    [string]$UnreachableIPv4 = ''
)
$ErrorActionPreference = 'Stop'
$harnessRoot = $PSScriptRoot
$repoRoot = Split-Path (Split-Path $harnessRoot -Parent) -Parent
if (-not $Go) {
    $portableGo = Join-Path $repoRoot '.tools/go/bin/go.exe'
    if (Test-Path -LiteralPath $portableGo) { $Go = $portableGo }
    else { $Go = (Get-Command go -ErrorAction Stop).Source }
}
$Go = (Resolve-Path -LiteralPath $Go).Path
$captureRoot = Join-Path $harnessRoot '.local'
New-Item -ItemType Directory -Force $captureRoot | Out-Null
$env:GOCACHE = Join-Path $repoRoot '.tools/go-build'
$env:GOTOOLCHAIN = 'local'
Push-Location $harnessRoot
try {
    & $Go version
    if ($UnreachableCheck) {
        if (-not $UnreachableIPv4) { throw 'Supply -UnreachableIPv4 for the controlled fixture' }
        $env:HOPVISTA_LIVE_UNREACHABLE = '1'
        $env:HOPVISTA_UNREACHABLE_TARGET = $UnreachableIPv4
        & $Go test -run '^TestLiveIPv4UnreachableReply$' -count=1 -v .
        if ($LASTEXITCODE -ne 0) { throw 'Controlled unreachable check failed; see .local/unreachable-fixture.json' }
        return
    }
    if ($DrainCapsCheck -or $LoadedDrainCapsCheck) {
        $env:HOPVISTA_DRAIN_CAPS = '1'
        $env:HOPVISTA_DRAIN_LOAD = if ($LoadedDrainCapsCheck) { '1' } else { '0' }
        & $Go test -run '^TestLiveIPv4DrainCaps$' -count=1 -v .
        if ($LASTEXITCODE -ne 0) { throw 'IPv4 drain-cap study failed; see .local/drain-caps.json or .local/loaded-drain-caps.json' }
        return
    }
    if ($RouteCutoffCheck) {
        $env:HOPVISTA_ROUTE_CUTOFF = '1'
        & $Go test -run '^TestLiveRouteThenProbeCutoff$' -count=1 -v .
        if ($LASTEXITCODE -ne 0) { throw 'Route/cutoff ordering check failed; see .local/route-cutoff.json' }
        return
    }
    if ($GuardDelayCheck) {
        $env:HOPVISTA_GUARD_DELAY = '1'
        & $Go test -run '^TestLiveGuardAfterSetupDelay$' -count=1 -v .
        if ($LASTEXITCODE -ne 0) { throw 'Post-setup guard check failed' }
        return
    }
    if ($RouteLatencyStudy) {
        $env:HOPVISTA_ROUTE_LATENCY = '1'
        & $Go test -run '^TestLiveRouteLatency$' -count=1 -v .
        if ($LASTEXITCODE -ne 0) { throw 'Route latency characterization failed; see .local/route-latency.json' }
        return
    }
    if ($MarginStudy) {
        $env:HOPVISTA_MARGIN_STUDY = '1'
        & $Go test -run '^TestLiveMarginStudy$' -count=1 -v .
        if ($LASTEXITCODE -ne 0) { throw 'Margin characterization failed; see .local/margin-study.json' }
        return
    }
    if ($MarginBoundaryCheck -or $LoadedMarginBoundaryCheck) {
        $env:HOPVISTA_MARGIN_BOUNDARY = '1'
        $env:HOPVISTA_MARGIN_LOAD = if ($LoadedMarginBoundaryCheck) { '1' } else { '0' }
        & $Go test -run '^TestLiveCandidateMarginBoundary$' -count=1 -v .
        if ($LASTEXITCODE -ne 0) { throw 'Candidate margin boundary check failed; see .local/margin-boundary.json or .local/loaded-margin-boundary.json' }
        return
    }
    if ($LoadedMarginSamplingCheck) {
        $env:HOPVISTA_LOADED_MARGIN_SAMPLING = '1'
        & $Go test -run '^TestLiveLoadedMarginSampling$' -count=1 -timeout 2m -v .
        if ($LASTEXITCODE -ne 0) { throw 'Loaded margin sampling check failed; see .local/loaded-margin-sampling.json' }
        return
    }
    if ($MinuteCutoffCheck) {
        $env:HOPVISTA_MINUTE_CUTOFF = '1'
        & $Go test -run '^TestLiveMinuteCutoffUnderSchedulerLoad$' -count=1 -timeout 3m -v .
        if ($LASTEXITCODE -ne 0) { throw 'Loaded minute cutoff check failed; see .local/minute-cutoff.json' }
        return
    }
    if ($CutoffCapsCheck) {
        $env:HOPVISTA_CUTOFF_CAPS = '1'
        & $Go test -run '^TestLiveCutoffCaps$' -count=1 -v .
        if ($LASTEXITCODE -ne 0) { throw 'Cutoff cap investigation failed; see .local/cutoff-cap-observations.json' }
        return
    }
    if ($GuardedCutoffCheck) {
        $env:HOPVISTA_GUARDED_CUTOFF = '1'
        & $Go test -run '^TestLiveGuardedCutoff$' -count=1 -v .
        if ($LASTEXITCODE -ne 0) { throw 'Guarded cutoff investigation failed; see .local/guarded-cutoff-observations.json' }
        return
    }
    if ($AsyncTimeoutCheck) {
        $env:HOPVISTA_ASYNC_TIMEOUT = '1'
        & $Go test -run '^TestLiveAsyncTimeout$' -count=1 -v .
        if ($LASTEXITCODE -ne 0) { throw 'Asynchronous timeout investigation failed; see .local/async-timeout-observations.json' }
        return
    }
    if ($TimeoutSweep) {
        $env:HOPVISTA_TIMEOUT_SWEEP = '1'
        & $Go test -run '^TestLiveTimeoutSweep$' -count=1 -v .
        if ($LASTEXITCODE -ne 0) { throw 'Timeout sweep capture failed' }
        return
    }
    if ($NativeTimeoutComparison) {
        & (Join-Path $harnessRoot 'debug/native-timeout-comparison.ps1') -OutputDirectory $captureRoot
        return
    }
    if ($TimeoutCheck) {
        $env:HOPVISTA_LIVE_TIMEOUT = '1'
        & $Go test -run '^TestLiveNativeTimeout$' -count=1 -v .
        if ($LASTEXITCODE -ne 0) { throw 'Native timeout regression check failed; see .local/timeout-observations.json' }
        return
    }
    & $Go test ./...
    if ($LASTEXITCODE -ne 0) { throw 'Go tests failed' }
    & $Go vet ./...
    if ($LASTEXITCODE -ne 0) { throw 'Go vet failed' }
    & $Go build -o .local/feasibility.exe .
    if ($LASTEXITCODE -ne 0) { throw 'Go build failed' }
    $capture = & ./.local/feasibility.exe -ipv4 $IPv4 -ipv6 $IPv6 -no-reply-ipv4 $NoReplyIPv4 -no-reply-ipv6 $NoReplyIPv6 -hops $Hops
    if ($LASTEXITCODE -ne 0) { throw 'Go capture failed' }
    $capture | Set-Content -LiteralPath (Join-Path $captureRoot 'go-capture.json') -Encoding utf8
    $result = $capture | ConvertFrom-Json
    # Independent Windows observations; not substitutes for actual Go calls.
    $routeErrors = @()
    $windowsRoutes = @()
    $adapters = @()
    try {
        $windowsRoutes = @(Get-NetRoute -ErrorAction Stop |
            Select-Object AddressFamily, DestinationPrefix, NextHop, InterfaceIndex, RouteMetric)
        $adapters = @(Get-NetAdapter -IncludeHidden -ErrorAction Stop |
            Select-Object InterfaceIndex, Status, InterfaceDescription)
    } catch { $routeErrors += $_ }
    $selected = @()
    foreach ($destination in @($result.routes | Where-Object api_called | Select-Object -ExpandProperty destination -Unique)) {
        try {
            $selected += @(Find-NetRoute -RemoteIPAddress $destination -ErrorAction Stop |
                Select-Object IPAddress, InterfaceIndex, DestinationPrefix, NextHop)
        } catch { $routeErrors += $_ }
    }
    [pscustomobject]@{
        captured_at = [DateTime]::UtcNow.ToString('o')
        routes = $windowsRoutes
        selected = $selected
        adapters = $adapters
        errors = @($routeErrors | ForEach-Object { $_.ToString() })
    } | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath (Join-Path $captureRoot 'windows-routes.json') -Encoding utf8
    $result.environment | ConvertTo-Json
    $result.echoes | Select-Object case, family, ttl_or_hop_limit, reply_count, reply_status, native_error, outcome, elapsed_ms, close_ok | Format-Table
    $result.drains | Select-Object case, pending_at_stop, calls_started, following_call_started, drain_ms, context_error | Format-Table
    Write-Output "Local raw captures: $captureRoot (may contain private network addresses). Review required; this script does not declare the gate passed."
} finally { Pop-Location }
