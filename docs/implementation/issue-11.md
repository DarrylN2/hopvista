# Issue 11: redirected HTTPS journeys

Implemented and checked on 2026-10-17 (America/Los_Angeles), continuing from
`b864870` on `codex/diagnostic-ui-refinements`. The approved React/TypeScript,
Vite and plain CSS design and schema version 1 are preserved. No UI source or
embedded asset changes were needed. [Issue #11](https://github.com/DarrylN2/hopvista/issues/11)
extends the existing acceptance evidence; this does not declare release acceptance.

## Controlled checks

| Requirement | Repeatable evidence and outcome |
| --- | --- |
| Relative, absolute same-host and allowed new-host HTTPS redirects | `TestHTTPSRelativeSameHostAndNewHostJourney` passes a four-exchange chain, including a same-host port change and fragment removal. It records request authority/path, SNI, certificate SHA-256 identity, literal dial address, selected/actual peer, response status and fresh connection flag. The ordered event ledger confirms resolution before the new-host dial. |
| Pin reuse and policy | The chain performs one resolver call for each hostname, retains the initial pin across both same-host redirects, and retains `dns_ref` without inventing a repeated DNS duration. The new host uses its own selected public address. Exact dial assertions reject unchecked lookup, fallback IP, or extra connection behavior. |
| Phase attribution and first byte | Each exchange retains phase IDs, methods, bounds, explicit outcomes, timestamps and completed durations. TCP/TLS/HTTP bounds are 2000/2000/3000 ms. The fixture sends the first response-header byte separately, waits 40 ms, then sends the rest; the recorded HTTP endpoint precedes complete headers. Later exchanges cannot reuse earlier phase timestamps. |
| Invalid, IPv6, sensitive-address and limit stops | `TestHTTPSJourneysThroughAPI` preserves two earlier HTTPS responses for invalid scheme, IPv6 literal, and private-address stops, and all four responses at the redirect limit. It asserts dial/lookup counts, no forbidden exchange, actual stop code, human-readable explanation, retained selected address, and attempted-host trace. The over-limit new hostname is never resolved. |
| Certificate verification on a redirect | A permitted new public hostname receives a certificate with the wrong name. The third TCP connection succeeds, TLS fails verification, no HTTP request reaches that server, and no HTTP duration/status is invented. Earlier exchanges and failed TLS timestamps/bound remain. The actual TCP peer now survives this failure. Existing untrusted and IP-literal certificate fixtures remain regression coverage. |
| Local API, engine and findings | Every representative chain runs through the real session/start/status handler, run manager, engine, selector and HTTP measurer. API snapshots preserve schema version 1, exchange order and DNS references; findings reference the final recorded exchange and resolvable measurement IDs. Route/echo calls use only the final approved or last attempted address, with the corresponding trace role. |
| Real TCP/TLS | Opt-in `TestLivePinnedHTTPSRedirectJourney` passes a same-host relative redirect over actual loopback TCP sockets, normal certificate verification and production TLS defaults. It checks two fresh connections, Host/SNI, one DNS selection and actual loopback peers. Existing single-exchange HTTP/HTTPS fixtures also pass. |

The default HTTPS fixture uses real Go TLS 1.2 and HTTP over `net.Pipe`, with
explicitly simulated DNS answers and socket peer addresses. Only the test build
exposes the existing private dial seam (`export_test.go`); production dial policy
is unchanged. TLS 1.2 prevents TLS 1.3 ticket/alert writes deadlocking on the
unbuffered pipe. The separate opt-in loopback fixture uses production TLS defaults.
These tests do not connect to the public addresses named by the simulated pins.
API path observations use the existing fake probe adapter and skip sampling waits.
They verify orchestration and identity, not Windows route/ICMP or network latency.

## Fixes established by failing tests

`TestHTTPSRedirectStopReasonThroughAPI` first failed because all stops returned
the same generic family error. Go now provides specific explanations for invalid
redirects, IPv6, a new host with a non-public selection, and the redirect limit.
The certificate case also now explains verification failure. The existing view
displays these messages without computing findings or changing presentation.

The redirected wrong-name certificate test first failed because `actual_peer_ip`
was absent after successful TCP and failed TLS. The measurer now records the
validated TCP peer at dial completion, before TLS and `GotConn`. Certificate and
peer verification remain enabled; no fallback address or DNS lookup was added.

## Reproduce

From the repository root in PowerShell (use the local Go installation or `go`):

```powershell
$env:GOCACHE = Join-Path $PWD '.tools/go-build'
$env:HOPVISTA_JOURNEY_CAPTURE = Join-Path $PWD '.tools/issue-11'
& ./.tools/go/bin/go.exe test ./internal/httpmeasure -run TestHTTPS -count=1 -v
# Optional controlled loopback sockets, with local socket access permitted:
$env:HOPVISTA_LIVE_HTTP = '1'
& ./.tools/go/bin/go.exe test ./internal/httpmeasure -run TestLive -count=1 -v
Remove-Item Env:HOPVISTA_LIVE_HTTP
# Start the production-asset replay, then open its printed loopback URL:
node tools/https-journey-preview.mjs
```

The capture command writes `https-journey.json`, six `run-*.json` API snapshots
and corresponding `requests-*.json` records under the ignored `.tools/issue-11/`.
Certificate keys and certificates are regenerated per fixture; their identities
are expected to differ between runs. Captures contain controlled observations,
not live public network measurements. The preview is loopback-only and displays
a persistent label describing the simulation. Submit `https://fixture.example/`
with `completed`, `invalid`, `ipv6`, `sensitive`, `limit`, or `certificate` as its
path. The preview replays captured JSON into the existing production view; it
does not initiate a diagnostic run or replace the real API integration checks.

## Verification outcomes

- Focused HTTPS tests and existing HTTP/engine regression tests pass.
- Production TypeScript/Vite build passes; regenerated embedded assets have no
  content changes. Full root `go test ./...`, `go vet ./...`, and executable build
  pass. The opt-in packaged loopback smoke check passes, including native IPv4
  path and embedded UI/security checks.
- The existing opt-in loopback tests initially failed in the socket-restricted
  sandbox, then passed outside it, alongside the new redirect fixture. This was
  an execution restriction, not an observed product regression.
- `go test -race` could not run because this Go environment lacks cgo. No race
  detector result is claimed.
- All six API-captured results were replayed at measured 1440px and 390px widths:
  ordered exchanges, selected address and public scope, specific policy stop
  explanations, and final-host versus attempted-host identity matched the API
  results. No document-wide horizontal overflow was observed in these 12 cases.
  A blocked private DNS selection remained inspectable without becoming the
  summary/trace target. The failed TLS exchange displayed its retained peer.
- Evidence controls for the final exchange HTTP phase, failed TLS phase, and
  blocked-host DNS record focused the exact evidence ID. Keyboard Enter on
  Return to measurement restored the originating button. The final replay
  browser reported no console errors. Screenshots and DOM check output are local
  ignored artifacts under `.impeccable/review/issue-11/`.
- The first replay helper rejected `/ipv6` because its fixture-ID route accepted
  only letters; that helper route was corrected and the full 12-case browser
  matrix passed afterward. The initial incomplete attempt is not passing evidence.

Broader accessibility, human comprehension, 200% zoom, and Windows/network
release conditions remain open as recorded in the acceptance matrix and approved
design verification record.
