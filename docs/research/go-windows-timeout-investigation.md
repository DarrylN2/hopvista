# Native ICMP timeout investigation

**Current decision, 2026-10-14:** the approved empirical development gate [passed final verification](ipv4-development-gate-final-review.md). The 250 ms initial allowance is accepted on named tested conditions, without a worst-case guarantee. Earlier provisional-margin/open-gate statements below describe historical stages. Failed timing experiments and actual unverified native/platform cases remain as recorded; release acceptance remains separate.

Follow-up to [milestone 1 evidence](go-windows-network-feasibility.md), on the same unelevated Windows 10.0.26200/amd64 environment with Go 1.27.2. **The revised IPv4-only v1 gate is still incomplete.** Remote IPv6 remains unavailable and is now deferred beyond v1 under [issue #10](https://github.com/DarrylN2/hopvista/issues/10). Issue #10 approves the full-cap start rule with an evidence-validated scheduling margin and explicit unstarted `deadline` outcome; only fixture versions of that rule have been tested, and no margin has been validated.

## Reproduction and hypotheses

The focused feedback command is:

```powershell
pwsh -NoProfile -File tools/windows-network-feasibility/run.ps1 -TimeoutCheck
```

It executes three actual synchronous `IcmpSendEcho2` calls to `192.0.2.1`, each requesting 150 ms, and fails when a confirmed no-reply timeout takes more than the requested bound plus a **50 ms investigation allowance**. That margin is a test threshold, not an approved product contract. All three attempts failed: **495.32, 486.14, 503.01 ms**, count 0, last error 11010, successful handle cleanup. This minimal repro has no cancellation timer, HTTP, DNS, orchestration, or route lookup in its measured IPv4 call.

Ranked hypotheses and discriminating checks:

1. **Windows/path completion behavior:** if responsible, changing timeout values will expose a pattern also seen through another binding. Sweep 50–1500 ms.
2. **Go interop or request-options error:** if responsible, independent C# P/Invoke with NULL request options will obey the short timeout when Go does not. Use the same DLL/API/address/payload/buffer size.
3. **One target's path:** if responsible, the short-timeout symptom will disappear on the two other documentation-address paths. Keep the API and timeout fixed.

## Actual observations

All measurements below had zero replies and timeout error 11010. These are observations from this machine, not guaranteed Windows timer behavior.

| Requested timeout | Go, same target | Independent C#, same target |
| --- | --- | --- |
| 50 ms | 471.63 ms | 308.62 ms |
| 150 ms | 486.38 ms | 500.28 ms |
| 300 ms | 497.41 ms | 499.50 ms |
| 500 ms | 503.75 ms | 500.04 ms |
| 750 ms | 501.23 ms | 499.01 ms |
| 1000 ms | 1000.53 ms | 1004.91 ms |
| 1500 ms | 1501.20 ms | 1497.29 ms |

Independent C# calls with a 150 ms timeout also took **479.50 ms** for `198.51.100.1` and **504.21 ms** for `203.0.113.1`. All are explicitly chosen documentation-address candidates; no no-reply result was assumed from their scope.

**Inference:** the evidence argues against a Go-only interop problem, nonzero request-options problem, or behavior unique to one destination. The half-second clustering is consistent with coarse completion timing somewhere in the Windows/path behavior. It does not identify a specific kernel timer, driver, filter, or network device. First-call timing varied, so do not describe 500 ms as a guaranteed minimum or maximum.

## Go asynchronous event experiment

The harness now has an **experimental** event-path binding, exercised by:

```powershell
pwsh -NoProfile -File tools/windows-network-feasibility/run.ps1 -AsyncTimeoutCheck
```

Actual Go calls used `CreateEventW`, `IcmpSendEcho2`/`Icmp6SendEcho2` with a non-NULL event, `WaitForSingleObject`, and family-specific `IcmpParseReplies`/`Icmp6ParseReplies`. All request/options/reply/source/destination buffers stayed pinned with `runtime.Pinner` until native completion and parsing. A GC was forced while pending. Only then were buffers unpinned and ICMP/event handles closed. The timeout check waits for completion rather than releasing storage when its preferred deadline passes.

| Target | Submission elapsed | Native completion elapsed | Result |
| --- | --- | --- | --- |
| `127.0.0.1` | 0.57 ms | 2.24 ms | Event signaled; one status-0 IPv4 reply. |
| `::1` | 0.71 ms | 1.61 ms | Event signaled; one status-0 IPv6 reply. |
| `192.0.2.1`, timeout argument 150 ms | 0.57 ms | **320.54 ms** | Event signaled; parser returned zero replies/error 11010. Timing check failed. |

Every observed case completed, unpinned its buffers, and closed both handles. This proves local event submission/completion/parsing and cleanup across a forced GC in these cases. It does **not** establish timely cancellation, remote IPv6, or an event-based worst-case bound. Simply returning control quickly does not make buffers reusable.

The public [IPv4 API](https://learn.microsoft.com/en-us/windows/win32/api/icmpapi/nf-icmpapi-icmpsendecho2) documents event submission and [reply parsing](https://learn.microsoft.com/en-us/windows/win32/api/icmpapi/nf-icmpapi-icmpparsereplies). [Go Pinner](https://pkg.go.dev/runtime#Pinner) describes keeping objects fixed through external ownership. [IcmpCloseHandle](https://learn.microsoft.com/en-us/windows/win32/api/icmpapi/nf-icmpapi-icmpclosehandle) documents handle closing without an explicit pending-request cancellation contract. No claim that closing a pending handle safely cancels it was made. `CancelIoEx` is documented for file I/O; no evidence was established that an opaque IP Helper handle supports it, and it was not used as a claimed fix.

If event completion remains unverified after the diagnostic's 2-second observation window, pinned buffers/handles are retained until process exit, subsequent probes stop, and the test fails. This is a diagnostic escape path, not accepted application cleanup behavior.

## Candidate policy and regression check

The caller policy is to **refuse a new native probe when less than its full cap plus a validated margin remains**, rather than trim the native timeout into the observed unreliable short range. This preserves the distinction between scheduling cutoff and native completion and leaves a measurement explicitly unstarted with a `deadline` outcome. It cannot interrupt an already pending call. Issue #10 approved this rule after the experiments below; the fixture's 50 ms allowance is not the validated margin required by the specification.

The candidate exists only in the harness's `guardedProbe` and `echoInternal` experiment path; the default capture continues to preserve the original short-timeout behavior. No production adapter or coordinator was changed.

```powershell
pwsh -NoProfile -File tools/windows-network-feasibility/run.ps1 -GuardedCutoffCheck
```

The test was run **before** applying the guard: a 150 ms remaining-budget case entered a native call, passed 142 ms, took **290.47 ms**, and failed the regression check. Three full-budget calls used 750 ms and completed in 489.77–513.28 ms.

After the guard, a 150 ms remaining budget was refused **before handle creation**, with `not_started`, no send timestamp, and no cleanup handle. With 900 ms remaining, three actual Go requests retained the full 750 ms timeout and completed in **78.44, 502.49, 496.24 ms**, each with timeout 11010 and successful handle close. The guarded regression test **passed**. A normal unit test also verifies refusal with 1 ns, 150 ms and 799 ms remaining without making a native call.

The experiment reserves 750 + 50 = 800 ms. **50 ms is an investigation allowance, not a justified production margin.** These samples demonstrate the caller mitigation; they do not prove a worst-case completion bound for every allowed cap or Windows build. The approved IPv4 policy still requires a validated conservative margin, explicit unstarted `deadline` evidence in an eventual adapter, and pending IPv4 cancellation/drain evidence under supported conditions. Pending IPv6 evidence belongs to the later IPv6 release. The underlying 150 ms direct-call test remains intentionally red; the guard avoids that call rather than changing Windows behavior.

### Post-setup guard check

A review found that the first guard ran before native handle and buffer setup, while `nativeTimeout` could still trim the timeout if that setup consumed the 50 ms allowance. The harness now checks the full cap plus investigation allowance again immediately before choosing the native timeout. An opt-in live test holds the call for 80 ms after handle creation. With an 800 ms fixture budget and a 750 ms cap, it returned `not_started` without entering `IcmpSendEcho2`, then closed the created handle successfully. The [reviewed capture](go-windows-guard-delay.capture.json) contains this result and no private network addresses. The standard unit test checks the exact threshold and one-nanosecond-short case. `go test ./...` and `go vet ./...` passed. This confirms the guard's refusal path under an induced setup delay on this machine. It does not validate 50 ms as a production margin, eliminate scheduling delay between the final check and native entry, or bound an in-flight call.

A later review found a narrower gap: the guarded path read the clock once for the full-budget check and again for `nativeTimeout`, which could trim a full-cap call if the process paused between those reads. The harness now uses one clock reading to choose either the full native cap or an unstarted refusal, then rechecks the budget immediately before the native-send boundary. A unit test simulates a 100 ms pause after an earlier guard and verifies refusal; it also rejects a fractional-millisecond full cap rather than silently shortening it. `go test ./...`, `go vet ./...`, and build passed. The 80 ms live post-setup check passed again in an ordinary unelevated process, with no send and successful handle close. A restricted command-sandbox attempt returned native error 5 during handle creation and was **not counted** as guard evidence. The second check narrows the scheduling window; it cannot make the check and Windows call entry atomic or validate a production margin.

## Follow-up: approved-cap cutoff experiment

The opt-in `-CutoffCapsCheck` extends the fixture-only guard to the three approved per-probe caps: **750 ms trace, 900 ms gateway, and 1,000 ms target**. It ran as an unelevated Go 1.27.2 Windows/amd64 process on Windows 10.0.26200. The sole destination was the previously observed candidate-silent `192.0.2.1`; every started call returned zero replies and native timeout 11010. The [reviewed capture](go-windows-cutoff-cap.capture.json) contains the fixture deadline, native send start/end, return after handle cleanup, and all 15 observations. It contains no private network addresses.

| Requested native cap | Refused budgets, before handle creation | Three full-budget native-call elapsed times | Full-budget fixture deadline |
| --- | --- | --- | --- |
| 750 ms | 150 and 799 ms | 204.48, 502.16, 496.34 ms | 900 ms after fixture start |
| 900 ms | 150 and 949 ms | 502.93, 486.25, 500.30 ms | 1,050 ms after fixture start |
| 1,000 ms | 150 and 1,049 ms | 1,005.83, 1,005.66, 992.23 ms | 1,150 ms after fixture start |

The nine started calls used their full configured native timeout, returned before the fixture deadline, and closed their handles successfully. The longest observed interval from native return to fixture return after handle cleanup was about 1.14 ms. The 1,000 ms calls show that native completion can exceed the nominal timeout by a few milliseconds even in this small sample. The 50 ms threshold and extra fixture budget are diagnostic choices, not validated scheduling margins or worst-case bounds.

**The earlier smaller-remaining-timeout contract was superseded in issue #10 after this experiment.** The revised rule declines to start when a full cap plus a validated margin cannot fit, and records an unstarted `deadline` outcome. This passing fixture establishes only the behavior of a 50 ms investigation allowance on this IPv4 path. It does not validate a production margin or establish IPv6, route-lookup, cancellation, and other-Windows-build bounds. Application infrastructure remains gated.

## Follow-up: local scheduling-margin characterization

The opt-in `-MarginStudy` ran 45 actual, unelevated Go `IcmpSendEcho2` calls on Windows 10.0.26200/amd64 with Go 1.27.2: five calls for each 750, 900, and 1,000 ms cap on each of three documentation-address candidate silent paths (`192.0.2.1`, `198.51.100.1`, and `203.0.113.1`). The harness recorded a point immediately before `echo`, native send start/end, and return after handle cleanup. Every call returned zero replies with native timeout 11010 and `close_ok: true`; none crossed its cap + 150 ms diagnostic observation deadline. This 150 ms window is not a proposed production margin.

| Native cap | Samples | Longest native call | Longest pre-call-to-cleanup interval | Largest observed pre-send interval | Largest observed native-return-to-cleanup interval |
| --- | ---: | ---: | ---: | ---: | ---: |
| 750 ms | 15 | 500.83 ms | 502.49 ms | 3.40 ms | 1.35 ms |
| 900 ms | 15 | 500.63 ms | 501.09 ms | 0.51 ms | 1.30 ms |
| 1,000 ms | 15 | 1,001.52 ms | **1,003.19 ms** | 0.66 ms | 1.17 ms |

The [reviewed capture](go-windows-margin-study.capture.json) contains all individual intervals, statuses, environment, and cleanup outcomes. Its only addresses are the three documentation destinations; it has no private route/source data. These samples show that the 1,000 ms cap needs some positive headroom even on this machine, but they cannot establish a conservative worst-case margin. The study does not exercise an actual 60-second diagnostic run, route lookup in the timed interval, cancellation during a pending call, IPv6, VPN paths, a separate standard-user account, load extremes, or another Windows build. Milestone 1 still needs a validated start margin; issue #10 now allows `stopping` to continue until native cleanup without a hard latest-return promise. Do not choose 50 or 150 ms as production policy from this capture alone.

## Follow-up: route-lookup timing sample

The opt-in `-RouteLatencyStudy` made 25 individually timed unelevated `GetBestRoute2` calls and a separate 1,000-call timed batch for each of `1.1.1.1`, `192.0.2.1`, `::1`, and `2606:4700:4700::1111` on the same Windows build. The first three destinations returned code 0; the remote IPv6 destination returned 1231, consistent with the still-absent IPv6 default route. The longest recorded whole-call intervals were 0.64 ms for the first IPv4 destination and 0.56 ms for remote IPv6. Many rapid calls recorded `0 ms` because their start and end fell on the same observed clock tick. This is a timer-resolution limit, not evidence of zero latency. The 1,000-call batches took 6.18, 6.34, 6.93, and 8.85 ms respectively; those totals describe throughput, not a per-call maximum.

The [redacted route-timing capture](go-windows-route-latency.capture.json) contains 100 individual timestamps and elapsed values, four batch totals, return codes, and the unelevated environment. It deliberately excludes private selected source, next-hop, and interface values. These fast, consecutive observations do not bound a worst-case route lookup, include an active VPN or remote IPv6 route, or prove that a synchronous `GetBestRoute2` call cannot delay a cutoff. The API exposes no timeout/cancellation parameter; a production strategy for a delayed lookup remains unverified.

## Follow-up: route-before-probe cutoff ordering

The opt-in `-RouteCutoffCheck` used actual unelevated Go `GetBestRoute2` for `192.0.2.1`, then inserted a controlled 100 ms delay before applying the harness-only full-cap start guard. The fixture began with cap + 80 ms remaining for each 750, 900, and 1,000 ms cap. All three route calls succeeded. After the delay, the guard returned `not_started` before handle creation; no native echo call began. A fourth fixture used a 50 ms deadline, performed a real lookup, held the simulated route phase until 30 ms after the deadline, then rechecked the guard. The stop signal was observed before route-phase return; the echo remained `not_started`, and quiescence was recorded after that return. Stop-to-quiescent was **30.09 ms in this one run**. The [reviewed capture](go-windows-route-cutoff.capture.json) contains only the documentation destination, route return code/timing, deadline, stop and quiescence timestamps, and probe observation; selected private route details are omitted.

This supports a narrow ordering rule: after route selection, check the remaining full-cap budget again before starting an echo, including when the route phase crosses the cutoff. The delay was inserted **after** the real lookup, so this does not show a slow `GetBestRoute2` call can be interrupted or bound. It does not validate the 50 ms investigation allowance as a production scheduling margin. A route change between lookup and later network work also remains possible. The route/cutoff strategy has a passing harness fixture, but its behavior during an actual stalled lookup remains unverified; the feasibility gate is still open.

## Follow-up: provisional 250 ms start-margin boundary

The earlier 45-call margin study had a largest observed check-to-cleanup interval of 1,003.19 ms at the 1,000 ms cap, or 3.19 ms above that cap. As a **candidate for further validation**, the new `-MarginBoundaryCheck` uses a 250 ms full-cap start allowance. This round value is far above the observed overrun, but the sample maximum does not define a worst case or justify production adoption by itself.

On the same unelevated Windows 10.0.26200/amd64 machine with Go 1.27.2, the harness refused budgets of cap + 249 ms for each of the 750, 900, and 1,000 ms caps before creating a native handle. It then made three full-cap calls for each cap on each of three documentation-address candidate silent paths, with cap + 270 ms fixture budgets. All **27** started calls requested their full native cap, returned zero replies with native timeout 11010, closed their handles, and returned before their fixture deadlines. The largest observed check-to-cleanup interval was **1,011.94 ms** at the 1,000 ms cap, or **11.94 ms above the cap**. The [reviewed capture](go-windows-margin-boundary.capture.json) records each refusal, send, result, deadline, and cleanup outcome without private route values.

This is a passing **local candidate test**, not a validated production margin. It has no induced CPU load, separate standard-user account, other Windows build, VPN path, actual 60-second run, or demonstrated stalled `GetBestRoute2`; native API documentation supplies no hard latest-return bound. The approved `stopping` contract handles work that remains in flight after cutoff. The feasibility gate stays open while the supported conditions and conservative margin are validated.

## Follow-up: loaded 60-second cutoff fixture

The opt-in `-MinuteCutoffCheck` ran a real 60-second deadline on the same unelevated Windows 10.0.26200/amd64 machine with Go 1.27.2. During the final five seconds, the test process set `GOMAXPROCS=1` and ran two CPU-bound Go workers; its prior setting was 20 and was restored after the test. This creates process-local Go scheduler contention, not a measured system-wide or worst-case load. At about 58.7 seconds, the fixture performed a real IPv4 `GetBestRoute2` lookup and used the provisional 250 ms guard for a 1,000 ms no-reply echo to `192.0.2.1`. The call used the full 1,000 ms native timeout, returned timeout 11010 with zero replies, closed its handle, and finished before the 60-second cutoff. A later guard check returned `not_started` without entering the native API.

A separate route phase began before the cutoff, performed a real route lookup, then deliberately paused **after** that lookup until past the deadline. Its post-phase guard refused the echo. The stop signal was observed about 6.93 ms after the nominal deadline, and quiescence after phase return and worker shutdown was **62.52 ms after stop observation** in this one run. The [reviewed capture](go-windows-minute-cutoff.capture.json) records those timestamps, API outcomes, cleanup, test conditions, and worker progress without private selected-route values.

This fixture confirms the candidate guard and no-new-echo ordering under one explicitly stated local load condition. It does not show an actually slow `GetBestRoute2`, a hard native return limit, behavior on other Windows builds/accounts/network paths, or a production `stopping` UI. The candidate margin and milestone 1 gate remain unapproved.

## Follow-up: pending IPv4 drain across configured caps

The opt-in `-DrainCapsCheck` requested cancellation 25 ms after an actual unelevated Go IPv4 send to each of three documentation-address candidate silent paths, once at each configured 750, 900, and 1,000 ms cap. All nine calls were still pending at the stop request, retained their full native timeout argument, returned zero replies with native timeout 11010, and closed their handles. The context stopped later scheduling; it did not interrupt the synchronous native call. The [reviewed capture](go-windows-drain-caps.capture.json) contains send/stop/quiescent timestamps and outcomes, with no private network addresses.

| Native cap | Pending samples | Observed stop-to-quiescent range |
| --- | ---: | ---: |
| 750 ms | 3 | 63.28–483.24 ms |
| 900 ms | 3 | 466.32–488.81 ms |
| 1,000 ms | 3 | 957.21–978.25 ms |

One 750 ms call completed much sooner than the other two, so these timings are observations, not a fixed scheduling rule. This experiment verifies finite pending drain and cleanup on this machine for the three caps. It does not establish a worst-case bound under load, other paths, or other supported Windows conditions. The start margin and route/cutoff strategy also remain unvalidated.

## Follow-up: loaded candidate-margin boundary matrix

On 2026-10-14 local date, `-LoadedMarginBoundaryCheck` repeated the existing 250 ms candidate boundary matrix under two CPU-bound Go workers with process `GOMAXPROCS=1` on the same unelevated Windows `10.0.26200/amd64`, Go 1.27.2. Both workers began before the matrix; progress counters increased around every attempted full-cap call. The workers were joined and the prior scheduler setting (20) was restored. This is one process-local contention condition, not measured system-wide stress. The [reviewed capture](go-windows-loaded-margin-boundary.capture.json) preserves the failed run; the earlier unloaded capture was not overwritten.

**Diagnostic result: failed**, because the fixture expected 27 started full-cap calls. All three cap + 249 ms budgets refused before handle creation as expected. Of the 27 cap + 270 ms attempts, the first 750 ms attempt to `192.0.2.1` spent 22.93 ms from check through cleanup and refused after handle setup consumed its 20 ms excess budget. It recorded no send, zero native timeout, `not_started`, and successful handle close. This is the intended second-guard safety behavior; it is not a native-call overrun or proof that the 250 ms candidate is too small. The fixture's assumption that every attempt can start with 20 ms setup headroom was not supported under this load.

| Native cap | Started no-reply calls | Largest check-to-cleanup interval | Largest interval above cap |
| --- | ---: | ---: | ---: |
| 750 ms | 8 | 505.46 ms | None (244.54 ms below cap) |
| 900 ms | 9 | 524.42 ms | None (375.58 ms below cap) |
| 1,000 ms | 9 | 1,026.57 ms | 26.57 ms |

All 26 started calls used their full cap, returned zero replies/native timeout 11010, closed handles, and returned before their fixture deadlines. The largest observed overrun exceeds the previous unloaded 11.94 ms sample maximum, illustrating why that earlier maximum is not a worst-case bound. No post-cutoff native start or cleanup failure was observed in this matrix. This check did not exercise pending cancellation under load, a real 60-second deadline, route selection, a stalled `GetBestRoute2`, a separate standard-user account, another Windows build, or alternate networks. The matrix is not counted as passing, the 250 ms candidate remains provisional, and application infrastructure remains gated. A follow-up must distinguish acceptable safe refusals from the desired number of actually started samples while preserving this failed capture; widening a sampling window alone would not validate the margin.

## Follow-up: loaded margin sampling with safe refusals

The next opt-in check, `-LoadedMarginSamplingCheck`, kept the 250 ms candidate and the same process-local condition (two CPU-bound Go workers, `GOMAXPROCS=1`) but recorded safe post-setup refusals separately. Its fixture budget was cap + 250 ms + 50 ms; the extra 50 ms is a sampling setup window, **not** a larger proposed margin. For each of three paths and three caps, it allowed up to five attempts to obtain three actually started calls. Worker progress was observed around every attempt, the workers were joined, and the prior scheduler setting (20) was restored. The [redacted capture](go-windows-loaded-margin-sampling.capture.json) is separate from the failed 20 ms-headroom matrix.

The fixture **passed its sampling checks** on the same unelevated Windows `10.0.26200/amd64`, Go 1.27.2 environment. It took 28 attempts: one 750 ms attempt refused safely after setup (no send, zero timeout argument, handle closed), and 27 used the full native caps. All started calls returned zero replies/native timeout 11010, closed successfully, and returned before their fixture deadlines. The largest check-to-cleanup intervals were 518.14 ms at the 750 ms cap, 514.11 ms at 900 ms, and 1,012.64 ms at 1,000 ms. The largest interval above a cap in this run was **12.64 ms**. The prior failed matrix's observed 26.57 ms overrun remains part of the evidence; this new maximum does not replace it.

This gives three actual loaded samples per cap/path and demonstrates safe refusal under longer setup delay. It does not validate 250 ms for every Windows 11 x64 condition, bound native return, or exercise pending cancellation under load or a stalled `GetBestRoute2`. The milestone 1 gate remains open.

## Follow-up: pending IPv4 drain under process-local scheduler load

The opt-in `-LoadedDrainCapsCheck` ran on the same unelevated Windows `10.0.26200/amd64`, Go 1.27.2 machine. It repeated the three-cap, three-path pending-cancellation matrix with two CPU-bound Go workers and `GOMAXPROCS=1`. Both workers started before probing and progressed during every stop-to-quiescent interval. They were joined and the original scheduler setting (20) was restored. The [reviewed capture](go-windows-loaded-drain-caps.capture.json) is separate from the earlier unloaded drain capture.

All **nine** actual `IcmpSendEcho2` calls were pending when cancellation was requested about 25 ms after send. Each returned a no-reply timeout (native error 11010), then closed its handle successfully before the harness recorded quiescence. No following call was scheduled by this fixture. The observed stop-to-quiescent intervals were **220.94–458.67 ms** at the 750 ms cap, **428.38–441.18 ms** at 900 ms, and **934.32–952.56 ms** at 1,000 ms. These finite observations support retaining ownership and showing `stopping` through cleanup on this named condition. The earlier unloaded maximum of 978.25 ms remains part of the evidence; neither maximum establishes a latest-return bound.

This fixture does not validate a production run slot or UI, another Windows build/account/network, system-wide stress, or a genuinely stalled `GetBestRoute2`. The 250 ms start margin and milestone 1 gate remain unapproved.

## Saved evidence and remaining work

[Timeout investigation capture](go-windows-timeout-investigation.capture.json) contains the actual Go baseline/sweep, independent C# comparison, Go async experiment, guarded policy results and environment. No private route/source addresses are included. Original files remain in ignored `.local/`. The C# script is explicitly retained under `tools/windows-network-feasibility/debug/` for diagnosis, not as application code or replacement acceptance evidence.

[Approved-cap cutoff capture](go-windows-cutoff-cap.capture.json) contains the later fixture-only 750/900/1,000 ms observations and environment. Its source capture remains in ignored `.local/`.

[Local scheduling-margin capture](go-windows-margin-study.capture.json) contains the 45 later full-cap IPv4 observations. Its source capture remains in ignored `.local/`.

[Provisional start-margin boundary capture](go-windows-margin-boundary.capture.json) contains three pre-handle refusals and 27 full-cap calls with a 250 ms candidate allowance. Its source capture remains in ignored `.local/`.

[Loaded 60-second cutoff capture](go-windows-minute-cutoff.capture.json) contains one full-cap echo, a later refusal, and an induced route-phase crossing under process-local Go scheduler load. Its source capture remains in ignored `.local/`.

[Route-lookup timing capture](go-windows-route-latency.capture.json) contains 100 later, redacted individual `GetBestRoute2` observations and four batch totals. Its source capture remains in ignored `.local/`.

[Route-before-probe cutoff capture](go-windows-route-cutoff.capture.json) contains three pre-cutoff refusals and one simulated route-phase crossing. Its source capture remains in ignored `.local/`.

[Pending IPv4 drain capture](go-windows-drain-caps.capture.json) contains nine stopped-while-pending calls across the three configured caps. Its source capture remains in ignored `.local/`.

The default `go test ./...`, `go vet ./...`, and build checks pass; opt-in direct and async timing checks expose the unresolved overrun. No temporary untagged debug logs or service infrastructure were added. The investigation did not change networking configuration.

The [controlled live IPv4 unreachable capture](go-windows-unreachable.capture.json) now verifies that reply decoder path on this machine. Still required for the revised IPv4 milestone 1 gate: a validated cutoff/start margin and review of the route/cutoff strategy under supported conditions. The simulated route-phase crossing passed; an actually stalled `GetBestRoute2` remains unverified. Issue #10 now permits `stopping` until a pending route lookup or echo finishes and cleanup completes, without a hard drain deadline. VPN route comparison and other-Windows-build evidence remain separately unverified; an active VPN is not required by issue #10's first gate. Remote IPv6 echo/transit and pending drain are deferred to the future IPv6 release, not passed. Follow revised issue #10 rather than treating these harness samples as a feasibility pass.
