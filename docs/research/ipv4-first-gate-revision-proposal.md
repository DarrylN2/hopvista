# Approved milestone 1 gate revision

Approved explicitly by the user on 2026-10-14: “I approve.” Applied to issue #10. **The empirical 250 ms initial policy is approved on tested development conditions; [final verification passed](ipv4-development-gate-final-review.md). Release acceptance remains separate.**

## Approved decision

Make milestone 1 a development-feasibility gate based on actual Go execution on the available Windows 11 x64 machine. Separate the decision to begin application implementation from claims of broad Windows 11 compatibility or release readiness. Keep Windows 11 x64 and IPv4 as the first-release targets.

The current [issue #10](https://github.com/DarrylN2/hopvista/issues/10) already speaks of an ordinary unelevated process and tested Windows conditions. A filtered administrator token is unelevated; it does not establish behavior under a separate standard-user account. The existing audit made additional account/build coverage a blocker to beginning application work. This proposal explicitly permits development with those coverage limits recorded.

## Replacement first-gate requirements

Before application infrastructure, execute and review the standalone actual-Go IPv4 harness on the available Windows 11 x64 machine. Record OS/build, architecture, Go version, token elevation, API arguments, statuses, durations, and cleanup. The tested process must be unelevated; native access must succeed outside the restricted command sandbox. No administrator-only API success may substitute.

Required local evidence remains:

- Local and remote IPv4 echo; low-TTL replies; continued probing after silence; a genuine unreachable reply distinguished from API failure.
- Destination-specific route source/interface/next hop and on-link behavior compared with Windows route information.
- Correct layouts, reply counts/statuses, timeout units, errors, and buffer/handle ownership through native completion.
- Full configured 750, 900, and 1,000 ms caps, insufficient-budget refusal before send, a real 60-second cutoff fixture, and pending-call completion/cleanup under the named unloaded and process-loaded conditions.
- Refusal after route/setup delay consumes the budget, with scheduling stop and quiescence recorded separately. A safe refusal is an expected guard result, not a native completion sample. Preserve failed fixtures and unexpected outcomes.

**Approved initial start allowance: 250 ms.** Accepted as an empirical development policy on the tested conditions, rather than a universal worst-case proof. The largest previously recorded loaded check-to-cleanup overrun above a cap is 26.57 ms; this sample does not establish a maximum. The allowance must be checked after route selection and immediately before native send. Guarded calls use the full cap or remain unstarted; they must not trim the cap. Preserve the remaining scheduler window between the final check and native API entry as a known limitation. This policy must be retested when the production adapter is integrated or its scheduling behavior changes.

**Route-delay acceptance:** actual route selection plus explicitly simulated route-phase cutoff crossing is sufficient for the development gate's ordering check. A genuinely stalled `GetBestRoute2` remains unverified and is not counted as passing. Its undocumented maximum duration and cancellation behavior remain accepted uncertainties for development. Started route/echo work retains ownership until it returns and cleanup finishes. The future application stays `stopping` and holds its active-run slot throughout, with no hard latest-return promise. An indefinitely pending fake route call must test that lifecycle behavior in the later coordinator milestone; it cannot establish native completion behavior.

**Coverage limits:** another Windows build, a separate standard-user account, alternate networks/VPN, and system-wide stress are not prerequisites for beginning application development. They remain unverified and must appear in support documentation and any evidence review. Repeating tests on this machine cannot pass those conditions.

## Release boundary

Application implementation and release acceptance remain separate. The production Windows adapter and packaged executable must pass the applicable local live checks and all required deterministic security, HTTP, lifecycle, results, and UI checks before distribution. The standalone harness is not that evidence.

Until broader Windows/account coverage is obtained, describe any distributed build as an **experimental Windows 11 x64 release, tested on version 25H2 build 26200 with an unelevated token**. Other builds and separate standard-user accounts remain unverified. Do not claim comprehensive Windows 11 support. Windows 10, ARM64, and IPv6 remain outside the first release.

No hard 60-second terminal response, hard native drain time, atomic check-to-send guarantee, or demonstrated recovery from a genuine native route stall is added by this revision. Preserve the approved `Stopping…` behavior and the distinction between a scheduling cutoff and native completion.

## Applying an approved revision

1. Record the explicit user decision in issue #10 and reconcile its timing, first-gate, release, and milestone wording with this revision. Keep the evidence reports' historical failures and limitations.
2. Run a final verification against the current harness source after the guard hardening: tests/vet/build, the full-cap boundary and loaded sampling checks, pending drain, route/cutoff, and the real minute fixture. Evaluate actual capture contents; an exit code alone cannot pass the gate. Reuse unchanged live decoding and independent route evidence, with its recorded scope.
3. Review whether every revised first-gate requirement is satisfied. Only then record the development gate as passed, with remaining unverified cases. A failed required check still blocks it.
4. Tell the user clearly that application coding is beginning before adding application code. Prioritize measurement engine, results, local API, lifecycle, and tests; keep the early UI minimal and usable.

Approval authorizes this revised gate and the empirical 250 ms initial policy; it does not itself pass final verification or the release gate.
