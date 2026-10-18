# Milestone 13: remote IPv4 route and tracert comparison

On 2026-10-15, an opt-in live test compared the production Windows IPv4 adapter with system `tracert.exe -d -4` for the same public IPv4 destination, `1.1.1.1`, using a four-hop limit. Both reported the same first responding hop at TTL 1 on the current unelevated Windows 11 x64 machine and network. HopVista's result was `hop_limit_reached`; it did not infer that the destination was unreachable. The test uses one HopVista echo per TTL and Windows `tracert`'s three probes per TTL, so it does not require identical later hops or RTTs. The test keeps private intermediate addresses out of this note.

The opt-in test also queries independent Windows `Find-NetRoute` information for the same destination. The production `GetBestRoute2` result agreed with Windows on the selected source IPv4 address, interface index, next hop, and whether the route was on-link. This is one current route condition; route choices can change with VPNs, interfaces, and network policy. Private source and gateway addresses are not copied into this note.

Run from the repository root with loopback and remote ICMP access permitted:

```powershell
$env:GOCACHE = Join-Path (Get-Location) '.tools/go-build'
$env:GOTOOLCHAIN = 'local'
$env:HOPVISTA_REMOTE_TRACE_IP = '1.1.1.1'
.\.tools\go\bin\go.exe test ./internal/pathmeasure -run TestLiveRemoteTracertComparison -count=1 -v
```

This adds one remote path and route comparison to the earlier loopback executable check. It does not validate later-hop identity or exact latency, alternate networks or VPNs, another Windows account/build, a genuinely stalled `GetBestRoute2`, or the empirical 250 ms margin across wider conditions. Release acceptance remains open, including the remaining functional and human UI checks.
