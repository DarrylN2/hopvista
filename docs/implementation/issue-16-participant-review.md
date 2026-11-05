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

### Connection-reuse wording recheck

After the UI explained that Chrome may keep a website connection open from
an earlier visit and use it again, the facilitator asked P1 to refresh the
page and answer: “If Chrome already has that connection open, what step could
it skip on the next visit?” The facilitator had already corrected the earlier
open-tab interpretation. Refresh was participant-performed and not
independently observed.

Actual response: “It might use the saved connection?”

Assessment: **assisted recheck passed narrowly for the reuse example**. P1
now identifies use of a connection already available, which implies skipping
opening another one. The question and revised text supplied the situation;
this does not establish first-use comprehension, the actual path Chrome took,
or understanding of alternate addresses/proxies. The broader direct-path
versus browser distinction remains only partially checked. Other issue #16
human tasks remain open.

### Dotted path connectors: first reading

Prompt: run `https://fixture.example/i13highsilent` and say what the dotted
line between hop markers in the IPv4 path diagram means. No expected answer
or explanation was supplied in this prompt. The UI does have a nearby muted
sentence about probe order, but the participant was not directed to it.

Actual response: “Its like the connection, like a train station”.

Assessment: **not yet passed for path-connector meaning**. The train-station
analogy suggests P1 reads the connected markers as a journey or connection
between stops. The answer does not establish the intended probe-order-only
meaning and may imply a physical or website-request route; do not assign a
more precise misunderstanding than the words support. Ask P1 to read the
nearby explanatory sentence and interpret it before deciding whether the
on-screen language requires revision.

### Dotted path connectors: note recheck

Facilitator follow-up directed P1 to the sentence below “Trace to…” about
dotted connectors and asked for its meaning in their own words. The exact
expected interpretation was not supplied in the prompt, but the on-screen
sentence explicitly states probe order.

Actual response: “It shows the order of hops”.

Assessment: **guided recheck passed for the order meaning**. P1 identified
the order of hop checks after attention was directed to the nearby note.
This does not establish that the visual alone conveys the intended meaning,
or that P1 understands the line does not prove physical links or the route
of the website request. Check those limits separately. No physical topology
or HTTP route is inferred from the replay.

### Dotted path connectors: link inference

Prompt after the order recheck: “Do the dots also prove there’s a direct
connection between each pair of hops? What makes you think so?” The nearby
on-screen sentence says they do not establish a physical link; the prompt
did not supply the expected answer.

Actual response: “I think so, cause it looks like that”.

Assessment: **not passed for the physical-link limit**. Even after locating
the order note, P1 inferred direct connections from the visual itself. This
is evidence that the diagram's connected appearance outweighs the current
technical caveat for this participant. The facilitator corrected the claim
after the answer. Clarify the visible adjacent text in ordinary language and
recheck; if the graphic still yields the inference, the connector treatment
will need reconsideration within the approved composition. No actual
hop-to-hop link or website-request route is established by this trace.

### Revised dotted-line note recheck

After the path note was made more explicit, P1 was asked to refresh Chrome,
rerun `i13highsilent`, and say what the dotted line alone reveals about a
direct connection between neighboring markers. Prior facilitator correction
and the new on-screen note both supplied the distinction; the browser action
was participant-performed and not independently observed.

Actual response: “Thats about the order”.

Assessment: **assisted recheck passed for order, but the direct-link limit
remains unconfirmed in this answer**. P1 again describes the dots as ordering
checks, consistent with the revised note, and does not repeat the earlier
direct-link claim. The answer does not explicitly say whether the line can
establish a direct connection. Ask that narrower question once before scoring
the limit. Do not infer a website-request route from this replay.

### Direct-link limit recheck

Prompt after the revised note and prior correction: “Can the dotted line tell
us whether two neighboring hop markers are directly connected?”

Actual response: “No cause there's a note”.

Assessment: **assisted recheck passed for rejecting the direct-link
inference**. P1 now says the line cannot tell us that neighboring markers
connect directly and cites the adjacent note. The answer confirms that the
revised explanation corrected the earlier visual inference for this
participant when the note was used. It does not prove the graphic alone is
unambiguous, unassisted first-use comprehension, or knowledge of the actual
website-request route. Keep the connector explanation visible and continue
other path/echo questions.

