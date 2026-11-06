# Issue #17: experimental candidate preparation

## Updated candidate — 2026-11-06

Prepared for review, **not accepted or published**. Issue #16 is closed with all
seven criteria checked after the user's explicit approval; its
[audit](issue-16-acceptance-audit.md) and exact assisted participant answers
retain their original scope. Issue #17 and release acceptance remain open for
the user's explicit decision on this candidate. PR #27 remains open and draft
against `main`; no merge, tag, upload or release publication is authorized.

### Identity and tested condition

| Item | Recorded value |
| --- | --- |
| Clean source commit | `28dd288fe84e0d4a3763d18be69382e784fa7c1b` on `accessibility-human-acceptance` |
| Go module build version | `v0.0.0-20261106215437-28dd288fe84e` |
| Local executable | `.tools/issue-17/rc-28dd288/hopvista-28dd288-windows-amd64.exe` |
| SHA-256 | `5A5606FA711AAD0D420CDAFEB903DC9C4AE480681691F784B8B12B2BD5C5847C` |
| Build | Go `go1.27.2`, `windows/amd64`, `GOAMD64=v1`, `CGO_ENABLED=0`, `vcs.modified=false` |
| Host | Windows 11 Pro 25H2, build **26200.9711**, x64; medium-integrity unelevated token, Administrators membership deny-only |
| Browser/tooling | Isolated headless Edge **156.0.4314.58**, Playwright **1.62.1**, Node **v20.16.0** |

`git fetch origin` succeeded outside the restricted shell sandbox. Local HEAD
and origin's branch both matched the full revision above; origin/main was
`b67eaee6e27a73085c864d7c090a122c85ab96ad`. GitHub independently confirmed #16
closed, #17 open and PR #27 draft/unmerged at that tip. `gh` was unavailable.
The production UI typecheck/rebuild left the tracked tree clean before building.
This executable includes the participant-driven UI improvements; the historical
`a159f9c` candidate below does not. Subsequent evidence documentation does not
change this executable's source identity. No production code, assets, schema v1
or Go-owned findings changed during this preparation.

Each packaged harness launches a copied executable from a disposable directory
with only Windows System32 on its PATH. No Go/Node runtime is needed by the app.
Sandbox permission escalation allowed sockets, native commands and headless
browser launch; the Windows user token remained unelevated. No desktop, existing
browser, system clipboard or screen reader was controlled.

### Exact-candidate checks

| Check | Outcome and evidence scope |
| --- | --- |
| Full UI suite with `HOPVISTA_PACKAGED_EXE` | **25 passed, 0 failed, 1 opt-in system-clipboard skip**. All three real packaged-browser journeys passed: redirects/503, advanced hops, native path and exact evidence/return, rejected input/retry, retained incomplete HTTP, cancellation, another run and usable UI after actual process restart. Other cases use intercepted API presentation fixtures. |
| Packaged loopback/API/security | **Passed** embedded UI/assets, exact Host/Origin/token/no-CORS checks, explicit IPv6 rejection, loopback route/echo/trace, same-IP tracert and pinned HTTP curl. `HOPVISTA_PACKAGED_DRAIN_IP=192.0.2.1` and `HOPVISTA_PACKAGED_HTTPS_URL=https://example.com/` explicitly enabled native stopping/busy/retained evidence and direct pinned HTTPS comparisons. |
| Packaged restart/privacy | **Passed**, three actual launches: completed/incomplete results cleared on restart, token rotation and old-token rejection, invalid inputs, empty stdout, listener-only stderr and unchanged isolated directory. |
| Packaged real-minute controlled cutoff | **Passed** `TestPackagedSuspendedMinuteCutoff`, unchanged 60-second production window, disposable process paused with native echo in flight, stable terminal snapshots and retained evidence, no new native probes after cutoff. This is induced scheduling delay, not a natural native stall. |
| Same-selected-IP public network helper | **Passed**, checksum guarded. Candidate selected/connected/path IP `104.20.23.154` for `https://example.com/`; Windows `Find-NetRoute` matched source `10.8.26.136`, interface **9**, next hop `10.8.26.1`, off-link. First-hop `time_exceeded` reply matched four-hop `tracert -d -4 -h 4 -w 750`. Proxy-disabled curl pinned to the same IP returned **200**, peer matched, certificate verification **0**. |
| Candidate UI replay/layout | **65 passed, 0 failed**, 13 modes at 1440/960/760/390/320 CSS pixels, including long URL/ID/64-hop stress. Evidence focus/return, exact Go finding preservation, narrow reading order and no page-wide overflow passed. Reduced-motion emulation matched with zero animations and no selected-evidence/arrow motion. These are captured/synthetic API replays, not fresh native runs or human review. |
| Candidate UI measurements | **20 sampled color pairs, 0 failures**. Synthetic clipboard rejection selected the literal manual-copy address at 320px without overflow. Existing approximately 144×37px wordmark remains the disclosed exception to the enhanced 44px coarse-pointer target. No system clipboard or native permission-denial claim. |

