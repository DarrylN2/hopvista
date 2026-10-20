# Issue 14: packaged cancellation and cutoff evidence

Checked on 2026-10-19 (America/Los_Angeles; captures use 2026-10-20 UTC) for
[issue #14](https://github.com/DarrylN2/hopvista/issues/14), with no comments
when read. Branch `codex/packaged-cutoff-acceptance` starts at freshly fetched
`origin/main`, `b1ef63141f655ae0ae8032a0cd37a3252d0244ff` (merged PR #21).
The workspace was clean. The approved UI, React/TypeScript, Vite, plain CSS,
schema version 1 and Go-owned findings are unchanged. **Issue #14 remains
partially verified; release acceptance remains open.**

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

## Outstanding acceptance

A repeatable native/network fixture that keeps the unmodified packaged app
active through its actual 60-second wall-clock cutoff is still required.
`192.0.2.1` on this Wi-Fi condition is not that fixture. The opt-in test retains
its failure rather than skipping or treating early full-cap refusal as success.
No actual stalled GetBestRoute2 was observed. Other Windows accounts/builds,
alternate networks/VPNs and broader load remain unverified. Do not close issue
#14 or claim release acceptance from this record.

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