### High intermediate reply: first interpretation

Prompt on `i13highsilent`: Hop 2 shows 450 ms and Hop 3 shows 12 ms; ask what,
if anything, P1 concludes about delay between them. No expected conclusion
was supplied. The row has a small caveat about ICMP response time versus
link delay/bottleneck, but the prompt did not direct P1 to read it.

Actual response: “Hop 2 is like the bottleneck”.

Assessment: **not passed for the intermediate-reply limit**. A single 450 ms
response from the device at Hop 2 followed by a 12 ms response at Hop 3 does
not establish a bottleneck or a 450 ms link delay. P1 inferred a causal
slowdown from the higher intermediate reply. This is a demonstrated
comprehension miss, not evidence of an actual network bottleneck. The
facilitator corrected the claim after the answer. Ask P1 to read the nearby
row note before deciding whether its technical language needs revision.

### High intermediate reply: row-note reading

Facilitator prompt: read the small note beside Hop 2's 450 ms reading and say
what the number can and cannot tell us; “not sure” was explicitly allowed.
The expected interpretation was not supplied in the prompt. The note uses
“router”, “ICMP”, “RTT”, “link delay” and “bottleneck”.

Actual response: “I'm not too sure of the techincal words”.

Assessment: **not passed; the note's language blocked comprehension** for
this participant. Do not infer that P1 understood the ICMP-versus-transit
distinction or revise their earlier bottleneck conclusion based on this
answer. Replace the static on-screen explanation with plain words that
distinguish the device's answer to HopVista from time between hops or website
slowdown. Keep the recorded RTT, outcome, trace and Go-owned findings intact;
recheck after the wording change with assistance noted.

### High intermediate reply: precise-note recheck (2026-11-04)

After a pause in the review, the local controlled replay was restarted at a
new loopback port. P1 was asked to rerun `i13highsilent`, read the revised Hop 2
note, and explain what its 450 ms measures. Prior facilitator explanations had
already corrected the bottleneck claim. The participant opened and operated
Chrome; the browser action was not independently observed by the agent.

Actual response: “The time it takes to get from hopvista to my device”.

Assessment: **not passed for the send-to-reply interval**. HopVista runs on
P1's computer; the recorded reply time starts when the check leaves that
computer and ends when a reply from the network device at Hop 2 returns.
The answer describes travel toward P1's device and omits the outward trip,
showing that “this device's reply” still leaves the endpoints ambiguous for
this participant. The facilitator clarified the round trip after the answer.
Revise the visible note to name the user's computer, the hop and the return
reply, while preserving the result and no-bottleneck limit. Any later success
will be assisted.

### High intermediate reply: round-trip wording recheck

After the UI explicitly named the user's computer as both the start and
return endpoint, P1 was asked to refresh the controlled `i13highsilent`
replay, read the Hop 2 note, and explain what 450 ms covers. The facilitator
had already explained the out-and-back meaning before this prompt, and the
browser refresh was participant-performed rather than agent-observed.

Actual response: “The time it takes to send a check from my computer to the
hop”.

Assessment: **partial; the return reply remains unconfirmed**. P1 now
identifies the outward direction and the correct starting computer, which
corrects the earlier reversal. The answer omits the reply from Hop 2 back to
the computer. Do not score this as a full round-trip understanding or infer
that P1 rejects the return leg; ask one narrow follow-up before changing the
UI again. The 450 ms remains a captured reply time, not proof of a link delay
or website bottleneck.

### High intermediate reply: return-leg follow-up

Facilitator follow-up after P1 named only the outward check: “What has to
happen after that for the 450 ms timer to stop?” Earlier UI and facilitator
explanations had named the return reply, so this is a guided recheck.

Actual response: “The reply needs to arrive to my computer”.

Assessment: **assisted recheck passed for the return endpoint**. Taken with
P1's preceding outward-leg answer, this identifies the round trip from the
computer toward Hop 2 and the reply back to the computer. The question
directed attention to what happens after the outward check, and P1 had
received prior help; this is not unassisted first-use comprehension. It does
not by itself establish understanding of whether the 450 ms proves link
delay or a website bottleneck. Check that causal limit separately.

