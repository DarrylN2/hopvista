# Issue 13: difficult path and echo results

Verified on 2026-10-18 (America/Los_Angeles) for
[issue #13](https://github.com/DarrylN2/hopvista/issues/13), which had no comments
when read. Branch `codex/path-echo-acceptance` starts at freshly fetched
`origin/main`, `93aa698536a5832a9ecf26d0905d3484ad941408` (merged PR #20).
The workspace was clean. The approved design, React/TypeScript, Vite, plain
CSS, schema version 1 and Go-owned findings are preserved. **Overall release
acceptance remains open.**

## Existing evidence and the gap

Existing `internal/pathmeasure` cases cover a silent hop followed by a reply,
on-link/separate gateways, an unreachable response, hop limits 1/30/64 and
cutoff after route selection. Existing engine cases cover HTTP succeeding with
silent ICMP and retained progress through cancellation. Issue #11 verifies
final/attempted-host identity across HTTPS redirects; issue #12 verifies partial
website phases and pending findings. Earlier UI review included an explicitly
synthetic long/high-RTT presentation case. These checks are reused, not counted
as new live path measurements.

The remaining gap was durable high-intermediate-RTT/silence and difficult-path
cases carried through the real engine/local API into the approved results view.
`TestDifficultPathsThroughAPI` extends issue #11/#12's TLS journey, API lifecycle
and capture helpers using the existing normalized fake probe adapter. The API
helper now permits a configured hop limit or omission to exercise its default.
The existing schema-version-1 `completed.json` fixture now records a 450 ms
intermediate reply, a silent hop and a later 10 ms destination reply, with
consistent elapsed timestamps; its HTTP status is 200 so its existing
dominant-wait example describes a successful HTTP response. No duplicate golden
fixture was added.

## Controlled scenarios

All cases retain completed HTTPS evidence from the same selected `8.8.8.8`
pin. Normal HTTP returns 204; the attempted-host case retains a 302 and a
new-host non-public-address policy stop. IDs and all DNS/finding/variation
references are checked for uniqueness/resolution.

| Replay path | API assertions and observed rendering |
| --- | --- |
| `/highsilent` | Separate route-specific gateway: three replies to `192.0.2.254`. Trace continues after a 750 ms unanswered hop to 450 ms and 12 ms time-exceeded replies, then a 14 ms destination reply. Target has five attempted/four replied samples, including measured zero and a miss. Successful RTTs 0/14/20/18 yield absolute changes 14/6/2 and median variation 6 ms. |
| `/onlink` | Same silent/high/lower trace with an on-link route. Gateway is skipped with zero samples/attempts; no separate gateway is fabricated. |
| `/limit1` | Explicit limit 1: HTTP completes, target echo has zero replies, and trace ends at hop limit with one unanswered hop. No target outage or HTTP failure is inferred. |
| `/default30` | Start request omits `max_hops`. API result and trace use 30; all 30 attempted hops are unanswered and the terminal trace is hop-limit-reached. |
| `/limit64` | Explicit limit 64: all 64 ordered unanswered hops survive API serialization and rendering. Target variation is unavailable, without a numeric substitute. |
| `/unreachable` | Separate gateway remains inspectable. Trace records a time-exceeded reply then an actual normalized unreachable reply from `192.0.2.253`, and stops at TTL 2. Target has one reply out of five attempts; variation is unavailable. HTTP still completes. |
| `/attempted` | An allowed starting host redirects to a different host selecting a non-public address. The response and blocked DNS selection remain inspectable; path/probes use the earlier approved `8.8.8.8` and say attempted host. No final URL/host or connection to the blocked IP is invented. |
| `/deadline` | A test-only 500 ms manager window cannot fit the target echo's full 1000 ms cap plus margin. The normalized adapter returns an unstarted deadline sample with no send/end time, address, duration or RTT. The run is completed with partial measurements, trace deadline and zero attempted trace hops; the view says Not sent (deadline). |
| `/cancelled` | API cancellation arrives during fake target-echo setup, before its send guard. The sample is not sent, has no timestamps/address/RTT, and is cancelled. Lifecycle is cancelled, trace is cancelled, completed HTTP survives and no final finding is published. The view says Not sent (cancelled). |

Normal findings remain HTTP observations or valid dominant **HTTP-phase** waits
if the real protocol timings meet that rule on a loaded machine. Their evidence
must reference that exchange's completed phases; ICMP RTTs cannot enter the
website comparison. The policy case keeps the Go redirect observation. The API
helper also rejects findings in nonterminal snapshots. No production defect was
demonstrated; the acceptance cases passed first, so no production red/green fix
or UI change was necessary.

## Reproduce

From the repository root in PowerShell:

```powershell
$env:GOCACHE = Join-Path $PWD '.tools/go-build'
$env:HOPVISTA_JOURNEY_CAPTURE = Join-Path $PWD '.tools/issue-13'
& ./.tools/go/bin/go.exe test ./internal/httpmeasure -run TestDifficultPathsThroughAPI -count=1 -v
& ./.tools/go/bin/go.exe test ./internal/result ./internal/pathmeasure ./internal/engine -count=1
node tools/https-journey-preview.mjs
```

Open the printed loopback URL and submit `https://fixture.example/highsilent`
or another captured path from the table. The persistent banner labels the
simulation. This helper replays exact API-captured JSON in production assets;
it does not start a new diagnostic run. Restart it after regenerating captures.
Captures are ignored local files under `.tools/issue-13/`.

The probe adapter scripts route/ICMP statuses, addresses and RTTs. DNS answers
and TCP peer addresses are simulated; Go TLS/certificate verification and HTTP
run over `net.Pipe`. No public addresses are contacted. Sampling waits are
skipped, so elapsed adapter durations do not simulate the scripted RTTs or
prove native timeout timing. The shortened deadline and controlled cancellation
setup are test-only: they do not satisfy issue #14's real packaged 60-second or
native-drain requirements, and production cutoff behavior is unchanged.

## Browser verification

- All nine terminal cases were replayed at measured 1440, 390 and 320 CSS pixels
  (27 checks), with no document-wide horizontal overflow. Narrow reading order
  remains summary, website timing, then path. The complete 1/30/64-hop content
  and the five distinct trace outcomes remain available.
- Hop order, raw reply addresses/RTTs, the 750 ms unanswered-hop bound, target
  zero versus absent RTT, gateway/on-link text, target reply counts and variation
  availability match the API. Counts are not labelled as general packet loss.
- The probe-order-only connector caveat, direct-path/browser caveat, and adjacent
  intermediate-ICMP caution remain visible. The 450 ms reply is not classified
  as a causal bottleneck or incorporated into website timing/findings.
- Eighteen desktop/narrow evidence checks focus the exact record and keyboard
  Enter on Return to measurement restores the originating control. These include
  unanswered/high RTT/zero samples, route, variation, unreachable reply, unsent
  deadline/cancelled samples and the 64th hop. Opening raw variation evidence
  exposes the formula, value and successful-sample reference pairs.
- No browser console warnings/errors were observed. DOM checks and representative
  screenshots are ignored artifacts under `.impeccable/review/issue-13/`.

This is controlled functional/rendered evidence, not a human comprehension,
screen-reader, 200% zoom or full accessibility acceptance result. It adds no
live Windows/native/remote path or broader platform claim. Existing opt-in live
and packaged evidence remains scoped to its recorded runs.

## Automated verification

The focused HTTP/API, result-schema, pathmeasure and engine suites pass. The
full root `go test ./... -count=1`, `go vet ./...`, executable build and repeated
TypeScript checks pass. Opt-in live/native/packaged/minute fixtures skip in the
ordinary full suite and were not enabled for this controlled path task. No
production UI source or embedded assets changed, so no Vite rebuild was needed.
No race-detector result is claimed in the cgo-disabled environment.

## Standards

Independent review of implementation commit `9e36023` against
`93aa698536a5832a9ecf26d0905d3484ad941408`: no findings. All five changed files
follow the documented standards/glossary, reuse existing helpers/fixtures and
stay at the engine/API/normalized-probe seams. No meaningful baseline code-smell
findings warranted changes.

## Spec

Independent review of the same diff against fetched issue #13 and existing
coverage: no actionable findings or scope creep. The requested scenarios,
evidence and noncausal wording are covered; simulation and native/release
limits remain explicit. Browser checks were not independently repeated by the
reviewer. The final commit adds this review record to the reviewed changes.

Review summary: Standards 0 findings; Spec 0 findings. Overall release
acceptance remains open.
