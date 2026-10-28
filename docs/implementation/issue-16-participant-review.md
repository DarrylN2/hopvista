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
no expected answers were supplied before the two phase attempts below. After
the revised-copy attempt, the facilitator supplied definitions and explicitly
excluded response-body download and page display. No comprehension pass or screen-reader
participant result has been inferred. Preserve first attempts and help. For
each demonstrated failure, link a defect/regression check and record the
participant's recheck.

| Task/case | Actual words or observed action | Help given | Outcome | Recheck |
| --- | --- | --- | --- | --- |
| Familiarity | “I'm new to networking and HopVist” | Review URL and scenario supplied. | Participant scope recorded. | Not applicable. |
| `i11completed`: locating information | “I can send you pics of what the website shows. I think its okay in terms of how easy it is to find” | No content explanation supplied. | General perceived ease recorded; observation, uncertainty and selected IP have not yet been identified by the participant. | Specific readings requested. |
| `i11completed`: phase meanings | “I think most of them are clear” | No definitions or expected answers supplied. | Perceived clarity recorded; actual phase comprehension remains unverified. | Participant explanations requested, allowing “not sure”. |
| `i11completed`: phase meanings, follow-up | “For each timing, I'm not sure.” | Asked for own-word meanings; no expected definitions supplied. | Comprehension criterion not met in this attempt. | Added a visible four-step guide and simpler row explanations; actual recheck below. |
| `i11completed`: revised-copy phase recheck | “Okay, so I don't really know, but I think DNS. DNS finds and then establishes TCP makes the connection, TLS, secures the connection, HTTP is a protocol. I don't really know. I also don't know if these timings include downloading and displaying the whole page.” | Asked to refresh, run the same case and read updated timing guidance. No definitions supplied before this answer. Refresh/version was not independently observed. | Partial TCP/TLS understanding; DNS boundary unclear, HTTP timing boundary not explained, full-page exclusion unknown. Revised wording has not established comprehension acceptance. | Facilitator then explained address lookup, connection opening, encryption/certificate setup and request-to-first-response-byte timing, excluding body download and page display. Any subsequent recheck must be labelled assisted; no assisted answer yet. |
| Browser and input scope | “I'm using google chrome, and I can navigate with mouse and keyboard” | No operation assistance supplied. | Browser/input familiarity recorded; this is not evidence that the specific keyboard evidence/copy tasks passed. | Specific tasks still pending. |
| Browser/direct path, missing vs zero, cutoff/completion, hop limit/unreachable and keyboard evidence/copy | No response yet. | No help yet. | Not checked. | Pending later rounds. |

## Resumed review: 2026-10-27

P1 explicitly resumed issue #16 in this chat: “ALright lets do 16”. The earlier
deferral is lifted for this review. Participant is the same repo owner/user;
the earlier networking explanations and subsequent personal use of HopVista
are prior exposure. Phase-meaning rechecks must retain that assisted history;
they cannot establish first-use comprehension for a new participant.

