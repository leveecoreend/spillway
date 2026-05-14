package dlq

import (
	"context"
	"fmt"
	"time"
)

// Manager orchestrates retry scheduling for dead-letter jobs.
type Manager struct {
	store  Store
	policy RetryPolicy
}

// NewManager creates a Manager with the given store and retry policy.
func NewManager(store Store, policy RetryPolicy) (*Manager, error) {
	if err := policy.Validate(); err != nil {
		return nil, fmt.Errorf("invalid retry policy: %w", err)
	}
	return &Manager{store: store, policy: policy}, nil
}

// Enqueue saves a new job to the dead-letter store.
func (m *Manager) Enqueue(ctx context.Context, j *Job) error {
	if j == nil {
		return fmt.Errorf("job must not be nil")
	}
	return m.store.Save(ctx, j)
}

// Retry attempts to reschedule a job for another execution.
// It increments the attempt counter and sets the next retry time.
func (m *Manager) Retry(ctx context.Context, id string) (*Job, error) {
	j, err := m.store.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get job %s: %w", id, err)
	}
	if !j.CanRetry(m.policy.MaxAttempts) {
		return nil, fmt.Errorf("job %s has reached max attempts (%d)", id, m.policy.MaxAttempts)
	}
	j.Attempts++
	delay := m.policy.NextInterval(j.Attempts)
	next := time.Now().Add(delay)
	j.NextRetryAt = &next
	if err := m.store.Save(ctx, j); err != nil {
		return nil, fmt.Errorf("save job %s: %w", id, err)
	}
	return j, nil
}

// Discard marks a job as terminal and removes it from the active store.
func (m *Manager) Discard(ctx context.Context, id string) error {
	if _, err := m.store.Get(ctx, id); err != nil {
		return fmt.Errorf("get job %s: %w", id, err)
	}
	return m.store.Delete(ctx, id)
}

// PendingRetries returns all jobs that are due for a retry.
func (m *Manager) PendingRetries(ctx context.Context) ([]*Job, error) {
	all, err := m.store.List(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	var pending []*Job
	for _, j := range all {
		if j.NextRetryAt != nil && !j.NextRetryAt.After(now) {
			pending = append(pending, j)
		}
	}
	return pending, nil
}
