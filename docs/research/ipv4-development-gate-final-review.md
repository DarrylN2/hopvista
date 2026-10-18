# Milestone 1 final development-gate review

**Decision on 2026-10-14: passed under the explicitly approved revised development gate. Release acceptance is not passed.** The user approved the [gate revision](ipv4-first-gate-revision-proposal.md) with “I approve.” [Issue #10](https://github.com/DarrylN2/hopvista/issues/10) records the change. Application implementation may proceed; no application code was added during verification.

The initial 250 ms allowance is accepted empirically on named tested conditions, without a universal timing bound or comprehensive Windows support claim.

## Final execution and review

Current hardened Go source ran unelevated on Windows `10.0.26200/amd64` (Windows 11 25H2), Go 1.27.2, outside the restricted command sandbox. Invocation: `run-gate-recheck.ps1 -ConditionLabel approved-development-gate-final`. Raw logs/captures remain private in `tools/windows-network-feasibility/.local/gate-recheck-20261014T194457Z`.

The runner's parameter forwarding was corrected to use named-parameter hashtable splatting. The final suite executed all eight intended steps and saved fresh captures. The [review JSON](ipv4-development-gate-final/review.json) records redacted base observations, independent route comparison, step times/status, limitations and source SHA-256 hashes. Fixture `gate` strings remain disclaimers; this reviewed decision comes from the approved criteria, not an exit code alone.

| Check | Reviewed result |
| --- | --- |
| Tests, vet, build, layouts/error handling | Passed on current source in the base step. |
| Live IPv4 echo and TTL | Localhost, next-hop and remote echo replied. TTL 1/2 returned time-exceeded, TTL 3 was silent, and TTL 4 replied; handles closed. |
| Independent route comparison | Four IPv4 routes matched Windows source/interface and next-hop/interface, including on-link cases. Two query errors concern deferred IPv6, not IPv4. |
| Genuine unreachable reply | Reused unchanged [live capture](go-windows-unreachable.capture.json): count 1, status 11003, native error 0, correct normalization, successful close. |
| [Unloaded boundary](ipv4-development-gate-final/margin-boundary.json) | Three short budgets refused; 27 full-cap calls closed before fixture deadlines. Largest interval above cap: 5.35 ms. |
| [Loaded sampling](ipv4-development-gate-final/loaded-margin-sampling.json) | 27 full-cap calls (three per cap/path), no refusals or deadline crossings in this run; all closed. Largest overrun: 25.68 ms. Two workers, `GOMAXPROCS=1`, progress recorded. |
| [Unloaded drain](ipv4-development-gate-final/drain-caps.json) | Nine pending calls at stop returned timeout 11010 and closed. Largest stop-to-quiescent: 976.77 ms. |
| [Loaded drain](ipv4-development-gate-final/loaded-drain-caps.json) | Nine pending calls returned/closed with load progress. Largest interval: 956.42 ms. |
| [Setup refusal](ipv4-development-gate-final/guard-delay.json) | Induced 80 ms setup delay: no send, handle closed. Its 50 ms diagnostic allowance is not the adopted 250 ms policy. |
| [Route-phase cutoff](ipv4-development-gate-final/route-cutoff.json) | Three post-route refusals and one simulated phase crossing; no echo started. Delay follows lookup, not a native stall. |
| [Loaded real minute](ipv4-development-gate-final/minute-cutoff.json) | Full 1,000 ms call completed/closed before cutoff; later attempt and simulated route crossing refused. Quiescence 21.11 ms after observed stop; workers joined and scheduler setting restored. |

The base includes intentionally unguarded short-timeout diagnostics, which do not validate the old remaining-timeout rule. Approved guarded calls use full caps or refuse without sending.

## Preserved limits and next milestone

The earlier failed loaded boundary matrix remains failed, including its safe setup refusal and 26.57 ms overrun. The earlier 978.25 ms drain remains larger than the final samples. Neither maximum is a bound; no failure was erased or reclassified.

A genuinely stalled `GetBestRoute2`, separate standard-user account, other builds, alternate networks/VPN and system-wide stress remain unverified. The last clock check and native entry are not atomic under arbitrary scheduling pauses. Started native work retains ownership until return and cleanup; the future application remains `stopping` and holds its slot throughout, with no hard latest terminal time.

Next is milestone 2: normalized results, route/echo interface, deterministic fake adapter and JSON fixtures. Announce application coding before adding it. Later lifecycle tests must exercise an indefinitely pending fake route, without claiming native completion evidence. Production adapter/packaged-executable live verification and security/HTTP/lifecycle/UI release checks remain required. Distribution remains experimental Windows 11 x64, tested on 25H2 build 26200 with an unelevated token, until broader coverage exists. Windows 10, ARM64 and IPv6 remain outside v1.
