# hopvista

The [v1 specification](https://github.com/DarrylN2/hopvista/issues/10) targets an IPv4-only first release on Windows 11 x64. IPv6, Windows 10, and ARM64 support are deferred. The user-approved milestone 1 development gate [passed final verification](docs/research/ipv4-development-gate-final-review.md); application implementation is unblocked, while release acceptance remains open.

The initial 250 ms start allowance is empirical on tested conditions, with no hard native completion promise. Another Windows build/account and a genuinely stalled route lookup remain unverified. Until broader coverage exists, distribution is experimental and described as tested on Windows 11 25H2 build 26200 with an unelevated token.

Milestone 1: [standalone Go Windows networking feasibility harness](tools/windows-network-feasibility/README.md).
See the [observed results and remaining validation gates](docs/research/go-windows-network-feasibility.md).

The current [v1 acceptance status](docs/implementation/v1-acceptance-status.md) maps issue #10 to the implemented checks and remaining release work.

## Build the local app

The approved React/TypeScript UI source is in `ui/`. Its Vite output is checked in under `internal/localapi/static/` so a Go build embeds the production UI and the executable needs no Node runtime. On a Windows x64 development machine with Go and Node installed, run from the repository root in PowerShell:

```powershell
npm.cmd --prefix ui ci
npm.cmd --prefix ui run build
New-Item -ItemType Directory -Force .tools | Out-Null
$env:CGO_ENABLED = '0'
go build -o .tools/hopvista.exe ./cmd/hopvista
go version -m .tools/hopvista.exe
Get-FileHash .tools/hopvista.exe -Algorithm SHA256
& ./.tools/hopvista.exe
```

Open the printed `http://127.0.0.1:<port>/` URL in your browser. Keep the executable running while using the UI; Ctrl+C stops it. The packaged executable needs no Go/Node installation. Relaunching creates a fresh listener and mutation token; open the newly printed URL. Run results are held in process memory. An already open browser tab can retain its rendered result until reloaded or closed.

The updated [experimental candidate record](docs/implementation/issue-17.md)
identifies clean source `28dd288fe84e0d4a3763d18be69382e784fa7c1b`, version
`v0.0.0-20261106215437-28dd288fe84e`, reproduction commands and exact-binary
checks. The local candidate is
`.tools/issue-17/rc-28dd288/hopvista-28dd288-windows-amd64.exe`, SHA-256
`5A5606FA711AAD0D420CDAFEB903DC9C4AE480681691F784B8B12B2BD5C5847C`.
It includes the participant-driven UI improvements; the older `a159f9c`
candidate is historical. Issue #16 is closed with human reading and accessibility
criteria covered at their [recorded scopes](docs/implementation/issue-16-acceptance-audit.md).
The user accepted this exact candidate for an experimental Windows 11 x64
release on 2026-11-06; [issue #17's record](docs/implementation/issue-17.md)
retains the decision and limits. PR #27 stays draft. Merge, tagging, asset
upload and publication require separate approval. See
[packaged checks](tools/app-smoke/README.md) to verify another build;
do not assign it this checksum without checking it.

For this version, run the executable on the laptop you want to measure from,
then open its printed loopback URL in a browser on that laptop. The existing
browser UI is embedded in the executable, and measurements originate on that
laptop. The executable can run from an otherwise empty folder without Go or
Node installed. Distribution remains experimental Windows 11 x64, tested on
Windows 11 25H2 build 26200.9711 with an unelevated token.
