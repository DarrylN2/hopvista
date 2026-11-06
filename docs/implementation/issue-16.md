# Issue 16: accessibility and human comprehension acceptance

Latest participant follow-up, 2026-10-27: human review has resumed. The same
participant reports Chrome at 100% zoom and difficulty understanding the main
observation, limits and selected IPv4. [Their actual answers](issue-16-participant-review.md)
are retained. The follow-up below adds plainer visible guidance and labels;
the participant's recheck and issue #16 acceptance remain pending.

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

P1's revised-copy recheck still did not establish phase comprehension: they
described TCP making the connection and TLS securing it, but were uncertain
about DNS, described HTTP as a protocol without explaining its measured wait,
and explicitly did not know whether download/display were included. Their full
answer is preserved in the participant record. The facilitator then supplied
plain definitions and explained the excluded body download/page display; any
later answer must be recorded as assisted. The automated guidance test remains
a regression check for available content, not evidence that this human defect
is resolved. No further production change or acceptance pass is claimed.

## Resumed participant readability follow-up: 2026-10-27

P1 resumed review after PR #26 merged, from clean `b67eaee`. They described the
example as hops/connections, reported technical wording and confused the
selected number address with the submitted website link. They explicitly asked
what IPv4 means. This is actual difficulty locating/interpreting the intended
finding and address, not a comprehension pass. The participant record preserves
the prompts, answers, previous help and self-reported Chrome/100% scope.

At the already approved rendered-production-UI seam, the focused regression
`node --test --test-name-pattern='result explains its purpose' tools/accessibility-ui.test.mjs`
first failed against the old intro (1 failure, 15 pattern skips). It passed
after the production rebuild (1 pass, 15 pattern skips). It checks visible
guidance before opening details and preservation of the supplied finding,
limit and literal address; it does not measure human understanding.

Implementation `1654ecd` changes the intro, finding/address/limit headings,
defines IPv4 beside the displayed number and explains its difference from the
entered link. Existing static guidance for completed response/dominant-wait
findings is simpler and precedes the exact recorded observation. Detailed
dominant-wait help uses “step” and “website request”. Go-owned finding strings,
limits, measurements, evidence references and schema version 1 are unchanged.
The approved composition/colors/typography and interactions are retained.

Verification:

- TypeScript/Vite build and final standalone typecheck passed.
- Full `npm.cmd --prefix ui test`, with the new executable enabled, passed
  **18 checks with 1 explicit system-clipboard skip**. This includes all three
  real packaged browser journeys, without replacing application API responses.
- The new executable is `.tools/issue-16/hopvista-reading-20261027.exe`, source
  `1654ecd7ef49c04b6864d79731c3ec0fdafd6d50`, Go 1.27.2 windows/amd64,
  `CGO_ENABLED=0`, `vcs.modified=false`; SHA-256
  `4AB6CBE6DACBFD6729573605D0BB13FC9F46A2F577918684F6074B5A45BD5732`.
  This is a new verification build, not the earlier `a159f9c` candidate/checksum.
- Fifteen existing captured/synthetic modes at five widths passed **75
  combinations** with preserved exact findings, evidence focus/return, reading
  order and no page-wide overflow. Runtime reduced motion matched with zero
  active animations. Report:
  `.impeccable/review/issue-16/observations-2026-10-28T03-30-52.541Z.json`.
- Twenty sampled color pairs passed; the manual-copy fallback remained selected
  and within 320px. The existing 144×37px wordmark exception remains disclosed.
- The normal replay was additionally inspected in isolated headless Edge at
  1440/390/320 CSS pixels; visible address guidance, closed technical details,
  exact observation/limit and no overflow passed. Desktop/narrow screenshots
  were visually inspected. No unexpected page errors were observed.
- Full root `go test ./... -count=1`, `go vet ./...`, and `git diff --check`
  passed. Ordinary-suite native/network/minute/packaged opt-ins were disabled;
  the packaged browser checks were enabled separately as above. No race result
  is claimed with cgo disabled. No Go source changes were made.

Local logs are `.tools/issue-16/resume-copy-{red,green,go-test,ui-test,look,observe,metrics}.log`.
Existing browser/desktop/clipboard sessions were not controlled. Impeccable
context loading could not install its engine in the restricted cache, so the
existing product/design context and clarify/craft-floor references were read
directly. This did not prevent the permitted rendered checks.

