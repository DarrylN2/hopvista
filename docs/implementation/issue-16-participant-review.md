# Issue 16 participant review

Status: prepared; **no participant observations recorded yet**. This script
does not establish human acceptance. Overall issue #10 release acceptance is
separate. Use the current production assets and the verified API captures from
issues #11–#14, with their documented native versus simulated provenance.

## Setup and reviewer scope

Run `node tools/accessibility-fixtures.mjs`, set `HOPVISTA_JOURNEY_CAPTURE` to
the absolute `.tools/issue-16` directory, then run
`node tools/https-journey-preview.mjs`. Open the printed loopback URL. The
persistent banner labels replay. Enter `https://fixture.example/` followed by
the named mode below in **Target URL**, then Start diagnostic run. Reload
between cases, particularly running/stopping replays, which remain static.
No request is sent to fixture.example.

Ask the participant to use their own words and think aloud. Do not explain the
expected answer first. Record their first response, difficulty locating content,
any facilitator help, and their answer after help separately. A repo owner or
developer can participate, but record their prior knowledge; that alone does
not establish comprehension for an unfamiliar nontechnical reader. An agent,
API fixture, expected answer or unreturned questionnaire is not a participant.

Record: date/time; participant's chosen identifier and role; familiarity with
HopVista, networking and these fixtures; browser/version, viewport/zoom and
input method; assistive technology/version if used; facilitator; case IDs;
actual words/observations; pass, difficulty, failure or not checked per task.
Do not record unnecessary personal information.

## Tasks (show prompts, keep the assessment notes for the facilitator)

| Case | Participant prompt | Facilitator assessment |
| --- | --- | --- |
| `i11completed` | Without opening details, point to the observation, its uncertainty and the selected IP. Explain DNS, TCP, TLS and HTTP timing. What part of loading a page is measured? | Finds observation/limit/IP and phase explanations. First response-header byte is distinct from body download/rendering. Exchanges remain separate. |
| Same case | Would this have to match your browser? What do the dotted path lines establish? | Can explain another browser IP/proxy/cached DNS/existing connection and probe order rather than physical links or the HTTP route. |
| `i12tcptimeout` | Which measurement did not finish? Does its missing duration mean zero? What earlier evidence remains useful? | Distinguishes timeout/unavailable from measured zero, retaining earlier exchanges. |
| `i12measuredzero` and `i13highsilent` | Compare the zero reading, missing RTT and 450 ms intermediate reply followed by a lower reply. What can you conclude? | Measured zero differs from no reply. Does not infer a bottleneck, packet-loss rate or website failure from intermediate ICMP. |
| `i13limit1` and `i13unreachable` | Did these stop for the same reason? Show what supports your answer. | Hop limit is a configured stopping rule; unreachable is an observed response. Successful HTTP can coexist with echo misses. |
| `i14pausedminute` | Is a completed run proof that every measurement completed? Explain cutoff and when cleanup ends. | Separates lifecycle completed from measurement deadline; cutoff stops new work and cleanup may finish later. Explain that this captured native fixture used an induced whole-process pause. |
| `i14cutoffstopping` and `i14cutoff` | What is pending? Can another run start? What changes when work finishes? | Findings pending while stopping, retained evidence, busy slot/no ETA until cleanup; terminal controls become available. Static replay is not a live timing check. |
| `i11sensitive` and `i11certificate` | Explain why the journey stopped, which address/path was used and what is known about later phases. | Distinguishes deliberate redirect policy from network failure; retained TCP peer from failed TLS; no invented HTTP duration/final host. |
| Any completed case | Using only keyboard, inspect HTTP evidence and return to where you came from. Copy the selected IPv4; if unavailable, show how you would copy manually. | Records actual focus/name/navigation experience. No facilitator mouse assistance counted as keyboard success. |

## Observation record

Participant: **pending**. No answers, comprehension pass or screen-reader
participant result has been inferred. Fill the record from the participant's
actual responses, preserving first attempts and help. For each demonstrated
failure, link a defect/regression check and record the participant's recheck.

| Task/case | Actual words or observed action | Help given | Outcome | Recheck |
| --- | --- | --- | --- | --- |
| Pending participant | — | — | Not checked | — |
