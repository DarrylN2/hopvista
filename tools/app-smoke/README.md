# Packaged executable smoke check

This opt-in Windows x64 check starts the built HopVista executable from an otherwise empty temporary directory. The child process receives a `PATH` containing only Windows System32, so it cannot find a Go or Node executable through `PATH`. The test verifies embedded UI assets, the local session endpoint, Host, Origin, and token rejection, explicit IPv6-literal rejection, and a full loopback HTTP 503 diagnostic run with native IPv4 route, echo, and trace evidence. It compares the loopback destination reply with system `tracert.exe -d -4` for the same selected `127.0.0.1` address. It then runs a `localhost` target and compares the application's selected and connected peer and HTTP 200 response with an independent system `curl.exe --resolve` request pinned to the same address and port, with proxies disabled. The check requires `tracert.exe` and `curl.exe` in Windows System32. It terminates the child and removes the temporary copy afterward.

From the repository root in PowerShell, where loopback sockets are permitted:

```powershell
$env:GOCACHE = Join-Path (Get-Location) '.tools/go-build'
$env:GOTOOLCHAIN = 'local'
.\.tools\go\bin\go.exe build -o .tools\hopvista.exe ./cmd/hopvista
$env:HOPVISTA_PACKAGED_EXE = (Resolve-Path -LiteralPath '.tools\hopvista.exe').Path
.\.tools\go\bin\go.exe test ./tools/app-smoke -run TestPackagedExecutableLoopback -count=1 -v
```

The check skips in ordinary `go test ./...` unless `HOPVISTA_PACKAGED_EXE` is set. It verifies this binary, pinned HTTP comparison, and one-hop `tracert` semantics on loopback on the current machine. It does not compare a remote route or its variable hop timings, or check signing, an installer, other Windows accounts or builds, VPNs, a genuinely stalled native route lookup, or the complete release acceptance matrix.

For the opt-in packaged cancellation check, set `HOPVISTA_PACKAGED_DRAIN_IP` to a candidate no-reply IPv4 address before the `go test` command above (for example, `$env:HOPVISTA_PACKAGED_DRAIN_IP = '192.0.2.1'` on the reviewed local network). After the loopback checks, this starts a run to that IP, cancels while a target echo is in flight, verifies `stopping` and a busy second start, and checks retained echo evidence after the native call returns. A network where the candidate replies promptly cannot satisfy this fixture and must use another controlled no-reply target. This is a finite observation, not a native latest-return bound.

For a direct public HTTPS comparison, set `HOPVISTA_PACKAGED_HTTPS_URL` to an HTTPS target that returns a final response without redirecting (for example, `$env:HOPVISTA_PACKAGED_HTTPS_URL = 'https://example.com/'`) before the same `go test` command. The packaged app records the selected and actual peer IPv4 address, completed TCP/TLS/HTTP phases, response status, and a fresh connection. Independent system `curl.exe` then uses `--resolve` for that exact host, port, and selected IP with proxies disabled. The check compares peer, response status, certificate verification, and first-response timing within a broad two-second range. It requires direct public HTTPS access and is one live network observation.