### Standards

Independent read-only review of `git diff b67eaee...1654ecd`: **0 findings**.
The reviewer checked AGENTS/agent docs, glossary, product/design boundaries,
public rendered seam, data preservation and the participant record. Retained
focused/layout/metrics logs were checked; no tests or browser operations were
independently performed.

### Spec

Independent read-only review of the same diff against issue #16 and the user's
reported confusion: **0 actionable findings**. The reviewer confirmed the
preserved Go findings/limits, unchanged data/schema, actual answers/assisted
history and pending human recheck. Retained UI/replay logs matched the stated
scope; no tests or browser operations were independently performed.

Standards **0**; Spec **0**. A successful participant recheck is still required;
these checks do not establish that the reported readability problem is resolved.
The new wording has no fresh real-screen-reader or native-zoom check in this
pass. Earlier observations keep their dated scope. #16 and overall release
acceptance remain open.

## Participant follow-up: timeout versus measured zero

In the resumed review, P1 read Exchange 2's connection timeout text correctly
but then answered “0” when asked whether the step had a recorded zero or no
completed time. The exact response and assistance history are in
`issue-16-participant-review.md`. This is an observed comprehension miss, not
evidence that the API recorded zero. The UI now says “Did not finish within
2000 ms (no completed time)” for that timeout and explains that earlier
finished steps still have usable times. The HTTP response-start timeout uses
the same missing-time distinction. A measured 0 ms remains a completed value.
Go-owned findings, result outcomes/bounds, schema v1 and evidence are unchanged.

Verification for this wording change: the rendered-UI test failed first on the
old “No response within 2000 ms” label, then passed after the change. The
TypeScript/Vite build passed. Full `npm.cmd --prefix ui test` with a newly
built local Windows x64 executable passed **19 checks, 1 opt-in system-clipboard
skip**, including the three packaged browser journeys. Full root
`go test ./... -count=1`, `go vet ./...`, and `git diff --check` passed. The
verification executable is `.tools/issue-16/hopvista-missing-zero-20261027.exe`
(Go 1.27.2, `CGO_ENABLED=0`, source tree dirty for this verification build),
SHA-256 `3295CF16392FB95E0E20F087C27BC0C95A33363185B483106B4FBE37F6D84ADA`.
It is not the issue #17 release candidate. Isolated headless checks for timeout
and measured-zero captures passed **10 mode/width combinations** at the five
existing widths, with no layout failures and reduced motion showing zero
animations. Report:
`.impeccable/review/issue-16/observations-2026-10-28T04-00-06.551Z.json`.
Logs: `.tools/issue-16/missing-zero-{red,green,ui-test,go-test,observe}.log`.
No new real Chrome, screen-reader, or native zoom check is claimed. A human
recheck of the new wording remains outstanding, so #16 and release acceptance
remain open.

Independent read-only review of `git diff dc51a32...7e67338` found **0
Standards findings** and **0 Spec findings**. Standards checked repository
language/design guidance, the established rendered-browser test seam, data
preservation and the participant record. Spec checked the observed “0” answer
against issue #16's missing-versus-zero criterion, unchanged findings/schema,
and the explicitly pending human recheck. Reviewers inspected retained logs;
they did not run tests or control a browser.

## Participant follow-up: direct connection versus Chrome

P1 correctly said HopVista and Chrome need not show matching times but could
not give a reason after reading the original caveat or after a facilitator
example. The exact responses and assistance are in
`issue-16-participant-review.md`. The visible note now says HopVista opens its
own direct connection and explains that Chrome or another browser may reuse
an open connection, use a saved or different number address, or go through
another service (a proxy), so their times can differ. It does not claim an
actual Chrome route or timing. The approved screen composition and Go-owned
findings, measurements and schema v1 are unchanged. Human recheck remains
pending; #16 and release acceptance remain open.