Review baseline: clean `b67eaee` (PR #26 merged); production UI source and
embedded assets are unchanged from candidate source `a159f9c`. Required local
captures were checked against `.tools/issue-16/manifest.json` SHA-256 entries.
The loopback replay listener was started and its HTTP 200 response/replay
banner verified. The participant opens and operates the browser themselves;
no desktop or existing browser is controlled by the agent. This uses fixed
captured examples, not fresh measurements of the participant's network.

The first prompt asks P1 to locate the observation, uncertainty and selected
IPv4 in `i11completed` without opening technical details, and to report their
browser/zoom. No expected answer was supplied. Initial socket-restricted server
readiness failed; the disposable listener was stopped and restarted with
permitted loopback access. HTTP 200 and the replay banner were then verified.

P1 reports **Google Chrome at 100% zoom**; version and exact CSS viewport are
unknown. Ambient in-app browser state is not substituted for this self-report.
No specific keyboard-only operation is reported in this round.

| Task | Actual response | Assessment / follow-up |
| --- | --- | --- |
| Main observation | “I think it observes the links or hops or connections between my computer and the fixture link.” | P1 described the path concept, but did not identify the particular returned response/finding in this example. The prompt was broad, so this is not evidence of an incorrect network diagnosis; locating/interpreting the intended main observation remains unverified. |
| Uncertainty | “I think the language here is still too technical, even for me. Maybe like simplify the language or dumb it down a little bit.” | Explicit readability difficulty; no interpretation of the limit supplied. Simplify the primary explanations and labels, then recheck. |
| Selected address | “I don't know what you mean by selected. Is is the \"https://start.example/start\"? What is even IPv4?” | The participant could not distinguish the website link from the selected numeric address; IPv4 is undefined for this reader. This criterion is not met in this attempt. Add a visible definition and plain address label without changing the recorded selection. |

Response to this feedback: the agent acknowledged a wording gap without giving
the expected numeric address or finding. The UI follow-up explains the purpose
in everyday words, defines IPv4 and its difference from the entered link,
renames the summary/address/limit headings, and simplifies existing static
finding guidance. Exact Go-owned observation/limit strings remain visible and
unchanged, with the same measurements, schema and evidence references. This
records a readability fix whose human recheck begins below. Results are scoped
to the particular task and assistance actually supplied.

Facilitator follow-up: the participant is told that IPv4 is a number address
computers use to reach a website, and asked to refresh/rerun the example and
find the address using the updated screen. This definition is assistance and
must remain part of the recheck scope. No expected numeric address is supplied.

### Number-address recheck

Prompt: refresh/rerun `i11completed`, look at “Address for path checks (IPv4)”,
and say which number appears and what it is used for. This directs attention to
the revised label. The visible UI and facilitator both supplied the IPv4
definition before this answer; the facilitator did not supply `1.1.1.1`.

Actual response: “It shows 1.1.1.1 public , and i think that basically that
number is like an address for a website”.

Assessment: **assisted recheck passed for locating the displayed number address
and describing its basic role as a website address**. The reported number and
public scope match the example's path selection. This answer does not establish
understanding of public/private scope, redirects or multiple addresses, and is
not first-use comprehension by an unfamiliar participant. The refresh and
browser operation were participant-performed and not independently observed.
Main finding, uncertainty, phase meanings and the other outstanding tasks remain
open. The next prompt asks what happened when HopVista contacted the website,
using the “What this run found” panel, without supplying the expected finding.

### Main-finding recheck

Prompt: “Now look at ‘What this run found.’ In your own words, what happened
when HopVista contacted the website?” No expected response or status was
supplied in this prompt.

Actual response: “It returned a response and an HTTP status of 204”.

Assessment: **guided recheck passed for identifying the recorded website
response and status 204** in `i11completed`. The answer matches the final
exchange and Go-owned finding in this captured example. The participant was
directed to the summary panel after the wording fix; this is not an unassisted
first-use finding-location check. It does not establish an understanding of
status-code classes, dominant-wait rules, overall page success, browser timings
or the finding's limits. No new live network observation is inferred from this
replay. The next prompt asks the participant to explain the limits in their own
words, without providing the expected explanation first.

### Finding-limit recheck

Prompt: “Now read ‘What this does not tell us.’ In your own words, what are we
still unable to conclude from this run?” No expected limit or timing answer was
supplied before the response.

Actual response: “It doesn't give us the timings or like how slow/fast it needed
to connect to the website?”

Assessment: **not yet passed for the finding's limit**. P1 interpreted the
limit as an absence of connection timing. This example does show measured
steps, including connection timing, so the answer reveals a distinction the
current wording did not convey to this reader: direct request timings versus
browser/full-page timing and the cause of waits. The question mark also leaves
some uncertainty about P1's confidence. Do not score this as a claim that the
whole website was fast or slow. The facilitator may explain the distinction and
ask a narrower assisted recheck; any later success must retain that assistance.

### Connection-timing recheck

Facilitator correction before this prompt: HopVista shows some measured times,
including connection time; the limit concerns full-page timing in Chrome and
the cause of a wait. Prompt: find the Connection row for the last website
request in Website timing and report its time.

Actual response: “0 ms? There are 4 exchanges”.

Assessment: **assisted recheck passed for locating the recorded connection
value, with residual uncertainty about which exchange to use**. The fourth
exchange is the final request (`https://next.example/final`), and its TCP
connection duration is recorded and displayed as 0 ms. P1 correctly noticed
four exchanges; the prompt said “last website request” without explaining
that the fourth exchange is the final one. The displayed 0 ms is a recorded
value in this controlled replay; it does not show that real-world connection
time is literally zero or that Chrome loaded the page instantly. The answer
does not by itself establish the finding's browser-timing or causal limits.

### Full-page timing recheck

Prompt after the facilitator identified Exchange 4 and explained that its
0 ms Connection value is one step of one website request: “Does that 0 ms tell
you how long Chrome took to load and display the whole page?”

Actual response: “No i don't think so? Does the connection mean that?”

Assessment: **assisted recheck passed narrowly for rejecting a full-page
timing inference**, with explicit uncertainty about what Connection means.
The answer follows a direct yes/no prompt and prior facilitator explanation;
it is not unassisted first-use comprehension. It does not establish an
understanding of the connection step, the cause-of-wait limit, or browser
versus direct-path differences. The facilitator should define Connection in
plain language before a later task.

### Response-begins timing prompt

Facilitator explanation: Connection is the time HopVista recorded for opening
a connection to the website's number address; it does not measure Chrome
loading or displaying the page. Prompt: “In Exchange 4, what time do you see
beside Response begins?”

Actual response: “the encryption,  connection, and address lookup”.

Assessment: **the requested Response begins value was not identified**. P1
named the other three displayed timing steps instead of a time. This may
reflect difficulty finding the fourth row or ambiguity in the facilitator's
prompt; it is not evidence that P1 misunderstood the value itself. Rephrase
with the exact row labels and ask again. Do not score phase-meaning comprehension
from this response.

### Response-begins value recheck

Facilitator follow-up identified the fourth row below Encryption in Exchange 4
as “Response begins” / “HTTP first byte” and asked for its displayed number;
the expected value was not supplied.

Actual response: “0 ms”.

Assessment: **assisted recheck passed for locating the displayed Response
begins value**. Exchange 4's HTTP-first-byte duration is recorded as 0 ms in
this controlled replay. The facilitator pointed to the precise row after an
earlier unsuccessful prompt. This does not establish unassisted finding of the
row, understanding of what starts at that point, full-page speed, or a causal
explanation. The 0 ms fixture value must not be generalized to real traffic.

### Response-begins meaning recheck

Prompt: “In your own words, what does Response begins measure? It’s fine to
use the short explanation beside that row.” The label and inline help were
available, and previous facilitator prompts had identified the row.

Actual response: “The time between our request and website's response”.

Assessment: **assisted recheck passed for the basic request-to-response wait**.
P1 described the interval in ordinary words. The visible label “Response
begins” supplies the start-of-response boundary, while the answer itself does
not explicitly distinguish the first response byte from the completed body
or rendered page. Prior explanation and row guidance mean this is not
unassisted first-use phase comprehension. Continue checking incomplete versus
zero timing in another example; do not infer whole-page timing from this answer.

### Incomplete-connection value check

Prompt: enter `https://fixture.example/i12tcptimeout`, start the replayed run,
and report what Connection shows in Exchange 2. The expected outcome and
number were not supplied in the prompt.

Actual response: “No response within 2000 ms”.

Assessment: **guided check passed for locating the displayed timeout text**.
It matches Exchange 2's connection timeout and 2000 ms bound in this captured
example. The participant was told the exact exchange and row. This answer
does not yet establish the meaning of timeout, missing versus zero duration,
or which earlier timings remain useful. Ask for an interpretation next.

### Incomplete-connection meaning check

Prompt: “What do you think ‘No response within 2000 ms’ means for that
Connection step?” No expected interpretation was supplied before the answer.

Actual response: “Opening a connection failed?”

Assessment: **partial understanding**. P1 associated the message with the
connection attempt rather than a full-page result, and expressed uncertainty.
The captured outcome is specifically a timeout: the connection did not
complete within the 2000 ms bound. This does not prove a permanent inability
to connect or say what a later attempt would do. The facilitator then
clarified that distinction. A later check should test whether P1 distinguishes
this missing completed duration from a recorded 0 ms duration.

### Missing-versus-zero check

Prompt: “For Exchange 2, would you read the Connection time as 0 ms, or as no
completed time recorded?” The facilitator had already explained that this
attempt timed out after a two-second limit.

Actual response: “0”.

Assessment: **not passed**. P1 treated the timed-out connection as a recorded
zero-duration connection despite reading the timeout text correctly and
receiving an explanation. This is direct evidence that the result screen does
not yet make missing completed time distinct enough from measured zero for
this participant. Clarify the visible timeout label and incomplete finding
context without changing the recorded outcome, bound, schema or Go finding.
Recheck after the wording change; any success remains assisted.

### Timeout wording human recheck

After the UI began displaying “Did not finish within 2000 ms (no completed
time)” for Exchange 2's timed-out Connection step, P1 was asked to refresh
Chrome, rerun `i12tcptimeout`, and explain what that row says about its
recorded time. This follows the facilitator's earlier explicit correction
that a timeout has no completed time; the browser refresh/run was
participant-performed and not independently observed by the agent.

Actual response: “It ran out of time, it didn't finish under 2 seconds”.

Assessment: **assisted recheck passed for the bounded timeout meaning**. P1
now describes a step that did not finish within the two-second limit, rather
than a completed 0 ms connection. The answer does not literally say “no
completed time”, so the exact missing-versus-measured-zero distinction should
still be checked against a separate completed-zero example. This is not a
new participant or unassisted first-use pass. No new network measurement is
inferred from the replay. Other issue #16 tasks remain open.

### Measured-zero comparison

Prompt: run `https://fixture.example/i12measuredzero`; in Exchange 2, report
what Response begins shows and how it differs from the previous Connection
timeout. The expected value was not supplied, but the participant had already
been coached on completed versus missing time.

Actual response: “It shows 0 ms and it;s different cause the response is
immediate”.

Assessment: **assisted recheck passed for distinguishing recorded zero from
timeout/missing completed time**. Exchange 2's response-start phase is
completed with duration 0 ms in this controlled capture; the previous
connection phase timed out without a completed duration. The word “immediate”
overstates what a recorded 0 ms proves, particularly outside this fixture and
for a whole page. The facilitator clarified that 0 ms is the recorded value
for the start of the response, not proof of literally instantaneous delivery
or page rendering. This comparison does not establish first-use comprehension
or other acceptance criteria.

### Direct connection versus Chrome

Prompt on the `i12measuredzero` result: read “A direct connection” and say
whether HopVista's timing would necessarily match Chrome's timing for the same
website, with a reason. No expected reason was supplied before the answer.

Actual response: “I think it wont match chrome's timing cause its from my
computer?”

Assessment: **partial understanding**. P1 correctly rejects guaranteed
matching timings but gives “from my computer” as the reason. Both HopVista and
Chrome run on the participant's computer; that phrase alone does not explain
the difference. The visible caveat's direct-versus-browser distinctions
(another address, intermediary/proxy, saved address answer or reused
connection) were not expressed. The facilitator should explain one or two
concrete differences and recheck as assisted comprehension. This answer does
not establish that the participant understood what the replay measures or
prove any actual Chrome path or timing.

### Browser difference assisted recheck

Facilitator explanation: both tools run on this computer; HopVista opens its
own direct connection, while Chrome may reuse an existing connection or use
saved address information. Prompt: “In your own words, what’s one reason
Chrome and HopVista could show different times?”

Actual response: “Not too sure”.

Assessment: **not passed after assistance**. P1 did not state a reason for
different times despite the visible technical caveat and a plain verbal
example. Do not treat the earlier correct rejection of guaranteed matching
as comprehension of the mechanism. Simplify the on-screen caveat to connect
the examples explicitly to different timings, preserve its direct-path
limits, and recheck. Any later success is assisted.

### Revised browser-note human recheck

After the note was simplified, P1 was asked to refresh Chrome, read “A direct
connection”, and explain in their own words one reason Chrome and HopVista
might show different times. The facilitator had already given a direct
connection/reused-connection explanation earlier. The participant's browser
refresh was not independently observed by the agent.

Actual response: “Cause "Chrome or another browser may reuse an open
connection, use a saved or different number address, or go through another
service (a proxy). Their times can differ."”

Assessment: **the revised explanation was found, but comprehension remains
unverified**. P1 copied the visible sentence rather than paraphrasing a
reason, despite the prompt asking for their own words. The quote accurately
matches the revised UI, so it supports visibility/locatability of the note
after the reported refresh, not independent understanding of reuse, saved
addresses, proxies, or why the times differ. This is an assisted participant
recheck, not a new-reader first-use pass. Ask about one concrete phrase in
ordinary language before scoring this criterion.

### Open-connection phrase check

Prompt: “When it says Chrome may ‘reuse an open connection,’ what do you think
Chrome is reusing?” No definition of that phrase was supplied before the
answer, though prior prompts had discussed direct versus browser connections.

Actual response: “Like i guess an opened tab?”

Assessment: **not passed for the connection-reuse mechanism**. P1 interpreted
“open connection” as an open browser tab. The intended meaning is a network
connection the browser may keep open to the site and use again. The wording
should explain that concrete action without requiring familiarity with the
term; preserve the direct-path caveat and other browser differences. The
facilitator clarified the term after this response, so any later success is
assisted. Do not infer P1's actual Chrome state from the answer.
