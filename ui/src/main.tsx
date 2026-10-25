import { useEffect, useMemo, useRef, useState } from "react";
import { createRoot } from "react-dom/client";
import type { Echo, Exchange, Family, Finding, Measurement, Path, RunResult } from "./types";
import "./style.css";

const directPathCaveat = "HopVista tests a direct connection from this computer. Your browser may use another IP, a proxy, cached DNS, or an existing connection.";
const terminal = new Set(["completed", "cancelled", "failed"]);
const phaseHelp = {
  DNS: "Time spent finding an IPv4 address for the website name. This computer's resolver may use a saved (cached) answer.",
  TCP: "Time spent opening a connection to the selected IPv4 address, before encryption and the website request.",
  TLS: "Time spent setting up encryption and checking the website's security certificate for this HTTPS connection.",
  HTTP: "Time from writing the request until the website's response starts (the first response-header byte). Includes network travel and server work, not downloading or displaying the page.",
};

function ms(value: number) { return `${Number(value.toFixed(value < 10 ? 2 : 1))} ms`; }
function outcome(value: string) { return value.replaceAll("_", " "); }
function runDescription(state: RunResult["state"]): string {
  switch (state) {
    case "running": return "Measurements appear as they arrive. The observed finding will be available when this run finishes.";
    case "stopping": return "No new measurements will start. Waiting for work already started to finish and release its resources. Completed evidence remains available.";
    case "completed": return "Measurement work has finished. Read each outcome below; completion does not mean every measurement received a response.";
    case "cancelled": return "You stopped this diagnostic run. Completed measurements remain available; unfinished measurements are marked cancelled.";
    case "failed": return "The diagnostic run could not finish coherently. Any recorded evidence remains available below.";
  }
}
function phaseValue(m: Measurement): string {
  if (m.outcome === "completed" && m.duration_ms !== undefined) return ms(m.duration_ms);
  if (m.outcome === "timeout" && m.bound_ms !== undefined) return `No response within ${ms(m.bound_ms)}`;
  if (m.outcome === "skipped") return m.method === "reused_selection" ? "Reused selection" : "Not applicable";
  return outcome(m.outcome);
}
function phaseRows(exchange: Exchange) {
  return [
    { label: "DNS", name: "Address lookup", measurement: exchange.dns, help: phaseHelp.DNS },
    { label: "TCP", name: "Connection", measurement: exchange.tcp, help: phaseHelp.TCP },
    { label: "TLS", name: "Encryption", measurement: exchange.tls, help: phaseHelp.TLS },
    { label: "HTTP first byte", name: "Response begins", measurement: exchange.http, help: phaseHelp.HTTP },
  ];
}
function phaseLabel(m: Measurement) {
  if (m.outcome === "timeout" && m.bound_ms !== undefined && m.method === "request_written_to_first_response_header_byte") return `No first response byte within ${ms(m.bound_ms)}`;
  return phaseValue(m);
}
function selectedIP(family: Family): string | undefined {
  return family.paths.at(-1)?.selected_ip || family.exchanges.at(-1)?.selected_ip || family.dns.find((d) => d.selected_ip)?.selected_ip;
}
function selectedScope(family: Family): string | undefined {
  return family.exchanges.at(-1)?.selected_scope || family.dns.find((d) => d.selected_scope)?.selected_scope;
}
function evidenceItems(run: RunResult): Map<string, Record<string, unknown>> {
  const found = new Map<string, Record<string, unknown>>();
  function visit(value: unknown) {
    if (!value || typeof value !== "object") return;
    if (Array.isArray(value)) { value.forEach(visit); return; }
    const item = value as Record<string, unknown>;
    if (typeof item.id === "string") found.set(item.id, item);
    Object.values(item).forEach(visit);
  }
  visit(run.families);
  return found;
}

async function readJSON<T>(response: Response): Promise<T> {
  const body = await response.json();
  if (!response.ok) throw new Error(body.error?.message || `HTTP ${response.status}`);
  return body as T;
}

