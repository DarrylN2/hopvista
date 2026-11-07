# Results clarity

Approved in chat on 2026-11-06 (America/Los_Angeles). Implementation baseline:
`f563919faa55438a6428eb580abd7be7a8786677`. Mode: Operate. This brief refines
the [existing results design](diagnostic-results.md) for networking beginners.

## Approved direction

- Preserve the current Clear field notes appearance, desktop path-left layout,
  narrow-screen finding/timing/path order, measurement values and Go-owned
  findings. Do not calculate a diagnosis in React or introduce a health score.
- Put the observation and its explicit limitation before address information.
  Remove supplementary text that repeats those statements. Retain relevant
  incomplete-step, redirect-policy and cutoff explanations.
- Lead timing rows with Address lookup, Connection, Encryption and Response
  begins; retain protocol names secondarily. Give each row a short definition
  and keep the exclusion of page download/display visible. Keep redirected
  website requests and their addresses, outcomes and timings separate.
- Use native expandable details for browser examples, address context,
  longest-wait thresholds, measurement mechanics, individual gateway/target
  samples and reply-time variation. Keep counts and the numbered path visible.
  Preserve raw records and JSON, exact evidence navigation and return focus.
- Explain round-trip reply timing once beside the path. Keep the connector
  warning, aligned Not measured values and timeout bounds visible. Distinguish
  recorded zero, silence, received unreachable responses and hop-limit stops.
  Preserve pending findings, stopping/cleanup wording and partial evidence.

The scope is presentation and comprehension, with no backend/schema changes,
merge, tag, upload or release. The previously accepted experimental executable
and its checksum remain historical; this branch does not inherit its acceptance.

## Participant lessons retained

The [participant record](../implementation/issue-16-participant-review.md) and
[final audit](../implementation/issue-16-acceptance-audit.md) document a passed
assisted reading, earlier misunderstandings and facilitator assistance. Missing
times were repeatedly read as zero; dotted markers were read as direct links;
connection reuse was confused with an open tab; phase meanings and the finding's
uncertainty needed clarification. The inline labels and warnings above preserve
those lessons. Browser examples still explicitly describe reusing a website
connection from an earlier visit. These UI changes do not establish a new human
comprehension pass or invalidate the historical scoped acceptance.

## Verification

The rendered production UI is the existing public regression seam, using
controlled schema-1 API snapshots. Reading-order and plain phase-label assertions
failed against the previous presentation and passed after the changes. Existing
checks cover keyboard evidence selection/return, unchanged observation/limit
strings, timeout bounds, missing values and stopping. Additional checks cover
recorded zero, visible echo counts, keyboard sample disclosures, and return to an
evidence source inside a disclosure closed while inspecting evidence.

Rebuilt production assets were checked against 13 existing captured scenarios:
normal redirects, connection timeout, measured zero, HTTP 503, no IPv4,
certificate failure, redirect-policy stop, high/silent replies, one-hop limit,
unreachable response, cutoff/stopping, terminal cutoff and a 64-hop long-URL/ID
stress variant. At 1440, 960, 760, 390 and 320 CSS pixels, with details closed
and open, all 130 combinations retained reading order without page-wide
overflow. One additional 200% CSS-zoom reflow check passed; this is not a new
native browser-zoom acceptance check. No page JavaScript errors were observed.

Desktop/narrow normal and high/silent-result screenshots were inspected directly;
the stopping captures were also retained. Local screenshots and measurements are
under `.impeccable/review/results-clarity/` (ignored artifacts). Captured API
replays are presentation evidence, not fresh network measurements.

Impeccable context loading was attempted once in this chat but its engine could
not install in the restricted cache. PRODUCT.md, DESIGN.md and the surface brief
were read directly, along with the clarify, distill and craft-floor guidance.
No detector result is claimed. The sandbox also prevented the first headless
Edge launch; the isolated-browser checks ran successfully outside that sandbox.
No personal browser, system clipboard or screen reader was operated.

Final validation: TypeScript/Vite build, root `go test ./...`, `go vet ./...`
and `git diff --check` passed. The full UI suite passed 27 tests with zero
failures and one explicit system-clipboard opt-in skip. All three packaged
loopback journeys were enabled against a newly built executable (local QA
artifact, not a release), SHA-256
`04B4C77507B0055B9BC9C4BB72F5845FC0761B7E4C686856C0DA1281B3C72E1D`.
They exercised real start/poll/cancel/restart APIs, redirected and incomplete
requests, retained findings, and evidence navigation. The first combined run
found an old assertion expecting “Actual peer”; updating it for “Connected to”
retained the same selected/actual-address checks, and the full rerun passed.
