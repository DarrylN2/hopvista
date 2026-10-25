# Issue #17: experimental candidate preparation

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
- **#16 remains open:** partial participant feedback and facilitator help remain
  recorded. Human comprehension is deferred; automation is not a substitute.
  No fresh real-reader check of changed timing wording or native 200% zoom was
  performed here; earlier checks retain their dated scope.
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
