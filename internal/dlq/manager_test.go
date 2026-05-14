package dlq

import (
	"context"
	"testing"
	"time"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	store := NewInMemoryStore()
	policy := DefaultRetryPolicy()
	m, err := NewManager(store, policy)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	return m
}

func TestNewManager_InvalidPolicy(t *testing.T) {
	_, err := NewManager(NewInMemoryStore(), RetryPolicy{MaxAttempts: 0})
	if err == nil {
		t.Fatal("expected error for invalid policy")
	}
}

func TestManager_Enqueue(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()
	j := NewJob("q1", "send-email", []byte(`{}`))
	if err := m.Enqueue(ctx, j); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	got, err := m.store.Get(ctx, j.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != j.ID {
		t.Errorf("expected ID %s, got %s", j.ID, got.ID)
	}
}

func TestManager_Enqueue_NilJob(t *testing.T) {
	m := newTestManager(t)
	if err := m.Enqueue(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil job")
	}
}

func TestManager_Retry(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()
	j := NewJob("q1", "process", []byte(`{}`))
	_ = m.Enqueue(ctx, j)

	updated, err := m.Retry(ctx, j.ID)
	if err != nil {
		t.Fatalf("Retry: %v", err)
	}
	if updated.Attempts != 1 {
		t.Errorf("expected attempts=1, got %d", updated.Attempts)
	}
	if updated.NextRetryAt == nil {
		t.Fatal("expected NextRetryAt to be set")
	}
	if updated.NextRetryAt.Before(time.Now()) {
		t.Error("NextRetryAt should be in the future")
	}
}

func TestManager_Retry_MaxAttemptsExceeded(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()
	j := NewJob("q1", "process", []byte(`{}`))
	j.Attempts = m.policy.MaxAttempts
	_ = m.Enqueue(ctx, j)

	_, err := m.Retry(ctx, j.ID)
	if err == nil {
		t.Fatal("expected error when max attempts reached")
	}
}

func TestManager_Discard(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()
	j := NewJob("q1", "cleanup", []byte(`{}`))
	_ = m.Enqueue(ctx, j)

	if err := m.Discard(ctx, j.ID); err != nil {
		t.Fatalf("Discard: %v", err)
	}
	_, err := m.store.Get(ctx, j.ID)
	if err == nil {
		t.Fatal("expected job to be removed after discard")
	}
}

func TestManager_PendingRetries(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()

	past := time.Now().Add(-time.Minute)
	future := time.Now().Add(time.Hour)

	jDue := NewJob("q1", "due-job", []byte(`{}`))
	jDue.NextRetryAt = &past
	_ = m.Enqueue(ctx, jDue)

	jNotDue := NewJob("q1", "future-job", []byte(`{}`))
	jNotDue.NextRetryAt = &future
	_ = m.Enqueue(ctx, jNotDue)

	pending, err := m.PendingRetries(ctx)
	if err != nil {
		t.Fatalf("PendingRetries: %v", err)
	}
	if len(pending) != 1 {
		t.Errorf("expected 1 pending job, got %d", len(pending))
	}
	if pending[0].ID != jDue.ID {
		t.Errorf("expected job %s, got %s", jDue.ID, pending[0].ID)
	}
}
