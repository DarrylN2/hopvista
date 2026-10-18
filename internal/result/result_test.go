package result_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"hopvista/internal/result"
)

// Golden fixtures are authored independently of the Go serializer. Strict
// decoding catches removed/renamed fields; structural round trips catch missing
// required fields, omitted zero values, and unexpected additions on the wire.
func TestVersionOneJSONFixtures(t *testing.T) {
	files, err := filepath.Glob("testdata/*.json")
	if err != nil || len(files) != 3 {
		t.Fatalf("fixtures: %v %v", files, err)
	}
	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			var run result.Run
			decoder := json.NewDecoder(bytes.NewReader(data))
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&run); err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(run)
			if err != nil {
				t.Fatal(err)
			}
			var want, got any
			if err := json.Unmarshal(data, &want); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(encoded, &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("schema changed\nwant: %s\ngot: %s", data, encoded)
			}
			checkContract(t, run)
		})
	}
}

func checkContract(t *testing.T, r result.Run) {
	t.Helper()
	if r.SchemaVersion != result.SchemaVersion || r.ConnectionMode != "direct" || r.MaxHops < 1 || r.MaxHops > 64 {
		t.Fatal("invalid run contract")
	}
	if !r.State.Terminal() || r.EndedAt == nil || len(r.Families) != 1 || r.Families[0].Family != "ipv4" {
		t.Fatal("invalid terminal IPv4 fixture")
	}
	ids := map[string]bool{}
	addID := func(id string) {
		t.Helper()
		if id == "" || ids[id] {
			t.Fatalf("empty/duplicate evidence id: %q", id)
		}
		ids[id] = true
	}
	measurement := func(m result.Measurement) {
		t.Helper()
		addID(m.ID)
		if m.Method == "" {
			t.Fatal("missing method")
		}
		if m.DurationMS != nil && (m.Outcome != result.Completed || m.StartedAt == nil || m.EndedAt == nil || *m.DurationMS < 0) {
			t.Fatalf("invented duration: %+v", m)
		}
		if m.Outcome == result.Completed && (m.DurationMS == nil || m.StartedAt == nil || m.EndedAt == nil) {
			t.Fatalf("missing completed duration: %+v", m)
		}
		for _, at := range []*time.Time{m.StartedAt, m.EndedAt} {
			if at != nil {
				_, offset := at.Zone()
				if offset != 0 {
					t.Fatal("timestamp must be UTC")
				}
			}
		}
	}
	echo := func(e result.Echo) {
		t.Helper()
		measurement(e.Measurement)
		if e.Outcome != result.Completed && (e.RTTMS != nil || e.ReplyAddress != "") {
			t.Fatalf("invented reply: %+v", e)
		}
		if e.Status == result.NotSent && (e.StartedAt != nil || e.EndedAt != nil) {
			t.Fatalf("unstarted probe has send time: %+v", e)
		}
		if e.Outcome == result.Completed && (e.RTTMS == nil || e.ReplyAddress == "") {
			t.Fatal("completed ICMP reply missing RTT/address")
		}
	}
	var refs []string
	for _, f := range r.Families {
		for _, d := range f.DNS {
			measurement(d.Measurement)
		}
		for _, e := range f.Exchanges {
			addID(e.ID)
			measurement(e.DNS)
			measurement(e.TCP)
			measurement(e.TLS)
			measurement(e.HTTP)
			if e.DNSRef != "" {
				refs = append(refs, e.DNSRef)
			}
		}
		for _, p := range f.Paths {
			addID(p.ID)
			measurement(p.Route.Measurement)
			if p.Route.OnLink != nil && *p.Route.OnLink && p.Route.NextHop != "" {
				t.Fatal("on-link has a separate gateway")
			}
			for _, s := range []result.EchoSeries{p.Gateway, p.Target} {
				addID(s.ID)
				attempted, replied := 0, 0
				for _, e := range s.Samples {
					echo(e)
					if e.StartedAt != nil {
						attempted++
					}
					if e.Status == result.EchoReply {
						replied++
					}
				}
				if s.Attempted != attempted || s.Replied != replied {
					t.Fatalf("counts do not match raw samples: %+v", s)
				}
				if s.Variation != nil {
					addID(s.Variation.ID)
					refs = append(refs, s.Variation.EvidenceRefs...)
					if s.Variation.Outcome == result.Unavailable && s.Variation.ValueMS != nil {
						t.Fatal("unavailable variation has numeric value")
					}
				}
			}
			addID(p.Trace.ID)
			if p.Trace.MaxHops != r.MaxHops {
				t.Fatal("effective max_hops changed")
			}
			for _, h := range p.Trace.Hops {
				echo(h)
			}
		}
	}
	for _, f := range r.Findings {
		if f.Limit == "" || f.Observation == "" || len(f.EvidenceRefs) == 0 {
			t.Fatal("finding missing evidence or uncertainty")
		}
		refs = append(refs, f.EvidenceRefs...)
		if f.ExchangeRef != "" {
			refs = append(refs, f.ExchangeRef)
		}
	}
	for _, ref := range refs {
		if !ids[ref] {
			t.Fatalf("dangling evidence reference: %s", ref)
		}
	}
}

