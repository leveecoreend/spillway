package dlq

import (
	"testing"
	"time"
)

func makeJob(id, queue string, status JobStatus, retries int, createdAt time.Time) *Job {
	j := &Job{
		ID:        id,
		Queue:     queue,
		Status:    status,
		Retries:   retries,
		CreatedAt: createdAt,
	}
	return j
}

func TestFilter_ByQueue(t *testing.T) {
	now := time.Now()
	jobs := []*Job{
		makeJob("1", "email", StatusFailed, 1, now),
		makeJob("2", "sms", StatusFailed, 1, now),
		makeJob("3", "EMAIL", StatusFailed, 1, now),
	}
	got := Filter(jobs, FilterOptions{Queue: "email"})
	if len(got) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(got))
	}
}

func TestFilter_ByStatus(t *testing.T) {
	now := time.Now()
	jobs := []*Job{
		makeJob("1", "email", StatusFailed, 0, now),
		makeJob("2", "email", StatusTerminal, 0, now),
		makeJob("3", "email", StatusPending, 0, now),
	}
	got := Filter(jobs, FilterOptions{Status: StatusFailed})
	if len(got) != 1 || got[0].ID != "1" {
		t.Fatalf("expected job 1, got %+v", got)
	}
}

func TestFilter_ByMinRetries(t *testing.T) {
	now := time.Now()
	jobs := []*Job{
		makeJob("1", "q", StatusFailed, 0, now),
		makeJob("2", "q", StatusFailed, 3, now),
		makeJob("3", "q", StatusFailed, 5, now),
	}
	got := Filter(jobs, FilterOptions{MinRetries: 3})
	if len(got) != 2 {
		t.Fatalf("expected 2 jobs with retries >= 3, got %d", len(got))
	}
}

func TestFilter_ByCreatedBefore(t *testing.T) {
	base := time.Now()
	jobs := []*Job{
		makeJob("1", "q", StatusFailed, 0, base.Add(-2*time.Hour)),
		makeJob("2", "q", StatusFailed, 0, base.Add(-30*time.Minute)),
		makeJob("3", "q", StatusFailed, 0, base.Add(1*time.Hour)),
	}
	cutoff := base.Add(-1 * time.Hour)
	got := Filter(jobs, FilterOptions{CreatedBefore: cutoff})
	if len(got) != 1 || got[0].ID != "1" {
		t.Fatalf("expected only job 1, got %+v", got)
	}
}

func TestFilter_NoOptions_ReturnsAll(t *testing.T) {
	now := time.Now()
	jobs := []*Job{
		makeJob("1", "a", StatusFailed, 0, now),
		makeJob("2", "b", StatusPending, 1, now),
	}
	got := Filter(jobs, FilterOptions{})
	if len(got) != 2 {
		t.Fatalf("expected all 2 jobs, got %d", len(got))
	}
}

func TestFilter_EmptyInput(t *testing.T) {
	got := Filter(nil, FilterOptions{Queue: "email"})
	if len(got) != 0 {
		t.Fatalf("expected empty result, got %d", len(got))
	}
}
