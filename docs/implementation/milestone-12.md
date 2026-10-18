# Milestone 12: accepted hop budget and loopback tracert comparison

On 2026-10-15, a read of the current GitHub issue #10 found that its approved hop budget is a default of 30 and an advanced range of 1–64. Milestone 8 had incorrectly described 16 as the issue's maximum and changed the application to match. The local API, path sampler, and minimal input now use the issue's 30/1–64 contract. Tests cover invalid values, the default, the 64-hop API setting, and traces at 1, 30, and 64 hops. The earlier milestone note is a historical implementation record, not the current contract.

The opt-in isolated executable check now compares HopVista's selected `127.0.0.1` IPv4 trace with Windows `tracert.exe -d -4` for the same address. Both showed a destination reply at TTL 1 on this machine. This is semantic loopback agreement; it does not establish agreement for a remote route or compare varying RTTs. The existing separate pinned `curl.exe --resolve` comparison still checks HTTP 200 and the actual connected peer.

The same executable check rejects a wrong mutation token and an IPv6-literal target with their specific API codes before any diagnostic run starts. Its existing Host and Origin checks remain in place.

The full root Go suite, `go vet ./...`, Windows executable build, and opt-in isolated executable check passed on 2026-10-15. The isolated check required running outside the socket-restricted command sandbox. Release acceptance remains open: a genuinely stalled `GetBestRoute2`, another Windows account, broader Windows 11 builds and network conditions, remote `tracert` comparison, full UI comprehension/accessibility review, and the remaining acceptance matrix are unverified. The 250 ms native start margin remains empirical on the named tested conditions.
