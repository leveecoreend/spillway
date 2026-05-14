package dlq

import (
	"context"
	"testing"
	"time"
)

// TestScheduler_Integration verifies the full Start/tick/Stop lifecycle
// using a real ticker at a short interval.
func TestScheduler_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	m, store := newTestManager(t)
	sched := NewScheduler(m, store, 20*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Insert a job that is immediately due for retry.
	job := NewJob("integration-queue", []byte(`{"task":"integrate"}`), nil)
	job.Attempts = 1
	job.NextRetryAt = time.Now().Add(-10 * time.Millisecond)

	if err := store.Save(ctx, job); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := sched.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer sched.Stop()

	// Poll until the job's attempt count increases or we time out.
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		updated, err := store.Get(ctx, job.ID)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if updated.Attempts > job.Attempts {
			// Success — scheduler picked up and retried the job.
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Error("scheduler did not retry the due job within the deadline")
}