function App() {
  const [token, setToken] = useState("");
  const [target, setTarget] = useState("");
  const [hops, setHops] = useState(30);
  const [run, setRun] = useState<RunResult | null>(null);
  const [runID, setRunID] = useState("");
  const [notice, setNotice] = useState("Ready.");
  const [submitting, setSubmitting] = useState(false);
  const [cancelling, setCancelling] = useState(false);
  const [requestError, setRequestError] = useState("");
  const [detailsOpen, setDetailsOpen] = useState(false);
  const [pendingEvidence, setPendingEvidence] = useState("");
  const [activeEvidence, setActiveEvidence] = useState("");
  const evidenceSource = useRef<HTMLButtonElement | null>(null);
  const statusRef = useRef<HTMLParagraphElement>(null);
  const detailsRef = useRef<HTMLDetailsElement>(null);
  const evidence = useMemo(() => run ? evidenceItems(run) : new Map<string, Record<string, unknown>>(), [run]);

  useEffect(() => {
    if (submitting || cancelling) statusRef.current?.focus();
  }, [submitting, cancelling]);

  useEffect(() => {
    let live = true;
    fetch("/api/v1/session", { cache: "no-store" })
      .then((response) => readJSON<{ mutation_token: string }>(response))
      .then((data) => { if (live) setToken(data.mutation_token); })
      .catch((error: Error) => { if (live) setRequestError(`Could not connect to HopVista. Reload this page to try again. ${error.message}`); });
    return () => { live = false; };
  }, []);

  useEffect(() => {
    if (!runID) return;
    let live = true;
    let timer: number | undefined;
    async function poll() {
      try {
        const response = await fetch(`/api/v1/runs/${runID}`, { cache: "no-store" });
        const data = await readJSON<RunResult>(response);
        if (!live) return;
        setRun(data);
        setRequestError("");
        setNotice(data.state === "stopping" ? "Stopping… completed evidence remains available." : `${outcome(data.state)}.`);
        if (terminal.has(data.state)) setRunID("");
        else timer = window.setTimeout(poll, 400);
      } catch (error) {
        if (!live) return;
        setRequestError(`Could not refresh this run. Retrying automatically; its last recorded state is shown below. ${(error as Error).message}`);
        timer = window.setTimeout(poll, 1000);
      }
    }
    void poll();
    return () => { live = false; window.clearTimeout(timer); };
  }, [runID]);

  useEffect(() => {
    if (!detailsOpen || !pendingEvidence) return;
    const target = document.getElementById(`evidence-${pendingEvidence}`);
    target?.focus();
    target?.scrollIntoView({ block: "center" });
    setPendingEvidence("");
  }, [detailsOpen, pendingEvidence, run]);

  async function start(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!token || submitting || runID) return;
    setSubmitting(true);
    setNotice("Starting…");
    setRequestError("");
    try {
      const response = await fetch("/api/v1/runs", {
        method: "POST", headers: { "Content-Type": "application/json", "X-HopVista-Token": token },
        body: JSON.stringify({ target_url: target, max_hops: hops }),
      });
      const data = await readJSON<{ run_id: string }>(response);
      setRun(null);
      setDetailsOpen(false);
      setActiveEvidence("");
      evidenceSource.current = null;
      setNotice("Running…");
      setRunID(data.run_id);
    } catch (error) {
      setNotice(run ? `${outcome(run.state)}.` : "Ready.");
      setRequestError(`Could not start the diagnostic run. ${(error as Error).message}`);
    }
    finally { setSubmitting(false); }
  }

  async function cancel() {
    if (!runID || cancelling) return;
    setCancelling(true);
    setNotice("Requesting stop…");
    try {
      const response = await fetch(`/api/v1/runs/${runID}/cancel`, { method: "POST", headers: { "X-HopVista-Token": token } });
      const data = await readJSON<RunResult>(response);
      setRun(data);
      setNotice(data.state === "stopping" ? "Stopping… completed evidence remains available." : `${outcome(data.state)}.`);
    } catch (error) {
      setNotice("Cancellation unconfirmed. Last recorded state retained.");
      setRequestError(`Could not confirm cancellation. The run may still be active. ${(error as Error).message}`);
    }
    finally { setCancelling(false); }
  }

  function showEvidence(id: string) {
    evidenceSource.current = document.activeElement instanceof HTMLButtonElement ? document.activeElement : null;
    setActiveEvidence(id);
    setDetailsOpen(true);
    setPendingEvidence(id);
  }

  function returnToMeasurement() {
    evidenceSource.current?.focus();
  }

  const family = run?.families.find((item) => item.family === "ipv4");
  const address = family && selectedIP(family);
  const path = family?.paths.at(-1);
  const finalFinding = run && terminal.has(run.state) ? run.findings.find((item) => item.family === "ipv4") : undefined;
  const incomplete = run?.state === "completed" && ((family && family.outcome !== "completed") || [...evidence.values()].some((item) => ["timeout", "deadline", "cancelled", "error", "unavailable"].includes(String(item.outcome))));

  return <main className={run || runID || submitting ? "has-run" : undefined} data-state={run?.state || (runID || submitting ? "running" : "ready")}>
    <a className="skip-link" href="#run-workspace">Skip to diagnostic run</a>
    <header className="masthead"><a className="wordmark" href="#"><svg aria-hidden="true" viewBox="0 0 32 32"><path d="M5 25V7m22 18V7M5 16h22"/><circle cx="5" cy="16" r="3"/><circle cx="16" cy="16" r="3"/><circle cx="27" cy="16" r="3"/></svg><span>HopVista</span></a><p className="scope">Local IPv4 diagnostic <span>Experimental · Windows 11 x64</span></p></header>
    <section className="intro" aria-labelledby="intro-title"><h1 id="intro-title">See the path.<br/><span>Understand the wait.</span></h1><p>Observed hop replies and website timings, from this computer.</p></section>
    <section className="workspace" id="run-workspace" aria-label="Diagnostic run">
    <form className="start-form" onSubmit={start}>
      <div className="target-input"><label htmlFor="target">Target URL</label><input id="target" type="url" required placeholder="https://example.com/" autoComplete="url" spellCheck={false} value={target} onChange={(event) => setTarget(event.target.value)} /></div>
      <button type="submit" disabled={!token || !!runID || submitting}>{submitting ? "Starting…" : "Start diagnostic run"}</button>
      {!!runID && <button type="button" className="secondary" disabled={cancelling || run?.state === "stopping"} onClick={cancel}>{cancelling ? "Requesting stop…" : run?.state === "stopping" ? "Stopping…" : "Cancel run"}</button>}
      <details className="advanced"><summary>Advanced</summary><div className="advanced-fields"><label htmlFor="hops">Maximum hops (1–64)</label><input id="hops" type="number" min="1" max="64" step="1" required value={hops} onChange={(event) => setHops(Number(event.target.value))} /></div></details>
    </form>
    <div className="status-line"><p className="status" role="status" aria-live="polite" tabIndex={-1} ref={statusRef}><span className="status-dot" aria-hidden="true" />{notice}</p><span className="local-note">Memory-only results. Closing HopVista clears them.</span></div>
    {requestError && <p className="request-error" role="alert">{requestError}</p>}
    </section>
    <p className="caveat"><strong>A direct connection.</strong> {directPathCaveat}</p>
    {!run && <section className="empty-state" aria-labelledby="empty-title"><h2 id="empty-title">{runID || submitting ? "Waiting for the first measurements…" : "One run. Two views of the connection."}</h2>{runID || submitting ? <p>The finding stays pending until this run finishes. Recorded measurements will appear here as they arrive.</p> : <div className="empty-guide"><div><div className="empty-path" aria-hidden="true"><i/><i/><i/></div><h3>IPv4 path</h3><p>Hop replies in probe order.</p></div><div><div className="empty-phases" aria-hidden="true"><span>DNS</span><span>TCP</span><span>TLS</span><span>HTTP</span></div><h3>Website timing</h3><p>Separate waits for each exchange.</p></div></div>}<p className="muted">Enter an HTTP or HTTPS target URL to begin. Page rendering is not measured.</p></section>}
    {run && <>
      <div className="run-heading"><div><h2>{run.target_url}</h2><p>Run result · IPv4 only · Direct connection · Hop limit {run.max_hops}</p></div><span className={`state state-${run.state}`}>{run.state === "stopping" ? "Stopping…" : outcome(run.state)}</span></div>
      <p className="run-description">{runDescription(run.state)}{incomplete && <> Some measurements are incomplete; their individual outcomes explain why.</>}</p>
      <div className="result-grid">
        <aside className="summary-column">
          <section className="panel answer" aria-labelledby="answer-title"><h3 id="answer-title">What this run observed</h3>
            <p className="selected"><strong>Selected IPv4</strong><span className="selected-address">{address || (terminal.has(run.state) ? "No IPv4 address selected" : "Pending")}</span>{family && selectedScope(family) && <span className="address-scope"><span className="visually-hidden">, address scope: </span>{outcome(selectedScope(family)!)}</span>}{address && <CopyAddress key={address} address={address} />}</p>
            {finalFinding ? <FindingView finding={finalFinding} exchange={family?.exchanges.find((exchange) => exchange.id === finalFinding.exchange_ref)} showEvidence={showEvidence} /> : terminal.has(run.state) ? <p>No observed finding is available for this run result.</p> : <p>Pending until the diagnostic run finishes.</p>}
            {run.error && <p className="error">{run.error.message}</p>}
            {family?.error && <p className="error">{family.error.message}</p>}
          </section>
          <section className="panel timing" aria-labelledby="timing-title"><h3 id="timing-title">Website timing</h3><p className="section-help">Four separate steps for each website request. DNS finds an address, TCP opens a connection, TLS secures it, and HTTP first byte waits for the response to start. These timings do not measure a full page load.</p>
            {family?.exchanges.length ? family.exchanges.map((exchange) => <ExchangeView key={exchange.id} exchange={exchange} showEvidence={showEvidence} />) : <p>No website exchange recorded yet.</p>}
            {family?.dns.filter((dns) => !family.exchanges.some((exchange) => exchange.dns_ref === dns.id)).map((dns) => <div className="exchange" key={dns.id}><h4>IPv4 selection for {dns.hostname}</h4><p>{phaseLabel(dns)}. {phaseHelp.DNS}</p><EvidenceButton id={dns.id} showEvidence={showEvidence} /></div>)}
          </section>
        </aside>
        <section className="panel path-column" aria-labelledby="path-title"><h3 id="path-title">IPv4 path</h3><p className="section-help">Observed replies, one probe at a time.</p>
          {path ? <PathView path={path} showEvidence={showEvidence} /> : <p>{family?.outcome === "unavailable" ? "No usable IPv4 path was available." : "Path measurements pending or unavailable."}</p>}
        </section>
      </div>
      <details className="panel technical" open={detailsOpen} onToggle={(event) => setDetailsOpen(event.currentTarget.open)} ref={detailsRef}>
        <summary>Technical evidence</summary>
        <p>Measurements use the selected IPv4 address for direct probes. The actual HTTP peer is checked against that selection. Missing durations and RTTs are not zero.</p>
        <div className="evidence-list">{[...evidence].map(([id, item]) => <EvidenceItem key={id} id={id} item={item} selected={activeEvidence === id} returnToMeasurement={returnToMeasurement} />)}</div>
        <details><summary>Full run result JSON</summary><pre>{JSON.stringify(run, null, 2)}</pre></details>
      </details>
    </>}
    <footer><span>HopVista</span><p>One diagnostic run records observations, not a root cause.</p></footer>
  </main>;
}