Native cancellation: stop `2026-11-06T22:16:12.9307055Z`, last send
`22:16:12.871864Z`, returned-and-closed `22:16:13.8628974Z`, quiescence
`22:16:13.8634032Z`. Controlled cutoff: deadline `22:17:59.8140135Z`,
stop observed `22:18:00.3145195Z`, quiescence `22:18:00.316384Z`, elapsed
**60.5023705 seconds**. The approximately 500ms late observation follows the
harness's deliberate pause; the API cannot answer while the process is paused.
It establishes neither ordinary scheduling lag nor a native completion bound.

The separate public run measured TCP/TLS/HTTP **6.7302/38.8395/14.6856ms**;
pinned curl start-transfer was **41.262ms** on a different fresh request.
The trace reached its four-hop limit while HTTPS succeeded. Identical timings,
later hop identities, browser-path equivalence and a causal diagnosis are not
required or inferred.

Source inspection of `cmd/hopvista`, `internal/localapi`, `internal/run` and
`ui/src` retains loopback-only serving, per-launch mutation tokens, in-memory
results, same-origin API fetches and no application database, browser result/token
storage, default telemetry or external enrichment. Output/directory tests support
that inspection; they are not OS-wide file/packet tracing or secure erasure.
An existing browser tab may retain rendered React state until reload/close.

### Source regression and skips

Full `go test ./... -p 1 -count=1 -v -timeout 180s` **passed**, with live
adapter, native pending cancellation and separate raw-return/handle-close
ordering, unloaded/loaded cap-plus-margin sampling, loopback HTTP/TLS, local
API/engine, real-minute manager-under-load and the exact packaged executable
explicitly enabled. Packages were serialized. `go vet ./...`, final TypeScript
checking and `git diff --check` **passed**. The production UI rebuild produced
no tracked asset changes, and the executable's checksum was unchanged after
verification.

Full Go suite: **three explicit skips** — standalone remote-tracert comparison
(the exact-candidate same-IP helper passed separately), natural packaged-minute
fixture (known early-finishing fixture, not a pass) and suspended packaged-minute
fixture (passed separately above). The focused packaged run explicitly enabled
public HTTPS/drain branches; these were not repeated in the full suite. UI:
**one explicit skip**, system clipboard, as requested. No race-detector result
is claimed with cgo disabled. The separate historical feasibility-harness module
was not rerun. No failed new check was suppressed or counted as passed.

### Reproduction and local evidence

Use a clean checkout of the recorded source commit on Windows x64 with Go
1.27.2, Node and pinned UI development dependencies on the build/test host.
The executable's user needs none of those development tools. A different build
environment can produce a different checksum; record it separately.