The focused rendered-browser check failed on the old note and passed after
the edit. TypeScript/Vite build passed. Full `npm.cmd --prefix ui test` with
the new local verification executable passed **20 checks with 1 opt-in
system-clipboard skip**, including three packaged browser journeys. Root
`go test ./... -count=1`, `go vet ./...`, and `git diff --check` passed. The
verification executable `.tools/issue-16/hopvista-browser-path-20261027.exe`
was built from the dirty working tree for this wording check (Go 1.27.2,
Windows amd64, `CGO_ENABLED=0`), SHA-256
`F7E5DC856CD76F15D5F8528E62082A91EB11227AE69972DAF46608E35C51EBE2`.
It is not a release candidate. Two replay modes at five widths passed **10
isolated headless combinations** with no layout failure and zero animations
under reduced motion; report
`.impeccable/review/issue-16/observations-2026-10-28T04-10-24.961Z.json`.
Logs: `.tools/issue-16/browser-path-{red,green,ui-test,go-test,observe}.log`.
No new real Chrome, screen-reader or native zoom claim is made.

Independent read-only review of `git diff a152838...7693202` found **0
Standards findings** and **0 Spec findings**. Standards checked the repo's
design/language guidance, direct-path uncertainty, public browser test seam,
and factual participant record. Spec checked the issue #16 browser-difference
criterion, preserved Go/schema boundaries, and pending human recheck.
Reviewers inspected retained logs; they did not run tests or control a browser.

The next human check found another wording problem: P1 read “reuse an open
connection” as an open browser tab. The note now says a browser may keep a
website connection open from an earlier visit and use it again instead of
opening another. The exact response is in the participant record. This is a
clarification of the same direct-path caveat; it changes no diagnostic result,
schema, finding or approved layout. Human recheck of this phrasing remains
pending.

The focused rendered-browser check failed on the previous phrase and passed
with the concrete explanation. TypeScript/Vite build passed. Full UI tests
passed **20 checks with 1 opt-in system-clipboard skip**, including three
packaged browser journeys, using local verification executable
`.tools/issue-16/hopvista-browser-reuse-20261027.exe` (dirty source tree,
Go 1.27.2, Windows amd64, `CGO_ENABLED=0`), SHA-256
`5298F14254EB363934B53BCC93BC3EFA58049D4F83D900386EFDD273403A05FD`.
It is not a release candidate. Full root `go test ./... -count=1`, `go vet
./...`, and `git diff --check` passed. Ten isolated headless mode/width
combinations had no layout failure; report
`.impeccable/review/issue-16/observations-2026-10-28T04-14-57.381Z.json`.
Reduced-motion animations were zero. Logs:
`.tools/issue-16/browser-reuse-{red,green,ui-test,go-test,observe}.log`.
No new real Chrome, screen-reader or native zoom check is claimed.

Independent read-only review of `git diff 4e9b54d...02cc56f` found **0
Standards findings** and **0 Spec findings**. Standards checked repo
language/design guidance, preserved data boundaries, the rendered UI seam
and factual participant record. Spec checked the observed tab/connection
confusion against issue #16 and the pending human recheck. Reviewers inspected
retained logs but did not run tests or control a browser.

## Participant follow-up: dotted path connectors

P1 first described the dotted line as a connection “like a train station”,
then identified hop order after being directed to the note, but still said
the dots proved a direct connection between markers “cause it looks like
that”. The exact first/guided responses are in
`issue-16-participant-review.md`. The nearby path note now says in plain
language that the dots order numbered checks, HopVista did not measure a
direct connection between adjacent markers, and the website request may
have taken another route. The note uses normal body text instead of muted
text. Connector geometry and the approved path composition remain in place
pending human recheck; all trace outcomes and Go-owned findings are unchanged.

The focused rendered-browser check failed on the former technical note and
passed with the revised note. TypeScript/Vite build passed. Full UI tests
passed **21 checks with 1 opt-in system-clipboard skip**, including three
packaged browser journeys, using local verification executable
`.tools/issue-16/hopvista-path-dots-20261027.exe` (dirty source tree,
Go 1.27.2, Windows amd64, `CGO_ENABLED=0`), SHA-256
`0F2D4FE47E0B53C69AE395A61F1E4C4111F3A166197DD1D7E562C05E174D12C2`.
It is not a release candidate. Full root `go test ./... -count=1`, `go vet
./...`, and `git diff --check` passed. Ten isolated headless mode/width
combinations for `i13highsilent` and 64-hop content had no layout failure;
report `.impeccable/review/issue-16/observations-2026-10-28T04-28-34.126Z.json`.
Reduced-motion animations were zero. Logs:
`.tools/issue-16/path-dots-{red,green,ui-test,go-test,observe}.log`.
No new real Chrome, screen-reader or native zoom check is claimed. Human
comprehension of the revised note is still pending. #16 and release
acceptance remain open.

