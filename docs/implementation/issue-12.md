# Issue 12: incomplete website measurements

Implemented on 2026-10-18 (America/Los_Angeles) for
[issue #12](https://github.com/DarrylN2/hopvista/issues/12), which had no comments
when read. Branch `codex/incomplete-website-measurements` starts at freshly fetched
`origin/main`, `b17ea9f129e508f9fc1f977cc8945d2d6f339d25`, the merged PR #18 commit.
The starting workspace was clean. The approved UI composition, React/TypeScript,
Vite, plain CSS, schema version 1, IPv4 pinning and Go-owned findings are preserved.
There are no ADR files in this repository. This extends the
[issue #11 record](issue-11.md), [acceptance matrix](v1-acceptance-status.md), and
[approved design verification](../design/diagnostic-results.md). **Release
acceptance remains open.**

## Demonstrated defects and fixes

Each production change followed a failing test at an approved seam, or a browser
reproduction followed by the same check against the rebuilt production assets:

- The existing `Client.Measure` connect-failure test now includes timeout. It
  failed because unattempted HTTP was marked error/timeout and a TCP timeout's
  journey code was `tcp_connect_failed`. Failed TCP/TLS now own the stop; later
  unattempted phases remain unavailable, without durations or invented starts.
- The existing target-selection table failed for `net.DNSError{IsTimeout:true}`.
  Resolver-reported network timeouts now produce `dns_timeout`, as do expired
  lookup contexts. A separate cooperative resolver waits for the real two-second
  selection bound and verifies no numeric duration or selected IP is invented.
- New cases through the real engine, manager and local API first failed because
  DNS/TCP/TLS stops had generic explanations. Go now supplies specific phase,
  bound, no-A, cancellation and cutoff explanations to the existing view.
- A failed lookup after a completed redirect was described as a URL-policy stop
  and linked only to the earlier HTTP response. It now leads with IPv4 selection
  for the attempted hostname and links to the failed DNS record. Earlier DNS,
  response and attempted-host path evidence remain available.
- A measured first header byte followed by stalled remaining headers incorrectly
  produced a completed journey and no finding. The first-byte measurement stays
  completed with its actual duration, while the journey is timeout/error and Go
  explains the incomplete remaining headers. No status or final host is invented.
- The shortened-cutoff HTTP case first failed because `deadline` was labelled
  `response_header_timeout`. Run-context cancellation/cutoff now has precedence in
  the HTTP error code; it stays distinct from a per-phase timeout.
- Browser replay of an HTTP timeout with successful path probes lacked the
  incomplete-measurements sentence. Its condition now includes timeout and the
  IPv4 family outcome. A completed no-A lookup also gets the notice; a completed
  HTTP 404/503 response with successful path probes does not. Only this condition
  changes in UI source; no layout, CSS or finding computation changes were made.

## Coverage and scope

| Scenario | Repeatable evidence |
| --- | --- |
| DNS failure, reported timeout and no usable A | `TestIncompleteMeasurementsThroughAPI` asserts explicit family outcome/code/message, DNS bound, missing duration on failure, completed no-A lookup, and zero dial/probe calls. The no-A answers include IPv6, mapped IPv6 and an unusable IPv4; no IPv6 fallback or fabricated IPv4 exchange occurs. |
| Actual DNS bound expiry | `TestLookupBoundExpiresWithoutNumericTiming` waits for the selection context to expire. Existing selection cache, mixed-answer and address-scope checks remain in `internal/target`. |
| IPv6 and mapped-literal validation | The existing pre-exchange cancellation test is extended as `TestRejectedOrCancelledTargetMakesNoNetworkCalls`: both literal syntaxes reject before DNS/dial. Existing `TestURLPolicy`, local API `TestValidationAndMethods`, redirect policy and packaged smoke checks are reused. |
| TCP failure and timeout | Extended existing `TestConnectFailureHasNoNumericTiming`; API cases retain the first completed HTTPS redirect and DNS pin, attribute the stop to the second TCP phase, preserve its bound/error/timestamps and leave TLS/HTTP unavailable. |
| TLS failure, certificate failure and timeout | API cases use a closed pipe or a server that consumes ClientHello without replying; the real transport enforces its two-second TLS bound. Issue #11's wrong-name redirected certificate case is reused, along with the existing real-socket wrong-name, untrusted and IP-literal certificate checks. TCP peer and completed phases survive. |
| No first response-header byte | A real TLS/HTTP fixture reads the request and withholds all response bytes until the transport's three-second bound closes it. HTTP has no duration/status; earlier DNS, TCP, TLS and redirect evidence survive. |
| Partial response headers | `TestFirstByteRetainedWhenHeadersTimeoutThroughAPI` sends one header byte and delays the rest beyond the three-second header bound. First-byte duration remains measured, including zero when recorded by the platform clock; family timeout, absent status/final host and the Go finding stay explicit. |
| Cancellation and cutoff | `TestInterruptedHTTPThroughAPI` holds the second HTTP response, cancels through the local API or uses a 300 ms manager window. Earlier phases/exchange survive; cancelled/deadline have no invented HTTP duration. Running/stopping have no findings. These are controlled lifecycle checks, not a native-drain or production 60-second acceptance claim. |
| Completed HTTP 404/503 | `TestCompletedErrorResponsesThroughAPI` uses real TLS/HTTP and asserts completed timings, statuses, no family error and the Go response observation. Existing direct 503 coverage is retained. |
| Dominant measured wait | `TestDominantWaitThroughAPI` adds ties, missing comparable phases/durations, measured zero, exclusion of unavailable phases and failure before ranking. An earlier 50,000 ms exchange cannot influence the final exchange. Finding IDs resolve to the compared/stopped phases. Existing `TestDominantWaitRuleEdges` retains the 99/100 ms and exact/below-2x boundaries. |

The API cases reuse issue #11's TLS journey, capture writer, fake path adapter
and API session/start/status helper. DNS answers, socket peer addresses, TCP
failure/timeout signals and ICMP are controlled simulations. TCP timeout signals
are injected; they do not demonstrate an actual stalled OS connect. TLS and
HTTP run real Go protocol code over `net.Pipe`, with actual timers and normal
certificate verification. The rule cases supply explicitly simulated recorded
timings through the engine's HTTP interface; they do not measure fixture speed.
One defensive rule case intentionally supplies a duration on an unavailable
phase to prove it is excluded. Path sampling waits are skipped. No test connects
to the simulated public IPs. The separate opt-in fixtures use real loopback TCP.

## Reproduce

From the repository root in PowerShell:

```powershell
$env:GOCACHE = Join-Path $PWD '.tools/go-build'
$env:HOPVISTA_JOURNEY_CAPTURE = Join-Path $PWD '.tools/issue-12'
& ./.tools/go/bin/go.exe test ./internal/target ./internal/httpmeasure ./internal/engine -count=1 -v
node ui/node_modules/typescript/bin/tsc --noEmit -p ui/tsconfig.json
npm.cmd --prefix ui run build
& ./.tools/go/bin/go.exe test ./... -count=1
& ./.tools/go/bin/go.exe vet ./...
& ./.tools/go/bin/go.exe build -o .tools/hopvista-issue-12.exe ./cmd/hopvista
$env:HOPVISTA_LIVE_HTTP = '1'
& ./.tools/go/bin/go.exe test ./internal/httpmeasure -run TestLive -count=1 -v
Remove-Item Env:HOPVISTA_LIVE_HTTP
$env:HOPVISTA_PACKAGED_EXE = (Resolve-Path '.tools/hopvista-issue-12.exe').Path
& ./.tools/go/bin/go.exe test ./tools/app-smoke -run TestPackagedExecutableLoopback -count=1 -v
node tools/https-journey-preview.mjs
```

The last command uses `HOPVISTA_JOURNEY_CAPTURE`, defaults to issue #11 when
unset, and prints a loopback URL. Open it and submit
`https://fixture.example/<capture-name>`, using a path listed in its simulation
banner. This replays API-captured JSON in the actual production assets; it does
not initiate a new diagnostic run. Regenerate captures and restart the helper
after changing fixtures. Certificates/timestamps and real measured durations may
vary. Captures are ignored local files under `.tools/issue-12`.

## Verification outcomes

- Focused suites, repeated local TypeScript checks, production Vite build, full
  root Go suite, vet and executable build pass. The opt-in real TCP/TLS tests and
  packaged executable loopback smoke check pass. The packaged check includes
  embedded UI/security, native IPv4 probes and independent pinned HTTP checks.
- Ordinary full-suite runs skip opt-in live/network/minute fixtures. Only the
  explicit loopback fixtures above were enabled for this issue. No new public
  network measurement, native stalled route, loaded cutoff, or wider Windows
  build/account/network claim is made. `CGO_ENABLED=0`; no race-detector result
  is claimed.
- Browser replay covered 20 scenarios at measured 1440px and 390px widths: DNS
  failure/timeout, no-A, TCP failure/timeout, TLS failure/certificate/timeout,
  first-byte timeout, partial headers, redirected DNS failure, 404, 503, zero,
  tie, missing comparables, deadline, cancelled, running and stopping. No
  document-wide horizontal overflow was observed. Narrow screens retain
  summary/timing/path order. Observations, uncertainty, phase explanations and
  the direct-path caveat remain readable. Running/stopping findings stay pending;
  start is disabled and stopping also disables cancellation.
- TCP/TLS timeout, redirected DNS and retained first-byte evidence controls
  focus the exact record, expose the bound/error and preserve missing duration
  versus measured zero. Keyboard Enter on Return to measurement restores the
  originating control. No browser console warnings/errors were observed.
- DOM outputs and representative screenshots are ignored local artifacts under
  `.impeccable/review/issue-12/`. Early captures exposed the incomplete-notice
  defect; only rebuilt-asset checks are counted above. Initial synthetic rule
  records had empty earlier phase labels; their fixture records were completed
  before final replay checks. This is controlled rendered verification, not a
  full accessibility, 200% zoom, screen-reader or human comprehension acceptance.

## Standards

Independent review of implementation commit `13a1cb9` against fixed baseline
`b17ea9f129e508f9fc1f977cc8945d2d6f339d25`: no findings. The reviewer checked
documented glossary/design rules, schema version 1, Go-owned findings, approved
test seams, reuse of fixture helpers, and the code-smell baseline.

## Spec

Independent review of the same diff against issue #12 and referenced records:
no actionable findings or scope creep. This was source and verification-record
review; the reviewer did not independently repeat the browser checks. The final
commit adds these review outcomes and the acceptance-matrix pointer to the
reviewed implementation.

Review summary: Standards 0 findings; Spec 0 findings; neither axis identified
an issue requiring a change. Remaining release limitations above still apply.
