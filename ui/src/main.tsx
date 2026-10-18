import { useEffect, useMemo, useRef, useState } from "react";
import { createRoot } from "react-dom/client";
import type { Echo, Exchange, Family, Finding, Measurement, Path, RunResult } from "./types";
import "./style.css";

const directPathCaveat = "HopVista tests a direct connection from this computer. Your browser may use another IP, a proxy, cached DNS, or an existing connection.";
const terminal = new Set(["completed", "cancelled", "failed"]);
const phaseHelp = {
  DNS: "Local resolver lookup; an answer may come from a cache.",
  TCP: "Connection setup to the selected IPv4 address.",
  TLS: "Encrypted connection setup and certificate verification for this URL.",
  HTTP: "From writing the request to the first response-header byte. Includes network transit and server work, not page rendering or body download.",
};

function ms(value: number) { return `${Number(value.toFixed(value < 10 ? 2 : 1))} ms`; }
function outcome(value: string) { return value.replaceAll("_", " "); }
function phaseValue(m: Measurement): string {
  if (m.outcome === "completed" && m.duration_ms !== undefined) return ms(m.duration_ms);
  if (m.outcome === "timeout" && m.bound_ms !== undefined) return `No response within ${ms(m.bound_ms)}`;
  if (m.outcome === "skipped") return m.method === "reused_selection" ? "Reused selection" : "Not applicable";
  return outcome(m.outcome);
}
function phaseRows(exchange: Exchange) {
  return [
    { label: "DNS", measurement: exchange.dns, help: phaseHelp.DNS },
    { label: "TCP", measurement: exchange.tcp, help: phaseHelp.TCP },
    { label: "TLS", measurement: exchange.tls, help: phaseHelp.TLS },
    { label: "HTTP first byte", measurement: exchange.http, help: phaseHelp.HTTP },
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
  const [detailsOpen, setDetailsOpen] = useState(false);
  const [pendingEvidence, setPendingEvidence] = useState("");
  const detailsRef = useRef<HTMLDetailsElement>(null);
  const evidence = useMemo(() => run ? evidenceItems(run) : new Map<string, Record<string, unknown>>(), [run]);

  useEffect(() => {
    let live = true;
    fetch("/api/v1/session", { cache: "no-store" })
      .then((response) => readJSON<{ mutation_token: string }>(response))
      .then((data) => { if (live) setToken(data.mutation_token); })
      .catch((error: Error) => { if (live) setNotice(error.message); });
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
        setNotice(data.state === "stopping" ? "Stopping… completed evidence remains available." : `${outcome(data.state)}.`);
        if (terminal.has(data.state)) setRunID("");
        else timer = window.setTimeout(poll, 400);
      } catch (error) {
        if (!live) return;
        setNotice((error as Error).message);
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
    try {
      const response = await fetch("/api/v1/runs", {
        method: "POST", headers: { "Content-Type": "application/json", "X-HopVista-Token": token },
        body: JSON.stringify({ target_url: target, max_hops: hops }),
      });
      const data = await readJSON<{ run_id: string }>(response);
      setRun(null);
      setDetailsOpen(false);
      setNotice("Running…");
      setRunID(data.run_id);
    } catch (error) { setNotice((error as Error).message); }
    finally { setSubmitting(false); }
  }

  async function cancel() {
    if (!runID) return;
    try {
      const response = await fetch(`/api/v1/runs/${runID}/cancel`, { method: "POST", headers: { "X-HopVista-Token": token } });
      const data = await readJSON<RunResult>(response);
      setRun(data);
      setNotice(data.state === "stopping" ? "Stopping… completed evidence remains available." : `${outcome(data.state)}.`);
    } catch (error) { setNotice((error as Error).message); }
  }

  function showEvidence(id: string) {
    setDetailsOpen(true);
    setPendingEvidence(id);
  }

  const family = run?.families.find((item) => item.family === "ipv4");
  const path = family?.paths.at(-1);
  const finalFinding = run && terminal.has(run.state) ? run.findings.find((item) => item.family === "ipv4") : undefined;

  return <main>
    <header className="masthead"><div><p className="eyebrow">Local IPv4 diagnostic</p><h1>HopVista</h1></div><p className="scope">Experimental Windows 11 x64 build</p></header>
    <form className="start-form" onSubmit={start}>
      <div className="target-input"><label htmlFor="target">Target URL</label><input id="target" type="url" required placeholder="https://example.com/" value={target} onChange={(event) => setTarget(event.target.value)} /></div>
      <details className="advanced"><summary>Advanced</summary><label htmlFor="hops">Maximum hops (1–64)</label><input id="hops" type="number" min="1" max="64" step="1" required value={hops} onChange={(event) => setHops(Number(event.target.value))} /></details>
      <button type="submit" disabled={!token || !!runID || submitting}>Start diagnostic run</button>
      <button type="button" className="secondary" disabled={!runID || run?.state === "stopping"} onClick={cancel}>Cancel run</button>
    </form>
    <p className="status" role="status" aria-live="polite">{notice}</p>
    <p className="caveat">{directPathCaveat}</p>
    {run && <>
      <div className="run-heading"><div><p className="eyebrow">Run result · IPv4 only</p><h2>{run.target_url}</h2></div><span className="state">{outcome(run.state)}</span></div>
      <div className="result-grid">
        <aside className="summary-column">
          <section className="panel answer" aria-labelledby="answer-title"><h3 id="answer-title">What this run observed</h3>
            <p className="selected"><strong>Selected IPv4:</strong> {family ? selectedIP(family) || "No IPv4 address selected" : terminal.has(run.state) ? "No IPv4 address selected" : "Pending"}{family && selectedScope(family) ? ` · ${outcome(selectedScope(family)!)}` : ""}</p>
            {finalFinding ? <FindingView finding={finalFinding} showEvidence={showEvidence} /> : terminal.has(run.state) ? <p>No observed finding is available for this run result.</p> : <p>Pending until the diagnostic run finishes.</p>}
            {run.error && <p className="error">{run.error.message}</p>}
            {family?.error && <p className="error">{family.error.message}</p>}
          </section>
          <section className="panel timing" aria-labelledby="timing-title"><h3 id="timing-title">Website timing</h3>
            {family?.exchanges.length ? family.exchanges.map((exchange) => <ExchangeView key={exchange.id} exchange={exchange} showEvidence={showEvidence} />) : <p>No website exchange recorded yet.</p>}
            {family?.dns.filter((dns) => !family.exchanges.some((exchange) => exchange.dns_ref === dns.id)).map((dns) => <div className="exchange" key={dns.id}><h4>IPv4 selection for {dns.hostname}</h4><p>{phaseLabel(dns)}. {phaseHelp.DNS}</p><EvidenceButton id={dns.id} showEvidence={showEvidence} /></div>)}
          </section>
        </aside>
        <section className="panel path-column" aria-labelledby="path-title"><h3 id="path-title">IPv4 path</h3>
          {path ? <PathView path={path} showEvidence={showEvidence} /> : <p>{family?.outcome === "unavailable" ? "No usable IPv4 path was available." : "Path measurements pending or unavailable."}</p>}
        </section>
      </div>
      <details className="panel technical" open={detailsOpen} onToggle={(event) => setDetailsOpen(event.currentTarget.open)} ref={detailsRef}>
        <summary>Technical evidence</summary>
        <p>Measurements use the selected IPv4 address for direct probes. The actual HTTP peer is checked against that selection. Missing durations and RTTs are not zero.</p>
        <div className="evidence-list">{[...evidence].map(([id, item]) => <EvidenceItem key={id} id={id} item={item} />)}</div>
        <details><summary>Full run result JSON</summary><pre>{JSON.stringify(run, null, 2)}</pre></details>
      </details>
    </>}
  </main>;
}

function EvidenceButton({ id, showEvidence }: { id: string; showEvidence: (id: string) => void }) {
  return <button type="button" className="evidence-link" onClick={() => showEvidence(id)}>Evidence {id}</button>;
}
function FindingView({ finding, showEvidence }: { finding: Finding; showEvidence: (id: string) => void }) {
  return <><p className="observation">{finding.observation}</p><p className="limit">{finding.limit}</p><div className="evidence-refs">{finding.evidence_refs.map((id) => <EvidenceButton key={id} id={id} showEvidence={showEvidence} />)}</div></>;
}
function ExchangeView({ exchange, showEvidence }: { exchange: Exchange; showEvidence: (id: string) => void }) {
  return <div className="exchange"><h4>Exchange {exchange.redirect_index + 1}</h4><p className="exchange-url">{exchange.requested_url}</p><p><strong>IPv4 {exchange.selected_ip}</strong> · {outcome(exchange.selected_scope)}{exchange.actual_peer_ip && <> · Actual peer {exchange.actual_peer_ip}</>}</p>
    <dl className="phases">{phaseRows(exchange).map(({ label, measurement, help }) => <div key={label}><dt>{label}</dt><dd><strong>{phaseLabel(measurement)}</strong><span>{help}</span>{measurement.error && <span>{measurement.error.message}</span>}<EvidenceButton id={measurement.id} showEvidence={showEvidence} /></dd></div>)}</dl>
    <p className="exchange-footer">{exchange.response_status !== undefined ? `HTTP status ${exchange.response_status}. ` : ""}{exchange.location ? `Redirect location: ${exchange.location}. ` : ""}{exchange.redirect_stop_reason ? `Redirect stopped: ${outcome(exchange.redirect_stop_reason)}.` : ""}</p>
  </div>;
}
function EchoText({ echo }: { echo: Echo }) {
  if (echo.status === "no_reply") return <>No reply within the recorded timeout{echo.bound_ms !== undefined ? ` (${ms(echo.bound_ms)})` : ""}. No address or RTT observed.</>;
  if (echo.status === "not_sent") return <>Not sent ({outcome(echo.outcome)}).</>;
  return <>{outcome(echo.status)}{echo.reply_address ? ` · ${echo.reply_address}` : ""}{echo.rtt_ms !== undefined ? ` · raw RTT ${ms(echo.rtt_ms)}` : ""}{echo.error ? ` · ${echo.error.message}` : ""}</>;
}
function PathView({ path, showEvidence }: { path: Path; showEvidence: (id: string) => void }) {
  const { route, trace } = path;
  return <><p><strong>Trace to {trace.target_role === "attempted_host" ? "attempted host" : "final host"}: {trace.selected_ip || path.selected_ip}</strong></p>
    <p className="muted">Dotted connectors show probe order only. They do not establish a physical link or the HTTP route.</p>
    <div className="route-block"><h4>Route-specific gateway</h4><p>{route.outcome === "completed" ? route.on_link ? "On-link route; no separate gateway." : route.next_hop ? `Next hop ${route.next_hop}.` : "No separate gateway recorded." : `Route selection ${outcome(route.outcome)}.`}</p><p>{path.gateway.attempted} gateway echoes attempted; {path.gateway.replied} replied.</p>{path.gateway.samples.map((sample) => <p key={sample.id}><EchoText echo={sample} /> <EvidenceButton id={sample.id} showEvidence={showEvidence} /></p>)}<EvidenceButton id={route.id} showEvidence={showEvidence} /></div>
    <ol className="hop-list">{trace.hops.map((hop) => <li key={hop.id}><div className="hop"><span className="ttl">Hop {hop.ttl}</span><span><EchoText echo={hop} />{hop.status === "time_exceeded" && hop.rtt_ms !== undefined && <small> This is one router’s ICMP response; its RTT is not link delay or proof of a bottleneck.</small>}</span><EvidenceButton id={hop.id} showEvidence={showEvidence} /></div></li>)}</ol>
    {trace.hops.length === 0 && <p>No trace hops recorded.</p>}
    <p><strong>Trace outcome:</strong> {outcome(trace.outcome)} · {trace.highest_attempted_ttl} of {trace.max_hops} allowed hops attempted. <EvidenceButton id={trace.id} showEvidence={showEvidence} /></p>
    <div className="echo-summary"><h4>Target echo</h4><p>{path.target.attempted} attempted; {path.target.replied} replied. ICMP echo is separate from website timing and does not measure page load.</p>{path.target.samples.map((sample) => <p key={sample.id}><EchoText echo={sample} /> <EvidenceButton id={sample.id} showEvidence={showEvidence} /></p>)}{path.target.variation && <p>RTT variation: {path.target.variation.value_ms !== undefined ? ms(path.target.variation.value_ms) : outcome(path.target.variation.outcome)}. <EvidenceButton id={path.target.variation.id} showEvidence={showEvidence} /></p>}</div>
  </>;
}
function EvidenceItem({ id, item }: { id: string; item: Record<string, unknown> }) {
  const text = (key: string) => typeof item[key] === "string" ? item[key] as string : undefined;
  const number = (key: string) => typeof item[key] === "number" ? ms(item[key] as number) : undefined;
  const labels = [text("outcome"), text("status"), text("method")].filter((value): value is string => !!value);
  const error = item.error as { message?: string } | undefined;
  return <section id={`evidence-${id}`} className="evidence-item" tabIndex={-1}><h4>{id}</h4><p>{labels.map(outcome).join(" · ") || "Recorded evidence"}</p><dl>{[["Duration", number("duration_ms")], ["Bound", number("bound_ms")], ["RTT", number("rtt_ms")], ["Started", text("started_at")], ["Ended", text("ended_at")], ["Selected IP", text("selected_ip")], ["Reply address", text("reply_address")]].filter((pair) => pair[1]).map(([label, value]) => <div key={label}><dt>{label}</dt><dd>{value}</dd></div>)}</dl>{error?.message && <p className="error">{error.message}</p>}<details><summary>Raw evidence</summary><pre>{JSON.stringify(item, null, 2)}</pre></details></section>;
}

createRoot(document.getElementById("root")!).render(<App />);