Independent read-only review of `git diff 7f5099e...11292fb` found **0
Standards findings** and **0 Spec findings**. Standards checked repo
language/design guidance, preserved Go/schema boundaries, the rendered UI
test seam and factual participant record. Spec checked the observed
physical-link inference and still-pending human recheck against issue #16.
Reviewers inspected retained logs but did not run tests or control a browser.

## Participant follow-up: high intermediate reply

In `i13highsilent`, P1 inferred Hop 2 was a bottleneck from its 450 ms reply
followed by Hop 3's 12 ms reply. After being directed to the existing row
note, P1 said they were unsure of the technical words. The actual responses
are in `issue-16-participant-review.md`. The static note now says the time
starts when HopVista sends a check and ends when it receives that device's
reply, not time between stops or proof the device made the website slow. The
recorded 450/12 ms replies,
missing reply, outcomes, schema and Go-owned findings remain unchanged.
This is a readability fix; human recheck remains pending.

The focused rendered-browser check failed on the former ICMP/RTT/bottleneck
copy and passed on the plain-language note. TypeScript/Vite build passed.
Full UI tests passed **22 checks with 1 opt-in system-clipboard skip**,
including three packaged browser journeys, using local verification
executable `.tools/issue-16/hopvista-hop-reply-20261027.exe` (dirty source
tree, Go 1.27.2, Windows amd64, `CGO_ENABLED=0`), SHA-256
`E67473FBFDEED3F9C165AB0947162F50E9B2B82FD63F546D96B001DFE0E65225`.
It is not a release candidate. Full root `go test ./... -count=1`, `go vet
./...`, and `git diff --check` passed. Ten isolated headless mode/width
combinations for `i13highsilent` and 64-hop content had no layout failure;
report `.impeccable/review/issue-16/observations-2026-10-28T04-40-26.942Z.json`.
Reduced-motion animations were zero. Logs:
`.tools/issue-16/hop-reply-{red,green,ui-test,go-test,observe}.log`.
No new real Chrome, screen-reader or native zoom check is claimed. #16 and
release acceptance remain open.

An independent Standards review of `2843428...900debc` found **1 wording
issue**: “how long this device took to answer” could be read as time spent
inside the device, while the recorded value spans sending the check to
receiving the reply. Spec review found **0 findings**. The displayed note and
rendered test were corrected to state both endpoints explicitly. The new
focused test failed on the first plain-language wording and passed on this
precise version. Human comprehension of this final wording remains pending.

For the precise version, TypeScript/Vite build passed; full UI tests passed
**22 checks with 1 opt-in system-clipboard skip**, including three packaged
browser journeys. Local verification executable
`.tools/issue-16/hopvista-hop-reply-precise-20261027.exe` was built from the
dirty source tree (Go 1.27.2, Windows amd64, `CGO_ENABLED=0`), SHA-256
`BD989C40E954767BCBDE40A81B573B12AF1166CA56D2BD18DA0B4E954008C403`.
It is not a release candidate. Full root `go test ./... -count=1`, `go vet
./...`, and `git diff --check` passed. Ten isolated headless path/width
combinations had no layout failure; report
`.impeccable/review/issue-16/observations-2026-10-28T04-43-05.778Z.json`.
Reduced-motion animations were zero. Logs:
`.tools/issue-16/hop-reply-precision-{red,green,ui-test,go-test,observe}.log`.

Independent read-only corrective review of `git diff 900debc...29421b7`
found **0 remaining Standards findings** and **0 Spec findings**. The
Standards reviewer confirmed the send-to-reply interval resolves the earlier
device-only timing ambiguity; Spec confirmed the unchanged data and
still-pending human comprehension check. Reviewers did not operate a browser.

## Participant follow-up: round-trip endpoints (2026-11-04)

