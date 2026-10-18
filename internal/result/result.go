// Package result defines the platform-independent schema_version 1 wire contract.
// Optional numeric measurements are pointers: absent evidence must not become zero.
package result

import "time"

const SchemaVersion = 1
const DirectPathCaveat = "HopVista tests a direct connection from this computer. Your browser may use another IP, a proxy, cached DNS, or an existing connection."

type Outcome string

const (
	Pending     Outcome = "pending" // phase scheduled but not yet finished in a running snapshot
	Completed   Outcome = "completed"
	Timeout     Outcome = "timeout"
	Deadline    Outcome = "deadline"
	Error       Outcome = "error"
	Skipped     Outcome = "skipped"
	Unavailable Outcome = "unavailable"
	Cancelled   Outcome = "cancelled"
)

type State string

const (
	Running      State = "running"
	Stopping     State = "stopping"
	RunCompleted State = "completed"
	RunCancelled State = "cancelled"
	Failed       State = "failed"
)

func (s State) Terminal() bool { return s == RunCompleted || s == RunCancelled || s == Failed }

// ErrorData uses portable codes. Native structs, handles and numeric statuses
// belong to adapter diagnostics, never to coordinator decisions.
type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Measurement struct {
	ID         string     `json:"id"`
	Outcome    Outcome    `json:"outcome"`
	Method     string     `json:"method"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	EndedAt    *time.Time `json:"ended_at,omitempty"`
	BoundMS    *float64   `json:"bound_ms,omitempty"`
	DurationMS *float64   `json:"duration_ms,omitempty"`
	Error      *ErrorData `json:"error,omitempty"`
}

type Run struct {
	SchemaVersion   int        `json:"schema_version"`
	RunID           string     `json:"run_id"`
	TargetURL       string     `json:"target_url"`
	ConnectionMode  string     `json:"connection_mode"`
	MaxHops         int        `json:"max_hops"`
	StartedAt       time.Time  `json:"started_at"`
	DeadlineAt      time.Time  `json:"deadline_at"`
	StopRequestedAt *time.Time `json:"stop_requested_at,omitempty"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
	State           State      `json:"state"`
	Error           *ErrorData `json:"error,omitempty"`
	Families        []Family   `json:"families"`
	Findings        []Finding  `json:"findings"`
}

type Family struct {
	Family    string     `json:"family"`
	Outcome   Outcome    `json:"outcome"`
	Error     *ErrorData `json:"error,omitempty"`
	DNS       []DNS      `json:"dns"`
	Exchanges []Exchange `json:"exchanges"`
	Paths     []Path     `json:"paths"`
	FinalURL  string     `json:"final_url,omitempty"`
	FinalHost string     `json:"final_host,omitempty"`
}

type DNS struct {
	Measurement
	Hostname      string   `json:"hostname"`
	Addresses     []string `json:"addresses"`
	SelectedIP    string   `json:"selected_ip,omitempty"`
	SelectedScope string   `json:"selected_scope,omitempty"`
}

type Exchange struct {
	ID                 string      `json:"id"`
	RedirectIndex      int         `json:"redirect_index"`
	RequestedURL       string      `json:"requested_url"`
	Hostname           string      `json:"hostname"`
	Port               int         `json:"port"`
	SelectedIP         string      `json:"selected_ip"`
	SelectedScope      string      `json:"selected_scope"`
	ActualPeerIP       string      `json:"actual_peer_ip,omitempty"`
	DNSRef             string      `json:"dns_ref,omitempty"`
	DNS                Measurement `json:"dns"`
	TCP                Measurement `json:"tcp"`
	TLS                Measurement `json:"tls"`
	HTTP               Measurement `json:"http"`
	ResponseStatus     *int        `json:"response_status,omitempty"`
	Location           string      `json:"location,omitempty"`
	RedirectStopReason string      `json:"redirect_stop_reason,omitempty"`
	ConnectionReused   *bool       `json:"connection_reused,omitempty"`
}

type Route struct {
	Measurement
	SelectedIP     string  `json:"selected_ip"`
	SourceIP       string  `json:"source_ip,omitempty"`
	InterfaceIndex *uint32 `json:"interface_index,omitempty"`
	NextHop        string  `json:"next_hop,omitempty"`
	OnLink         *bool   `json:"on_link,omitempty"`
}

type EchoStatus string

const (
	EchoReply        EchoStatus = "echo_reply"
	TimeExceeded     EchoStatus = "time_exceeded"
	UnreachableReply EchoStatus = "unreachable_reply"
	NoReply          EchoStatus = "no_reply"
	NotSent          EchoStatus = "not_sent"
	CallError        EchoStatus = "call_error"
)

type Echo struct {
	Measurement
	SelectedIP   string     `json:"selected_ip"`
	TTL          int        `json:"ttl"`
	Status       EchoStatus `json:"status"`
	ReplyAddress string     `json:"reply_address,omitempty"`
	RTTMS        *float64   `json:"rtt_ms,omitempty"`
}

type EchoSeries struct {
	ID        string     `json:"id"`
	Outcome   Outcome    `json:"outcome"`
	Attempted int        `json:"attempted"`
	Replied   int        `json:"replied"`
	Samples   []Echo     `json:"samples"`
	Variation *Variation `json:"variation,omitempty"`
}

type Variation struct {
	ID           string   `json:"id"`
	Outcome      Outcome  `json:"outcome"`
	ValueMS      *float64 `json:"value_ms,omitempty"`
	Formula      string   `json:"formula"`
	EvidenceRefs []string `json:"evidence_refs"`
}

type TraceOutcome string

const (
	TracePending        TraceOutcome = "pending"
	DestinationReplied  TraceOutcome = "destination_replied"
	UnreachableResponse TraceOutcome = "unreachable_response"
	HopLimitReached     TraceOutcome = "hop_limit_reached"
	TraceCancelled      TraceOutcome = "cancelled"
	TraceDeadline       TraceOutcome = "deadline"
	TraceUnavailable    TraceOutcome = "unavailable"
	TraceError          TraceOutcome = "error"
)

type Trace struct {
	ID                  string       `json:"id"`
	SelectedIP          string       `json:"selected_ip,omitempty"`
	Family              string       `json:"family"`
	TargetRole          string       `json:"target_role"` // final_host or attempted_host
	MaxHops             int          `json:"max_hops"`
	HighestAttemptedTTL int          `json:"highest_attempted_ttl"`
	Outcome             TraceOutcome `json:"outcome"`
	Hops                []Echo       `json:"hops"`
}

type Path struct {
	ID         string     `json:"id"`
	SelectedIP string     `json:"selected_ip"`
	Route      Route      `json:"route"`
	Gateway    EchoSeries `json:"gateway"`
	Target     EchoSeries `json:"target"`
	Trace      Trace      `json:"trace"`
}

type Finding struct {
	Type         string   `json:"type"`
	Family       string   `json:"family"`
	ExchangeRef  string   `json:"exchange_ref,omitempty"`
	Observation  string   `json:"observation"`
	Limit        string   `json:"limit"`
	EvidenceRefs []string `json:"evidence_refs"`
}
