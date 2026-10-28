param(
    [Parameter(Mandatory)][string]$Executable,
    [Parameter(Mandatory)][string]$ExpectedSHA256,
    [Parameter(Mandatory)][string]$OutputDirectory
)

# Exact packaged public-API evidence, not a broader Windows support test.
$ErrorActionPreference = 'Stop'
$image = (Resolve-Path -LiteralPath $Executable).Path
if ((Get-FileHash -LiteralPath $image -Algorithm SHA256).Hash -ne $ExpectedSHA256) {
    throw 'Candidate checksum differs from the recorded executable.'
}
$output = [IO.Path]::GetFullPath($OutputDirectory)
New-Item -ItemType Directory -Force -Path $output | Out-Null
$directory = Join-Path ([IO.Path]::GetTempPath()) ('hopvista-candidate-' + [guid]::NewGuid())
New-Item -ItemType Directory -Path $directory | Out-Null
$copy = Join-Path $directory 'hopvista.exe'
$process = $null
try {
    Copy-Item -LiteralPath $image -Destination $copy
    $start = [Diagnostics.ProcessStartInfo]::new($copy)
    $start.WorkingDirectory = $directory
    $start.UseShellExecute = $false
    $start.CreateNoWindow = $true
    $start.RedirectStandardError = $true
    $start.RedirectStandardOutput = $true
    $start.Environment['PATH'] = Join-Path $env:SystemRoot 'System32'
    $process = [Diagnostics.Process]::Start($start)
    $stdout = $process.StandardOutput.ReadToEndAsync()
    $announcement = $process.StandardError.ReadLineAsync()
    if (-not $announcement.Wait(10000)) { throw 'Candidate listener announcement timed out.' }
    if ($announcement.Result -notmatch 'HopVista: (http://127\.0\.0\.1:\d+)/') {
        throw 'Candidate did not announce its loopback listener.'
    }
    $base = $Matches[1]
    $stderr = $process.StandardError.ReadToEndAsync()
    $session = Invoke-RestMethod "$base/api/v1/session" -NoProxy
    $headers = @{ Origin = $base; 'X-HopVista-Token' = $session.mutation_token }
    $accepted = Invoke-RestMethod "$base/api/v1/runs" -NoProxy -Method Post -Headers $headers -ContentType 'application/json' -Body '{"target_url":"https://example.com/","max_hops":4}'
    $until = [DateTime]::UtcNow.AddSeconds(75)
    do {
        Start-Sleep -Milliseconds 100
        $run = Invoke-RestMethod ($base + $accepted.status_url) -NoProxy
    } while ($run.state -in @('running', 'stopping') -and [DateTime]::UtcNow -lt $until)
    $run | ConvertTo-Json -Depth 40 | Set-Content (Join-Path $output 'public-run.json')
    if ($run.state -ne 'completed' -or $run.schema_version -ne 1) { throw 'Public diagnostic run did not complete.' }
    $family = @($run.families | Where-Object family -eq 'ipv4')[0]
    if ($family.exchanges.Count -ne 1 -or $family.paths.Count -ne 1) { throw 'Expected one HTTPS exchange and path.' }
    $exchange = $family.exchanges[0]
    $path = $family.paths[0]
    $ip = $exchange.selected_ip
    if ($exchange.actual_peer_ip -ne $ip -or $path.selected_ip -ne $ip -or $path.route.outcome -ne 'completed') { throw 'Selected IP, peer and path disagree.' }
    foreach ($phase in @('tcp', 'tls', 'http')) {
        if ($exchange.$phase.outcome -ne 'completed') { throw "Incomplete HTTPS phase: $phase" }
    }
    $rows = @(Find-NetRoute -RemoteIPAddress $ip | Select-Object InterfaceIndex, IPAddress, NextHop)
    $rows | ConvertTo-Json | Set-Content (Join-Path $output 'windows-route.json')
    $source = @($rows | Where-Object IPAddress)[0]
    $route = @($rows | Where-Object NextHop)[0]
    $nextHop = if ($route.NextHop -eq '0.0.0.0') { '' } else { $route.NextHop }
    if ($path.route.source_ip -ne $source.IPAddress -or $path.route.interface_index -ne $source.InterfaceIndex -or $path.route.interface_index -ne $route.InterfaceIndex -or $path.route.next_hop -ne $nextHop -or $path.route.on_link -ne ($nextHop -eq '')) { throw 'Candidate route disagrees with Find-NetRoute.' }
    $trace = & "$env:SystemRoot/System32/tracert.exe" -d -4 -h 4 -w 750 $ip
    if ($LASTEXITCODE -ne 0) { throw 'Windows tracert failed.' }
    $trace | Set-Content (Join-Path $output 'tracert.txt')
    $first = @($trace | Where-Object { $_ -match '^\s*1\s' })[0]
    if ($first -notmatch '(\d+\.\d+\.\d+\.\d+)\s*$') { throw 'Windows first hop unavailable; semantic comparison remains open.' }
    if ($path.trace.hops[0].status -ne 'time_exceeded' -or $path.trace.hops[0].reply_address -ne $Matches[1]) { throw 'First-hop semantics disagree.' }
    $curl = & "$env:SystemRoot/System32/curl.exe" --silent --show-error --max-time 8 --ipv4 --noproxy '*' --resolve "example.com:443:$ip" --output NUL --write-out '%{http_code}|%{remote_ip}|%{ssl_verify_result}|%{time_starttransfer}' 'https://example.com/'
    if ($LASTEXITCODE -ne 0) { throw 'Pinned curl failed.' }
    $curl | Set-Content (Join-Path $output 'curl.txt')
    $parts = $curl.Split('|')
    if ($parts.Count -ne 4 -or $parts[0] -ne [string]$exchange.response_status -or $parts[1] -ne $ip -or $parts[2] -ne '0') { throw 'Pinned curl peer/status/certificate disagrees.' }
    Write-Output "PASS: candidate $ExpectedSHA256; selected $ip; Find-NetRoute source/interface/next-hop, first-hop tracert semantics, pinned HTTPS curl status $($parts[0])/peer/certificate. Timings and later hops need not match."
} finally {
    if ($process) {
        if (-not $process.HasExited) { $process.Kill() }
        $process.WaitForExit()
        $process.Dispose()
    }
    # Remove only this invocation's known file and empty directory.
    if (Test-Path -LiteralPath $copy) { Remove-Item -LiteralPath $copy }
    Remove-Item -LiteralPath $directory
}
