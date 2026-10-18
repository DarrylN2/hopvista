// Package run owns the in-memory diagnostic-run lifecycle. Its executor can be
// replaced by the measurement engine without changing the local API contract.
package run

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"hopvista/internal/result"
)

const MeasurementWindow = 60 * time.Second

var ErrBusy = errors.New("a diagnostic run is already active")
var ErrNotFound = errors.New("diagnostic run not found")

type Input struct {
	TargetURL string
	MaxHops   int
}

// Executor must return only after all synchronous native work and its cleanup
// finish. Publish may retain completed evidence while work is still pending.
// Findings are deliberately absent from all nonterminal snapshots.
type Executor interface {
	Run(context.Context, Input, func([]result.Family)) ([]result.Family, []result.Finding, error)
}

type entry struct {
	value         result.Run
	ctx           context.Context
	cancel        context.CancelFunc
	userCancelled bool
}

type Manager struct {
	mu       sync.Mutex
	executor Executor
	window   time.Duration
	active   *entry
	latest   *entry
}

func NewManager(executor Executor) *Manager {
	return &Manager{executor: executor, window: MeasurementWindow}
}

// NewManagerWithWindow is for deterministic cutoff tests; production uses 60 s.
func NewManagerWithWindow(executor Executor, window time.Duration) *Manager {
	return &Manager{executor: executor, window: window}
}

func (m *Manager) Start(input Input) (result.Run, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.active != nil {
		return result.Run{}, ErrBusy
	}
	if m.executor == nil || m.window <= 0 {
		return result.Run{}, errors.New("run executor is unavailable")
	}
	id, err := randomID()
	if err != nil {
		return result.Run{}, err
	}
	started := time.Now().UTC()
	deadline := started.Add(m.window)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	e := &entry{value: result.Run{
		SchemaVersion: result.SchemaVersion, RunID: id, TargetURL: input.TargetURL,
		ConnectionMode: "direct", MaxHops: input.MaxHops, StartedAt: started,
		DeadlineAt: deadline, State: result.Running, Families: []result.Family{}, Findings: []result.Finding{},
	}, ctx: ctx, cancel: cancel}
	m.active = e
	m.latest = nil // restart evicts the preceding terminal result
	go m.execute(ctx, e, input)
	go m.observeStop(ctx, e)
	return clone(e.value), nil
}

func (m *Manager) execute(ctx context.Context, e *entry, input Input) {
	defer e.cancel()
	families, findings, err := m.executor.Run(ctx, input, func(partial []result.Family) {
		m.mu.Lock()
		defer m.mu.Unlock()
		if m.active == e {
			e.value.Families = cloneFamilies(partial)
		}
	})
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.active != e {
		return
	}
	// When context ends, stop time must precede quiescence even if the
	// observer goroutine has not yet acquired the lock.
	if ctx.Err() != nil {
		m.markStopping(e)
	}
	if families != nil {
		e.value.Families = cloneFamilies(families)
	}
	ended := time.Now().UTC()
	e.value.EndedAt = &ended
	switch {
	case e.userCancelled:
		e.value.State = result.RunCancelled
	case err != nil && !errors.Is(err, context.DeadlineExceeded) && ctx.Err() != context.DeadlineExceeded:
		e.value.State = result.Failed
		e.value.Error = &result.ErrorData{Code: "orchestration_failed", Message: "The diagnostic run could not produce a coherent result."}
	default:
		e.value.State = result.RunCompleted
	}
	if e.value.State == result.RunCompleted {
		e.value.Findings = cloneFindings(findings)
	}
	m.active, m.latest = nil, e
}

func (m *Manager) observeStop(ctx context.Context, e *entry) {
	<-ctx.Done()
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.active == e {
		m.markStopping(e)
	}
}

func (m *Manager) markStopping(e *entry) {
	if e.value.StopRequestedAt == nil {
		now := time.Now().UTC()
		e.value.StopRequestedAt = &now
	}
	e.value.State = result.Stopping
}

// Cancel is idempotent for the current run. It releases no active slot until
// the executor has returned and all resources it owns are quiescent.
func (m *Manager) Cancel(id string) (result.Run, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.active != nil && m.active.value.RunID == id {
		e := m.active
		if e.value.State != result.Stopping {
			e.userCancelled = e.ctx.Err() == nil && time.Now().Before(e.value.DeadlineAt)
			m.markStopping(e)
			e.cancel()
		}
		return clone(e.value), nil
	}
	if m.latest != nil && m.latest.value.RunID == id {
		return clone(m.latest.value), nil
	}
	return result.Run{}, ErrNotFound
}

func (m *Manager) Get(id string) (result.Run, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.active != nil && m.active.value.RunID == id {
		return clone(m.active.value), nil
	}
	if m.latest != nil && m.latest.value.RunID == id {
		return clone(m.latest.value), nil
	}
	return result.Run{}, ErrNotFound
}

func randomID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw[:]), nil
}

// JSON round trips detach nested slices and pointers before a caller encodes
// the snapshot outside the lock. All stored values are JSON contract values.
func clone(v result.Run) result.Run {
	b, _ := json.Marshal(v)
	var out result.Run
	_ = json.Unmarshal(b, &out)
	return out
}
func cloneFamilies(v []result.Family) []result.Family {
	if v == nil {
		return []result.Family{}
	}
	b, _ := json.Marshal(v)
	var out []result.Family
	_ = json.Unmarshal(b, &out)
	return out
}
func cloneFindings(v []result.Finding) []result.Finding {
	if v == nil {
		return []result.Finding{}
	}
	b, _ := json.Marshal(v)
	var out []result.Finding
	_ = json.Unmarshal(b, &out)
	return out
}