```powershell
git switch --detach 28dd288fe84e0d4a3763d18be69382e784fa7c1b
npm.cmd --prefix ui ci
npm.cmd --prefix ui run build
git status --porcelain # Must be empty before building.
$out = '.tools/issue-17/rc-28dd288'
New-Item -ItemType Directory -Force $out | Out-Null
$env:GOCACHE = Join-Path $PWD '.tools/go-build'
$env:GOTOOLCHAIN = 'local'
$env:CGO_ENABLED = '0'
$env:GOOS = 'windows'
$env:GOARCH = 'amd64'
# Substitute your Go 1.27.2 executable if needed.
& ./.tools/go/bin/go.exe build -o "$out/hopvista-28dd288-windows-amd64.exe" ./cmd/hopvista
& ./.tools/go/bin/go.exe version -m "$out/hopvista-28dd288-windows-amd64.exe"
Get-FileHash "$out/hopvista-28dd288-windows-amd64.exe" -Algorithm SHA256
$env:HOPVISTA_PACKAGED_EXE = (Resolve-Path "$out/hopvista-28dd288-windows-amd64.exe").Path
npm.cmd --prefix ui test
$env:HOPVISTA_PACKAGED_DRAIN_IP = '192.0.2.1'
$env:HOPVISTA_PACKAGED_HTTPS_URL = 'https://example.com/'
& ./.tools/go/bin/go.exe test ./tools/app-smoke -run 'TestPackagedExecutableLoopback|TestPackagedRestartPrivacy' -count=1 -v -timeout 180s
./tools/candidate-network-check.ps1 -Executable $env:HOPVISTA_PACKAGED_EXE -ExpectedSHA256 5A5606FA711AAD0D420CDAFEB903DC9C4AE480681691F784B8B12B2BD5C5847C -OutputDirectory $out
$env:HOPVISTA_PACKAGED_CUTOFF_IP = '192.0.2.1'
$env:HOPVISTA_CUTOFF_CAPTURE = Join-Path $PWD $out
& ./.tools/go/bin/go.exe test ./tools/app-smoke -run '^TestPackagedSuspendedMinuteCutoff$' -count=1 -v -timeout 100s
Remove-Item Env:HOPVISTA_PACKAGED_CUTOFF_IP
Remove-Item Env:HOPVISTA_PACKAGED_DRAIN_IP
Remove-Item Env:HOPVISTA_PACKAGED_HTTPS_URL
$env:HOPVISTA_LIVE_ADAPTER = '1'
$env:HOPVISTA_LIVE_DRAIN_IP = '192.0.2.1'
$env:HOPVISTA_LIVE_HTTP = '1'
$env:HOPVISTA_LIVE_LOOPBACK = '1'
$env:HOPVISTA_LIVE_ENGINE = '1'
$env:HOPVISTA_LIVE_MINUTE_CUTOFF = '1'
& ./.tools/go/bin/go.exe test ./... -p 1 -count=1 -v -timeout 180s
& ./.tools/go/bin/go.exe vet ./...
node ui/node_modules/typescript/bin/tsc --noEmit -p ui/tsconfig.json
```

The detach command is for a separate reproduction checkout, not a request to
move the working branch. Review artifacts remain local and ignored under
`.tools/issue-17/rc-28dd288/`: `build-info.log`, `sha256.txt`, `ui-test.log`,
`packaged.log`, `cutoff.log`, `network.log`, `public-run.json`,
`windows-route.json`, `tracert.txt`, `curl.txt`, `replay.log`, `metrics.log`,
`go-test.log`, `go-vet.log` and cutoff JSON snapshots. Nothing was uploaded.
Replay report:
`.impeccable/review/issue-16/observations-2026-11-06T22-17-25.144Z.json`.
Metrics report:
`.impeccable/review/issue-16/metrics-2026-11-06T22-17-39.174Z.json`.
Replay uses the existing fixture-generation/observe/metrics commands described
below, adding `stress` to the twelve selected modes and pointing
`HOPVISTA_REVIEW_URL` at a temporary candidate copy with System32-only PATH.
The prior captures must be locally available; these are not bundled in a fresh
clone. The current UI suite contains durable synthetic presentation checks.

### Acceptance limits and decision

Distribution remains **experimental Windows 11 x64 tested on Windows 11 25H2
build 26200.9711 with an unelevated token**. Another account/build/network/VPN,
Windows 10, ARM64, IPv6 and comprehensive Windows 11 support are unverified.
The 250ms start margin is empirical; cutoff stops scheduling, and started native
work may finish later. No hard drain or latest-return bound is claimed.

The known natural cutoff fixture previously finished early and remains failed
historical evidence; it is not rerun or counted as passed here. Natural native
work spanning cutoff and a genuinely stalled `GetBestRoute2` remain unverified
limits of the experimental scope. The controlled-pause pass stays separate.
Existing human reading, actual NVDA and native zoom evidence is reused only at
its recorded scope; no new participant, screen-reader or native zoom pass is
claimed or required by this preparation. Personal clipboard/fallback and live
cleanup observation are not extra human acceptance gates.

