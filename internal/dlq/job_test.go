package dlq

import (
	"testing"
	"time"
)

func TestNewJob(t *testing.T) {
	payload := []byte(`{"task":"send_email"}`)
	job := NewJob("job-001", "emails", "redis", payload, "connection refused", 3)

	if job.ID != "job-001" {
		t.Errorf("expected ID job-001, got %s", job.ID)
	}
	if job.Queue != "emails" {
		t.Errorf("expected queue emails, got %s", job.Queue)
	}
	if job.Backend != "redis" {
		t.Errorf("expected backend redis, got %s", job.Backend)
	}
	if job.Status != StatusPending {
		t.Errorf("expected status pending, got %s", job.Status)
	}
	if job.Attempts != 0 {
		t.Errorf("expected 0 attempts, got %d", job.Attempts)
	}
	if job.MaxAttempts != 3 {
		t.Errorf("expected max_attempts 3, got %d", job.MaxAttempts)
	}
	if job.Metadata == nil {
		t.Error("expected non-nil metadata map")
	}
	if job.CreatedAt.IsZero() || job.UpdatedAt.IsZero() {
		t.Error("expected non-zero timestamps")
	}
}

func TestJob_CanRetry(t *testing.T) {
	job := NewJob("j1", "q", "b", nil, "err", 3)

	job.Attempts = 0
	if !job.CanRetry() {
		t.Error("expected CanRetry true when attempts=0, max=3")
	}

	job.Attempts = 3
	if job.CanRetry() {
		t.Error("expected CanRetry false when attempts=max")
	}

	job.Attempts = 4
	if job.CanRetry() {
		t.Error("expected CanRetry false when attempts exceed max")
	}
}

func TestJob_IsTerminal(t *testing.T) {
	job := NewJob("j1", "q", "b", nil, "err", 3)

	nonTerminal := []JobStatus{StatusPending, StatusRetrying}
	for _, s := range nonTerminal {
		job.Status = s
		if job.IsTerminal() {
			t.Errorf("expected IsTerminal false for status %s", s)
		}
	}

	terminal := []JobStatus{StatusSucceeded, StatusFailed, StatusDiscarded}
	for _, s := range terminal {
		job.Status = s
		if !job.IsTerminal() {
			t.Errorf("expected IsTerminal true for status %s", s)
		}
	}
}

func TestJob_NextRetryAt(t *testing.T) {
	job := NewJob("j1", "q", "b", nil, "err", 3)
	if job.NextRetryAt != nil {
		t.Error("expected nil NextRetryAt on new job")
	}

	next := time.Now().Add(5 * time.Minute)
	job.NextRetryAt = &next
	if job.NextRetryAt == nil {
		t.Error("expected non-nil NextRetryAt after assignment")
	}
}
