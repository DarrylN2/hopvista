# Windows networking feasibility harness

**Current status, 2026-10-14:** the user-approved empirical development gate [passed final verification](../../docs/research/ipv4-development-gate-final-review.md). The initial 250 ms allowance is accepted on the tested local conditions. Production/release acceptance and the recorded wider-platform/native-stall uncertainties remain open. Earlier incomplete-gate/provisional-margin wording below documents historical experiments and their limits.

Standalone milestone 1 of [approved specification #10](https://github.com/DarrylN2/hopvista/issues/10), section 8. The first release is now **IPv4-only**; this disposable Windows/amd64 module retains IPv6 experiments as future research, not v1 gate evidence. It uses only the Go standard library and actual `iphlpapi.dll` calls. It is not the production adapter or application.

## Run

To collect the existing checks on a separate ordinary Windows 11 x64 condition, use [`run-gate-recheck.ps1`](run-gate-recheck.ps1) and the [external-condition procedure](../../docs/research/ipv4-external-condition-recheck.md). It records a manifest and fresh raw captures under ignored `.local/`; it does not declare the feasibility gate passed.

From the repository root in an **unelevated** PowerShell process:

```powershell
pwsh -NoProfile -File tools/windows-network-feasibility/run.ps1
```

The script uses `.tools/go/bin/go.exe` when present, otherwise Go on PATH. Supply `-Go C:/path/to/go.exe` to choose another toolchain. It runs tests, vet, builds, executes the Go harness, and captures independent Windows route information. Go 1.26 or later is required; the executed toolchain version is recorded. No C compiler, raw socket, or third-party Go dependency is used. Windows ARM64/386 are unverified and intentionally excluded by filename constraints.

Override live destinations for a capable IPv6 or controlled test network:

```powershell
pwsh -NoProfile -File tools/windows-network-feasibility/run.ps1 `
  -IPv4 1.1.1.1 -IPv6 2606:4700:4700::1111 `
  -NoReplyIPv4 192.0.2.1 -NoReplyIPv6 2001:db8::1 -Hops 4
```

Use a scoped IPv6 literal such as `fe80::1234%12` when needed. The numeric scope is the Windows interface index; an interface name can also be resolved. A missing link-local zone is rejected before native calls. Destinations must be literal unicast addresses in their requested family. These explicit diagnostic arguments are not a complete production address-security policy; avoid subnet-directed broadcasts. No DNS or HTTP measurement is part of this gate.

The default destinations are Cloudflare's public resolver addresses and documentation addresses used as **candidate** silent paths. A documentation address may return an immediate error or an ICMP response instead of timing out. Silence, an unreachable reply, and pending cancellation must be observed, not assumed from an address.

## Evidence and review

Raw captures are written to ignored `.local/go-capture.json` and `.local/windows-routes.json`. They can contain private source/gateway/interface details. Review and redact before sharing. The console summary omits addresses. Exit 0 means capture completed, **not** that the feasibility gate passed.

The Go capture includes OS/build, architecture, actual Go version, token elevation, route source/interface/next hop, API arguments, TTL/Hop Limit, native timeout in milliseconds, reply count, native error versus reply status, parsed reply address/RTT, call elapsed time, and handle-close result. All native calls use NULL event/APC/context arguments. The first synchronous reply is decoded; multiple-reply parsing and echo payload pointer dereferencing are not implemented.

Checks include localhost v4/v6; a local link-local IPv6 address with its actual interface scope when available; selected next-hop echo and on-link route; low TTL/Hop Limit steps followed by ordinary echo; consecutive candidate-silent steps; invalid-handle and short-buffer failures; and missing IPv6 zone preflight. A unit test decodes a constructed IPv4 unreachable reply buffer and distinguishes it from a native API error with the same code. The opt-in controlled WSL fixture later produced a live unreachable reply. The script independently captures `Find-NetRoute`, `Get-NetRoute`, and adapter information for route/VPN comparison.

Cancellation checks request stop 25 ms after the send boundary; cutoff checks use a **150 ms relative deadline** to exercise remaining-time conversion without waiting 60 seconds. An in-flight synchronous native call retains its buffers and handle until return. Capture records the stop time, quiescent time, native timeout, whether the call was still pending at stop, and whether a subsequent call started. Only a pending, observed no-reply call provides cancellation/drain evidence. Immediate failure is not that evidence. The relative cutoff is not a full application 60-second orchestration test.

Review every required **IPv4** case from revised issue #10. The controlled live unreachable-response check passed; a validated start margin and route/cutoff ordering remain open. The approved v1 contract shows `stopping` until an in-flight route lookup or echo finishes and cleanup completes, with no hard latest-return promise. VPN comparison is separately unverified but is not required by the first gate. Remote IPv6 echo/transit, scoped live routes, and pending v6 cancellation are deferred beyond v1, not passed. Do not proceed to application infrastructure while the revised IPv4 gate remains incomplete. Current observations and unverified checks are in [the evidence report](../../docs/research/go-windows-network-feasibility.md).

## Focused timeout investigation

These opt-in live checks run through the same unelevated runner. Default unit tests skip network tests. Use one switch per invocation:

```powershell
pwsh -NoProfile -File tools/windows-network-feasibility/run.ps1 -TimeoutCheck
pwsh -NoProfile -File tools/windows-network-feasibility/run.ps1 -TimeoutSweep
pwsh -NoProfile -File tools/windows-network-feasibility/run.ps1 -NativeTimeoutComparison
pwsh -NoProfile -File tools/windows-network-feasibility/run.ps1 -AsyncTimeoutCheck
pwsh -NoProfile -File tools/windows-network-feasibility/run.ps1 -GuardedCutoffCheck
pwsh -NoProfile -File tools/windows-network-feasibility/run.ps1 -GuardDelayCheck
pwsh -NoProfile -File tools/windows-network-feasibility/run.ps1 -CutoffCapsCheck
pwsh -NoProfile -File tools/windows-network-feasibility/run.ps1 -MarginStudy
pwsh -NoProfile -File tools/windows-network-feasibility/run.ps1 -MarginBoundaryCheck
pwsh -NoProfile -File tools/windows-network-feasibility/run.ps1 -LoadedMarginBoundaryCheck
pwsh -NoProfile -File tools/windows-network-feasibility/run.ps1 -LoadedMarginSamplingCheck
pwsh -NoProfile -File tools/windows-network-feasibility/run.ps1 -MinuteCutoffCheck
pwsh -NoProfile -File tools/windows-network-feasibility/run.ps1 -RouteLatencyStudy
pwsh -NoProfile -File tools/windows-network-feasibility/run.ps1 -RouteCutoffCheck
pwsh -NoProfile -File tools/windows-network-feasibility/run.ps1 -DrainCapsCheck
pwsh -NoProfile -File tools/windows-network-feasibility/run.ps1 -LoadedDrainCapsCheck
pwsh -NoProfile -File tools/windows-network-feasibility/run.ps1 -UnreachableCheck -UnreachableIPv4 $wslIPv4
```

`TimeoutCheck` reproduces the 150 ms native-call overrun with three probes. `TimeoutSweep` varies the actual Go timeout argument. `NativeTimeoutComparison` runs independent C# interop with NULL options on the same and two additional documentation-address paths; this is a diagnostic comparison, not substitute Go acceptance evidence. `AsyncTimeoutCheck` calls actual Go event-based IPv4/local IPv6 APIs, pins buffers through completion with `runtime.Pinner`, forces a GC while pending, parses completed replies, then unpins and closes handles. It never treats an elapsed caller deadline as proof that native buffers are reusable. If the event fails to signal within 2 seconds, the diagnostic retains its pinned buffers/handles until process exit, stops further probes, and fails the check.

`GuardedCutoffCheck` tests the now-approved **full-cap start rule** only in the live test fixture: refuse to start when less than the full 750 ms probe cap plus a 50 ms investigation allowance remains. It checks that the 150 ms remaining-budget case makes no native call, then measures three full-budget calls. The 50 ms value is not an approved margin; this is not production adapter adoption or a worst-case latency guarantee.

`GuardDelayCheck` induces an 80 ms pause after handle creation to verify the second full-budget check refuses a shortened native call and closes that handle. It tests the harness guard under one controlled setup delay; it does not validate a production margin or bound scheduling and native completion.

The guarded experiment computes its full native timeout from the same clock reading as its budget check and rechecks the budget immediately before the native send boundary. A unit check simulates a pause after an earlier guard and requires refusal rather than a shortened timeout. This hardens the harness rule; it does not make clock checking and Windows call entry atomic or validate the 250 ms candidate across Windows conditions.

`CutoffCapsCheck` extends that **harness-only experiment** to the approved 750 ms trace, 900 ms gateway, and 1,000 ms target caps. For each cap it records refusal at 150 ms and at cap + 49 ms remaining, then three actual calls with cap + 150 ms remaining. Its capture records the fixture deadline, native send start/end, return after handle cleanup, native timeout/error, and whether the deadline had passed at return. It uses one previously observed candidate-silent IPv4 documentation address; a reply or immediate error fails the no-reply prerequisite. The 50 ms allowance remains an investigation threshold, not an approved product margin. The revised issue #10 still requires a validated start margin and honest stop/cleanup ordering before the feasibility gate passes.

`MarginStudy` measures five full-cap calls for each approved cap on each of three documentation-address candidate silent paths. It records the complete interval from a policy-check point through native completion and handle cleanup, along with the native interval and a 150 ms diagnostic observation window. A reply or immediate API error fails the no-reply prerequisite. This local sample is not a production margin or worst-case guarantee; it does not exercise IPv6, cancellation, route lookup, other Windows builds, or a separate standard-user account.

`MarginBoundaryCheck` tests a provisional 250 ms start allowance on the same three IPv4 paths and three caps. It checks refusal before handle creation at cap + 249 ms and makes 27 full-cap calls near the candidate boundary, recording send, deadline, result, and cleanup. A passing local fixture does not validate the margin for production or establish a worst-case native return time.

`LoadedMarginBoundaryCheck` repeats that unchanged boundary matrix with two CPU-bound Go workers and process `GOMAXPROCS=1`. It waits for both workers to begin, records worker progress around each attempted full-cap call, joins the workers, and restores the original scheduler setting. It writes a separate `.local/loaded-margin-boundary.json`, preserving the unloaded capture. This is process-local contention, not system-wide stress. The first execution **failed the 27-started-call expectation**: one attempt safely refused after setup consumed the 20 ms fixture headroom, closing its created handle; the other 26 calls completed before their fixture deadlines with successful cleanup. Keep this failure visible rather than counting the matrix as passed or changing the candidate margin to make it pass. See the [recorded analysis](../../docs/research/go-windows-timeout-investigation.md#follow-up-loaded-candidate-margin-boundary-matrix).

`LoadedMarginSamplingCheck` uses the same two-worker, `GOMAXPROCS=1` condition with cap + candidate margin + 50 ms fixture budgets. It records safe post-setup refusals and allows up to five attempts to obtain three actual full-cap calls per cap/path. The first execution recorded one safe refusal and 27 completed calls, with no fixture deadline crossing or close failure. Its `.local/loaded-margin-sampling.json` and [reviewed capture](../../docs/research/go-windows-loaded-margin-sampling.capture.json) are separate from both earlier boundary captures. The fixture passes its sampling rules; this does not establish a production margin or worst-case time.

`MinuteCutoffCheck` runs a real 60-second deadline with two CPU-bound Go workers and `GOMAXPROCS=1` for the final five seconds. It checks one full-cap IPv4 probe, a later refusal, and an induced post-lookup route phase that crosses the cutoff without starting an echo. It records stop and quiescence times; the induced delay does not represent an actually stalled `GetBestRoute2` or establish a production timing bound.

`RouteLatencyStudy` makes 25 individually timed `GetBestRoute2` calls and one 1,000-call timed batch for each of two IPv4 and two IPv6 destinations. It saves only destination literals, family, timestamps, elapsed times, and return codes; private selected source/next-hop information stays out of the shared capture. Rapid individual calls can measure below this machine's clock resolution and appear as `0 ms`; that means unresolved duration, not an instantaneous lookup. The batch measures aggregate throughput, not worst-case single-call latency. The API has no timeout or cancellation argument, so these samples do not establish a worst-case bound.

`RouteCutoffCheck` performs real IPv4 route lookups, adds a controlled 100 ms delay in the route phase, then applies the harness's full-cap start guard. It verifies refusal before handle creation for the 750, 900, and 1,000 ms caps when the route phase consumes the 50 ms investigation allowance. A separate 50 ms cutoff fixture keeps the simulated route phase active for 30 ms past the deadline, records stop observation and quiescence, and confirms no echo starts. The capture omits selected source, next hop, and interface details. The induced delay does not bound or cancel `GetBestRoute2` or validate a production margin.

`DrainCapsCheck` requests cancellation 25 ms after an actual IPv4 send for each configured native cap and each of three documentation-address candidate silent paths. It records whether the call was pending at stop, stop-to-quiescent time, native timeout/error, and handle cleanup. A successful run confirms finite local samples; v1 now shows `stopping` until cleanup without promising a worst-case drain time.

`LoadedDrainCapsCheck` repeats those nine pending-call checks with two CPU-bound Go workers and process `GOMAXPROCS=1`, records worker progress from stop through quiescence, and writes a separate `.local/loaded-drain-caps.json`. It joins the workers and restores the scheduler setting. The first run had nine pending calls at stop, nine native returns and successful handle closes afterward, and a largest observed stop-to-quiescent interval of 952.56 ms. This is [one named local condition](../../docs/research/go-windows-loaded-drain-caps.capture.json), not a latest-return guarantee or evidence for a separate account/build.

`UnreachableCheck` requires a separately started controlled responder. The one-shot `debug/unreachable_fixture.py` ran as root inside an existing Ubuntu WSL instance: it temporarily suppressed normal echo replies, sent ICMP host-unreachable for one matching `hv-check` request, and restored the original echo setting on exit. The setting was read back as `0`. The Go check then observed reply count 1, status 11003, native error 0, and successful handle close. Run this only in a controlled WSL environment; its raw `.local/unreachable-fixture.json` contains private WSL addresses. The [saved capture](../../docs/research/go-windows-unreachable.capture.json) redacts them.

The native and async short-timeout checks currently **fail intentionally** when the observed overrun exceeds the 50 ms investigation allowance. This failure is the recorded signal; the normal unit suite still passes. Captures are `.local/timeout-observations.json`, `timeout-sweep.json`, `native-timeout-comparison.json`, `async-timeout-observations.json`, `guarded-cutoff-observations.json`, `guard-delay.json`, `cutoff-cap-observations.json`, `margin-study.json`, `margin-boundary.json`, `minute-cutoff.json`, `route-latency.json`, `route-cutoff.json`, and `drain-caps.json`. A successful timing capture is not a passed timing contract. [Recorded diagnosis and candidate mitigation](../../docs/research/go-windows-timeout-investigation.md).

## Interop references

- [IcmpSendEcho2](https://learn.microsoft.com/en-us/windows/win32/api/icmpapi/nf-icmpapi-icmpsendecho2) and [Icmp6SendEcho2](https://learn.microsoft.com/en-us/windows/win32/api/icmpapi/nf-icmpapi-icmp6sendecho2): synchronous calls, reply count, errors, timeout units, buffer sizing.
- [IP_OPTION_INFORMATION32](https://learn.microsoft.com/en-us/windows/win32/api/ipexport/ns-ipexport-ip_option_information32) and [ICMP_ECHO_REPLY32](https://learn.microsoft.com/en-us/windows/win32/api/ipexport/ns-ipexport-icmp_echo_reply32): fixed-width layouts on a 64-bit platform.
- [IPV6_ADDRESS_EX](https://learn.microsoft.com/en-us/windows/win32/api/ipexport/ns-ipexport-ipv6_address_ex) and [ICMPV6_ECHO_REPLY](https://learn.microsoft.com/en-us/windows/win32/api/ipexport/ns-ipexport-icmpv6_echo_reply_lh): compact reply address, scope, status, RTT.
- [GetBestRoute2](https://learn.microsoft.com/en-us/windows/win32/api/netioapi/nf-netioapi-getbestroute2) and [MIB_IPFORWARD_ROW2](https://learn.microsoft.com/en-us/windows/win32/api/netioapi/ns-netioapi-mib_ipforward_row2): source selection and route-dependent next hop; unspecified means on-link.

Buffers have 8-byte aligned backing storage; route row size 104 and next-hop offset 44 are checked. Echo reply buffers are 512 bytes, exceeding the documented one-reply header + 8-byte request + 8-byte ICMP error + IPv6 IO_STATUS_BLOCK requirement. Only required scalar fields are decoded; Go pointers are never embedded in native layouts. The compact IPv6 reply address starts at offset 6, scope at 22, and status at 28; this differs from the input `SOCKADDR_INET` scope offset 24. The reply scope uses Microsoft's documented network byte order; nonzero reply scope remains unverified because local replies returned zero. `runtime.KeepAlive` keeps every call buffer reachable through synchronous completion. Every successfully created echo handle is closed with a recorded result on success and failure.
