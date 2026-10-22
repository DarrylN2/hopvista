# Diagnostic results

Status: design brief approved by the user on 2026-10-17; first implementation completed, limited rendered verification performed, human acceptance outstanding.

Accessibility follow-up: [issue #16 verification](../implementation/issue-16.md)
records actual NVDA, native 200% zoom and runtime reduced-motion checks, plus
keyboard focus, manual-copy fallback and input-boundary contrast fixes. The
approved composition and Go-owned findings are preserved. Participant review
and overall release acceptance remain open.

Mode: Operate. Direction: Clear field notes with guided reading.

## Approved brief

Help curious non-specialists locate the observation, its limits, website timings, and supporting IPv4 path. Redesign the original functional interface freely while preserving issue #10's product truth.

Use a light neutral canvas, clear ink typography, restrained blue actions, generous spacing, and fine dividing lines. The path is the main desktop area on the left; the observation, limit, selected IPv4/scope, and website timing are adjacent on the right. Narrow screens put summary and timing before the vertically ordered path.

Show measured addresses, ICMP statuses, and raw RTTs without inferring physical links, HTTP routes, link delay, or bottlenecks. Label connectors as probe order only. Keep gateway, target echo, and per-exchange website timings distinct. Retain the exact direct-path caveat and Go-provided finding wording. Missing measurements are explicit outcomes, never numeric zero substitutes.

Findings remain pending during running and stopping. Stopping has no ETA and keeps the run slot occupied. Completed, cancelled, and failed runs retain evidence. Measurement incompleteness is separate from lifecycle. API errors identify refresh uncertainty rather than changing the last recorded state.

Use contextual evidence links to expand, focus, and highlight exact records, with return navigation. Retain raw evidence and full JSON. Keep keyboard controls, visible focus, text statuses, coarse-pointer targets and reduced-motion support. Keep React/TypeScript, Vite and plain CSS.

## Implementation verification

- Production TypeScript/Vite build passed and regenerated embedded Go assets.
- Root Go tests and vet passed; optional packaged check skips in ordinary test runs.
- Final rebuilt executable passed the opt-in packaged loopback smoke check with socket access outside the sandbox. Its first sandboxed attempt failed with a socket access restriction; the successful retry is not evidence of visual verification.
- Independent source review found an in-progress IPv4 selection incorrectly described as absent. The fix was reviewed as resolved; missing active selection now says Pending.
- Initial browser attempts timed out. Starting the final executable outside the sandbox resolved local access. A real loopback run was inspected at 1440 and 390 pixels; evidence navigation focused the exact record and restored focus to its source.
- A separate loopback-only review server served explicitly labeled synthetic results from existing result fixtures and UI-only state variants. Running, stopping, completed, partial, cancelled, failed, unsupported IPv6, and a 64-hop long-URL/high-intermediate-RTT result were captured at 1440 and 390 pixels; the long result was also captured at 320 pixels. All 17 combinations had no document-wide horizontal overflow. Running/stopping findings remained pending; stopping disabled both start and cancel controls. These variants exercise presentation, not native lifecycle correctness.
- DOM layout checks at 960, 760, 390 and 320 pixels confirmed desktop side-by-side and narrow summary/timings/path order without horizontal overflow. Keyboard Enter on Return to measurement restored focus to the source evidence button. Full keyboard/screen-reader and human comprehension review remain outstanding.
- Review screenshots are generated local artifacts under .impeccable/review/ and are ignored by Git. No whole-release acceptance is claimed.
- Independent visual disposition: ship at the representative scope inspected. The reviewer inspected actual completed loopback desktop/mobile, synthetic completed desktop/mobile, stopping mobile, partial desktop and failed mobile captures. No material visual fixes were requested. This does not cover uninspected captures, 200% zoom, full accessibility or human comprehension acceptance.
- Impeccable context launcher could not install its engine in the restricted cache; existing project context was read directly. No detector ran.

## Remaining rendered review

Inspect at 1440, 960, 760, 390 and 320 CSS pixels and at 200% zoom. Check no page-wide overflow, long URLs/evidence IDs, 1/30/64-hop sequences, explicit focus/return and opened disclosures. Review running, stopping, completed, partial, cancelled, run failure, API refresh failure and empty states. Exercise silent/high-RTT hops, HTTP–ICMP disagreement, on-link routes, redirects including policy stops, HTTP 4xx/5xx, missing IPv4, unsupported IPv6, and unavailable/cancelled/deadline timing outcomes.

Human review must confirm readers can explain phase meanings, cutoff versus terminal completion, hop limit versus unreachable response, missing values versus zero, and the limits of ICMP and direct-path measurements. This remains part of release acceptance, not an inferred consequence of passing builds.

## Crisp graphic refinement

The user approved and requested bolder, typeset, clarify and polish passes after critique, selecting crisp graphic character plus tighter introductory copy and clearer finding explanations. The approved desktop/narrow composition remains intact.

The implementation strengthens blue actions, the wordmark, section rules and numbered hop markers. Fixed heading roles contrast with supporting prose; addresses and recorded numeric values use a system monospace stack. Missing/timeout outcomes remain sans-serif text. The entry copy is shorter, and the empty state explains path versus website timing without fabricated samples. The intro compacts when a run begins.

Go-provided finding observations and limits remain verbatim. Static presentation guidance explains finding categories, and a disclosure defines the existing dominant-wait rule without recalculating a finding. Timing subtitles explain phase purpose. Finding evidence links name the exact associated phase when available; original IDs remain in accessible names, titles, focus targets and technical evidence.

Verification: final production build, Go tests/vet, and the packaged loopback smoke check passed. A real local run exercised the new response explanation. Seventeen synthetic scenario/width checks again found no document-wide horizontal overflow, including a 64-hop long result at 320px; active findings stayed pending. Layout checks at 960/760/390/320px confirmed the approved reading order. Both ordinary and phase-named evidence links focused the exact record and restored source focus with Enter. Sampled new text/background contrast ranged from 5.34:1 to 11.44:1, with the primary button at 6.83:1. No browser console errors were observed in the exercised fixture flow.

An independent reviewer inspected completed desktop/narrow, stopping narrow, partial desktop and failed narrow screenshots and returned ship at that scope. A final verdict pass reviewed recaptured desktop/narrow phase-named evidence links and found no actionable issue. Captures live under the ignored .impeccable/review/graphic/ directory. Typography detector invocation was attempted but its missing engine could not load; no automated clean-scan claim is made. Keyboard zoom shortcuts produced no observable zoom change in this browser connection, so 200% zoom acceptance remains unverified. Full human comprehension and screen-reader acceptance remain open.

## Motion, color and interaction feedback

The user requested animate, colorize and delight. Blue now groups path/evidence, teal groups website timings, and amber continues to identify stopping/cancelled states. Color does not classify a hop or RTT as healthy, faulty or causal. Finding supporting text uses blue-tinted ink; selected evidence has a persistent blue wash.

The focal motion acknowledges evidence selection with a single 520ms wash. Directional arrows transition over 160ms on hover/focus. Newest-hop markers, copy confirmation and the running-state dot use single 220–240ms marks. No loops, fake progress, delayed content or measurement animation were added. Reduced-motion CSS removes animations, transitions and arrow displacement while retaining static feedback; the CSSOM rule was inspected, but this browser exposes no runtime media-preference emulation.

A Copy IPv4 button appears only when an address is selected. Keyboard copy wrote the exact displayed IP and announced success; the acknowledgement resets after 2400ms. A synthetic rejecting clipboard exercised the visible manual-copy fallback. The permission-policy header was not reliably enforced, and replay of the synthetic denial was inconsistent in the browser bridge; native permission-denial behavior is not accepted from that fixture. A copy failure does not change run state or measurements.

Seventeen measured scenario/width combinations had no horizontal overflow at 1440/390px plus the 64-hop long result at 320px. Running/stopping findings remained pending and unsupported IPv6 showed no copy action. Early viewport attempts affected the wrong tab and were discarded; checks.json under .impeccable/review/motion/ contains the corrected measured widths. Evidence selection still focused the exact record and Enter returned to its source. Sampled text/background contrasts for new colored roles were 5.60:1–6.83:1. Normal fixture flow had no observed console errors. Build, Go checks and packaged loopback smoke passed. The final packaged app also completed a real loopback diagnostic run, copied its selected IP via Enter and reset its Copied label after 2400ms. This pass was checked directly in-browser, not by another independent review; previous independent graphic-pass review remains historical.
