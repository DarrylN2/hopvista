# Issue 16 participant review

Status: in progress; **partial participant feedback recorded below**. This script
does not yet establish human acceptance. Overall issue #10 release acceptance is
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

Participant: **P1, the user of this chat**, 2026-10-21 (America/Los_Angeles).
Self-reported familiarity: new to networking and HopVista. Browser: Google
Chrome (version unknown); participant reports they can use mouse and keyboard.
Exact CSS viewport and zoom have not been supplied. Six participant-provided
screenshots (3456×2040 source pixels) show the normal replay, website timing and
opened technical evidence. Image dimensions are not CSS viewport measurements,
and screenshots do not establish keyboard-only operation. It is unknown whether
the phase explanation response preceded or followed opening details. The facilitator
provided the loopback review URL, the `i11completed` replay path and questions;
no expected answers were supplied. No comprehension pass or screen-reader
participant result has been inferred. Preserve first attempts and help. For
each demonstrated failure, link a defect/regression check and record the
participant's recheck.

| Task/case | Actual words or observed action | Help given | Outcome | Recheck |
| --- | --- | --- | --- | --- |
| Familiarity | “I'm new to networking and HopVist” | Review URL and scenario supplied. | Participant scope recorded. | Not applicable. |
| `i11completed`: locating information | “I can send you pics of what the website shows. I think its okay in terms of how easy it is to find” | No content explanation supplied. | General perceived ease recorded; observation, uncertainty and selected IP have not yet been identified by the participant. | Specific readings requested. |
| `i11completed`: phase meanings | “I think most of them are clear” | No definitions or expected answers supplied. | Perceived clarity recorded; actual phase comprehension remains unverified. | Participant explanations requested, allowing “not sure”. |
| `i11completed`: phase meanings, follow-up | “For each timing, I'm not sure.” | Asked for own-word meanings; no expected definitions supplied. | Comprehension criterion not met in this attempt. | Added a visible four-step guide and simpler row explanations; participant asked to refresh and explain again. Response pending. |
| Browser and input scope | “I'm using google chrome, and I can navigate with mouse and keyboard” | No operation assistance supplied. | Browser/input familiarity recorded; this is not evidence that the specific keyboard evidence/copy tasks passed. | Specific tasks still pending. |
| Browser/direct path, missing vs zero, cutoff/completion, hop limit/unreachable and keyboard evidence/copy | No response yet. | No help yet. | Not checked. | Pending later rounds. |