The requested decision is whether to accept this exact revision/checksum for an
experimental release at these tested conditions and limits. Until that explicit
decision, #17 stays open, PR #27 stays draft, and merge/tag/publication/assets
remain unapproved. Candidate preparation alone grants none of those actions.

### Updated-candidate review

Independent read-only review of `git diff 28dd288...bf9d5dd` (the candidate
record, README and acceptance-matrix update):

#### Standards

**0 findings.** The three documentation changes follow the repository's
issue-source and domain-language rules, preserve approved product/design
boundaries, and distinguish the updated executable from historical evidence.
Build identity, checksum, packaged/UI outcomes and replay counts match the
inspected artifacts. No applicable baseline code-smell findings; documentation
only. The reviewer edited no files and reran no tests.

#### Spec

**0 actionable findings.** The reviewer independently checked the candidate
checksum and clean Windows amd64 build metadata. Packaged, cutoff, network,
UI and replay logs support the recorded outcomes; skips and experimental limits
remain distinct from passes. Instructions/matrix are reconciled without changing
assets, reopening #16 or implying release authority. Explicit human release
approval remains outstanding. No tests or desktop/browser operations were run
by the reviewer.

Review summary: Standards **0 findings**; Spec **0 findings**. This review
appendix is documentation only and leaves the identified executable unchanged.

## Historical candidate — 2026-10-24

Everything below retains the earlier candidate's dated evidence and acceptance
disposition. Its pending #16 statements are historical; the current candidate
and disposition are above.