P1 resumed the controlled `i13highsilent` reading and described Hop 2's
450 ms as “The time it takes to get from hopvista to my device”. This omits
the outward trip and suggests the row's “this device's reply” did not make
the endpoints clear. The exact answer, assistance and assessment are in
`issue-16-participant-review.md`. The row now says the timer starts when
HopVista sends a check from the user's computer toward this hop and stops
when its reply gets back to that computer. It explicitly says the time
includes both directions. The time-between-hops and website-slowdown limits
remain visible. Measurements, schema v1, Go-owned findings and the approved
composition are unchanged. The participant recheck is pending.

Verification: the focused rendered-browser check failed on the former note
and passed after the edit. TypeScript/Vite build passed. Full UI tests passed
**22 checks with 1 opt-in system-clipboard skip**, including three packaged
browser journeys. The clean-source verification executable is
`.tools/issue-16/hopvista-hop-roundtrip-20261104.exe`, source `b319d85`,
Go 1.27.2, Windows amd64, `CGO_ENABLED=0`, `vcs.modified=false`; SHA-256
`16397C663E50A80082205813B177C2AAD0F8A4E95D760CE45C37568F750E4CDF`.
It is not the issue #17 release candidate. Full root `go test ./...
-count=1`, `go vet ./...`, and `git diff --check` passed. Ten isolated
headless combinations for `i13highsilent` and 64-hop content had no layout
failure and zero reduced-motion animations; report
`.impeccable/review/issue-16/observations-2026-11-05T03-37-36.440Z.json`.
Logs: `.tools/issue-16/hop-roundtrip-{red,green,ui-test,go-test,observe}.log`.
No real Chrome, screen-reader or native zoom check is claimed for this edit.

Independent read-only review of `git diff 209b756...b319d85` found **0
Standards findings** and **0 Spec findings**. Both reviewers checked the
send-and-return meaning, unchanged result/schema boundary and honest pending
participant recheck. They did not edit or operate a browser.

## Participant follow-up: silent hop versus zero (2026-11-04)

On `i13highsilent`, P1 read the old Hop 1 text “No address or RTT observed”
but answered “Treat it as 0 ms” when asked what time to assign a silent hop.
The exact answer and assistance are in `issue-16-participant-review.md`.
The path row now states that no reply time was recorded, explains that a
0 ms reading would require an observed reply, and retains the time limit
and lack of a reply address. It uses “reply time” in the primary text
instead of “RTT”. Outcomes, bounds, schema v1 and Go-owned findings remain
unchanged. A participant recheck is pending.

Verification: the focused rendered-browser test failed on the prior
“No address or RTT observed” row and passed on the explicit no-time row.
TypeScript/Vite build passed. Full UI tests passed **23 checks with 1 opt-in
system-clipboard skip**, including three packaged browser journeys. The
clean-source verification executable is
`.tools/issue-16/hopvista-silent-zero-20261104.exe`, source `51e55a4`,
Go 1.27.2, Windows amd64, `CGO_ENABLED=0`, `vcs.modified=false`; SHA-256
`61582CD3E2DEDFC04AAC9525AC4EDAEAFC5C2CB3D46A63FB1CB5A787D3B8F07D`.
It is not the issue #17 release candidate. Full root `go test ./...
-count=1`, `go vet ./...`, and `git diff --check` passed. Ten isolated
headless path/width combinations had no layout failure and zero
reduced-motion animations; report
`.impeccable/review/issue-16/observations-2026-11-05T03-57-55.443Z.json`.
Logs: `.tools/issue-16/silent-zero-{red,green,ui-test,go-test,observe}.log`.
No real Chrome, screen-reader or native zoom check is claimed for this edit.

Independent read-only review of `git diff c95c956...51e55a4` found **0
Standards findings** and **0 Spec findings**. Both reviewers checked the
missing-versus-zero meaning, unchanged data/schema boundary, factual
participant record and pending human recheck. They did not edit or
operate a browser.

## Participant follow-up: aligned reply-time value

After the silent-hop copy change, P1 still answered “Probably just 0 ms”
when asked whether Hop 1 had a measured reply time. P1 then confirmed the
bold “No reply time recorded” text was visible. The exact responses and
assistance are in `issue-16-participant-review.md`. The issue is therefore
not explained by a stale preview asset. The path row now uses the same
**Reply time** field for an observed hop and a silent hop: the former shows
the recorded raw RTT, while the latter says **Not measured**. The no-reply
time limit, missing address and explanation of what 0 ms means remain
visible. Numeric RTTs, outcomes, bounds, evidence, schema v1 and Go-owned
findings are unchanged. A human recheck of the revised layout is pending.

Verification: the focused rendered-browser test failed on the former
unmatched “Raw RTT”/no-reply prose and passed after both rows exposed a
**Reply time** value. TypeScript/Vite build passed. Full UI tests passed
**23 checks with 1 opt-in system-clipboard skip**, including three packaged
browser journeys. The clean-source verification executable is
`.tools/issue-16/hopvista-aligned-reply-20261104.exe`, source `14fe376`,
Go 1.27.2, Windows amd64, `CGO_ENABLED=0`, `vcs.modified=false`; SHA-256
`6B9DE1E135164315ABC9F01F18962FCFBBCEE47EDEC88F29AE7C3D647B4C6FE5`.
It is not the issue #17 release candidate. Full root `go test ./...
-count=1`, `go vet ./...`, and `git diff --check` passed. Ten isolated
headless path/width combinations had no layout failure and zero
reduced-motion animations; report
`.impeccable/review/issue-16/observations-2026-11-05T04-08-13.165Z.json`.
Logs: `.tools/issue-16/aligned-reply-{red,green,ui-test,go-test,observe}.log`.
No real Chrome, screen-reader or native zoom check is claimed for this edit.

Independent read-only review of `git diff eb641c8...14fe376` found **0
Standards findings** and **0 Spec findings**. Reviewers checked the raw
RTT/missing distinction, unchanged data/schema boundary, truthful
participant scope, public rendered test seam and pending human recheck.
They did not edit or operate a browser.

P1 then correctly read the aligned values as **Not measured** for Hop 1 and
**450 ms** for Hop 2, an assisted missing-versus-measured recheck. P1 also
requested a colon after the field label. The visible label now reads
**Reply time:** in both hop states. The values, outcomes and evidence are
unchanged. The exact participant answer is in the participant review record.

The focused rendered-browser test failed before the colon was added and
passed afterward. TypeScript/Vite build passed. Full UI tests passed **20
checks with 4 opt-in skips** (three packaged browser journeys and system
clipboard); full root `go test ./... -count=1` and `git diff --check` passed.
The preceding aligned-value change had already passed the three packaged
browser journeys; they were not rerun for this punctuation-only follow-up.
Logs: `.tools/issue-16/reply-colon-{red,green,ui-test,go-test}.log`.
Independent read-only review of `git diff 43b4ace...1e442b7` found **0
Standards findings** and **0 Spec findings**. Both reviewers checked the
requested colon in both states, exact participant quote and unchanged data.

## Participant follow-up: unreachable reply versus silence

P1 correctly read `i13limit1` as a configured one-hop stop, then read
`i13unreachable` as “unreachable response” but said HopVista received no
answer. The exact responses and assistance are in
`issue-16-participant-review.md`. This capture records a reply at hop 2
from `192.0.2.253` (6 ms) reporting unreachable; it is not silence.
The trace outcome now has an adjacent plain explanation that HopVista
received a device reply and that the website request is a separate check.
The outcome, hop limit, evidence, Go-owned findings, schema v1 and approved
composition remain unchanged. Human recheck is pending.

Verification of the unreachable explanation: TypeScript/Vite build passed;
the focused rendered test failed before the explanation and passed afterward.
Full UI tests passed **24 checks with 1 system-clipboard opt-in skip**,
including three packaged browser journeys. Full root Go tests, vet and diff
check passed. The clean verification executable
`.tools/issue-16/hopvista-unreachable-note-20261104.exe` is source `7987807`,
Go 1.27.2 Windows amd64, cgo disabled, `vcs.modified=false`; SHA-256
`4C0CF2A424717F403E35329428C8D2635A30FC0F07B7F9BC67A5D790F4346BB6`.
It is not the release candidate. Ten isolated headless layout combinations
passed; report
`.impeccable/review/issue-16/observations-2026-11-05T04-23-11.279Z.json`.
Logs: `.tools/issue-16/unreachable-{red,green,ui-test,go-test,observe}.log`.

Independent review of `ab54d1b...7987807`: Spec **0 findings**; Standards
found **1 test-fixture inconsistency**. The completed unreachable hop lacked
`duration_ms`, and a later hop remained after the unreachable stop. The
fixture was corrected with `duration_ms: 0` and two hops only. The focused
rendered check passed afterward (1 pass, 21 pattern skips), log
`.tools/issue-16/unreachable-fixture-fix.log`. Production UI was unaffected
by this test-data correction. Human recheck of the new explanation remains
pending. Reviewers did not operate a browser.
The Standards reviewer confirmed the correction read-only: **0 remaining
actionable findings**.

Continuation point (2026-11-05): refresh the controlled preview and rerun
`https://fixture.example/i13unreachable`. Ask P1 to read the new explanation
beside Trace outcome and describe what reply HopVista received. Record the
actual response with prior coaching disclosed. The last participant answer
was that no answer was received, before this explanation was added. Remaining
human cases include cutoff versus completion, stopping/cleanup, redirect
policy/certificate stops, keyboard evidence/return/copy, and unfinished normal
timing/uncertainty interpretations. Consult the participant task matrix;
do not infer blanket acceptance from these partial assisted rechecks.

