package dlq

import (
	"context"
	"testing"
)

func newTestJob(id, queue, backend string) *Job {
	return NewJob(id, queue, backend, []byte(`{}`), "timeout", 3)
}

func TestInMemoryStore_SaveAndGet(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	job := newTestJob("j1", "orders", "rabbitmq")
	if err := store.Save(ctx, job); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	fetched, err := store.Get(ctx, "j1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.ID != job.ID {
		t.Errorf("expected ID %s, got %s", job.ID, fetched.ID)
	}
}

func TestInMemoryStore_GetNotFound(t *testing.T) {
	store := NewInMemoryStore()
	_, err := store.Get(context.Background(), "nonexistent")
	if err != ErrJobNotFound {
		t.Errorf("expected ErrJobNotFound, got %v", err)
	}
}

func TestInMemoryStore_List(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	_ = store.Save(ctx, newTestJob("j1", "orders", "redis"))
	_ = store.Save(ctx, newTestJob("j2", "emails", "redis"))
	_ = store.Save(ctx, newTestJob("j3", "orders", "rabbitmq"))

	jobs, err := store.List(ctx, ListFilter{Queue: "orders"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(jobs))
	}

	jobs, err = store.List(ctx, ListFilter{Backend: "redis"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestInMemoryStore_Delete(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	_ = store.Save(ctx, newTestJob("j1", "q", "b"))
	if err := store.Delete(ctx, "j1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, err := store.Get(ctx, "j1")
	if err != ErrJobNotFound {
		t.Error("expected job to be deleted")
	}

	if err := store.Delete(ctx, "j1"); err != ErrJobNotFound {
		t.Errorf("expected ErrJobNotFound on double delete, got %v", err)
	}
}

func TestInMemoryStore_CountByStatus(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	j1 := newTestJob("j1", "q", "b")
	j2 := newTestJob("j2", "q", "b")
	j3 := newTestJob("j3", "q", "b")
	j2.Status = StatusRetrying
	j3.Status = StatusDiscarded

	_ = store.Save(ctx, j1)
	_ = store.Save(ctx, j2)
	_ = store.Save(ctx, j3)

	counts, err := store.CountByStatus(ctx)
	if err != nil {
		t.Fatalf("CountByStatus failed: %v", err)
	}
	if counts[StatusPending] != 1 {
		t.Errorf("expected 1 pending, got %d", counts[StatusPending])
	}
	if counts[StatusRetrying] != 1 {
		t.Errorf("expected 1 retrying, got %d", counts[StatusRetrying])
	}
	if counts[StatusDiscarded] != 1 {
		t.Errorf("expected 1 discarded, got %d", counts[StatusDiscarded])
	}
}