Pre-recheck audit, 2026-11-06: human review has resumed and P1's assisted
participation counts. The [#16 acceptance audit](issue-16-acceptance-audit.md)
identifies only the remaining normal phase-meaning/uncertainty recheck, without
requiring a new reader or P1 keyboard/fallback/live-cleanup tasks. #17 remains
open for completed blockers, an updated identifiable candidate and its required
checks/reconciliation, and explicit human release approval. The candidate below
is historical and does not contain later participant-driven UI changes.
References below to deferred review describe the preparation date.

Prepared on 2026-10-24 (America/Los_Angeles) for
[issue #17](https://github.com/DarrylN2/hopvista/issues/17), after
[PR #25](https://github.com/DarrylN2/hopvista/pull/25) merged. This is a
**pre-acceptance candidate**, not an accepted or published release. The user
explicitly deferred human comprehension. Issue #16, issue #17 and overall
release acceptance remain open. No desktop, existing browser, clipboard or
screen-reader operation was performed during this pass.

## Candidate identity

| Item | Recorded value |
| --- | --- |
| Source | `a159f9c083f3adf00d4eb41e4a6f49f83c926d66`, PR #25 merge |
| Local executable | `.tools/issue-17/hopvista-a159f9c-windows-amd64.exe` |
| SHA-256 | `6D7A7EEC1280354F6DCC8905E6A38F8C141D973DA731562D5F0C14344ABE4B08` |
| Go build version | `v0.0.0-20261025033858-a159f9c083f3`; Go `go1.27.2`, `windows/amd64`, `GOAMD64=v1`, `CGO_ENABLED=0` |
| VCS build information | `vcs.revision=a159f9c083f3adf00d4eb41e4a6f49f83c926d66`, `vcs.modified=false` |
| Tested host | Windows 11 Pro 25H2, build 26200.9711, AMD64; current token unelevated |
| Browser/tooling | Headless Edge 155.0.4283.62, Playwright 1.62.1, Node v20.16.0 |

The clean current branch was fast-forwarded to the merged commit. UI typecheck
and production rebuild produced no tracked asset changes. The executable was
built **before** adding this verification helper and documentation; later
documentation/helper commits do not change the candidate's source identity.
The older #24 binary/checksum is not this candidate. No production Go, UI,
embedded assets, schema version 1 or Go-owned observed findings changed here.
The executable and raw logs/captures are local ignored artifacts, not uploaded
release assets. The committed record retains identity, outcomes and commands.

## Checks against the exact executable

| Check | Outcome and scope |
| --- | --- |
| `npm.cmd --prefix ui test` with `HOPVISTA_PACKAGED_EXE` | **17 passed, 1 explicit skip**. Includes all three real packaged-browser journeys: submission/advanced hops, 302/503, exact evidence, IPv6 rejection/retry, retained partial evidence, cancellation, another run and usable fresh UI after actual process restart. The other browser cases intercept API responses and are presentation checks. System clipboard opt-in remains disabled. |
| `TestPackagedExecutableLoopback` | **Passed**, including embedded assets, Host/Origin/token/no-CORS boundary, loopback native route/echo/trace, tracert and pinned HTTP curl. `HOPVISTA_PACKAGED_DRAIN_IP=192.0.2.1` enabled native stopping/busy/evidence retention; `HOPVISTA_PACKAGED_HTTPS_URL=https://example.com/` enabled direct pinned HTTPS/curl comparison. |
| `TestPackagedRestartPrivacy` | **Passed**, all three actual launches, old result absence before another run, token rotation/old-token rejection, completed/incomplete runs, invalid input, empty stdout, listener-only stderr, unchanged isolated directory. |
| `TestPackagedSuspendedMinuteCutoff` | **Passed**. Real unchanged 60-second window, externally paused disposable process with native echo in flight, retained evidence, stable terminal snapshots and no new probes after cutoff. This is an induced scheduling delay, not a natural native stall. |
| `tools/candidate-network-check.ps1` | **Passed** against the checksum above. A separate real candidate run selected `104.20.23.154` for `https://example.com/`. Peer and path used that IP. Windows `Find-NetRoute` matched source, interface 8, next hop and off-link status. First-hop `time_exceeded` address matched four-hop `tracert -d -4 -h 4 -w 750`. Independent proxy-disabled, pinned curl returned `200`, peer `104.20.23.154`, certificate verification `0`. Later hop identities/RTTs need not match. |
| Representative captured/synthetic edge-state replay in candidate UI | **60 scenario/width checks passed**, twelve modes at 1440/960/760/390/320 CSS pixels, evidence focus/return, preserved findings, reading order and no page-wide overflow. Runtime reduced motion matched with zero animations and no evidence/arrow motion. These are intercepted API replays, not new native measurements. |
| Browser metrics in candidate UI | **20 sampled color pairs, 0 failures**; selected manual-copy fallback at 320 CSS pixels with no overflow. The existing 144×37px wordmark remains the disclosed exception to the enhanced 44px coarse-pointer target. |

The native cancellation observation: scheduling stop `2026-10-25T03:44:31.5067082Z`,
last native send `03:44:31.4495216Z`, returned-and-closed
`03:44:32.449472Z`, terminal quiescence `03:44:32.4501202Z`.
The controlled pause crossed cutoff `03:46:11.4224477Z`; stop was observed at
`03:46:11.9234307Z` and quiescence at `03:46:11.9271322Z` (60.5046845 seconds
after start). The approximately 500ms late observation follows the harness's
external pause. It is not an ordinary scheduling-lag or native latest-return
bound. The API cannot respond while the whole child process is suspended.

In the additional public run, TCP/TLS/HTTP measured 16.5634/54.9208/18.788ms;
curl start-transfer was 74.729ms. These are distinct fresh requests; neither
identical millisecond timings nor browser-path equivalence is claimed. The
four-hop trace reached its configured limit while HTTPS completed successfully.

Disposable packaged children run from fresh temporary directories with only
Windows System32 on PATH, without Go/Node. Headless browsers use temporary
profiles, not personal profiles. The network helper checks the input checksum,
captures the public result before cross-checking the same IP, stops and waits
for its child, then removes only its known copy and empty temporary directory.
The original helper acceptance run passed; no production defect or product
red/green fix is claimed.

PowerShell parser validation passed. A negative checksum invocation was rejected
before creating a temporary directory or launching a process. The candidate's
SHA-256 was rechecked unchanged after verification.

## Source and regression checks

`go test ./... -p 1 -count=1 -v -timeout 180s` **passed** with live adapter,
native pending cancellation and separate raw-return/handle-close ordering,
unloaded/loaded cap-plus-margin sampling, loopback HTTP/TLS, local API/engine,
real-minute manager-under-load and packaged checks explicitly enabled. Package
serialization avoids concurrent fixture loads. `go vet ./...`, final TypeScript
typechecking and `git diff --check` also **passed**.

The full suite had **three explicit skips**: standalone remote-tracert
comparison (superseded here by the exact-candidate same-IP helper), natural
packaged-minute fixture, and suspended packaged-minute fixture. The suspended
check passed separately as recorded above; the natural fixture is not counted
as passed. Clipboard access stayed disabled. No race-detector result is claimed
with `CGO_ENABLED=0`. The separate historical feasibility-harness module was
not rerun. One sandboxed network-condition lookup lacked CIM access; the
permitted condition query and Windows route comparison succeeded. Sandbox
permission escalation did not elevate the Windows user token.

Source review of the current entry point, in-memory manager/local API and `ui/src` retains the
earlier #15 conclusion: loopback-only listener, per-launch token, no application
run database/files, browser result/token storage, default telemetry or external
enrichment calls. UI fetches only the same-origin local API. Runtime output and
directory observations complement this source inspection; they are not OS-wide
packet/file tracing or secure-erasure proof. A browser tab already rendering a
result can retain its React state until reload/close after process exit.

## Reproduction

Use a clean checkout of the source commit above on Windows x64. Go 1.27.2 and
Node/the pinned UI development dependencies are needed on the **test host**.
The recorded host used `.tools/go/bin/go.exe`; substitute your Go executable.
Do not rebuild the candidate after modifying sources and still label it with
the checksum above. A different toolchain/build environment can yield a
different checksum; record its identity separately.

```powershell
npm.cmd --prefix ui ci
npm.cmd --prefix ui run build
# Confirm no tracked source/assets changed before building.
git status --porcelain
New-Item -ItemType Directory -Force .tools/issue-17 | Out-Null
$env:GOCACHE = Join-Path $PWD '.tools/go-build'
$env:GOTOOLCHAIN = 'local'
$env:CGO_ENABLED = '0'
& ./.tools/go/bin/go.exe build -o .tools/issue-17/hopvista-a159f9c-windows-amd64.exe ./cmd/hopvista
& ./.tools/go/bin/go.exe version -m .tools/issue-17/hopvista-a159f9c-windows-amd64.exe
Get-FileHash .tools/issue-17/hopvista-a159f9c-windows-amd64.exe -Algorithm SHA256
$env:HOPVISTA_PACKAGED_EXE = (Resolve-Path .tools/issue-17/hopvista-a159f9c-windows-amd64.exe).Path
npm.cmd --prefix ui test
$env:HOPVISTA_PACKAGED_DRAIN_IP = '192.0.2.1'
$env:HOPVISTA_PACKAGED_HTTPS_URL = 'https://example.com/'
& ./.tools/go/bin/go.exe test ./tools/app-smoke -run 'TestPackagedExecutableLoopback|TestPackagedRestartPrivacy' -count=1 -v -timeout 180s
$env:HOPVISTA_PACKAGED_CUTOFF_IP = '192.0.2.1'
$env:HOPVISTA_CUTOFF_CAPTURE = Join-Path $PWD '.tools/issue-17'
& ./.tools/go/bin/go.exe test ./tools/app-smoke -run '^TestPackagedSuspendedMinuteCutoff$' -count=1 -v -timeout 100s
Remove-Item Env:HOPVISTA_PACKAGED_CUTOFF_IP
# Enable the other applicable live fixtures for the full suite.
$env:HOPVISTA_LIVE_ADAPTER = '1'
$env:HOPVISTA_LIVE_DRAIN_IP = '192.0.2.1'
$env:HOPVISTA_LIVE_HTTP = '1'
$env:HOPVISTA_LIVE_LOOPBACK = '1'
$env:HOPVISTA_LIVE_ENGINE = '1'
$env:HOPVISTA_LIVE_MINUTE_CUTOFF = '1'
# The recorded full-suite invocation did not repeat the optional public/drain
# branches already exercised in the focused packaged run.
Remove-Item Env:HOPVISTA_PACKAGED_DRAIN_IP
Remove-Item Env:HOPVISTA_PACKAGED_HTTPS_URL
& ./.tools/go/bin/go.exe test ./... -p 1 -count=1 -v -timeout 180s
& ./.tools/go/bin/go.exe vet ./...
node ui/node_modules/typescript/bin/tsc --noEmit -p ui/tsconfig.json
```

The added helper is in the later verification commit, not in the clean binary
source commit. From a checkout containing the helper, point it at the already
built candidate (PowerShell 7, network/native access and Windows commands needed):

```powershell
./tools/candidate-network-check.ps1 -Executable .tools/issue-17/hopvista-a159f9c-windows-amd64.exe -ExpectedSHA256 6D7A7EEC1280354F6DCC8905E6A38F8C141D973DA731562D5F0C14344ABE4B08 -OutputDirectory .tools/issue-17
```

For presentation replay, use `node tools/accessibility-fixtures.mjs`, then run
an isolated copy of this candidate and set `HOPVISTA_REVIEW_URL` to its listener.
Set `HOPVISTA_REVIEW_MODES` to
`i11completed,i11sensitive,i11certificate,i12tcptimeout,i12measuredzero,i13highsilent,i13limit1,i13limit64,i13unreachable,i14pausedminute,i14cutoffstopping,i14cutoff`.
Run `node tools/accessibility-observe.mjs` and
`node tools/accessibility-metrics.mjs`, then stop/wait for the disposable child.
The tools intercept API responses using existing labeled captures; the helper
scripts require the locally retained prior captures. Those artifacts are not
bundled in a fresh clone; regenerate the originating #11–#14 fixtures using
their linked records before attempting replay.

Local logs: `.tools/issue-17/{build-info.log,sha256.txt,ui-test.log,packaged.log,cutoff.log,network.log,public-run.json,windows-route.json,tracert.txt,curl.txt,replay.log,go-test.log,go-vet.log}`.
Replay report: `.impeccable/review/issue-16/observations-2026-10-25T03-48-00.907Z.json`.

## Acceptance disposition

- **#24 functional scope covered:** merged implementation/review through
  `6238991`, with current-candidate journeys passing after the Advanced fix.
  Closed on GitHub on 2026-10-24 with the evidence disposition recorded there.
- **#15 stated scope covered:** restart/token/output/directory/security evidence
  plus #24's usable fresh browser session and current source inspection. No
  broader persistence/telemetry claim is inferred from absence in the launch
  directory alone. See [restart/privacy evidence](packaged-restart-privacy.md).
  Closed on GitHub on 2026-10-24 with the evidence disposition recorded there.
- **#16 remains open:** at this candidate's preparation date, human review was
  deferred; subsequent assisted P1 review and the remaining reading gap are
  mapped in the audit above. No human reading or native 200% zoom recheck was
  performed in this candidate pass; earlier checks retain their dated scope.
- **#17 remains open:** the candidate is available for review, but its criterion
  requiring completed blockers and an explicit human release decision is not
  satisfied while #16 is open. No release tag, signing, installer, upload or
  publication is performed.

Distribution remains **experimental Windows 11 x64 tested on the named Windows
11 25H2 build 26200 unelevated condition**. No additional account, Windows build,
VPN/network, Windows 10, ARM64, IPv6 or comprehensive Windows 11 support is
claimed. Natural native work spanning cutoff and a genuine `GetBestRoute2`
stall remain unverified. The known earlier natural 38.35-second cutoff fixture
remains failed evidence and was not rerun; the passing controlled pause is
separately labeled. No hard native drain/completion bound is introduced.

## Review

### Standards

Independent read-only review of `git diff a159f9c...f0c235b`: **0 documented
standard violations and 0 actionable heuristic findings**. The reviewer checked
AGENTS/agent docs, glossary, product/design boundaries and the retained binary
identity/network logs. No tests were independently rerun; no files were edited.

### Spec

Independent read-only review of the same diff against issue #17 and the user's
authorized pre-acceptance scope: **0 actionable findings**. The reviewer
independently checked the local executable checksum and inspected packaged,
cutoff, public-network, UI/replay and full-suite logs. Candidate identity,
same-selected-IP checks, skips and acceptance limits matched the record.
No tests or desktop/browser operations were independently performed.

Review summary: Standards **0 findings**; Spec **0 findings**. This final
documentation-only update records the review dispositions and issue closure;
production sources and the identified executable remain unchanged.
