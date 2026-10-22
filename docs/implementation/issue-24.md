# Issue #24: packaged UI functional journeys

Verified on 2026-10-21 (America/Los_Angeles) for
[issue #24](https://github.com/DarrylN2/hopvista/issues/24), following the
[approved scope and public seams](issue-24-scope.md). The starting branch was
`codex/accessibility-human-acceptance` at clean commit `01692fa`.
Existing issue #16 work through that commit was preserved. Human comprehension
review is explicitly deferred; issue #16 and overall release acceptance remain
open. No participant questions or desktop/existing-browser control were used.

## Functional result

`tools/packaged-ui.test.mjs` joins the embedded production UI, actual same-origin
session/start/poll/cancel endpoints, production manager/engine and Windows IPv4
adapter in repeatable isolated headless browser journeys. No application API
response is intercepted or fabricated. Target servers use real loopback TCP.
The UI's full JSON export is compared with the terminal public API result, and
the terminal wait requires the newly accepted run ID to avoid observing the
previous run's completed screen.

| Journey | Observed outcome |
| --- | --- |
| Form submission, advanced hops, redirect and error response | A form-submitted loopback URL with hop limit 4 produces a real 302 relative redirect followed by 503. Both exchanges show selected and actual peer `127.0.0.1`, matching API statuses. The native trace records hops. The displayed Go observation and limitation match the API verbatim. |
| Evidence navigation | The final exchange's HTTP link focuses its exact evidence record; its raw evidence matches the API measurement. Return to measurement restores focus to the originating button. |
| Rejection and incomplete HTTP | An IPv6 literal returns the real 400 `ipv6_not_supported_in_v1` rejection, displayed with a usable retry. A subsequent loopback redirect to a connection closed before headers retains the completed 302 and completed TCP evidence, with HTTP error and absent duration/status. The UI distinguishes incomplete measurements from completed lifecycle. |
| Another run | A successful 204 run follows the incomplete run in the same UI. It has a new run ID and replaces the prior exchange/target content. |
| Cancellation | A loopback server redirects to a held-header request. Clicking Cancel sends the actual mutation; the terminal run is cancelled, earlier 302 evidence survives, unfinished HTTP is cancelled without a duration, and no final finding is fabricated. A subsequent 200 run succeeds. |
| Process restart | The first disposable process is stopped and waited for, then the same executable is relaunched from the same isolated directory. The same disposable browser context navigates to the new listener and shows an empty target/results view with usable Start. Before another run, the previous run ID returns 404. A fresh form-submitted run completes with HTTP 200. |

All functional slices passed initially once browser launch was permitted. No
production defect was demonstrated; no product red/green fix is claimed. These
are new integration checks at the approved seams, added incrementally one
journey at a time. UI source/assets, schema version 1 and Go-owned findings did
not change. `npm --prefix ui test` now includes the new checks, with explicit
opt-in skips when the executable is absent; `test:packaged` runs them alone.

## Exact executable and conditions

- Executable: `.tools/hopvista-ui-e2e.exe`, built from clean implementation commit
  `9687962e2b39d5735f4b72676712925fcd5b4940`. `go version -m` confirms that revision
  and `vcs.modified=false`. The later change tightens only the test wait.
- SHA-256: `C5A23808D9BE2177A1F6BC74D07B9D8F23D35FB1DC57904374AF1D3FA8B8973C`.
- Windows version `10.0.26200.0`, x64; executable/browser checks used an
  unelevated medium-integrity token. Go `go1.27.2`, `CGO_ENABLED=0`, Node
  `v20.16.0`, Playwright `1.62.1`, installed Edge `155.0.4283.62` in headless mode.
- Each test copies the executable into a fresh temporary directory. Child PATH
  contains only Windows System32. Tests own and close their browser contexts,
  target servers and disposable app processes; process closure precedes removal
  of the copied executable. No personal profile or clipboard is accessed.
- Initial restricted-sandbox Edge startup failed before any journey executed.
  The permitted isolated headless retry outside that restriction passed. A
  sandboxed CIM OS lookup was unavailable; OS/architecture/token checks were
  recorded separately in the same permitted context as the executable checks.

## Reproduction and checks

Follow [the packaged browser command](../../tools/app-smoke/README.md). Build the
UI, build the executable, set its absolute `HOPVISTA_PACKAGED_EXE` path and run
`npm.cmd --prefix ui run test:packaged`. The supplied binary's checksum is printed.
Building the recorded source commit reproduces the source used here; another
Go/environment may yield a different binary checksum. The toolchain is needed
on the test host, not on the packaged child PATH.

Verification results:

- `npm.cmd --prefix ui run build`: typecheck and production build passed;
  generated assets have no tracked content changes.
- Focused runs of each new journey: passed. The final combined UI run with
  `HOPVISTA_PACKAGED_EXE` enabled passed **17 tests**, with **1 explicit skip**
  for the pre-existing opt-in system-clipboard check. Clipboard opt-in was not
  enabled. Log: `.tools/issue-24-ui-test.log`.
- Against the exact executable above, both `TestPackagedExecutableLoopback` and
  `TestPackagedRestartPrivacy` passed with their opt-in path enabled. These retain
  Host/Origin/token/no-CORS, native loopback route/echo/trace, tracert, pinned curl,
  actual restart/token rotation, output and launch-directory checks. Log:
  `.tools/issue-24-packaged-test.log`.
- Full root `go test ./... -count=1` and `go vet ./...`: passed with
  `CGO_ENABLED=0`. Log: `.tools/issue-24-go-test.log`. Ordinary-suite optional
  network/native/minute/packaged checks remained disabled; the two applicable
  packaged checks were enabled separately as described above. No race-detector
  result is claimed. The separate feasibility-harness module was not rerun.
- Final TypeScript check, `node --check tools/packaged-ui.test.mjs` and
  `git diff --check`: passed.
- After independent Spec review identified a possible stale-terminal test wait,
  the wait was tightened to require the accepted run ID in the public JSON export
  as well as terminal state. All three packaged browser journeys passed again,
  **3 passed, 0 skipped**, against the same executable. Log:
  `.tools/issue-24-packaged-ui-recheck.log`. The change affects only test
  synchronization, so the Go suite was not repeated.

## Standards

Independent source review of `git diff 01692fa...9687962`: **0 findings**.
The reviewer confirmed the agreed browser/API seams, controlled target boundary,
glossary vocabulary, cleanup registration and preserved product boundaries. No
meaningful baseline code smell warranted a change. Tests were not independently
rerun by the reviewer.

Independent follow-up review of `git diff 9687962...d8a78f0` found **0 new
actionable Standards findings**, including the synchronization fix and this
verification record. It confirmed that the run-ID check stays at the public
JSON-export seam and the record distinguishes checks, skips and remaining limits.

## Spec

Independent source review of the same diff: **1 repeatability finding**. The
terminal wait could accept the preceding run's completed screen before the new
POST response was consumed. Assertions prevented a false pass, but a false
failure remained possible. The wait now checks the new run ID and terminal state.
No missing functional scenarios or scope creep were identified. Independent
recheck of `d8a78f0` confirmed the repeatability finding **resolved**, with **0
remaining actionable Spec findings**. The reviewer also checked the recorded
logs and confirmed that the evidence wording matches their scope. No tests or
browser operations were independently rerun.

Review summary: Standards **0 findings**; Spec **1 original finding resolved,
0 remaining**. This final documentation update records both review dispositions.

## Remaining boundaries

This milestone verifies functional journeys on the recorded loopback Windows
condition. Held HTTP headers exercise context-aware cancellation; they do not
establish native drain or the 60-second cutoff. Existing issue #14 evidence
remains separate. No public HTTPS, alternate network/account/build, genuine
native stall, human comprehension or fresh screen-reader check is claimed.
Distribution remains experimental Windows 11 x64 under the existing tested-
condition limits. Issue #15's other evidence is retained; issue #16 and issue
#17's release decision remain open. No release was published.
