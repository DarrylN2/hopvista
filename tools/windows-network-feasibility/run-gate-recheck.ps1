param(
    [string]$Go = '',
    [string]$IPv4 = '1.1.1.1',
    [Parameter(Mandatory = $true)][string]$ConditionLabel,
    [switch]$PlanOnly
)

$ErrorActionPreference = 'Stop'
$harnessRoot = $PSScriptRoot
$runner = Join-Path $harnessRoot 'run.ps1'
$platform = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString()
$osVersion = [System.Environment]::OSVersion.Version.ToString()
$identity = [System.Security.Principal.WindowsIdentity]::GetCurrent()
$principal = [System.Security.Principal.WindowsPrincipal]::new($identity)
if ($platform -ne 'X64' -or [System.Environment]::OSVersion.Version.Build -lt 22000) {
    throw 'This recheck requires Windows 11 x64; record other platforms separately.'
}
if ($principal.IsInRole([System.Security.Principal.WindowsBuiltInRole]::Administrator)) {
    throw 'Run the recheck unelevated.'
}

$steps = @(
    [pscustomobject]@{ name = 'base'; switch = ''; captures = @('go-capture.json', 'windows-routes.json') },
    [pscustomobject]@{ name = 'margin-boundary'; switch = '-MarginBoundaryCheck'; captures = @('margin-boundary.json') },
    [pscustomobject]@{ name = 'loaded-margin'; switch = '-LoadedMarginSamplingCheck'; captures = @('loaded-margin-sampling.json') },
    [pscustomobject]@{ name = 'loaded-drain'; switch = '-LoadedDrainCapsCheck'; captures = @('loaded-drain-caps.json') },
    [pscustomobject]@{ name = 'unloaded-drain'; switch = '-DrainCapsCheck'; captures = @('drain-caps.json') },
    [pscustomobject]@{ name = 'setup-guard'; switch = '-GuardDelayCheck'; captures = @('guard-delay.json') },
    [pscustomobject]@{ name = 'route-cutoff'; switch = '-RouteCutoffCheck'; captures = @('route-cutoff.json') },
    [pscustomobject]@{ name = 'minute-cutoff'; switch = '-MinuteCutoffCheck'; captures = @('minute-cutoff.json') }
)
if ($PlanOnly) {
    $steps | Select-Object name, switch, @{Name='expected_captures'; Expression={$_.captures -join ', '}} | Format-Table
    return
}

$captureRoot = Join-Path $harnessRoot '.local'
$runFolder = Join-Path $captureRoot ('gate-recheck-' + [DateTime]::UtcNow.ToString('yyyyMMddTHHmmssZ'))
New-Item -ItemType Directory -Path $runFolder -Force | Out-Null
$results = @()
foreach ($step in $steps) {
    $started = [DateTime]::UtcNow
    $log = Join-Path $runFolder ($step.name + '.log')
    $arguments = @{}
    if ($Go) { $arguments.Go = $Go }
    if ($step.name -eq 'base') { $arguments.IPv4 = $IPv4 }
    if ($step.switch) { $arguments[$step.switch.TrimStart('-')] = $true }
    $status = 'passed'
    try {
        & $runner @arguments *>&1 | Out-File -LiteralPath $log -Encoding utf8
        if ($LASTEXITCODE -ne 0) { throw "runner exit code $LASTEXITCODE" }
    } catch {
        $status = 'failed'
        $_ | Out-File -LiteralPath $log -Encoding utf8 -Append
    }
    $saved = @()
    foreach ($name in $step.captures) {
        $source = Join-Path $captureRoot $name
        if ((Test-Path -LiteralPath $source) -and (Get-Item -LiteralPath $source).LastWriteTimeUtc -ge $started) {
            Copy-Item -LiteralPath $source -Destination (Join-Path $runFolder $name)
            $saved += $name
        }
    }
    if ($saved.Count -ne $step.captures.Count) { $status = 'incomplete' }
    $results += [pscustomobject]@{
        step = $step.name
        status = $status
        started_at = $started.ToString('o')
        ended_at = [DateTime]::UtcNow.ToString('o')
        expected_captures = $step.captures
        saved_captures = $saved
        log = [System.IO.Path]::GetFileName($log)
    }
}

[pscustomobject]@{
    condition_label = $ConditionLabel
    account = $identity.Name
    os_version = $osVersion
    architecture = $platform
    account_kind = 'unverified; manually confirm whether this is a separate standard-user account'
    gate = 'open; this runner collects evidence and does not approve the margin or application infrastructure'
    steps = $results
} | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath (Join-Path $runFolder 'manifest.json') -Encoding utf8

$results | Select-Object step, status, @{Name='saved_captures'; Expression={$_.saved_captures -join ', '}} | Format-Table
Write-Output "Raw local evidence: $runFolder"
Write-Output 'Review and redact account, route, source, next-hop, and interface details before sharing. This script does not declare the gate passed.'
if (@($results | Where-Object status -ne 'passed').Count) { throw 'At least one recheck step failed or lacked a fresh capture; inspect the local logs and manifest.' }