### High intermediate reply: causal-limit recheck

Prompt after the round-trip explanation: “Now that Hop 3 shows 12 ms, does
Hop 2's 450 ms prove Hop 2 slowed the website? Why or why not?” The revised
row note explicitly says the reply time does not prove a website slowdown,
and the facilitator had corrected P1's earlier bottleneck inference.

Actual response: “No it does not”.

Assessment: **assisted recheck passed narrowly for rejecting the causal
claim**. P1 no longer treats the higher intermediate reply as proof that
Hop 2 slowed the website. They did not give a reason despite the prompt, so
understanding of why device reply time differs from website-request delay
remains unverified. Prior teaching and the visible note mean this is not
unassisted first-use comprehension. Ask for the reason once; do not infer
an actual network bottleneck or lack of one from this replay.

### High intermediate reply: reason follow-up

Prompt: “What's the difference between Hop 2 taking 450 ms to answer
HopVista and Hop 2 slowing the website request? It's fine to say if you're
unsure.” Prior UI and facilitator help had already supplied the no-link-delay
and no-website-cause limits.

Actual response: “It doesn;t measure the time between hops?”

Assessment: **partial assisted understanding**. P1 identifies a true limit:
the reply time does not measure the interval between hops. The answer does
not explain that this is a separate check from the website request, so it
does not yet give the reason the 450 ms cannot establish website slowdown.
The question mark indicates uncertainty. The facilitator may explain the
separate-check distinction and ask one narrow follow-up; do not score a
full causal explanation from this response.

### Silent hop: displayed result

Prompt on `i13highsilent`: “In the same path, Hop 1 says ‘No reply.’ What do
you think that means for its time reading?” The facilitator had just
explained that the Hop 2 number times a separate check and that it cannot
establish website-request delay.

Actual response: “No address or RTT observed”.

Assessment: **guided check passed for locating the displayed missing
address/reply-time outcome**. P1 repeated the path row's explanation that no
reply address or RTT was observed. This does not yet establish whether they
interpret the missing time as zero or understand that a silent hop can be
followed by later replies. Ask for the missing-versus-zero meaning next.

### Silent hop: missing versus zero

Prompt: “If Hop 1 gave no reply, should we treat its reply time as 0 ms? Why
or why not?” The preceding prompt had drawn attention to the no-reply row;
the expected distinction was not supplied before this response.

Actual response: “Treat it as 0 ms”.

Assessment: **not passed**. P1 read the displayed “No address or RTT
observed” text, then treated the missing reply time as a completed zero.
This repeats the earlier timeout-versus-zero confusion in a different path
measurement. The current path row uses technical “RTT” and does not
explicitly distinguish missing time from a recorded 0 ms reply. The
facilitator corrected the distinction after the answer. Clarify the visible
no-reply row without changing the outcome, bound, evidence or Go-owned
findings, then recheck with assistance recorded.

### Silent hop: revised wording recheck

After the no-reply row was changed to distinguish a missing reply time from
measured zero, P1 was asked to refresh `i13highsilent` and say what time, if
any, they would record for Hop 1. Prior facilitator correction had explicitly
explained missing versus zero; the browser operation was participant-performed
and not independently observed by the agent.

Actual response: “There is no reply within 750 ms. Also, can you let me know
how far or like the percentage until we're done with this?”

Assessment: **guided recheck passed for reading the no-reply bound, but the
missing-duration interpretation remains unconfirmed**. P1 identified the
750 ms no-reply result from the revised row. They did not state whether to
record 0 ms or no completed reply time, which was the task's distinction.
Do not infer a comprehension pass from the quoted row. The facilitator gave
a rough progress update for the human walkthrough and should ask one narrow
follow-up on the time value. Issue #16 and release acceptance remain open.

### Silent hop: revised wording value follow-up

Prompt after the revised-row bound reading: “Should its reply time be recorded
as 0 ms or no time recorded?” The on-screen row already says “No reply time
recorded” and explains that a 0 ms reading would require a reply. The
facilitator had previously supplied the same distinction verbally.

Actual response: “Probably just 0 ms”.