function CopyAddress({ address }: { address: string }) {
  const [copied, setCopied] = useState(false);
  const [copying, setCopying] = useState(false);
  const [error, setError] = useState("");
  const manualCopyRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (error) manualCopyRef.current?.focus();
  }, [error]);

  useEffect(() => {
    if (!copied) return;
    const timer = window.setTimeout(() => setCopied(false), 2400);
    return () => window.clearTimeout(timer);
  }, [copied, copying]);

  async function copy() {
    if (copying) return;
    setCopying(true);
    setError("");
    try {
      await navigator.clipboard.writeText(address);
      setCopied(true);
    } catch {
      setCopied(false);
      setError("Copy is unavailable. The address is selected in the manual-copy field; use your copy shortcut.");
    } finally {
      setCopying(false);
    }
  }

  return <span className="copy-address"><button type="button" className="copy-button" aria-disabled={copying} aria-busy={copying} data-copied={copied} aria-label={`Copy selected IPv4 address ${address}`} onClick={copy}><svg aria-hidden="true" viewBox="0 0 16 16">{copied ? <path d="m3 8 3 3 7-7"/> : <><rect x="5.5" y="5.5" width="7" height="8" rx="1"/><path d="M3 10.5H2.5V2.5h7V3"/></>}</svg><span>{copied ? "Copied" : copying ? "Copying…" : "Copy IPv4"}</span></button><span className={error ? "copy-feedback error" : "visually-hidden"} role="status">{error || (copied ? "IPv4 address copied." : "")}</span>{error && <label className="manual-copy">IPv4 address for manual copying<input ref={manualCopyRef} readOnly value={address} onFocus={(event) => event.currentTarget.select()} /></label>}</span>;
}

