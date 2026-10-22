# Issue 16: accessibility and human comprehension acceptance

Independent checks first performed on 2026-10-20 (America/Los_Angeles) for
[issue #16](https://github.com/DarrylN2/hopvista/issues/16), which had no comments
when read. **Issue #16 remains open: the real participant review is incomplete.**
The first round has partial real participant feedback; no comprehension pass or
human acceptance has been inferred. **Issue #10 release acceptance remains separate
and open.**

## Baseline and scope

The workspace was clean. A fresh `git fetch origin` advanced main from
`b108ae1` to `eab3cecf75f843a75cbf60935bb12858f1ae18b8`, the merge of PR #23.
Branch `codex/accessibility-human-acceptance` starts there. No pre-existing
workspace changes were discarded. Read AGENTS.md, GLOSSARY.md, the agent docs,
PRODUCT.md, DESIGN.md, issue #10, the approved design/verification record,
the acceptance matrix and the issue #11–#14 verification records. No ADR
directory exists.

Condition: Windows 11 25H2 build 26200.9711 x64; Edge 155.0.4283.62; Node
20.16.0; Go 1.27.2 windows/amd64; portable NVDA 2026.2 AMD64, add-ons disabled.
Sandbox permission escalation used for browser/socket access is not a claim
that the app ran with an administrator token. No wider Windows/network claim
is added. React/TypeScript, Vite, plain CSS, the approved composition, schema
version 1 and Go-owned findings are retained. No Go production code changes.

`tools/accessibility-fixtures.mjs` collects the already verified local API
captures in `.tools/issue-11` through `.tools/issue-14`. Its manifest records
source paths and SHA-256 hashes; copied captures are byte-identical. The 64
source modes include duplicate cases retained by the earlier capture helpers.
One additional **synthetic presentation-only** stress variant lengthens URLs
and IDs/references consistently, preserving structured outcomes, numbers and
Go finding strings. It is not another native diagnostic run.

Representative cases retain #11's redirected HTTPS chain, policy stops and
certificate failure; #12's DNS/TCP/TLS/HTTP failures/bounds, partial headers,
missing values, measured zero and 404/503; #13's silence, high/lower RTTs,
on-link gateway, 1/30/64 hops, unreachable and unsent cancelled/deadline samples;
and #14's native stopping/cancelled, held-route simulations and paused-minute
capture. The minute capture used an induced whole-process pause, not a natural
native stall. Replay does not re-establish native timing or scheduling claims.

## Demonstrated defects and TDD

The user confirmed the rendered production UI seam: keyboard input, focus,
accessible names/content and browser API responses. Tests use the committed
schema-1 fixture through the browser API boundary; they do not call component
internals or calculate findings.

| Failure observed before the fix | Change and recheck |
| --- | --- |
| Start disabled its focused button and left focus on BODY, including after a fast terminal result. | Start/cancel requests focus the persistent status message, with a visible outline. Focus survives stopping and removal of the cancel control. The failing start test passed after rebuilding. |
| New Starting/Requesting stop progress feedback stayed stale after a failed request. | Separate failing tests established the status-recovery gap during implementation. Failed starts restore the last state/Ready; unconfirmed cancellation says the last recorded state is retained. Alerts explain uncertainty and keyboard retry remains available. |
| Clipboard failure only told keyboard users to select nonfocusable text. | A failing test established the missing keyboard manual-copy field. Failure now exposes a labelled read-only field, focuses it and selects the exact address. At 320 CSS pixels the field stays within the page. |
| A pending clipboard write disabled its focused button and lost focus. | A separate failing pending-write test passed after using aria-disabled/aria-busy with the existing duplicate-call guard. Focus survives the success acknowledgement. |
| The input boundary was 2.60:1 against its white surface. | A failing rendered contrast check passed after changing only the border from `#96a49a` to `#819187` (3.32:1). The design token was updated. See [W3C non-text contrast guidance](https://www.w3.org/WAI/WCAG22/Understanding/non-text-contrast.html). |
| Actual NVDA output joined the selected IP and scope (`8.8.8.8public`). A whitespace-only fix passed DOM inspection but NVDA still joined the words. | The strengthened failing regression now requires a hidden textual separator and explicit scope label. Actual final NVDA output was `1.1.1.1, address scope:public`. Visible scope styling and literal copied address stay intact. |

The test helper initially expected the wrong fixture IP (`1.1.1.1` rather than
the committed `192.0.2.10`), and its first phase-evidence loop used `dns-1`
instead of `exchange-1-dns`. Those failed checks were corrected from the
fixture, not by changing product data. They are harness errors, not extra
product defects. Every final UI test passes.

## Independent check outcomes

| Check | Evidence and outcome |
| --- | --- |
| Keyboard entry/settings/errors | Passed skip link, target entry, Advanced disclosure, hop field, invalid 0/65/fractional values and submitted 64. Verified start/cancel, error/retry, pending/final controls, visible focus and no trap in the exercised flows. Go's existing setting-policy tests remain regression coverage. |
| Evidence | All four website phases open their exact measurement, expose technical content and return keyboard focus. The capture matrix additionally opens phase evidence, returns to its source and opens full JSON at every tested width. Native NVDA navigation returned from `exchange-4-http` to that exact evidence button. |
| Layout | 64 source modes × 5 widths = 320 baseline combinations; the extended collection including stress passed 325 combinations at 1440/960/760/390/320 CSS pixels. No page-wide overflow. Summary/timing precede path on narrow screens; path is left on desktop. Hop counts, exact Go finding text, pending findings and evidence identity matched the captures. |
| Affected layout recheck | 20 additional combinations (`i11completed`, `i12running`, `i13limit64`, `stress`) passed after the first scope-text edit. The subsequent scope fix adds visually hidden text only; the final NVDA recheck and regression cover it. |
| Native zoom | Actual Edge Ctrl+plus reached devicePixelRatio 2 with no viewport emulation, from baseline ratio 1. At 200%, the window had 460 CSS pixels with 453px document width (scrollbar excluded). Inspected 64-hop content and opened long-ID evidence directly in the native window. Evidence and Return remained readable. Restored 100% afterward. |
| Reduced motion | The running Edge browser matched `prefers-reduced-motion: reduce`; zero active animations, selected-record animation `none`, arrow transform `none`. Static selection/focus and text feedback remained. This uses browser media emulation, not only CSS/source inspection or a changed Windows preference. |
| Contrast and color | Twenty sampled rendered foreground/background roles passed applicable text thresholds; lowest active-text ratio 4.79:1 (placeholder). The input boundary now passes 3:1. State words, outcomes, missing-value explanations and focus/selection remain textual or outlined; color does not classify RTT health. This is scoped measurement, not a blanket WCAG certification. |
| Pointer targets | Fine-pointer evidence/copy targets are at least 32px high; coarse-pointer emulation measured them at 44px. Inputs, action buttons and disclosures are at least 44px. The wordmark link is approximately 144×37px; its height is below the 44px enhanced target, with ample width and separation. No physical touch-device claim. |
| Clipboard | Real Edge clipboard copying wrote exactly `192.0.2.10` and Copied reset. The disposable context had granted permission; the test first wrote known disposable text and never read pre-existing user clipboard contents. Synthetic rejection passed fallback checks. A separate browser-enforced `Permissions-Policy: clipboard-write=()` check also passed without replacing navigator.clipboard. Neither is a user denying a native permission prompt. |

The initial long layout harness failed at its final reduced-motion selector;
its completed-looking but unsaved matrix was not accepted. The corrected
320-case measurement and subsequent 325-case assertion run passed. The
one-off PowerShell script-conversion attempt had a quoting error and did not
produce a runnable file; the committed ES modules passed syntax checks.

## Real screen-reader record

The official NV Access download's Authenticode signature was Valid, signer
NV Access Limited. It was extracted as a portable copy under ignored `.tools/`,
with no system-wide installation. Ran the actual NVDA process against a
dedicated native Edge window and inspected NVDA's input/output speech logs,
not an agent-created accessibility-tree substitute. The speech log is actual
screen-reader-generated output; no human listening, blind-user usability
review or human comprehension pass is claimed. See the
[NVDA user guide](https://download.nvaccess.org/documentation/userGuide.html)
for portable and logging options.

- Target URL was read as a required edit; Advanced as collapsed/expanded;
  Maximum hops (1–64) as a required editable spin button with selected 30.
- The final UI announced Starting, Running, completed, stopping and cancelled.
  The refresh-failure alert explicitly said it was retrying and retained the
  last recorded state. Native speech at 21:59:08 local included the controlled
  refresh error; stopping followed at 21:59:38 and cancelled at 21:59:59.
- Heading navigation reached What this run observed, Website timing and IPv4
  path in that order. Actual output read the selected address, Copy button,
  `Pending until the diagnostic run finishes.` in the running capture, then
  the final Go HTTP-204 observation in the completed HTTPS capture. The
  scope-separator final recheck was at 22:02:50 local.
- NVDA reported a four-item path list, hop 1 with no reply within 750 ms and
  no address/RTT, followed by hop 2's time-exceeded response, address and raw
  450 ms RTT. The earlier 64-hop pass reported a list with 64 items, hop-64
  evidence and the hop-limit outcome in order.
- Selecting final-exchange HTTP evidence read its ID, completed method,
  measured zero, 3000 ms bound and timestamps. Native Tab announced Return
  to measurement; Enter returned to `View evidence exchange-4-http`.

UI Automation intermittently returned stale omnibox focus and COM warnings;
NVDA also logged a recovered watchdog pause and an initial action fallback.
Those tool warnings are retained in ignored logs, not promoted to successful
focus evidence. Used actual NVDA speech, browser DOM focus checks and direct
native screenshots for the reported outcomes. An automation screenshot at
200% was blank and discarded; the native capture visibly showed wrapped
long-ID evidence and its Return control. Zoom can change the current scroll
position; selecting the measurement at the zoomed size brought it into view.
Portable NVDA was stopped after testing; native zoom was restored to 100%.

## Commands and verification

Run from the repository root in PowerShell. Build before browser tests; these
checks intentionally exercise embedded production assets rather than a Vite
development server.

```powershell
node ui/node_modules/typescript/bin/tsc --noEmit -p ui/tsconfig.json
npm.cmd --prefix ui run build
npm.cmd --prefix ui test
node tools/accessibility-fixtures.mjs
$env:HOPVISTA_JOURNEY_CAPTURE = Join-Path $PWD '.tools/issue-16'
node tools/https-journey-preview.mjs
# In a second shell, use the actual URL printed by the preview:
$env:HOPVISTA_REVIEW_URL = 'http://127.0.0.1:<printed-port>/'
node tools/accessibility-observe.mjs
node tools/accessibility-metrics.mjs
# Optional affected-case recheck:
$env:HOPVISTA_REVIEW_MODES = 'i11completed,i12running,i13limit64,stress'
node tools/accessibility-observe.mjs
Remove-Item Env:HOPVISTA_REVIEW_MODES

$env:GOCACHE = Join-Path $PWD '.tools/go-build'
& ./.tools/go/bin/go.exe test ./internal/result ./internal/pathmeasure ./internal/engine -count=1
& ./.tools/go/bin/go.exe test ./... -count=1 -v
& ./.tools/go/bin/go.exe vet ./...
& ./.tools/go/bin/go.exe build -o .tools/hopvista-issue-16.exe ./cmd/hopvista
$env:HOPVISTA_PACKAGED_EXE = Join-Path $PWD '.tools/hopvista-issue-16.exe'
& ./.tools/go/bin/go.exe test ./tools/app-smoke -run '^TestPackagedExecutableLoopback$' -count=1 -v
```

Repeated focused tests and TypeScript checks, final production build, all 12
UI tests, full root Go suite, vet, executable build and explicit packaged
loopback smoke passed. The ordinary Go suite skipped opt-in live HTTP/TLS,
engine/API, native adapter/tracert, real-minute, packaged cutoff, smoke and
restart/privacy fixtures. Only packaged loopback smoke was explicitly enabled
again here; no new paused-minute, live remote, native stall or broader account
test is claimed. Existing #11–#14 native evidence retains its original scope.
No race-detector result is claimed in the cgo-disabled environment.

`gh` was unavailable; GitHub app tools read the issue/comments and merged PR.
The first git fetch could not write FETCH_HEAD in the sandbox; the permitted
retry succeeded. The first sandboxed replay was unreachable from the browser,
and a sandboxed Edge launch failed; allowed loopback/browser retries succeeded.
Impeccable context could not install its engine in the restricted cache, so
existing PRODUCT/DESIGN context was read directly; no detector result is claimed.
Adding the development-only Playwright dependency reported an existing high
Vite audit result. `npm --prefix ui audit --json` identified the already-pinned
Vite 6.3.6; it was not auto-upgraded as part of accessibility acceptance.

Automatic approval review rejected a test that would read and restore the
pre-existing user clipboard. That test was never run. The approved safer
replacement uses only known disposable text. Review artifacts and raw logs are
ignored under `.impeccable/review/issue-16/` and `.tools/issue-16/`.

## Human acceptance and review disposition

Use [the concrete participant script](issue-16-participant-review.md).
Round 1 has been requested from the user. Participant identifier/familiarity,
actual answers, difficulties/help and affected-case rechecks are recorded as
supplied in that script. Initial feedback describes ease/clarity in general;
concrete readings and explanations remain pending. No agent or fixture is
counted as human acceptance.

The required independent Standards and Spec reviews are recorded below. Do not
close issue #16 or infer whole-release acceptance from this independent work
alone.

## Resumed follow-up: 2026-10-21

The user paused before the participant review, then requested a workflow without
desktop control. On resume, only repository/terminal work and isolated headless
browser checks were used. No open browser, pointer, keyboard, Windows preference
or screen reader was operated. The review server only serves loopback captures;
the participant opens and operates the browser themselves.

The last metrics invocation before the pause actually **failed**: Start was
sampled with white text on an intermediate background during its 160 ms
disabled-to-enabled transition, yielding 3.47:1. This supersedes the earlier
stable-color pass for that animated state. A new deterministic rendered test
settles the disabled state, completes the controlled diagnostic run and samples
the actual CSS transition 40 ms into re-enabling. It failed at **2.22:1** before
the fix and passed afterward. Start/cancel and Copy now change their foreground
and background states immediately; evidence, arrow and acknowledgement motion
remain. A separate copy-acknowledgement check passes with the approved colors.

After rebuilding, `npm.cmd --prefix ui test` passed **13 of 14** tests with no
failures. The remaining test is explicitly opt-in with
`HOPVISTA_LIVE_CLIPBOARD=1` because it writes the system clipboard; it was not
enabled on resume. Its earlier real-browser pass remains historical evidence.
Synthetic clipboard cases and the actual browser-enforced clipboard policy case
still ran, without reading or writing the system clipboard. The original
rendered-UI TDD seam approval remains in force.

`node tools/accessibility-metrics.mjs`, using the new replay listener printed
at `http://127.0.0.1:50853/`, passed with 20 sampled color pairs, no contrast
failures, the selected manual-copy address at 320px and no page overflow. The
wordmark remains the disclosed 144×37px coarse-pointer exception to the enhanced
44px target. Typechecking and production build passed again. Script syntax
checks, `go test ./... -count=1`, `go vet ./...` and `git diff --check` also
passed. The Go suite's opt-in native/network/packaged checks remained disabled,
as above. Full-suite output is retained in `.tools/issue-16/resume-go-test.log`.

The participant invitation was reissued with the new local URL and unprimed
normal-result questions. P1 replied that they are new to networking/HopVista,
find the information generally easy to locate and think most phases are clear.
These are real perceived-ease observations, not yet evidence of the requested
phase/limit comprehension. A follow-up asks for the actual selected IP,
observation, uncertainty and phase explanations, without providing answers.

## Participant timing comprehension follow-up

P1 supplied six Chrome screenshots of the normal replay, timing rows and opened
technical evidence, and stated: “For each timing, I'm not sure.” This is a real
comprehension failure for the phase task, superseding their earlier perceived
clarity. Their reported mouse/keyboard familiarity does not establish a specific
keyboard task pass. The screenshots do not establish their interpretation of
the finding, selected IP or limits.

At the approved rendered-UI seam, a new regression check required a visible
four-step guide before opening technical evidence. The focused command
`node --test --test-name-pattern="website timing explains" tools/accessibility-ui.test.mjs`
failed against the previous generic introduction (1 failure; 14 pattern skips).
The introduction now explains DNS/address, TCP/connection, TLS/encryption and
HTTP/response start together, and individual row descriptions use plainer
language while preserving first-response-header-byte scope and exclusion of
body download/rendering. No schema, measurement or Go finding was changed.
This test verifies available guidance, not human understanding.

`npm.cmd --prefix ui run build`,
`npx.cmd --prefix ui tsc --noEmit -p ui/tsconfig.json`, and
`npm.cmd --prefix ui test` passed (14 passed, 1 opt-in system-clipboard skip).
With `HOPVISTA_REVIEW_URL=http://127.0.0.1:50853/` and
`HOPVISTA_REVIEW_MODES=i11completed,i12tcptimeout,i12measuredzero,i13limit64`,
`node tools/accessibility-observe.mjs` passed all 20 affected scenario/viewport
cases, including preserved findings, evidence navigation and reduced motion.
Report: `.impeccable/review/issue-16/observations-2026-10-22T04-01-29.546Z.json`.
`node tools/accessibility-metrics.mjs` passed 20 sampled color pairs and the
320px manual-copy fallback; the previously disclosed wordmark target exception
remains. No desktop/browser control or system clipboard access was used.

The initial full-suite retry used bare `go`, which was absent from PATH and
therefore did not run. With `CGO_ENABLED=0`,
`& ./.tools/go/bin/go.exe test ./... -count=1` and
`& ./.tools/go/bin/go.exe vet ./...` then passed. Full-suite output is retained
in `.tools/issue-16/participant-copy-go-test.log`; the previously described
opt-in checks remain disabled. `git diff --check` also passed.
The participant has been asked to refresh and reread the updated production UI;
actual comprehension recheck answers remain pending. This wording change has
not yet been rechecked with real NVDA under the current no-desktop-control
workflow; prior real-reader checks apply to the earlier production wording.

## Standards

Independent source review of implementation commit
`7fe06e457caf847a636b6302f538627be6641987`, using
`git diff eab3cecf75f843a75cbf60935bb12858f1ae18b8...7fe06e457caf847a636b6302f538627be6641987`:
no actionable findings. The reviewer found no documented-standard violations or
material code-smell findings. Schema version 1, Go-owned findings, the approved
rendered-UI seam and explicit verification scope were preserved. Tests, desktop,
browser and clipboard checks were not independently rerun by the reviewer.

## Spec

Independent source review of the same diff against issue #16 and its referenced
records: no new actionable implementation defects or scope creep. The reviewer
confirmed that the real participant review and recording of actual responses
remain required. At review time the first invitation was awaiting a response;
the subsequent P1 feedback and concrete follow-up request are recorded above
and in the participant script. These documentation additions do not claim a
comprehension pass or change reviewed production behavior. The reviewer did not
rerun checks or operate the desktop.

Review summary: Standards 0 actionable findings; Spec 0 new implementation
findings. **Known issue #16 blocker: completion of the real participant review,
plus any demonstrated defect fixes and affected-case rechecks. Issue #10 release
acceptance remains separate and open.**

Follow-up source review of `git diff a58b2e8...4426b9c` completed after the
participant-driven timing change. Standards: no documented violations or
actionable heuristic smells. Spec: no new implementation defects or scope
creep; timing boundaries, schema and Go-owned findings preserved. Spec explicitly
retained the pending participant recheck, remaining participant tasks and lack
of a fresh real-reader check of the changed wording. Neither reviewer operated
a browser, desktop or clipboard or independently reran checks.
