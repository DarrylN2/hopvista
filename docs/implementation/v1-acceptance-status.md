# HopVista v1 acceptance status

Latest candidate preparation: [issue #17 evidence](issue-17.md) identifies a
single Windows x64 executable from clean merged commit `a159f9c`, including the
Advanced layout fix. Packaged browser/API/privacy/security, native cancellation,
controlled paused-minute cutoff and same-selected-IP Windows route/tracert/pinned
HTTPS curl checks pass at the recorded condition. The record reconciles #15 and
#24 with their completed checks, including usable browser restart. Historical
matrix entries below retain their original dates and are superseded by these
linked follow-ups where applicable. **Candidate preparation is not release
acceptance: human comprehension remains deferred, #16 and #17 remain open, and
no release is published.**

Latest functional follow-up: [issue #24](issue-24.md) joins the current embedded
production UI to the real packaged backend in isolated headless browser journeys.
Form submission, redirect/error/partial results, evidence navigation, rejected
input/retry, cancellation, subsequent runs and actual process restart pass on
the recorded loopback Windows condition. These checks complement the earlier
mocked-API UI and process/API checks. Human comprehension review is **explicitly
deferred** at the user's request; issue #16 and overall release acceptance remain
open. The historical ordering below does not override this priority change.

Latest accessibility follow-up: [issue #16](issue-16.md) records keyboard and
clipboard fixes, actual NVDA speech checks, layout/zoom, reduced-motion and
contrast/target measurements against the verified #11–#14 scenarios. Independent
checks pass at that scope. **Real participant comprehension review is pending;
issue #16 and overall issue #10 release acceptance remain open.**

Latest lifecycle follow-up: [issue #14 evidence](issue-14.md) rechecks native
cancellation and cap/margin behavior, carries held-route stops through the
engine/API and current UI, and records the **failed natural packaged-minute
candidate**: its 64-hop trace completed in 38.35 seconds. A later test-only
process pause spanned the full, unchanged 60-second packaged window while a
real native echo was in flight; its pass and interruption history are recorded
in the same document. A final opt-in native check recorded raw echo return and
handle closure separately without changing schema version 1. Together with
the earlier cancellation, held-route and UI checks, **issue #14's stated
acceptance criteria are covered on the recorded condition**. The pause is
induced by the harness, so a naturally stalled native call, broader conditions
and issue #10 release acceptance remain open.

Latest path/echo follow-up: [issue #13 verification](issue-13.md) carries silent,
high-intermediate-RTT, unreachable, hop-limit and unstarted deadline/cancelled
observations through the real engine/API into the approved results view. The
record distinguishes simulated probes, real TLS/HTTP, browser checks and
remaining release gates.

Latest process-level follow-up: [packaged restart/privacy verification](packaged-restart-privacy.md)
checks actual executable relaunch, loss of prior run results, per-launch token
rotation, complete default stdout/stderr, and the isolated launch directory.
These checks pass on the recorded unelevated Windows x64 condition; overall
release acceptance remains open.

Latest controlled follow-up: [issue #12 verification](issue-12.md) adds incomplete
website-measurement cases through the real engine/local API and approved results
view. DNS/TCP/TLS/HTTP stops, retained partial evidence, absent versus measured
zero values, completed 404/503 responses, dominant-wait gaps and evidence
navigation now have recorded checks. Full Go tests/vet, UI build, loopback TLS
and packaged smoke pass; independent Standards and Spec reviews found no issues.
Simulations and remaining human/platform limits are explicit in that record.
This follow-up does not declare release acceptance or replace the historical
matrix below.

Reviewed 2026-10-15 against [GitHub issue #10](https://github.com/DarrylN2/hopvista/issues/10), its linked [acceptance plan](https://github.com/DarrylN2/hopvista/issues/7#issuecomment-6051380625), the current workspace, and milestone notes through [milestone 19](milestone-19.md). The IPv4-first and no-hard-native-drain revisions in issue #10 supersede the older dual-stack and hard-deadline wording in issue #7. The current issue specifies a 30-hop default and an advanced range of 1–64; [milestone 12](milestone-12.md) corrected the earlier 16-hop implementation.

**Release acceptance: open.** The milestone 1 *development gate* passed under its approved tested-condition scope; it was not release acceptance. Current test evidence comes principally from one Windows 11 25H2 build 26200 x64 machine with an unelevated token and its current network. Keep any distribution claim experimental and tied to that condition until wider coverage is obtained. The 250 ms native start margin is empirical, and no hard terminal time or native drain bound is claimed.

| Issue #10 acceptance area | Current evidence | Status / remaining work |
| --- | --- | --- |
| Versioned results, normalized probe seam, orchestration and findings | Schema fixtures and evidence-reference checks in `internal/result`; fake route/echo, engine progress, HTTP–ICMP disagreement, finding-rule edges, and lifecycle tests. See [milestones 2–9](milestone-2.md) and [11](milestone-11.md). | **Controlled path/partial-outcome gaps covered at the recorded scope.** [Issue #13](issue-13.md) extends the durable high-RTT fixture and engine/API/UI checks for silence, lower later RTTs, HTTP/ICMP disagreement, 1/30/64-hop limits and distinct terminal outcomes. [Issue #12](issue-12.md) covers partial website phases. Existing fixtures do not establish human UI comprehension. |
| URL and redirect policy | `internal/target` tests cover IPv4 scopes, first usable A selection, mixed A/AAAA, AAAA-only, changed answers, sensitive redirects, IPv6 and mapped-literal rejection, and invalid URLs. `internal/httpmeasure` checks stopped redirects make no second dial; see [milestones 4–5](milestone-4.md). | **Controlled policy checks pass.** [Issue #11](issue-11.md) completed the redirected HTTPS fixture with pinned Host/SNI/certificate and per-exchange timing evidence; [issue #13](issue-13.md) rechecks attempted-host path identity through the current view. |
| Direct pinned HTTP/TLS | Controlled HTTP/TLS fixtures cover actual peer, fresh connections, headers-only first byte, wrong-name/untrusted/IP-literal certificate behavior, SNI, 503, connect and header failure. Packaged loopback HTTP and one public HTTPS request agreed with proxy-disabled `curl --resolve`; see [milestones 5](milestone-5.md), [11](milestone-11.md), and [18](milestone-18.md). | **Strong current-condition evidence.** The public HTTPS comparison is one host/address/network, not browser-path equivalence or general availability. [Issue #11](issue-11.md) covers the controlled redirected HTTPS gap; [issue #12](issue-12.md) covers incomplete website phases. |
| Windows route, ICMP and trace | Production adapter layout/status tests, live loopback, a remote first-hop `tracert` semantic comparison, and matching source/interface/next-hop with `Find-NetRoute`; see [milestones 6](milestone-6.md), [12](milestone-12.md), and [13](milestone-13.md). Silent and unreachable cases are also covered by fake fixtures and the milestone 1 live unreachable capture. | **Current-condition IPv4 evidence.** Later-hop identities and RTTs may vary. A genuinely stalled native `GetBestRoute2`, another account/build, alternate networks/VPN, and system-wide stress have not been observed; retain the experimental scope and no-bound wording. |
| Cancellation, cutoff and native ownership | Fake held-route/echo tests, live pending production echo cancellation, packaged `stopping`/busy-slot/evidence retention, unloaded and process-loaded cap-plus-margin samples, and a real 60-second manager cutoff under process load. [Issue #14](issue-14.md) adds a packaged 60-second cutoff with a test-only whole-process pause while real native echo was in flight, plus separate native return/handle-close observer timestamps. See [milestones 14–17](milestone-14.md) and [19](milestone-19.md). | **Issue #14 checked at the named condition; partial release evidence.** The process pause is an induced scheduling delay, and the held-route executor is simulated. No natural native call spanning cutoff or genuinely stalled route has been demonstrated. Earlier failed loaded boundary evidence remains failed. Do not infer a latest-return bound. |
| Local API security, packaging and privacy | `cmd/hopvista` binds `127.0.0.1`; API tests and the isolated executable check cover exact Host/Origin/token checks, no CORS, IPv6-literal rejection, busy/restart behavior, and embedded assets without Go/Node on the child `PATH`. The server logs its loopback URL, not submitted URLs or result bodies. See [milestones 3](milestone-3.md), [10](milestone-10.md), [12](milestone-12.md), and [15](milestone-15.md). | **Current-condition boundary and restart/privacy checks pass.** Actual process relaunch loses completed/incomplete results, rotates the token and rejects the old token; default stdout is empty and stderr contains only the listener announcement. See [packaged restart/privacy verification](packaged-restart-privacy.md) for API/output/directory evidence, source review and limits. No OS-wide persistence trace, graceful-console-signal, installer or signing claim is made. |
| User-visible IPv4 path and answer | The first minimal React/TypeScript Vite view now embeds in the Go executable. It shows a path sequence, selected IP and scope, adjacent terminal observation and limit, separate per-exchange phase outcomes, per-hop text, gateway/target echo, and links to technical evidence. See [milestone 20](milestone-20.md). | **Functional first pass, release review open.** Inspect the rendered desktop/narrow layouts, keyboard flow, evidence navigation, and normal/edge-case comprehension with a human. The local in-app browser could not reach the app's loopback listener during this pass, so no visual browser review is claimed. |
| Human review and supported-condition claim | No documented human comprehension/accessibility review of normal and edge-case views. Broader Windows 11 accounts/builds/networks and a genuine native route stall remain unverified. | **Release blocker for UI review.** Run the human review after the required view exists. Wider platform coverage can be recorded as a limit under issue #10's experimental Windows 11 x64 tested-condition wording; do not call it comprehensive Windows 11 support. Resolve any explicit release-scope decision before claiming more. |

## Next work in order

1. Controlled redirected HTTPS, partial website phases and difficult path/echo gaps now have the recorded follow-ups in [issue #11](issue-11.md), [issue #12](issue-12.md) and [issue #13](issue-13.md). Broader lifecycle/native-condition and human review gates remain open.
2. Packaged-process restart and default-output checks completed on 2026-10-18 at the [recorded current-condition scope](packaged-restart-privacy.md). Existing local security boundary tests remain regression checks.
3. Review the minimal path-first UI on desktop and narrow screens, then run keyboard, text-label, evidence-inspection, and human comprehension review against normal and edge-case fixtures. Refine visual design when resources arrive.
4. Recheck on additional Windows 11 account/build/network conditions when available, preserving the exact condition and any failures. Continue to label `GetBestRoute2` native-stall behavior unverified unless it is actually observed; the simulated held-route checks establish lifecycle ordering only.

This matrix records evidence scope; it does not change issue #10 or declare release acceptance. Historical milestone notes remain as observations from their respective implementation stages.

## Redirected HTTPS follow-up: 2026-10-17

[Issue #11 evidence](issue-11.md) closes the controlled redirected-HTTPS gap in
the URL-policy and direct HTTP/TLS rows above. Relative/same-host and allowed
new-host journeys, per-exchange phase attribution and first-byte semantics,
policy stops, a redirected certificate failure, API/engine result retention and
final/attempted path identity now have repeatable checks. New-host public pins
are simulated socket/DNS addresses with real TLS/HTTP; a separate loopback check
uses real TCP and normal TLS verification. The current production view was
checked with all six API-captured results at desktop and narrow widths, including
evidence focus/return. The approved UI is preserved. Go now explains policy stops
and retains the validated TCP peer if TLS verification fails. The existing
single-exchange TLS/HTTP checks, full Go suite/vet, UI build and packaged smoke
check pass. Other functional fixtures, privacy/restart review and human/release
acceptance listed above remain open; the original ordered next-work list is
historical, and its redirected-HTTPS item is now complete at this controlled scope.

## UI design follow-up: 2026-10-17

The user-approved Clear field notes with guided reading direction has been implemented in React/Vite and plain CSS. See the [approved brief and verification record](../design/diagnostic-results.md). Production build, Go tests/vet and the final packaged loopback smoke check passed. Starting the preview outside the sandbox resolved the earlier browser access restriction. Live loopback results and explicitly labeled synthetic edge-state fixtures were inspected; an independent representative visual review requested no material fixes. DOM checks confirmed narrow ordering, evidence focus/return, stopping controls and no horizontal overflow across the recorded state/width matrix. Full accessibility, 200% zoom, human comprehension, and the other release gates above remain open; this follow-up does not declare release acceptance.

The subsequent user-requested crisp graphic refinement strengthens typography, blue accents, hop markers and phase-specific evidence labels, tightens introductory copy and adds static finding explanations while preserving recorded findings and limitations. The production build, Go checks and packaged smoke check passed again. Seventeen synthetic state/width combinations had no horizontal overflow; representative independent visual review and a final evidence-label review found no material fixes. The linked verification record retains exact scope and outstanding acceptance work.

The user then requested motion, color and delight. The current pass adds bounded evidence/state acknowledgments, teal timing-category accents and an accessible Copy IPv4 action while preserving measurement semantics. Production build, Go checks, packaged smoke and a real packaged loopback interaction check passed. Seventeen corrected, measured viewport checks had no horizontal overflow; copy value/keyboard/reset and exact evidence focus/return were exercised. The linked record distinguishes simulated clipboard failure and source-checked reduced motion from unavailable native-permission and reduced-preference acceptance. This pass has direct browser verification; independent review of the previous graphic pass is not extended to these later changes.