Assessment: **not passed after assistance**. P1 again assigned a completed
zero to a hop with no observed reply. “Probably” indicates uncertainty. The
loopback preview serves the current embedded `app.js` with `Cache-Control:
no-store` and contains the revised phrase, but the agent has not observed
P1's Chrome screen or confirmed which parts of the row P1 read. Ask P1 to
report the phrase immediately after the 750 ms message before deciding
whether the wording or its placement needs further change. Do not record a
0 ms reply or infer packet loss from this capture.

### Silent hop: visible wording confirmation

Prompt: after the 750 ms message in Hop 1, do you see the bold words “No reply
time recorded”? The facilitator had already stated the correct missing-time
interpretation in the preceding reply.

Actual response: “Yes i do”.

Assessment: **visibility confirmed by participant self-report; comprehension
still not passed**. The current no-reply text is present on P1's Chrome view,
so stale asset delivery does not explain the repeated 0 ms answer. The
participant has not yet interpreted the value correctly in their own words.
Move the reply-time outcome into a consistent, labeled value position beside
other hop readings so the missing state is visually comparable with measured
reply times. This is a UI presentation change only; preserve all recorded
outcomes, bounds and Go-owned findings. Recheck with assistance disclosed.

### Silent hop: aligned-value recheck

Prompt after the path presentation change: refresh `i13highsilent` and report
the **Reply time** values for Hop 1 and Hop 2. The prior discussion had
repeatedly taught that a silent hop has no measured reply time, so this is an
assisted recheck. The participant operated Chrome; the agent did not observe
the screen directly.

Actual response: “Reply time says Not measured for Hop 1 and 450 ms for Hop
2. Also, maybe, write "Reply time:" with the :.”

Assessment: **assisted recheck passed for distinguishing missing from
measured reply time** in the same path. P1 correctly located the aligned
values and no longer assigned 0 ms to the silent hop. This does not
establish unassisted first-use comprehension, packet-loss rate or website
health. The participant requested a punctuation improvement to make the
field label read naturally; add a colon in both hop states without altering
the values or evidence.

### Configured hop-limit reading

Prompt: refresh the controlled preview, run `i13limit1`, read **Trace
outcome**, and explain why the path check stopped. The prompt supplied the
case URL but not its expected outcome. Earlier path explanations and
facilitator help make this an assisted participant review.

Actual response: “It says that hop limit reached. I think cause its limited
to 1 hops”.

Assessment: **guided recheck passed for a configured stopping rule**. P1
identified `hop_limit_reached` and associated it with the one-hop setting,
matching this capture's `max_hops: 1` and highest attempted hop 1. The
answer does not infer an unreachable reply or website failure. The next
case should test whether P1 distinguishes this configured stop from an
observed unreachable response; no fresh network measurement is inferred.

### Unreachable-response comparison

Prompt: run `i13unreachable`, read **Trace outcome**, and explain how its
stopping reason differs from the one-hop limit. The expected answer was not
provided in the prompt; prior path guidance and the immediately preceding
hop-limit case were available.

Actual response: “It says unreachable response, the limit is 4 hops, it
stopped cause the response was unreachable”.

Assessment: **partial understanding**. P1 located the outcome and correctly
noticed this case allows four hops, distinguishing it from the one-hop
configuration. “The response was unreachable” does not show whether P1
understands that HopVista actually received a device's unreachable message,
rather than receiving no reply. The facilitator clarified that distinction
after the answer. Ask a narrow response-versus-silence follow-up before
scoring the observed-response criterion. The capture does not establish a
general website failure or physical path break.

### Unreachable-response versus silence

Prompt after P1 named the outcome: “In the unreachable response example, did
a device send HopVista an answer, or did HopVista receive no answer at all?”
The facilitator's preceding message had explicitly explained that a device
sent an unreachable message, so this is an assisted check.

Actual response: “HopVista received no answer at all i think”.

Assessment: **not passed after assistance**. The `i13unreachable` capture
records an `unreachable_reply` at hop 2 from `192.0.2.253` with a 6 ms reply
time; the trace outcome is `unreachable_response`. P1 interpreted it as
silence instead of a received response. The visible raw outcome label and
hop status do not explain this in everyday words. The facilitator corrected
the interpretation after the answer. Add an adjacent plain explanation of
the observed unreachable reply without changing the outcome, hop limit,
evidence, website finding or schema; recheck with prior help disclosed.