P1's 2026-11-05 recheck of the revised `i13unreachable` explanation correctly
identified a received device reply saying the target address was unreachable.
The exact answer, prior assistance and limited assessment are in
`issue-16-participant-review.md`. This resolves that specific comprehension
gap for an assisted recheck; the remaining human tasks and issue #16 acceptance
remain open.

In `i14pausedminute`, P1 said “Completed means that the process is done, but
maybe not all metrics have a value.” This supports the basic run-state versus
measurement-completeness distinction. Cutoff and cleanup interpretation remain
unchecked; see the participant record for the exact prompt and scope.

The subsequent cutoff question pointed P1 to a finding panel that does not
explain cutoff in `i14pausedminute`. P1 correctly described its incomplete
connection step and retained earlier timings, but said underway work could not
finish afterward. This was a facilitator prompt mismatch, so cutoff
comprehension is unassessed rather than failed. The exact answer and correction
are in the participant record.

In `i14cutoffstopping`, P1 answered “I think its stopping for a response to get
back to HopVista?” when asked what the **Stopping…** message meant. The old
description did not explicitly separate waiting for a reply from waiting for
started checks to end and clear up. The new text says a started check may
receive a reply or end without one and that another run can start when this
one finishes, without promising a cleanup time. The exact answer and prior
assistance are recorded in `issue-16-participant-review.md`; human recheck is
pending. A focused rendered-browser regression failed before the copy change
and passed after rebuilding the production UI.