func TestMissingTimingIsAbsentAndMeasuredZeroIsRetained(t *testing.T) {
	zero := 0.0
	for _, tc := range []struct {
		m    result.Measurement
		want string
	}{
		{result.Measurement{ID: "literal", Outcome: result.Skipped, Method: "literal_ip"}, `{"id":"literal","outcome":"skipped","method":"literal_ip"}`},
		{result.Measurement{ID: "reply", Outcome: result.Completed, Method: "icmp_echo", DurationMS: &zero}, `{"id":"reply","outcome":"completed","method":"icmp_echo","duration_ms":0}`},
	} {
		got, err := json.Marshal(tc.m)
		if err != nil || string(got) != tc.want {
			t.Fatalf("%s (%v), want %s", got, err, tc.want)
		}
	}
}

func TestLifecycleNamesAndTerminalMeaning(t *testing.T) {
	for _, state := range []result.State{result.Running, result.Stopping} {
		if state.Terminal() {
			t.Fatalf("%s is still owned", state)
		}
	}
	for _, state := range []result.State{result.RunCompleted, result.RunCancelled, result.Failed} {
		if !state.Terminal() {
			t.Fatal(state)
		}
	}
	if result.State("unknown").Terminal() {
		t.Fatal("unknown state is terminal")
	}
}

func TestStableOutcomeAndStatusNames(t *testing.T) {
	outcomes := []result.Outcome{result.Completed, result.Timeout, result.Deadline, result.Error, result.Skipped, result.Unavailable, result.Cancelled}
	want := `["completed","timeout","deadline","error","skipped","unavailable","cancelled"]`
	got, err := json.Marshal(outcomes)
	if err != nil || string(got) != want {
		t.Fatalf("%s (%v)", got, err)
	}
	traces := []result.TraceOutcome{result.DestinationReplied, result.UnreachableResponse, result.HopLimitReached, result.TraceCancelled, result.TraceDeadline, result.TraceUnavailable, result.TraceError}
	want = `["destination_replied","unreachable_response","hop_limit_reached","cancelled","deadline","unavailable","error"]`
	got, err = json.Marshal(traces)
	if err != nil || string(got) != want {
		t.Fatalf("%s (%v)", got, err)
	}
	statuses := []result.EchoStatus{result.EchoReply, result.TimeExceeded, result.UnreachableReply, result.NoReply, result.NotSent, result.CallError}
	want = `["echo_reply","time_exceeded","unreachable_reply","no_reply","not_sent","call_error"]`
	got, err = json.Marshal(statuses)
	if err != nil || string(got) != want {
		t.Fatalf("%s (%v)", got, err)
	}
}
