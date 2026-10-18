# Milestone 2: normalized result and probe contract

Application coding began on 2026-10-14 after reading issue #10 and the final milestone 1 development-gate review. This milestone does not pass release acceptance. The standalone harness and reviewed evidence remain separate and unchanged.

The root Go module contains `internal/result` (schema version 1) and `internal/probe` (synchronous IPv4 route/echo interface). `internal/probe/probefake` is a scripted adapter for deterministic, network-free tests. All wire durations use milliseconds; optional durations and RTTs use pointers so missing measurements stay absent and a measured zero survives serialization. Every evidence item has an ID. UTC timestamps distinguish native start, return/cleanup, run scheduling stop, and terminal quiescence. Route lookups have no fabricated bound. On-link routes have no separate gateway. ICMP reply statuses are separate from measurement outcomes and API errors.

The fake adapter implements the same full-cap start policy that the Windows adapter must use: positive integral-millisecond caps up to one second, with the empirical initial 250 ms allowance, cancellation/cutoff checks before work and again after echo setup. A refused echo has no send timestamp, address, or RTT. A started route/echo remains owned until its synchronous call and cleanup return, including after cancellation. Blocking hooks model an indefinitely pending call until the test explicitly releases it. These are fake-interface checks, not new native stall evidence or a timing bound.

Golden fixtures cover an HTTP 503 completed response, on-link gateway, raw zero RTT, silent intermediate hop followed by a reply, unavailable variation, a route returning after cutoff with an unstarted echo refusal and attempted-host trace, and explicit unsupported IPv6. Contract tests check exact round-trip field preservation, omitted values, sample counts, UTC measurements and resolvable evidence references. They establish serialization, not measurement execution or finding-rule correctness.

## Verification

Use Go 1.26 or newer, or the existing repo-local `.tools/go/bin/go.exe`. Run `go test ./...`, `go test -race ./...`, and `go vet ./...` from the root. The harness remains its own nested module and is intentionally not included by these app tests.

Local verification on 2026-10-14: root-module tests and vet passed. Race instrumentation was unavailable: this portable Windows Go environment has CGO disabled and no GCC/Clang compiler on PATH. No race-detector pass is claimed; rerun with a supported C toolchain before relying on concurrent lifecycle behavior.

## Next independently testable outcomes

Follow issue #10's small-milestone sequence: the local start/poll/cancel API and lifecycle with fake data next, then target validation/DNS pinning and pinned HTTP, then the production IPv4 Windows adapter and orchestration. Keep the early UI minimal until design resources arrive. Lifecycle must hold its active slot and show stopping while a fake route remains pending, and only emit final findings after quiescence. No hard terminal-time bound is promised.

Production adapter and packaged executable verification, security/HTTP/lifecycle/UI acceptance, and broader Windows 11/account/network conditions remain open. The 250 ms allowance is empirical; an actually stalled native `GetBestRoute2` and another Windows account remain unverified. Experimental distribution scope remains Windows 11 x64 tested on 25H2 build 26200 with an unelevated token.
