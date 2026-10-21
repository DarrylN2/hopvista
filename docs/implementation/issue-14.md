# Issue 14: packaged cancellation and cutoff evidence

First checked on 2026-10-19 (America/Los_Angeles; captures use 2026-10-20 UTC) for
[issue #14](https://github.com/DarrylN2/hopvista/issues/14), with no comments
when read. Branch `codex/packaged-cutoff-acceptance` starts at freshly fetched
`origin/main`, `b1ef63141f655ae0ae8032a0cd37a3252d0244ff` (merged PR #21).
The workspace was clean. The approved UI, React/TypeScript, Vite, plain CSS,
schema version 1 and Go-owned findings are unchanged. **Issue #14 remains
partially verified. Later follow-ups below complete the issue's stated checks
at the named tested-condition scope. **Release acceptance remains open.**

## Condition and provenance

Windows 11 Pro 25H2, build **26200.9711**, x64/amd64, unelevated account; bundled
Go 1.27.2 windows/amd64, CGO disabled. CIM reports Windows 11 Pro; the registry's
legacy ProductName says Windows 10 Pro. Wi-Fi uses Intel Wi-Fi 6E AX211 160MHz,
reported link speed 9 Mbps, default gateway `10.8.26.1`. Cisco AnyConnect's
virtual Ethernet adapter was disabled and Bluetooth disconnected. This records
the observed adapter condition, not a guarantee about upstream routing.

Native tests use `192.0.2.1` as a stated no-reply candidate, real Windows
GetBestRoute2/IcmpSendEcho2 calls, and TCP port 1. The packaged executable is
copied into an isolated temporary directory with only System32 on child PATH;
it uses the production engine, adapter and unchanged **60-second** manager
window. No fixture build, production switch or deadline reduction was added.

The unloaded checks ran alongside ordinary desktop/agent activity. The loaded
adapter check uses GOMAXPROCS=1 and two busy Go workers, requiring both workers
to progress during accepted calls. The separate real-minute manager simulation
adds the same process-local contention during its final five seconds. These
checks overlapped in separate test processes; they do not establish an idle
baseline, system-wide saturation or a worst-case scheduling/native drain bound.

## Results, including the failed candidate

| Check | Result and scope |
| --- | --- |
| New `TestPackagedMinuteCutoff` | **Failed fixture condition.** Run `47d5c972f07c0ba76a31ee517fe99212` started `03:47:43.3179261Z`, configured cutoff `03:48:43.3179261Z`, ended `03:48:21.6712805Z`: **38.3533544 s**, no stop-request timestamp. All 64 hops completed: two time-exceeded replies and 62 timeouts. Trace was `hop_limit_reached`, IPv4 family `timeout`, lifecycle `completed`; this is not cutoff acceptance. |
| Why this candidate was insufficient | Native 750 ms trace requests commonly returned no-reply near 500 ms on this condition. For example TTL 64 ran `03:48:21.1750854Z`–`03:48:21.6678903Z`. Five native target timeouts and the completed gateway replies were retained. Do not pad waits, shorten production time or relabel this result as a minute cutoff. |
| Existing packaged finite echo cancellation, extended | **Pass.** Cutoff `03:51:22.1657411Z`; cancellation/stop `03:50:26.2191473Z`; second target echo started `03:50:26.1638437Z`, returned and closed `03:50:27.1633213Z`; run quiesced `03:50:27.1656921Z`. Exactly two target samples survived, both native timeouts; no third send. `stopping` withheld findings and returned API busy; terminal `cancelled` retained website/path evidence and immutable hop limit 1/window. Repeated terminal GET was identical. |
| Production adapter live cancellation | **Pass.** A pending 1000 ms native echo remained owned after cancellation and returned/closed 579.8 ms after stop. A second echo requested after cancellation was not sent. This is a finite observation, not a latest-return promise. |
| Existing unloaded cap/margin sampling | **Pass.** Three started no-reply calls at each configured cap 750/900/1000 ms, budgets cap + empirical 250 ms margin + 50 ms; all returned/closed before fixture deadlines. Short budgets refused without send timestamps, address or RTT. Largest positive check-to-return overrun above cap was 0/0/0.5411 ms respectively. |
| Existing loaded cap/margin sampling | **Pass.** Three started calls at each cap, zero safe refusals in this run; both workers progressed, all calls returned/closed before deadlines. No late starts were observed. Earlier standalone loaded-boundary failures in milestones 16/17 remain failed evidence; this pass does not supersede them. |
| Existing real-minute manager check | **Pass, simulated executor.** Actual production window 60 s, stop observed **23.0476 ms** after deadline and held executor quiesced **20.7374 ms** after stop. Busy slot, pending findings and stopping timestamps held until release. This is not a packaged/native route-stall observation. |
| New held-route engine/API cases | **Pass, simulated route boundary.** Cancellation and cutoff retain a real completed TLS/HTTP exchange, keep findings withheld and API busy while the route is held, then retain its late completed route without authorizing any echo. Cancellation stays lifecycle cancelled; cutoff stays lifecycle completed with per-measurement deadline. Terminal snapshots are stable. The cutoff case alone uses the existing test-only 500 ms manager constructor. |

The native adapter's `started_at` marks entry into IcmpSendEcho2, while its
`ended_at` is assigned after deferred IcmpCloseHandle. The native return and
handle close are not independently timestamped by schema version 1: the record
therefore calls that field **returned-and-closed**, not a raw return timestamp.
Route end marks GetBestRoute2 return. Run `ended_at` is recorded after the
executor returns and marks quiescence; `stop_requested_at` records observed
scheduling stop separately from the configured `deadline_at`. No schema fields
or instrumentation claiming a separate raw native return instant were invented.

## Coverage and browser verification

Reuse milestones 14–19, the production adapter's existing cap/refusal tests,
issue #11's TLS/capture helpers, issue #12's interrupted website cases and issue
#13's pre-send setup cancellation and unstarted deadline evidence. New cases
extend these helpers at the engine/local API boundary. Their fake route Finish
hook models a synchronous call owning resources until released; it is not an
observed GetBestRoute2 stall. Acceptance cases passed first. The failed native
candidate demonstrates inadequate fixture duration, not a production defect;
no production red/green fix is claimed.

The unchanged production assets were inspected through the existing loopback
API replay helper with native stopping/cancelled captures and simulated
route-cancellation/cutoff stopping/terminal captures. Its banner now explicitly
labels replay and directs readers to these provenance notes. This is a
presentation check, not live browser/native timing evidence.

All six cases at 1440 and 390 CSS pixels (12 combinations) had no document-wide
overflow. Stopping disabled Start and Stopping controls, showed no ETA and kept
findings pending. Terminal controls became available; lifecycle and measurement
outcomes stayed distinct. TCP evidence links focused the exact record and Enter
on Return to measurement restored the source focus in all 12 combinations.
Full run JSON retained configured cutoff, stop and terminal timestamps; it does
not fabricate a terminal timestamp while stopping. The displayed hop setting
remained the recorded setting. A narrow native-stopping screenshot was inspected;
no console errors were observed. Ignored artifacts are under
`.impeccable/review/issue-14/`; API captures are under `.tools/issue-14/`.
No human comprehension, full accessibility or release acceptance is implied.

## Reproduce

From the repository root in PowerShell, allowing socket/native execution:

```powershell
$env:GOCACHE = Join-Path $PWD '.tools/go-build'
npm.cmd --prefix ui run build
& ./.tools/go/bin/go.exe build -o .tools/hopvista.exe ./cmd/hopvista
$env:HOPVISTA_PACKAGED_EXE = Join-Path $PWD '.tools/hopvista.exe'
$env:HOPVISTA_CUTOFF_CAPTURE = Join-Path $PWD '.tools/issue-14'
$env:HOPVISTA_PACKAGED_CUTOFF_IP = '192.0.2.1'
& ./.tools/go/bin/go.exe test ./tools/app-smoke -run '^TestPackagedMinuteCutoff$' -count=1 -v -timeout 100s
# Expected to fail on the recorded condition: the candidate finishes early.
# A passing candidate must actually reach wall-clock cutoff, not just refuse
# a full-cap send before it. Preserve every failed candidate's capture first.

$env:HOPVISTA_PACKAGED_DRAIN_IP = '192.0.2.1'
& ./.tools/go/bin/go.exe test ./tools/app-smoke -run '^TestPackagedExecutableLoopback$' -count=1 -v
$env:HOPVISTA_LIVE_ADAPTER = '1'
$env:HOPVISTA_LIVE_DRAIN_IP = '192.0.2.1'
& ./.tools/go/bin/go.exe test ./internal/probe/windowsipv4 -run '^TestLive' -count=1 -v
$env:HOPVISTA_LIVE_MINUTE_CUTOFF = '1'
& ./.tools/go/bin/go.exe test ./internal/run -run '^TestLiveMinuteCutoffUnderProcessLoad$' -count=1 -v
$env:HOPVISTA_JOURNEY_CAPTURE = $env:HOPVISTA_CUTOFF_CAPTURE
& ./.tools/go/bin/go.exe test ./internal/httpmeasure -run '^TestHeldRouteStopThroughAPI$' -count=1 -v
node tools/https-journey-preview.mjs
```

Submit `https://fixture.example/nativestopping`, `/nativecancel`,
`/cancelroutestopping`, `/cancelroute`, `/cutoffstopping` or `/cutoff` to the
printed replay listener. Restart the helper after regenerating captures.
The 75-second observation limit in the opt-in packaged test and Go's test
timeout are harness limits, not production native drain guarantees.

## Outstanding acceptance after the first pass

A repeatable native/network fixture that keeps the unmodified packaged app
active through its actual 60-second wall-clock cutoff is still required.
`192.0.2.1` on this Wi-Fi condition is not that fixture. The opt-in test retains
its failure rather than skipping or treating early full-cap refusal as success.
No actual stalled GetBestRoute2 was observed. Other Windows accounts/builds,
alternate networks/VPNs and broader load remain unverified. This first pass
did not complete issue #14 or claim release acceptance.

## Verification

- Focused held-route engine/API checks passed, followed by the existing
  interrupted-HTTP and difficult-path regressions with the extended helper.
- TypeScript no-emit checking and the production Vite build passed. Embedded
  assets are unchanged; the production executable build passed.
- Full root `go test ./... -count=1` and `go vet ./...` passed. Opt-in packaged,
  live adapter and real-minute checks skip without their environment variables;
  their explicit outcomes are recorded above, including the failed candidate.
- `git diff --check` passed. The race detector was not run: CGO is disabled and
  the required Windows C toolchain is unavailable in this workspace.
- Required parallel reviews inspected implementation commit `e8477b6` against
  fixed baseline `b1ef63141f655ae0ae8032a0cd37a3252d0244ff`. Standards: no
  findings. Spec: no actionable implementation defects or scope creep; the
  acknowledged packaged wall-clock cutoff gap remains incomplete. The reviews
  confirmed the disclosed combined native returned-and-closed timestamp limit.
  Browser/native checks were not independently repeated by the reviewers.
  This final documentation update records those reviews without changing code.

## Follow-up: controlled full-minute packaged process pause

On the same unelevated Windows 11 Pro 25H2 build 26200.9711 x64 condition, a
new opt-in harness exercises the **unchanged production packaged executable**
and its real 60-second run window. It first observes a no-reply native target
echo to `192.0.2.1` through the packaged local API. After the first target
timeout, it waits 50 ms for the second echo to enter IcmpSendEcho2, then uses
NtSuspendProcess on **only the disposable child** until just after the run's
configured deadline. NtResumeProcess allows the actual native call to return.
No production code, deadline, cap, adapter, schema, or UI asset changes.

This is a **harness-induced process pause**, not a naturally stalled native
route or echo. During suspension the entire child, including its local API,
cannot answer; the check does not demonstrate a continuously responsive UI
through the paused interval. It demonstrates real packaged cutoff behavior
once scheduling resumes, alongside the already recorded native cancellation
and simulated held-route checks. It must not be described as a native stall.

The first run reached the actual deadline while a native echo was in flight:
started `04:35:41.5690454Z`, cutoff `04:36:41.5690454Z`, stop observed
`04:36:42.0688422Z`, terminal quiescence `04:36:42.0709716Z`. The target's
second native echo began before suspension and ended after cutoff. This run's
**test failed on an incorrect assertion** that at least two trace hops had to
exist. The path correctly stopped at target echo and scheduled no trace hops.
The test was corrected to accept partial evidence from the phase actually in
flight; no production behavior changed.

A corrected run passed: run `1b6e3acff4a395fd79e5655a4f4b7d41` started
`04:37:30.9220925Z`, cutoff `04:38:30.9220925Z`, stop observed
`04:38:31.4213834Z`, and quiesced `04:38:31.4231234Z` (60.5010309 s).
Target echo 2 started `04:37:34.0293829Z`, returned and closed
`04:38:31.4218963Z`; its no-reply timeout, the first target timeout and all
three completed gateway replies remained in schema-1 API evidence. The trace
outcome and IPv4 family were `deadline`, the run lifecycle was `completed`,
and no trace probe started after cutoff. Findings were withheld before
terminal, then Go's HTTP finding appeared. Repeated terminal GETs were stable;
the next start was accepted only after the terminal snapshot.

A further repeat was interrupted by a long host/tool pause and Go's test
timeout. It left a suspended disposable child, which was identified by its
temporary test path and stopped. A subsequent run reached the same cutoff
behavior but its test failed during cleanup because closing a new Windows
kill-on-close job preceded the existing process stopper. Cleanup order was
corrected. The job now ensures the OS kills the child if the test process dies
without running Go cleanups, while normal cleanup stops the child before
closing the job. **Neither failed test attempt is counted as a pass.**

The final run passed with both job ownership and normal cleanup. Run
`9449ac1183792a0a5757f8c451f0de4e` started `05:39:15.608934Z`, cutoff
`05:40:15.608934Z`, stop observed `05:40:16.1101784Z`, and quiesced
`05:40:16.1107191Z` (60.5017851 s). Harness suspension was logged at
`05:39:19.0913676Z` and resume at `05:40:16.1096085Z`. The actual native
target echo 2 entered at `05:39:19.035041Z` and returned/closed at
`05:40:16.1107191Z`. The observed 501 ms stop delay is caused by the
external pause; it is not a normal scheduling-lag measurement. The process
was absent after the test, and normal cleanup produced no error.

The final captured API result was replayed into the unchanged production UI
at 1440 and 390 CSS pixels. Both widths showed lifecycle `completed`, a
distinct IPv4/trace `deadline`, hop limit 64, two retained target timeouts,
no trace hops, and no page-wide overflow. The target echo 2 evidence link
focused the exact timestamped record; Enter on Return to measurement restored
the source focus. Full JSON exposed cutoff, observed stop and quiescence times.
The helper banner identified replay and points to this native-versus-simulation
record. The replay does not establish a continuously visible stopping state
during the externally paused period.

The run result exposes the configured deadline, observed stop time, native
call entry and combined returned-and-closed time, plus run quiescence. A
separate raw IcmpSendEcho2 return instant and IcmpCloseHandle completion
instant remain unavailable in schema version 1. The process pause timing is
logged by the harness, not presented as a measured network delay.

Reproduce from the repository root after building the production executable:

```powershell
$env:GOCACHE = Join-Path $PWD '.tools/go-build'
& ./.tools/go/bin/go.exe build -o .tools/hopvista-minute.exe ./cmd/hopvista
$env:HOPVISTA_PACKAGED_EXE = Join-Path $PWD '.tools/hopvista-minute.exe'
$env:HOPVISTA_PACKAGED_CUTOFF_IP = '192.0.2.1'
$env:HOPVISTA_CUTOFF_CAPTURE = Join-Path $PWD '.tools/issue-14'
& ./.tools/go/bin/go.exe test ./tools/app-smoke -run '^TestPackagedSuspendedMinuteCutoff$' -count=1 -v -timeout 100s
```

The natural no-pause 64-hop candidate remains a failed 38.35-second fixture.
At this stage an actually stalled GetBestRoute2, natural native work spanning
the cutoff, separately timed native return and handle close, and broader
accounts/builds/networks remained unverified. This induced-pause check did not
by itself complete issue #14 or release acceptance.

The final follow-up check passed explicit packaged paused-cutoff testing,
standalone TypeScript checking, the production Vite build, full root
`go test ./... -count=1`, `go vet ./...`, and a fresh production executable
build. Opt-in native fixtures still skip in the ordinary Go suite.

## Final issue #14 check: native return and cleanup timing

On 2026-10-20 (America/Los_Angeles; timestamps below are 2026-10-21 UTC), the
same unelevated Windows 11 Pro 25H2 build 26200.9711 x64 condition had one
active Intel Wi-Fi 6E AX211 adapter at a reported 9 Mbps, with IPv4 default
gateway `10.8.26.1`. The opt-in live adapter checks used the same stated
`192.0.2.1` candidate. Two sandboxed attempts could not open an ICMP handle,
including loopback; these are permission failures, not native timing results.
The elevated-permission retry passed, as did all four opt-in `TestLive*`
checks. There were no deliberately added CPU workers outside the loaded cap
case; ordinary desktop activity continued. The loaded case used GOMAXPROCS=1
and two progressing busy Go workers, as in the original record.

A package-local test observer now records the exact instant immediately after
`IcmpSendEcho2` returns and the instant immediately after `IcmpCloseHandle`
returns. Production constructs `Adapter{}` with no observer. The schema-1
`ended_at` remains the handle-closed instant, so the public run result and UI
remain unchanged. In the focused native cancellation pass, the echo entered at
`04:10:55.0102183Z`, cancellation was requested at `04:10:55.0311731Z`, the
native call returned at `04:10:55.5833103Z`, and the handle closed at
`04:10:55.5837292Z`. The test asserts that order, that no post-cancel echo
starts, and that the public echo end matches handle closure. A second pass of
all live checks confirmed the same ordering; its return and closure fell in
the same reported clock tick. Three full-cap no-reply samples at each of 750,
900 and 1000 ms and the corresponding loaded samples passed; no loaded
pre-send refusals occurred in this repeat. This is observed ordering, not a
promise that every native call returns within a fixed time.

The rebuilt production executable, with the observer nil, passed
`TestPackagedSuspendedMinuteCutoff` again. Its run started at
`04:12:24.2833521Z`, reached the unchanged cutoff at `04:13:24.2833521Z`,
recorded scheduling stop at `04:13:24.7839657Z`, and quiesced at
`04:13:24.7856906Z` (60.5023385 s). The target's second actual native echo
entered at `04:12:27.5976732Z` and returned-and-closed at
`04:13:24.7844749Z`. The harness alone paused the disposable process at
`04:12:27.6490488Z` and resumed it at `04:13:24.7836169Z`. Its pause explains
the approximately 500 ms late stop observation. The check retained earlier
gateway and target evidence, scheduled no trace probes after cutoff, held
findings until terminal, returned stable terminal snapshots, and released the
run slot after quiescence. Existing finite packaged cancellation and simulated
held-route checks cover the responsive `stopping` and busy-slot interval; the
whole-process pause cannot expose a live API view while suspended. The
unchanged UI replay remains the presentation check recorded above.

Reproduce the native ordering and packaged cutoff checks from the root after
building the current executable, using the documented native-network access:

```powershell
$env:GOCACHE = Join-Path $PWD '.tools/go-build'
$env:HOPVISTA_LIVE_ADAPTER = '1'
$env:HOPVISTA_LIVE_DRAIN_IP = '192.0.2.1'
& ./.tools/go/bin/go.exe test ./internal/probe/windowsipv4 -run '^TestLive' -count=1 -v
& ./.tools/go/bin/go.exe build -o .tools/hopvista-minute-current.exe ./cmd/hopvista
$env:HOPVISTA_PACKAGED_EXE = Join-Path $PWD '.tools/hopvista-minute-current.exe'
$env:HOPVISTA_PACKAGED_CUTOFF_IP = '192.0.2.1'
$env:HOPVISTA_CUTOFF_CAPTURE = Join-Path $PWD '.tools/issue-14'
& ./.tools/go/bin/go.exe test ./tools/app-smoke -run '^TestPackagedSuspendedMinuteCutoff$' -count=1 -v -timeout 100s
```

The ticket asks for a repeatable stated fixture and explicitly allows a
test-only seam. Taken with the prior finite native cancellation, held-route
engine/API, and UI replay checks, the above covers **issue #14's acceptance
criteria at this named condition**. The packaged minute fixture induces a
process scheduling delay, not a network-caused native stall. An actual stalled
GetBestRoute2, natural native work spanning the cutoff, additional Windows
accounts/builds/networks, full accessibility and human comprehension remain
unverified; issue #10 release acceptance stays open. These limits are not
reported as demonstrated failures or silently promoted to passes.
