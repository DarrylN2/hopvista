# Issue #16 acceptance audit — 2026-11-06

## Authority and disposition

Audited the live [issue #16](https://github.com/DarrylN2/hopvista/issues/16)
body (no comments), [issue #17](https://github.com/DarrylN2/hopvista/issues/17)
and its candidate comment, draft [PR #27](https://github.com/DarrylN2/hopvista/pull/27),
[accessibility evidence](issue-16.md) and the complete
[participant record](issue-16-participant-review.md). This audit assesses existing
evidence; it adds no participant answers, browser checks or native measurements.

**#16 remains open for a narrow human-reading gap: the normal DNS/TCP/TLS
meanings and the finding's uncertainty have not yet been adequately explained
in the recorded answers.** P1's actual assisted participation counts. Neither
a fresh reader nor an unassisted first-use pass is required by the issue.
Earlier statements limiting first-use claims describe evidence scope, not an
acceptance gate. Earlier task-level pending notes are historical; the mapping
below controls the current disposition.

## Required criteria mapped to existing evidence

Criterion numbers follow the seven bullets in the live issue, without changing
their wording or checking off the issue itself.

| Criterion | Existing evidence | Disposition |
| --- | --- | --- |
| 1. Review supplied HTTPS redirect, website-phase, path/echo and cancellation/cutoff scenarios; preserve outcomes/design | `issue-16.md`: Baseline and scope, Independent check outcomes, and participant/UI follow-ups. Verified #11–#14 captures, 325 layout combinations including stress, retained schema/outcomes/Go findings and approved composition. P1 reviewed normal and edge-case replays. | Covered at the recorded native/simulated/replay scopes. Replay is not a new network measurement. |
| 2. Keyboard-only entry/settings/start/cancel/disclosures/phase evidence/return/Copy IPv4/errors/retries; focus/names/order/no trap | `issue-16.md`: Demonstrated defects and TDD, Independent check outcomes, Real screen-reader record. Rendered keyboard tests cover the flows, all four phase evidence targets and exact return, copy pending/success/failure, start/cancel and retry. Later Target Echo return regression adds coverage. | Covered by the recorded keyboard checks and actual NVDA navigation. P1's mouse task neither satisfies nor invalidates these checks. No requirement says P1 must repeat them. |
| 3. Real screen-reader labels, announcements, pending/final findings, path order and evidence navigation | `issue-16.md`: Real screen-reader record documents actual portable NVDA 2026.2 speech, native navigation and the corrected IP/scope separator. | Covered at that dated scope. This is real screen-reader evidence, not a synthetic accessibility tree; no blind-user or human-listening study is claimed or required. Later UI wording is not retroactively included in those speech captures. |
| 4. 200% zoom, reading order, 320px, long URLs/IDs, 1/30/64 hops and opened evidence | `issue-16.md`: Independent check outcomes records native 200% zoom and 325 browser layout combinations, with later affected layout checks retaining their own scope. | Covered at the recorded revisions/conditions; no new zoom or complete matrix rerun is established here. |
| 5. Running-browser reduced motion, non-color meaning, contrast/targets and copy failure fallback | Same outcomes table records browser reduced-motion emulation, sampled contrast/target measurements, textual states and copy checks. Synthetic rejection and browser-enforced clipboard policy both exercised fallback. | Covered at recorded scope. The issue expressly permits honest synthetic-failure evidence; it does not demand native permission denial or P1's personal clipboard/fallback check. |
| 6. Human reading: observation, uncertainty, selected IP, DNS/TCP/TLS/HTTP meanings without details; browser differences, missing/zero, cutoff/completion and hop-limit/unreachable | See the subcriterion mapping below. | Partial: normal phase meanings and uncertainty remain narrow gaps. The completed assisted readings count; do not restart them. |
| 7. Participant/reviewer scope, actual observations, fixes and affected-case rechecks; no agent/fixture as human substitute | Full participant record retains P1's Chrome/100% zoom scope, actual words, early errors, facilitator assistance/error, UI fixes and rechecks. | Record/fix/recheck process covered for the completed tasks; final affected readings for criterion 6 remain outstanding. No recruited participant is necessary. |

Accessibility evidence is dated and scoped. The latest full UI suite recorded
before this audit is 22 passes, 0 failures and 4 explicit opt-in skips (three
packaged journeys without `HOPVISTA_PACKAGED_EXE`, and system clipboard). Earlier
enabled packaged/native/NVDA checks retain their exact revision and scope. Those
skips do not erase earlier evidence, and earlier binaries do not contain later UI
changes. This audit makes no claim of a fresh comprehensive accessibility pass
over every later string.

## Human-reading subcriteria

| Required reading | Recorded response and scope | Audit result |
| --- | --- | --- |
| Locate observation | Main-finding recheck: “It returned a response and an HTTP status of 204”. Directed to the revised summary, without technical details. | Covered as a guided location/reading; knowledge of status-code classes is not required. |
| Locate and explain uncertainty | Finding-limit recheck interpreted it as missing connection times; later full-page answer was “No i don't think so? Does the connection mean that?” after correction. P1 later rejected direct-link and website-bottleneck claims, but the normal finding's own limit was not adequately explained. | Partial. Locating the panel is recorded; one normal finding-limit explanation remains needed. Do not demand an explanation of every uncertainty in every fixture. |
| Selected IP | Number-address recheck identified `1.1.1.1 public` and described its basic website-address role after a visible definition and facilitator guidance. | Covered as assisted location/basic meaning. No additional public/private-address theory test is required. |
| DNS/TCP/TLS meanings | Initial phase attempts: “For each timing, I'm not sure”; then “DNS finds and then establishes TCP makes the connection, TLS, secures the connection, HTTP is a protocol. I don't really know.” Later value readings and certificate-row outcomes did not complete the normal phase-meaning recheck. | Partial. TCP/TLS have basic partial explanations; DNS remains unclear. One bundled explanation of the three normal steps will settle the affected reading without repeating all scenarios. |
| HTTP meaning | Response-begins meaning recheck: “The time between our request and website's response”; earlier assisted full-page answer rejects page-load equivalence. | Covered narrowly for basic response-start timing in context of the visible label and prior help. No first-byte terminology quiz is required; the remaining uncertainty answer can clarify the page-load limit. |
| Direct-path/browser differences | First rejected guaranteed matching; after revised reuse wording, “It might use the saved connection?” identifies an existing connection as the concrete difference. | Covered narrowly as assisted understanding of one valid reason. Explaining every alternative address/proxy/cache mechanism is optional. No actual Chrome route observation is required. |
| Missing versus zero | Timeout recheck plus measured-zero comparison; aligned reply-time recheck correctly says “Not measured” for Hop 1 and “450 ms” for Hop 2. | Covered as assisted distinction. Preserve earlier repeated zero mistakes and the limitation that recorded 0 ms is not literally instantaneous page loading. |
| Cutoff versus completion | “Completed means that the process is done, but maybe not all metrics have a value”; final cutoff answer states no new checks, underway checks end with or without a reply, another run after finish. | Covered as assisted reading. Preserve the mismatched cutoff prompt as facilitator error. A live cleanup transition, ETA or drain bound is not an additional human-reading requirement. |
| Hop limit versus observed unreachable | One-hop stop attributed to configured limit; revised unreachable reading: “HopVista got a reply from a device saying the target address was unreachable.” | Covered as guided/assisted comparison; earlier silence misunderstanding remains recorded. |

The record's additional dotted-link, high-RTT, redirect/certificate and
mouse evidence/copy tasks remain useful evidence with their exact scope. A
quoted visible sentence establishes an accurate reading; it does not automatically
prove deeper understanding. Conversely, the issue does not require independent
expert explanations of every technical term. P1's mouse return and visible
“Copied” feedback count for those observations; human keyboard-only operation,
clipboard contents and manual fallback remain unperformed observations, not new
acceptance blockers.

## Smallest remaining human step

Use the same P1 and one normal `i11completed` replay, with technical details
closed and the visible explanations available. Ask this single bundled prompt:

> In Website timing, what are Address lookup (DNS), Connection (TCP) and
> Encryption (TLS) each timing? Then, using “What this does not tell us”, what
> can this run still not tell you about loading the whole page in Chrome or
> what caused a wait? Short answers are enough; say “not sure” where needed.

Record the actual response once, with all earlier assistance disclosed. This is
an affected-case recheck, not a new-reader or first-use test. Do not supply an
expected answer before collecting it. If it demonstrates those meanings and
limits, reconcile #16 without repeating completed readings or P1's keyboard task.
If a specific meaning is still unclear, address that demonstrated gap only.
No participant recheck was conducted during this audit.

## Repository, PR and release state

At audit start the local branch `accessibility-human-acceptance` was clean at
`80192ff5a98a9e477eebeb25ee75862e47a8cc04`. GitHub branch metadata independently
matched that tip and `main` at `b67eaee6e27a73085c864d7c090a122c85ab96ad`;
the old `codex/accessibility-human-acceptance` branch was absent. PR #27 was open,
draft, unmerged and mergeable against main. `gh` was unavailable and shell Git
remote access failed, so the connected GitHub tools supplied live issue/PR/ref
state. No branch rename, merge or release action is authorized by this audit.

#17 still requires a candidate containing completed blocking work, checks of
that exact candidate, matrix/instruction reconciliation and explicit human
release approval. Historical `a159f9c` is a pre-acceptance candidate and predates
the participant-driven UI improvements. Do not relabel it as containing them.
Keep #16, #17 and release acceptance open for the actual remaining criteria;
preparing evidence does not authorize publication.

## Audit validation

Documentation-only change; no production assets or candidate binary changed.
`git diff --check` and TypeScript checking passed. The full UI suite's first
sandbox attempt failed in the shared Edge launch hook (22 hook failures,
4 skips), before browser assertions could run. An authorized isolated headless
retry outside that sandbox passed **22 tests, 0 failures, 4 opt-in skips**:
three packaged journeys without `HOPVISTA_PACKAGED_EXE` and system clipboard.
No existing desktop browser, clipboard or screen reader was operated. This
retry is UI regression evidence, not new human or native acceptance evidence.

Independent code-review of `git diff 80192ff...4b9d5f6`:

### Standards

**0 findings.** The documentation follows issue-source and domain-language
rules, preserves participant history, and links current dispositions to one
audit. No applicable code-smell findings. Reviewer made no edits or test runs.

### Spec

**0 actionable findings.** The seven-criterion mapping and narrow remaining
phase-meaning/uncertainty recheck follow the actual evidence. Assisted readings
count without invented participant/input/live-cleanup gates. Exact-candidate
requirements and human release approval remain separate. Reviewer performed no
tests or desktop/browser operations.

Review summary: Standards **0 findings**; Spec **0 findings**. The subsequent
validation/review appendix records outcomes only and changes no disposition.
