package dlq

import (
	"context"
	"testing"
	"time"
)

func newTestScheduler(t *testing.T) (*Scheduler, *Manager, Store) {
	t.Helper()
	m, store := newTestManager(t)
	sched := NewScheduler(m, store, 50*time.Millisecond)
	return sched, m, store
}

func TestNewScheduler_DefaultInterval(t *testing.T) {
	m, store := newTestManager(t)
	s := NewScheduler(m, store, 0)
	if s.interval != 30*time.Second {
		t.Errorf("expected default interval 30s, got %v", s.interval)
	}
}

func TestScheduler_StartStop(t *testing.T) {
	sched, _, _ := newTestScheduler(t)
	ctx := context.Background()

	if err := sched.Start(ctx); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	// Second Start should be a no-op
	if err := sched.Start(ctx); err != nil {
		t.Fatalf("second Start returned error: %v", err)
	}
	sched.Stop()
	// Second Stop should be a no-op
	sched.Stop()
}

func TestScheduler_RetriesDueJobs(t *testing.T) {
	sched, manager, _ := newTestScheduler(t)
	ctx := context.Background()

	// Enqueue a job that has already exceeded its next retry time.
	job := NewJob("test-queue", []byte(`{"action":"send_email"}`), nil)
	job.Attempts = 1
	job.NextRetryAt = time.Now().Add(-1 * time.Second) // due in the past

	if err := manager.store.Save(ctx, job); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Run a single tick directly to avoid timing flakiness.
	sched.tick(ctx)

	updated, err := manager.store.Get(ctx, job.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if updated.Attempts <= job.Attempts {
		t.Errorf("expected attempts to increase after retry tick, got %d", updated.Attempts)
	}
}

func TestScheduler_SkipsTerminalJobs(t *testing.T) {
	sched, manager, _ := newTestScheduler(t)
	ctx := context.Background()

	job := NewJob("test-queue", []byte(`{"action":"noop"}`), nil)
	// Exhaust all retries so the job is terminal.
	job.Attempts = manager.policy.MaxRetries
	job.NextRetryAt = time.Now().Add(-1 * time.Second)

	if err := manager.store.Save(ctx, job); err != nil {
		t.Fatalf("Save: %v", err)
	}

	sched.tick(ctx)

	updated, err := manager.store.Get(ctx, job.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if updated.Attempts != job.Attempts {
		t.Errorf("terminal job should not be retried, attempts changed to %d", updated.Attempts)
	}
}
