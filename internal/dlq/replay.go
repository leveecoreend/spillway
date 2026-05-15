package dlq

import (
	"context"
	"fmt"
	"time"
)

// ReplayResult summarizes the outcome of a replay operation.
type ReplayResult struct {
	Total    int
	Replayed int
	Skipped  int
	Failed   int
	Errors   []error
}

// Replayer re-enqueues dead-letter jobs that match a given filter back into
// the live queue via the Router, resetting their retry state.
type Replayer struct {
	store  Store
	router *Router
}

// NewReplayer creates a Replayer backed by the provided store and router.
func NewReplayer(store Store, router *Router) (*Replayer, error) {
	if store == nil {
		return nil, fmt.Errorf("replay: store must not be nil")
	}
	if router == nil {
		return nil, fmt.Errorf("replay: router must not be nil")
	}
	return &Replayer{store: store, router: router}, nil
}

// Replay iterates over all jobs matching f, dispatches each through the router,
// and deletes it from the DLQ on success. Terminal jobs are skipped unless
// forceTerminal is true.
func (r *Replayer) Replay(ctx context.Context, f Filter, forceTerminal bool) (*ReplayResult, error) {
	jobs, err := r.store.List(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("replay: listing jobs: %w", err)
	}

	result := &ReplayResult{Total: len(jobs)}

	for _, j := range jobs {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}

		if j.IsTerminal() && !forceTerminal {
			result.Skipped++
			continue
		}

		// Reset retry state so the job gets a fresh start.
		j.Retries = 0
		j.LastError = ""
		j.Status = StatusPending
		j.UpdatedAt = time.Now().UTC()

		if dispatchErr := r.router.Dispatch(ctx, j); dispatchErr != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Errorf("job %s: %w", j.ID, dispatchErr))
			continue
		}

		if delErr := r.store.Delete(ctx, j.ID); delErr != nil {
			// Non-fatal: job was re-enqueued but we couldn't clean up.
			result.Errors = append(result.Errors, fmt.Errorf("job %s delete: %w", j.ID, delErr))
		}

		result.Replayed++
	}

	return result, nil
}
