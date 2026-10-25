# Issue #24: packaged UI functional scope

Status: scope and public test seams approved by the user on 2026-10-21
(America/Los_Angeles), published as [issue #24](https://github.com/DarrylN2/hopvista/issues/24).
Parent specification: [issue #10](https://github.com/DarrylN2/hopvista/issues/10).

## Why this milestone

Inspection baseline: clean `codex/accessibility-human-acceptance` at
`01692fa213d7a05214a66aa198911c8b2790200b`.

Issues #11–#14 are closed with controlled and packaged verification records.
Issue #15 remains open, although `packaged-restart-privacy.md` records extensive
process/API/output checks. Its browser clean-session usability criterion is
related to this work. Issue #17 covers the broader release candidate and is
blocked by issue #16; it is not an independently scoped functional ticket.

`tools/accessibility-ui.test.mjs` serves production assets with intercepted API
responses. `tools/app-smoke` exercises a real isolated executable through HTTP
and checks asset delivery, native probes, cancellation and restart. Earlier
records also describe interactive packaged loopback runs. The remaining gap is
a repeatable automated browser journey joining the current embedded UI to the
real packaged backend without replacing its session/start/poll/cancel responses.
No production defect is established by this inspection.

## Functional outcome

A person can start a diagnostic run in the packaged UI, inspect its actual
result, recover from rejected input, cancel active measurement work, start
another run after completion, and use a fresh session after process restart.

## Acceptance criteria

- Build the Windows x64 executable from the current approved UI and record the
  source commit and checksum. Launch only disposable child processes from
  isolated temporary directories, with Go/Node absent from the child PATH.
- Use an isolated headless Playwright browser context and the executable's
  embedded production UI. Do not intercept or fabricate application API
  responses. Use only controlled loopback target servers; no public target or
  personal browser profile is required.
- Submit an HTTP target through the form and advanced hop setting. Observe real
  start/poll traffic and terminal completion; compare the rendered selected IPv4,
  hop setting, HTTP outcomes and Go finding with the actual schema-1 API result.
  Open an evidence link and verify it identifies the actual referenced record.
- Exercise a relative redirect and HTTP error response as recorded measurements,
  plus a controlled target connection failure with retained partial evidence.
  Verify an unsupported IPv6 literal produces a visible rejection and permits
  a subsequent valid run.
- Cancel a run while a controlled loopback HTTP server holds response headers.
  Verify the real cancellation request, terminal cancelled state, retained
  completed evidence, and ability to start another run afterward. Observe
  stopping if delivered; do not require an ephemeral stopping snapshot or claim
  that this fixture proves native drain or the full 60-second cutoff.
- Stop and relaunch the isolated executable. Navigate the disposable browser
  context to the new listener, verify a usable empty UI and successful new run,
  and verify the preceding run ID is unavailable before starting that run.
  Reuse existing token-rotation/privacy checks rather than duplicating them.
- Fail on unexpected browser errors or unexplained API failures. Ensure child
  processes, browser contexts and loopback servers are cleaned up on failures.
  Provide one documented opt-in command and record actual outcomes/skips.

## Approved public test seams

1. The packaged application's browser surface: accessible form/actions,
   rendered run state, measurements, finding and evidence navigation, backed
   by unmodified same-origin API requests.
2. The packaged process's public HTTP API across an actual restart: compare
   schema-1 evidence with the UI and check the old run is absent.

Controlled target servers supply responses and a held-header condition;
production engine, manager, security boundary and Windows adapter remain real.
No private React state or internal Go collaborator assertions are proposed.

## Implementation and review

Use incremental TDD at the agreed seams for demonstrated product defects.
Integration checks may pass initially; record that honestly rather than
inventing a product regression. Run focused checks regularly, TypeScript
typechecking, build, vet and the full test suite once at the end. Enable the
applicable packaged regression checks explicitly.

Approved review fixed point: `01692fa213d7a05214a66aa198911c8b2790200b`.
Complete the implement skill's independent Standards and Spec reviews and
commit on the current branch, preserving the existing issue #16 work.

## Boundaries

Preserve the approved UI, schema version 1, Go-owned findings, empirical native
margin and stopping/cleanup contract. Fix only demonstrated functional defects.
Do not resume participant questions, wording reviews, screen-reader operation,
desktop control or use of an existing browser. Human comprehension remains
explicitly deferred. Issue #16 and overall release acceptance remain open.
This milestone does not publish a release, complete issue #17, or establish
broader Windows/native-stall coverage.
