# hopvista

The [v1 specification](https://github.com/DarrylN2/hopvista/issues/10) targets an IPv4-only first release on Windows 11 x64. IPv6, Windows 10, and ARM64 support are deferred. The user-approved milestone 1 development gate [passed final verification](docs/research/ipv4-development-gate-final-review.md); application implementation is unblocked, while release acceptance remains open.

The initial 250 ms start allowance is empirical on tested conditions, with no hard native completion promise. Another Windows build/account and a genuinely stalled route lookup remain unverified. Until broader coverage exists, distribution is experimental and described as tested on Windows 11 25H2 build 26200 with an unelevated token.

Milestone 1: [standalone Go Windows networking feasibility harness](tools/windows-network-feasibility/README.md).
See the [observed results and remaining validation gates](docs/research/go-windows-network-feasibility.md).

The current [v1 acceptance status](docs/implementation/v1-acceptance-status.md) maps issue #10 to the implemented checks and remaining release work.

## Build the local app

The minimal React/TypeScript UI source is in `ui/`. Its Vite output is checked in under `internal/localapi/static/` so a Go build embeds the production UI and the executable needs no Node runtime. After changing UI source, run `npm ci` and `npm run build` from `ui/`, then build `cmd/hopvista` with Go. Run the executable and open the loopback URL it prints. The UI remains a first functional pass pending visual and human review; release acceptance is still open.