function EvidenceButton({ id, label = "View evidence", showEvidence }: { id: string; label?: string; showEvidence: (id: string) => void }) {
  return <button type="button" className="evidence-link" aria-label={`${label} ${id}`} title={id} onClick={() => showEvidence(id)}>{label}<svg aria-hidden="true" viewBox="0 0 16 16"><path d="M3 8h10M9 4l4 4-4 4"/></svg></button>;
}
function FindingView({ finding, exchange, showEvidence }: { finding: Finding; exchange?: Exchange; showEvidence: (id: string) => void }) {
  const context: Record<string, string> = {
    dominant_measured_wait: "This compares completed phases in one exchange. The longest measured wait does not locate the cause.",
    http_response_observed: "A recorded response describes this exchange, not the website’s overall speed. HTTP 4xx and 5xx statuses are still responses.",
    http_phase_incomplete: "Earlier completed timings remain useful. The unfinished phase has its own outcome; it is not a zero-duration measurement.",
    redirect_stopped: "A redirect asks for another URL. Stopping under the URL policy is deliberate and does not establish a network failure.",
    ipv4_selection_incomplete: "This run needs a usable IPv4 address. No IPv6 connection is attempted in v1.",
    measurement_cutoff: "The cutoff stops new measurement work. Work already started may finish later; no maximum cleanup time is promised.",
  };
  return <><p className="observation">{finding.observation}</p>{context[finding.type] && <p className="finding-context">{context[finding.type]}</p>}{["dominant_measured_wait", "http_response_observed"].includes(finding.type) && <details className="finding-help"><summary>What does “dominant wait” mean?</summary><p>At least two completed phases in the same exchange must be comparable. The longest must be at least 100 ms and at least twice the next longest. If no phase meets that rule, it does not establish that the website or browser was fast.</p></details>}<p className="limit"><strong>What this cannot establish</strong>{finding.limit}</p><div className="evidence-refs">{finding.evidence_refs.map((id) => { const phase = exchange && phaseRows(exchange).find((row) => row.measurement.id === id); return <EvidenceButton key={id} id={id} label={phase ? `View ${phase.label} evidence` : undefined} showEvidence={showEvidence} />; })}</div></>;
}
function ExchangeView({ exchange, showEvidence }: { exchange: Exchange; showEvidence: (id: string) => void }) {
  return <div className="exchange"><h4>Exchange {exchange.redirect_index + 1}</h4><p className="exchange-url">{exchange.requested_url}</p><p><strong>IPv4 {exchange.selected_ip}</strong> · {outcome(exchange.selected_scope)}{exchange.actual_peer_ip && <> · Actual peer {exchange.actual_peer_ip}</>}</p>
    <dl className="phases">{phaseRows(exchange).map(({ label, name, measurement, help }) => <div key={label}><dt>{label}<span>{name}</span></dt><dd><strong className={measurement.outcome === "completed" && measurement.duration_ms !== undefined ? "measured-value" : undefined} data-outcome={measurement.outcome}>{phaseLabel(measurement)}</strong><span>{help}</span>{measurement.error && <span>{measurement.error.message}</span>}<EvidenceButton id={measurement.id} showEvidence={showEvidence} /></dd></div>)}</dl>
    <p className="exchange-footer">{exchange.response_status !== undefined ? `HTTP status ${exchange.response_status}. ` : ""}{exchange.location ? `Redirect location: ${exchange.location}. ` : ""}{exchange.redirect_stop_reason ? `Redirect stopped: ${outcome(exchange.redirect_stop_reason)}.` : ""}</p>
  </div>;
}
function EchoText({ echo }: { echo: Echo }) {
  if (echo.status === "no_reply") return <>No reply within the recorded timeout{echo.bound_ms !== undefined ? ` (${ms(echo.bound_ms)})` : ""}. No address or RTT observed.</>;
  if (echo.status === "not_sent") return <>Not sent ({outcome(echo.outcome)}).</>;
  return <span className="echo-reading"><span className="echo-status">{outcome(echo.status)}</span>{echo.reply_address && <strong className="reply-address">{echo.reply_address}</strong>}{echo.rtt_ms !== undefined && <span className="raw-rtt">Raw RTT <strong>{ms(echo.rtt_ms)}</strong></span>}{echo.error && <span className="error">{echo.error.message}</span>}</span>;
}
function PathView({ path, showEvidence }: { path: Path; showEvidence: (id: string) => void }) {
  const { route, trace } = path;
  return <><p><strong>Trace to {trace.target_role === "attempted_host" ? "attempted host" : "final host"}: {trace.selected_ip || path.selected_ip}</strong></p>
    <p className="muted">Dotted connectors show probe order only. They do not establish a physical link or the HTTP route.</p>
    <div className="route-block"><h4>Route-specific gateway</h4><p>{route.outcome === "completed" ? route.on_link ? "On-link route: this address needs no separate gateway." : route.next_hop ? <>Next hop <span className="data-address">{route.next_hop}</span>.</> : "No separate gateway recorded." : `Route selection ${outcome(route.outcome)}.`}</p><p>{path.gateway.attempted} gateway echoes attempted; {path.gateway.replied} replied.</p>{path.gateway.samples.map((sample) => <p key={sample.id}><EchoText echo={sample} /> <EvidenceButton id={sample.id} showEvidence={showEvidence} /></p>)}<EvidenceButton id={route.id} showEvidence={showEvidence} /></div>
    <ol className="hop-list">{trace.hops.map((hop) => <li key={hop.id}><div className="hop"><span className="ttl">Hop <span>{hop.ttl}</span></span><span><EchoText echo={hop} />{hop.status === "time_exceeded" && hop.rtt_ms !== undefined && <small> This is one router’s ICMP response; its RTT is not link delay or proof of a bottleneck.</small>}</span><EvidenceButton id={hop.id} showEvidence={showEvidence} /></div></li>)}</ol>
    {trace.hops.length === 0 && <p>No trace hops recorded.</p>}
    <p><strong>Trace outcome:</strong> {outcome(trace.outcome)} · {trace.highest_attempted_ttl} of {trace.max_hops} allowed hops attempted. <EvidenceButton id={trace.id} showEvidence={showEvidence} /></p>
    <div className="echo-summary"><h4>Target echo</h4><p>{path.target.attempted} attempted; {path.target.replied} replied. ICMP echo is separate from website timing and does not measure page load.</p>{path.target.samples.map((sample) => <p key={sample.id}><EchoText echo={sample} /> <EvidenceButton id={sample.id} showEvidence={showEvidence} /></p>)}{path.target.variation && <p>RTT variation: {path.target.variation.value_ms !== undefined ? ms(path.target.variation.value_ms) : outcome(path.target.variation.outcome)}. <EvidenceButton id={path.target.variation.id} showEvidence={showEvidence} /></p>}</div>
  </>;
}
function EvidenceItem({ id, item, selected, returnToMeasurement }: { id: string; item: Record<string, unknown>; selected: boolean; returnToMeasurement: () => void }) {
  const text = (key: string) => typeof item[key] === "string" ? item[key] as string : undefined;
  const number = (key: string) => typeof item[key] === "number" ? ms(item[key] as number) : undefined;
  const labels = [text("outcome"), text("status"), text("method")].filter((value): value is string => !!value);
  const error = item.error as { message?: string } | undefined;
  return <section id={`evidence-${id}`} className={`evidence-item${selected ? " is-selected" : ""}`} tabIndex={-1}><h4>{id}</h4>{selected && <button type="button" className="evidence-link" onClick={returnToMeasurement}>Return to measurement</button>}<p>{labels.map(outcome).join(" · ") || "Recorded evidence"}</p><dl>{[["Duration", number("duration_ms")], ["Bound", number("bound_ms")], ["RTT", number("rtt_ms")], ["Started", text("started_at")], ["Ended", text("ended_at")], ["Selected IP", text("selected_ip")], ["Reply address", text("reply_address")]].filter((pair) => pair[1]).map(([label, value]) => <div key={label}><dt>{label}</dt><dd>{value}</dd></div>)}</dl>{error?.message && <p className="error">{error.message}</p>}<details><summary>Raw evidence</summary><pre>{JSON.stringify(item, null, 2)}</pre></details></section>;
}

createRoot(document.getElementById("root")!).render(<App />);
