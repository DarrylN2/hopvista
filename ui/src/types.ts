export type Outcome = "pending" | "completed" | "timeout" | "deadline" | "error" | "skipped" | "unavailable" | "cancelled";
export type RunState = "running" | "stopping" | "completed" | "cancelled" | "failed";

export interface ErrorData { code: string; message: string }
export interface Measurement {
  id: string;
  outcome: Outcome;
  method: string;
  started_at?: string;
  ended_at?: string;
  bound_ms?: number;
  duration_ms?: number;
  error?: ErrorData;
}
export interface DNS extends Measurement {
  hostname: string;
  addresses: string[];
  selected_ip?: string;
  selected_scope?: string;
}
export interface Exchange {
  id: string;
  redirect_index: number;
  requested_url: string;
  hostname: string;
  port: number;
  selected_ip: string;
  selected_scope: string;
  actual_peer_ip?: string;
  dns_ref?: string;
  dns: Measurement;
  tcp: Measurement;
  tls: Measurement;
  http: Measurement;
  response_status?: number;
  location?: string;
  redirect_stop_reason?: string;
  connection_reused?: boolean;
}
export interface Echo extends Measurement {
  selected_ip: string;
  ttl: number;
  status: "echo_reply" | "time_exceeded" | "unreachable_reply" | "no_reply" | "not_sent" | "call_error";
  reply_address?: string;
  rtt_ms?: number;
}
export interface EchoSeries {
  id: string;
  outcome: Outcome;
  attempted: number;
  replied: number;
  samples: Echo[];
  variation?: { id: string; outcome: Outcome; value_ms?: number; formula: string; evidence_refs: string[] };
}
export interface Route extends Measurement {
  selected_ip: string;
  source_ip?: string;
  interface_index?: number;
  next_hop?: string;
  on_link?: boolean;
}
export interface Trace {
  id: string;
  selected_ip?: string;
  family: string;
  target_role: "final_host" | "attempted_host";
  max_hops: number;
  highest_attempted_ttl: number;
  outcome: string;
  hops: Echo[];
}
export interface Path { id: string; selected_ip: string; route: Route; gateway: EchoSeries; target: EchoSeries; trace: Trace }
export interface Family {
  family: string;
  outcome: Outcome;
  error?: ErrorData;
  dns: DNS[];
  exchanges: Exchange[];
  paths: Path[];
  final_url?: string;
  final_host?: string;
}
export interface Finding {
  type: string;
  family: string;
  exchange_ref?: string;
  observation: string;
  limit: string;
  evidence_refs: string[];
}
export interface RunResult {
  schema_version: number;
  run_id: string;
  target_url: string;
  connection_mode: "direct";
  max_hops: number;
  started_at: string;
  deadline_at: string;
  stop_requested_at?: string;
  ended_at?: string;
  state: RunState;
  error?: ErrorData;
  families: Family[];
  findings: Finding[];
}
