# IPv4 first-gate recheck on another Windows condition

**Current status, 2026-10-14:** the revised empirical development gate [passed on the available machine](ipv4-development-gate-final-review.md). This kit remains useful for future coverage checks; another account/build is unverified, not a prerequisite for beginning development under the approved revision. Prior open-gate language below records the earlier requirement.

The [milestone 1 gate](ipv4-milestone-1-gate-audit.md) remains open. This procedure gathers actual Go evidence on a **separate ordinary standard-user account** or another **Windows 11 x64 build**. It does not validate a production margin by itself and does not authorize application infrastructure.

The current machine's personal accounts are members of Administrators. The current Go evidence comes from a filtered, unelevated token for one of them. The separate Codex sandbox accounts are restricted and have produced `ACCESS_DENIED` for native handles, so they do not substitute for an ordinary standard-user session. No account was created or changed for this check.

## Run on the target condition

Use an ordinary, **unelevated** PowerShell session on Windows 11 x64. On a standard-user account on this machine, place the harness in a folder that account can read and write. On another machine, copy or extract the harness source with its `tools/windows-network-feasibility/` directory layout. The portable kit at `tools/windows-network-feasibility/.local/hopvista-gate-recheck.zip` contains the harness source and scripts, but no Go runtime, private captures, or account credentials. Go 1.26 or later must be available on that account, either on PATH or passed as `-Go`.

From the extracted kit root or repository root:

```powershell
pwsh -NoProfile -File tools/windows-network-feasibility/run-gate-recheck.ps1 `
  -ConditionLabel 'standard-user-win11-x64-home' `
  -Go 'C:\path\to\go.exe'
```

Omit `-Go` if Go is on PATH or the verified portable Go archive is under `.tools/go/bin/go.exe` at that root. Use `-IPv4 <reachable-public-IPv4>` only if the default `1.1.1.1` cannot be probed from that network. Run `-PlanOnly` first to list steps without probing. The `ConditionLabel` is a descriptive claim supplied by the runner; account type must be verified separately by the person operating the Windows session.

The script runs the base route/echo capture, unloaded 250 ms margin boundary, loaded full-cap sampling, loaded pending-call drain, route/cutoff ordering, and the loaded real 60-second cutoff. It keeps each step's log and fresh raw JSON in a timestamped `tools/windows-network-feasibility/.local/gate-recheck-*` folder, then writes `manifest.json`. The manifest records the account name, OS version and architecture, step statuses, and which captures were actually saved. Any failed or missing step leaves the run incomplete; individual JSON files must still be inspected for whether the required response, no-reply prerequisite, and cleanup actually occurred. The runner does not turn its status into a gate decision.

**Privacy:** `go-capture.json` and `windows-routes.json` can include local source addresses, gateway, interface, route, and account details. Keep the raw folder private. Review and redact copies before sharing or committing; preserve original measurements, statuses, and timing fields. Redaction must not fabricate a reply or successful step.

## What this can and cannot settle

An ordinary standard-user run tests native access without relying on a filtered administrator token. A run on another Windows 11 x64 build expands build coverage. Both are needed if the first-release claim is meant to cover those conditions. A failure on either is evidence to investigate, not a reason to discard the capture.

The route/cutoff check inserts its delay **after** a real `GetBestRoute2` call; it cannot create or prove safe cleanup of a genuinely stalled lookup. The script also does not establish a hard latest-return time, a system-wide load envelope, VPN behavior, or production UI/run-slot behavior. The approved contract keeps `Stopping…` visible until started native work and cleanup finish, without a hard drain-time promise. The 250 ms margin remains provisional until the evidence is reviewed against the chosen support conditions.
