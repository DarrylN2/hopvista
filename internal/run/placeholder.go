package run

import (
	"context"
	"time"

	"hopvista/internal/result"
)

// Placeholder makes the early app usable for local API review while recording
// honestly that no network measurement engine is connected yet.
type Placeholder struct{}

func (Placeholder) Run(ctx context.Context, _ Input, _ func([]result.Family)) ([]result.Family, []result.Finding, error) {
	t := time.NewTimer(200 * time.Millisecond)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
	code := "measurement_not_connected"
	message := "Network measurements are not connected in this milestone."
	outcome := result.Unavailable
	if ctx.Err() == context.Canceled {
		code, message, outcome = "cancelled", "The diagnostic run was cancelled.", result.Cancelled
	}
	if ctx.Err() == context.DeadlineExceeded {
		code, message, outcome = "deadline", "The measurement cutoff was reached.", result.Deadline
	}
	return []result.Family{{Family: "ipv4", Outcome: outcome, Error: &result.ErrorData{Code: code, Message: message}, DNS: []result.DNS{}, Exchanges: []result.Exchange{}, Paths: []result.Path{}}}, []result.Finding{}, nil
}