P1 then read the revised `i14cutoffstopping` message as “Its waiting for checks
already started to end and clear up”. This is an assisted recheck pass for the
stopping wait. Reply versus no-reply, starting another run, and cleanup timing
were not established by that answer; the participant record preserves the
prompt, prior help and exact wording.

P1 then answered “I can start another run after this run finishes” in the
stopping replay. This is an assisted pass for the stated next-run timing. P1
did not identify an on-screen control or attempt a new run, and the static
capture does not demonstrate the live busy-slot transition. Those checks
remain open in the participant record.

In `i11sensitive`, P1 identified the HTTP 302 response and the redirect stop,
but did not explain the deliberate URL-policy reason or what remains unknown
about the redirected website. The exact first reading and assistance are in
`issue-16-participant-review.md`; that part of the human review is partial.

After being directed to the visible explanation, P1 said the redirect was
stopped because it selected a non-public IPv4 address on a different host.
This guided recheck matches the `i11sensitive` captured policy reason. It
does not yet establish what P1 concludes about the redirected website's
reachability or the distinction from a failed connection; see the participant
record for the exact answer and assistance.

Asked directly whether this result establishes the redirected website's
reachability, P1 answered “No it doesn't tell me”. This guided yes-or-no
recheck supports the finding's limit narrowly. A spontaneous distinction
between deliberate policy stop and failed connection remains unverified;
see the participant record for the exact prompt and context.

For `i11certificate`, P1 read that the run stopped during the TLS handshake
phase. This locates the recorded stop but does not yet establish understanding
of the completed connection, failed certificate verification, or absent final
HTTP response. The exact answer and prior assistance are in the participant
record; the certificate case remains partial.
