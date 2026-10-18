# Go Windows networking feasibility — milestone 1

**Current decision, 2026-10-14:** the explicitly revised development gate [passed final review](ipv4-development-gate-final-review.md). The initial 250 ms policy is approved empirically on this machine. Release readiness, wider Windows/account coverage and an actual stalled route call remain unverified. Earlier incomplete-gate statements below preserve historical research and have been superseded by the approved gate revision.

**Gate status: incomplete; application infrastructure remains gated.** Actual unelevated Go interop works for IPv4 and local IPv6 on this machine. The user revised [issue #10](https://github.com/DarrylN2/hopvista/issues/10) to ship an **IPv4-only first release** and, on 2026-10-14 local date, to show **“Stopping…” until native cleanup** instead of promising a hard maximum drain time. Remote IPv6, pending IPv6 drain, and scoped remote IPv6 evidence are deferred to a later release; they are not v1 gate checks or treated as passed. The full-cap start rule still needs a validated conservative scheduling margin; an actually stalled route lookup remains unverified.

The [current IPv4 gate audit](ipv4-milestone-1-gate-audit.md) separates observed Go evidence from the remaining controlled-network and timing-contract checks.

The subsequent [loaded boundary matrix](go-windows-timeout-investigation.md#follow-up-loaded-candidate-margin-boundary-matrix) **failed its 27-started-call expectation** under two busy Go workers with `GOMAXPROCS=1`: one attempt safely refused after setup, while 26 full-cap calls completed and closed before their fixture deadlines. Largest observed check-to-cleanup overrun was 26.57 ms. This measured condition does not validate the 250 ms candidate; the failed capture is preserved separately and the gate remains open.

A [follow-up loaded sampling fixture](go-windows-timeout-investigation.md#follow-up-loaded-margin-sampling-with-safe-refusals) then passed its own rule: 27 actual full-cap calls and one safe refusal in 28 attempts. It used an extra 50 ms sampling setup window while retaining the 250 ms candidate guard. Its 12.64 ms maximum overrun does not supersede the prior 26.57 ms observation or validate the margin. The gate stays open.

A later [loaded pending-drain fixture](go-windows-timeout-investigation.md#follow-up-pending-ipv4-drain-under-process-local-scheduler-load) observed nine calls still pending when stop was requested, then returning and closing under the same process-local scheduler load. Its largest stop-to-quiescent interval was 952.56 ms; the prior unloaded 978.25 ms observation remains larger. No latest-return guarantee or gate pass follows from these samples.

**Follow-up:** [focused timeout investigation](go-windows-timeout-investigation.md) reproduced the overrun through both Go and independent C# interop, including two additional targets. Actual Go event-based completion also overran. A test-fixture policy that refuses short remaining budgets passed its live check; issue #10 now approves that start rule, but no margin or production adapter has been validated and the gate remains incomplete.

The harness subsequently closed a gap in that fixture policy: it now rechecks the full budget after setup and before native send. An [induced 80 ms setup-delay check](go-windows-guard-delay.capture.json) refused the shortened call and closed its handle. This verifies one refusal path, not the required production margin.

The later [approved-cap cutoff experiment](go-windows-timeout-investigation.md#follow-up-approved-cap-cutoff-experiment) exercised 750, 900, and 1,000 ms fixture guards on the same unelevated Windows build. Its nine full-budget IPv4 calls completed with cleanup before their fixture deadlines, while six short-budget cases were refused before handle creation. These samples do not validate a production start margin or remote IPv6; the gate remains incomplete.

A further [local scheduling-margin study](go-windows-timeout-investigation.md#follow-up-local-scheduling-margin-characterization) captured 45 full-cap IPv4 calls across three candidate silent destinations. Its longest pre-call-to-cleanup interval was 1,003.19 ms for a 1,000 ms cap. This helps characterize this one machine but does not validate a production margin, IPv6, route-lookup latency, or a worst-case bound.

A later [provisional 250 ms boundary check](go-windows-timeout-investigation.md#follow-up-provisional-250-ms-start-margin-boundary) refused three just-short budgets before handle creation and completed 27 full-cap calls before their fixture deadlines. The largest observed check-to-cleanup overrun above a cap was 11.94 ms. The candidate still needs validation under stated supported conditions; it is not adopted as production policy.

A [loaded 60-second fixture](go-windows-timeout-investigation.md#follow-up-loaded-60-second-cutoff-fixture) then exercised the provisional guard with one full-cap IPv4 call and a later refusal while two CPU-bound Go workers contended on one Go scheduler processor. An induced post-lookup route phase crossed the real cutoff and did not start an echo. This supports local ordering under that named condition; it does not validate a production margin or an actually stalled `GetBestRoute2`.

A [route-lookup timing sample](go-windows-timeout-investigation.md#follow-up-route-lookup-timing-sample) then captured 100 `GetBestRoute2` calls. The observed calls were brief, but many were below the clock's measurable resolution; remote IPv6 still returned no route. These samples do not establish a worst-case route-lookup bound or cancellation strategy.

A later [route-before-probe cutoff check](go-windows-timeout-investigation.md#follow-up-route-before-probe-cutoff-ordering) confirmed that the harness refuses a native send after an induced route-phase delay consumes its investigation allowance, for all three approved IPv4 caps. A fourth fixture kept the simulated route phase active beyond its cutoff; stop was observed first, then the phase returned, and no echo started. A delayed `GetBestRoute2` call itself remains unbounded and unverified.

A [pending IPv4 drain study](go-windows-timeout-investigation.md#follow-up-pending-ipv4-drain-across-configured-caps) then stopped nine actual no-reply calls after send across the 750, 900, and 1,000 ms caps. Every call was pending at stop and closed its handle after native completion; the largest observed stop-to-quiescent interval was 978.25 ms. These observations support showing `stopping` until cleanup; they do not imply a maximum completion time.

A [controlled WSL fixture](go-windows-unreachable.capture.json) then returned a live IPv4 ICMP host-unreachable message to the actual unelevated Go call. Windows reported one reply, status 11003, native error 0, `unreachable_reply`, and successful handle close. The fixture restored its normal echo setting after the request. This verifies the live unreachable decoder on this machine; the timing and route/cutoff gate remains open.

Recorded on 2026-10-10 local date (2026-10-11 UTC), following [approved specification #10, section 8](https://github.com/DarrylN2/hopvista/issues/10). Read the root `AGENTS.md`, issue-tracker/triage/domain instructions, `GLOSSARY.md`, the probe and networking research notes, and the decisions in [#3](https://github.com/DarrylN2/hopvista/issues/3#issuecomment-6021816427), [#8](https://github.com/DarrylN2/hopvista/issues/8#issuecomment-6022027830), and [#7](https://github.com/DarrylN2/hopvista/issues/7#issuecomment-6051380625). No ADRs exist. The approved #10 cutoff and 30-hop production default supersede older planning notes' exact deadline and 16-hop proposal. This harness's four steps are a feasibility sample, not a production trace default.

## Toolchain and execution environment

- No `go` or `gh` executable was found on PATH or in the checked common Go locations. GitHub issue bodies/comments were read with the connected GitHub tools.
- Resolved Go availability with an official portable **Go 1.27.2 windows/amd64** archive, downloaded from [Go downloads](https://go.dev/dl/) into ignored `.tools/`. Verified SHA-256 against the [official release metadata](https://go.dev/dl/?mode=json): `1314008898bd40df77af4b014f777f08873dbdfbcd3d92308728ee03304fe04f`. The archive is `go1.27.2.windows-amd64.zip`, 79,008,081 bytes. No system installation or permanent PATH change was made.
- `go version` and the executable's `runtime.Version()` both reported `go1.27.2`. Build cache is workspace-local; `GOTOOLCHAIN=local` prevents automatic toolchain downloads. The isolated module has no third-party dependencies and declares Go 1.26; only 1.27.2 was executed here.
- `RtlGetVersion` reported **Windows 10.0.26200**, Windows/amd64. `GetTokenInformation(TokenElevation)` returned **false** in both sandbox and ordinary-process execution. This establishes that the successful Go process was unelevated; it does not substitute for testing a separate standard-user account or other Windows versions.
- The restricted sandbox returned native error **5** from IPv4/IPv6 handle creation and denied PowerShell CIM route queries. The same actual Go code ran outside that sandbox, still unelevated, and succeeded. Sandbox `ACCESS_DENIED` is not evidence of a Windows administrator requirement.

## Implementation and reproducible evidence

The isolated module and runner are in [tools/windows-network-feasibility](../../tools/windows-network-feasibility/README.md). It calls `IcmpCreateFile`, `Icmp6CreateFile`, `IcmpSendEcho2`, `Icmp6SendEcho2`, `IcmpCloseHandle`, and `GetBestRoute2` directly from Go through the standard library's Windows syscall support. No app server, UI, production run-result types, platform interface, or application coordinator was added.

Run from an unelevated PowerShell process:

```powershell
pwsh -NoProfile -File tools/windows-network-feasibility/run.ps1
```

The runner performs `go test ./...`, `go vet ./...`, and `go build`, captures actual Go observations, then independently captures Windows route information. All passed for the final source. Tests cover native layout/scoped-address conversion, remaining-time flooring and refusal at cutoff, refusal before native handle creation after cancellation, and reply-versus-API-error classification. These controlled tests supplement live native evidence.

The [redacted JSON capture](go-windows-network-feasibility.capture.json) retains environment, arguments, timestamps, raw statuses/errors, RTTs, elapsed times, handle-close outcomes, drain observations, and route comparison results. Original captures remain in ignored `tools/windows-network-feasibility/.local/`, including earlier live and sandbox captures. Private source/gateway/link-local and router addresses are omitted from the shared snapshot. This evidence JSON is a disposable harness format, not the versioned application run result.

Fixed native arguments: event/APC/context are NULL; request payload is eight ASCII bytes `hv-check`; TOS/flags/option length/option pointer are zero. Ordinary probes request 750 ms, TTL/Hop Limit as recorded, and 512-byte reply storage; the intentional short-buffer case supplies 8 bytes. `GetBestRoute2` receives NULL interface LUID/source, interface index 0, and sort options 0, with the selected literal destination. API failure uses the syscall's immediate last-error result only when reply count is zero; a nonzero count is decoded as a reply carrying its own status.

## Observed checks

| Check | Actual Go observation | Evidence limit |
| --- | --- | --- |
| IPv4 localhost | One reply from `127.0.0.1`, status 0; raw RTT 0 ms. | Local IPv4 interop only. |
| IPv6 localhost | One reply from `::1` at Hop Limit 64 and 1, status 0; raw RTT 0 ms. | Does not establish a remote IPv6 path. |
| Scoped local IPv6 | Echo to this machine's own link-local address with `%8`, Hop Limit 1, status 0; returned address bits match. Route-selected source includes scope 8, interface 8, unspecified next hop. | Native echo reply scope is **0**; retain the requested/route interface separately. Decoding a nonzero reply scope remains unverified. |
| Route-selected gateway | IPv4 next hop for `1.1.1.1` echoed with status 0. A route lookup to that gateway returned an unspecified next hop/on-link. | Selected for this destination at lookup time; no VPN comparison. |
| Remote IPv4 echo | `1.1.1.1`, TTL 64, one destination reply with status 0, roughly 38–39 ms raw RTT in the captures. | These are individual samples, not a reliability estimate. |
| IPv4 TTL stepping | TTL 1 and 2 returned status 11013 (`IP_TTL_EXPIRED_TRANSIT`). TTL 3 timed out; TTL 4 then returned another status-11013 router reply. | Demonstrates low-TTL replies and continuation after silence; no physical-link or HTTP-path claim. |
| Candidate silent IPv4 path | `192.0.2.1`, TTL 64, count 0, last error 11010 (`IP_REQ_TIMED_OUT`), roughly 0.5 s with a 750 ms argument. Candidate trace TTL 3 timed out and TTL 4 was still sent. | A timeout is not an unreachable reply and says nothing about HTTP availability. |
| Remote IPv6 route/echo | `2606:4700:4700::1111` and `2001:db8::1`: `GetBestRoute2` returned **1231**; echo count 0, native error **1231**, no fabricated RTT/reply address. | Windows independently reported **no `::/0` route**. This is route unavailability, not remote echo or Hop Limit success. |
| Native failures | Invalid echo handle and intentionally short reply buffer each returned count 0, error **87** (`ERROR_INVALID_PARAMETER`). | Separate API errors, not received ICMP unreachable statuses. |
| Native layouts and cleanup | Actual replies decoded IPv4/IPv6 addresses, status and RTT. Route row size 104, next-hop offset 44, sockaddr size 28; aligned 512-byte reply buffers. Every successfully created echo handle closed successfully, including failure/drain paths. | First synchronous reply only; payload pointers are not dereferenced. No stress/leak proof or other architecture claim. |
| Independent Windows routes | Selected source, interface and next hop match `Find-NetRoute`/`Get-NetRoute` for localhost, remote IPv4, candidate silent IPv4, scoped local IPv6 and gateway. Localhost, scoped local IPv6 and gateway are on-link. | Comparison is a point-in-time observation. No active VPN route was demonstrated. |
| Route lookup elapsed time | Captured synchronous lookups returned within a few milliseconds. | No API timeout/cancellation parameter; these samples establish no worst-case bound. |

The IPv6 reply's compact `IPV6_ADDRESS_EX` uses address offset 6 and scope offset 22, unlike input `SOCKADDR_INET` address offset 8/scope offset 24. Reply status/RTT are at 28/32. The initial scope decoder used offset 24; the scoped check prompted correction to 22. Live scope remains zero at the corrected location. Microsoft's [IPV6_ADDRESS_EX reference](https://learn.microsoft.com/en-us/windows/win32/api/ipexport/ns-ipexport-ipv6_address_ex) documents reply scope in network byte order; nonzero-scope parsing is source-based, not claimed as live-verified.

## Cancellation and cutoff evidence

The disposable scheduling checks send one actual native request, request cancellation or reach a short relative cutoff, then refuse the next request and wait for the first call to return. Buffers and handle remain owned until completion; `close_ok` is recorded afterward. A Go context does **not** interrupt the synchronous Windows call. Both observed pending IPv4 cases returned timeout 11010 and retained one started call with no subsequent call.

| Check | Native timeout argument | Measured behavior |
| --- | --- | --- |
| Cancellation during silent IPv4 call | 750 ms | Stop requested about 25 ms after send boundary. Call returned around 0.5 s after start; about 0.47 s of drain after stop. Handle closed successfully. |
| Call near cutoff | Remaining relative deadline about 149–150 ms, below the 750 ms cap | Repeated captures returned around **488–508 ms** after start, with about **338–358 ms** of drain after observed cutoff. Native call exceeded its supplied remaining-time bound by hundreds of milliseconds. |
| IPv6 cancellation/cutoff | 750 ms or remaining cutoff | Native send failed immediately with 1231; it was **not pending** at stop. This does not verify pending IPv6 drain. |

**The original remaining-native-timeout plus small-overhead contract failed these observations and was later superseded in issue #10.** Remaining time was converted/floored correctly, but these Windows calls did not honor a 150 ms argument closely enough to claim that contract. Returning around 0.5 s is finite and below 1 second in these samples; it is not proof of a worst-case bound. Do not infer that the cause is timer granularity, the network, or the Go runtime from these measurements alone.

The cutoff fixture uses a 150 ms relative deadline to exercise a native call with little remaining time; it is not a completed application 60-second run test. No application slot/state/cleanup-after-deadline promise is validated by this harness. Stop and quiescent timestamps are recorded separately.

This observed bound problem justified investigating the documented asynchronous event path before choosing the production adapter. [IcmpSendEcho2](https://learn.microsoft.com/en-us/windows/win32/api/icmpapi/nf-icmpapi-icmpsendecho2) and [Icmp6SendEcho2](https://learn.microsoft.com/en-us/windows/win32/api/icmpapi/nf-icmpapi-icmp6sendecho2) describe asynchronous calls. The initial capture above used synchronous calls. The [follow-up](go-windows-timeout-investigation.md) now records an experimental Go event binding with local completion/cleanup evidence, but no validated pending-operation cancellation method or production adapter selection.

## Remaining unverified checks and gate consequence

The first-release gate now evaluates IPv4. The IPv6 rows document deferred evidence for a future release; they do not block the IPv4-only v1 gate. Other rows remain unverified as stated, with production HTTP/UI checks belonging to later milestones.

| Remaining check | Required next evidence |
| --- | --- |
| Remote IPv6 echo and low Hop Limit transit | Unelevated Go run on a Windows network with a working external IPv6 route. Localhost and local scoped echo cannot substitute. |
| Pending IPv6 no-reply cancellation/cutoff | A routed, genuinely silent IPv6 target with stop and drain timestamps. Immediate no-route failure does not qualify. |
| Native cutoff/start contract | Validate the revised full-cap start rule with a conservative measured margin and explicit unstarted `deadline` evidence. Show `stopping` after cancellation/cutoff until a pending route lookup or echo returns and cleanup finishes; retain resources and the run slot. No hard latest terminal time is promised. The 149–150 ms direct overrun remains diagnostic evidence; event-based completion also overran. |
| Live ICMP unreachable decoding | **Observed** with the controlled WSL fixture: actual Go `IcmpSendEcho2` returned reply count 1, status 11003, native error 0, and a successful handle close. The constructed-buffer unit test remains supplemental. This local fixture does not establish behavior across other Windows conditions. |
| Nonzero IPv6 reply scope; remote scoped next hop | Actual replies/routes on a suitable IPv6 network. Current input/source scopes work, but the local echo reply's scope field is zero. |
| VPN destination-specific route comparison | An active VPN with explicit destination/source/interface/next-hop comparison. Miniport/virtual adapters alone are not VPN evidence. |
| Worst-case route/handle/native operation latency and cleanup under stress | Controlled repetitions under the intended Windows conditions. Current short observations and successful closes are limited evidence. |
| Other Windows versions, separate standard-user account, ARM64/386 | Separate execution and ABI/permission validation. Only this unelevated amd64 process/build was checked. |
| Independent `tracert` comparison and pinned Go HTTP/TLS | Not performed in this milestone; retain the broader acceptance checks and milestone 5's HTTP transport validation. Prior .NET HTTPS is not Go HTTP evidence. |

Milestone 1 supplies a runnable actual Go harness and useful evidence, but its revised IPv4 exit check is **not complete**. Live IPv4 unreachable decoding and simulated route-phase cutoff ordering now have controlled evidence; the validated start margin and behavior of an actually stalled `GetBestRoute2` remain explicit gaps. Pending IPv4 calls have returned and cleaned up in local samples, with no hard latest-return claim. VPN destination-specific route comparison and other-Windows-build coverage are separately unverified; issue #10 does not require an active VPN for the first gate. The IPv6 rows above are deferred research for a later release. Follow revised issue #10 and keep application infrastructure gated until the remaining IPv4 checks pass.
