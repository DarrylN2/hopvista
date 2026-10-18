# Packaged process restart and privacy

Checked on 2026-10-18 (America/Los_Angeles), from merged PR #19 at
`a5fd594c14e01b92405ca614ba4e7017a8536722`, on branch
`codex/packaged-restart-privacy`. The starting workspace was clean.

This addresses the explicit process-restart/default-output gap in the
[v1 acceptance matrix](v1-acceptance-status.md), under
[issue #10's local application and privacy requirements](https://github.com/DarrylN2/hopvista/issues/10).
It extends the existing packaged smoke checks rather than duplicating their
embedded-assets, local API security, route/tracert and pinned-curl coverage.
No production Go, schema, UI source or embedded asset changes were needed.
**This is current-condition evidence; overall release acceptance remains open.**

## Tested conditions

- Windows NT `10.0.26200.0`, x64; current token not elevated.
- Go `go1.27.2 windows/amd64`; built directly from the source above.
- Executable SHA-256:
  `BE97768021B195F2FFE6C545C043BECBB2B81DE94848DD63CE90B329D479CC67`.
- Each fixture copies the single executable to a fresh temporary directory.
  All three privacy-check launches use that same copy/directory. Child `PATH`
  contains only Windows System32; no Go/Node runtime is needed by the child.
- Actual loopback HTTP sockets, real API/manager/engine, and the production
  Windows IPv4 adapter are used. The server deliberately redirects, returns
  HTTP 503, or closes its connection before response headers. Its path/query
  values are synthetic canaries, not personal data. No public target is used.
- Socket/process checks ran outside the restricted command sandbox. Test
  cleanup stops and waits for each child before removing its temporary copy.

## Reproduce

From the repository root in PowerShell:

```powershell
$env:GOCACHE = Join-Path $PWD '.tools/go-build'
& ./.tools/go/bin/go.exe build -o .tools/hopvista-restart-privacy.exe ./cmd/hopvista
$env:HOPVISTA_PACKAGED_EXE = (Resolve-Path '.tools/hopvista-restart-privacy.exe').Path
& ./.tools/go/bin/go.exe test ./tools/app-smoke -run 'TestPackagedRestartPrivacy|TestPackagedExecutableLoopback' -count=1 -v
```

The tests skip in an ordinary full-suite run unless the opt-in executable path
is set. Use a newly built executable; they do not build it implicitly. The
privacy test prints aggregate outcomes, without printing captured process
streams, submitted markers, tokens or result bodies. The normal smoke test is
still independently runnable. Full-suite/typechecking checks are separate:

```powershell
Remove-Item Env:HOPVISTA_PACKAGED_EXE
& ./.tools/go/bin/go.exe test ./... -count=1
& ./.tools/go/bin/go.exe vet ./...
node ui/node_modules/typescript/bin/tsc --noEmit -p ui/tsconfig.json
```

## Observed outcomes

`TestPackagedExecutableLoopback` passed, retaining its existing checks for the
embedded UI, Host/Origin/token boundary, unsupported IPv6, real native loopback
route/echo/trace, independent tracert and pinned HTTP comparison.

`TestPackagedRestartPrivacy` passed all three launches:

| Check | Outcome |
| --- | --- |
| Retained completed result before exit | The first process returned a completed two-exchange redirect/HTTP 503 diagnostic run, with the submitted synthetic URL/query and a Go finding present in the API result. |
| Rejected requests | Unknown-path, URL-with-credentials and malformed-JSON requests returned their expected 404/400 codes without reflecting the marker. |
| Actual process restart | The first child was killed and waited for. The same executable was relaunched from the same directory; reading/cancelling its old run ID returned 404 before starting any new run. This is distinct from the existing new-diagnostic-run eviction test. |
| Per-launch token | The second launch issued a different 64-character mutation token and rejected the first token with `invalid_token`, using the second listener's correct Origin. A fresh token then successfully started a diagnostic run. The third launch changed the token again. |
| Retained incomplete result before exit | The second process returned a coherent terminal run with HTTP exchange error and inspectable evidence after the controlled server closed the connection. |
| Second process restart | A third launch could read neither earlier run ID; both returned 404 before any new diagnostic run was started. |
| Default output | Across each entire child lifetime, stdout had zero bytes and stderr contained exactly one correctly formatted timestamped loopback listener announcement. No submitted URL/query marker, mutation token, run ID, measurements or result JSON was logged. Output was fully drained after process exit; no stderr lines were dropped. |
| Launch-directory persistence | The directory still contained only `hopvista.exe` after each exit. Its SHA-256 was unchanged. |
| API caching | Every exercised API success/error response retained `Cache-Control: no-store`. |

The acceptance cases passed on their first run. No product defect was
demonstrated, so there was no production red/green fix. The existing smoke
launcher was extracted into shared test support to capture complete stdout and
stderr and reuse the same isolated binary across launches.

## Source review and limits

The production entry point creates a new in-memory run manager and local API
handler on each launch. The handler generates a random per-launch token. The
manager stores only the active/latest run in memory; there is no result-file or
database read/write path. The only normal process log is the listener URL.
Entry-point fatal logs concern listener/handler/server startup or serving errors;
request-server logging is discarded and measured errors are returned as API
data, not printed. The UI stores the current result in React state and fetches
only the same-origin local API. Source inspection found no application
telemetry, external enrichment, browser storage or result-persistence calls.
The direct HTTP transport disables proxy use and dials the selected target.

This combines observable output/API/directory checks with source review. It is
not OS-wide file-access or packet capture, proof about paging/crash dumps,
browser history/autofill/caches, secure memory erasure, or a graceful-console-
signal/installer/uninstaller test. Process exits here are forced with
`Process.Kill`; no claim of graceful-signal acceptance is made. Another Windows
build/account/network was not tested. No race-detector result is claimed in
this cgo-disabled environment. The remaining human/accessibility, functional
fixture and native-condition gates in the matrix remain open.

The disappearance check concerns the process/API result store. A browser tab
already displaying a terminal result keeps its rendered React state until it
is reloaded or closed; terminating the process does not remotely clear that tab.

The full root Go suite, `go vet ./...` and TypeScript typechecking passed. The
ordinary full-suite run skipped opt-in network/minute/packaged fixtures; the
two packaged checks were enabled explicitly in the focused run above. The
production executable build passed. No UI source/assets changed, so the
existing production assets were used without rebuilding them.

## Standards

Independent review of implementation commit `459d0e2` against
`a5fd594c14e01b92405ca614ba4e7017a8536722`: no findings. All six changed files
follow the documented standards/glossary. Shared launcher reuse and the
packaged-process/API/output test boundary were appropriate; no meaningful
baseline code-smell findings warranted changes.

## Spec

Independent review of the same diff: no actionable findings or scope creep.
The reviewer confirmed the agreed restart/output checks and the evidence
record's distinction between source inspection, runtime observations and
remaining limits. The reviewer did not independently rerun the opt-in tests.
The final commit adds this review record to the reviewed implementation.

Review summary: Standards 0 findings; Spec 0 findings. Overall release
acceptance remains open.
