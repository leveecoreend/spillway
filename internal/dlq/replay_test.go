package dlq

import (
	"context"
	"errors"
	"testing"
	"time"
)

func newReplayStore(t *testing.T) Store {
	t.Helper()
	s, err := NewInMemoryStore()
	if err != nil {
		t.Fatalf("NewInMemoryStore: %v", err)
	}
	return s
}

func TestNewReplayer_NilStore(t *testing.T) {
	r, err := NewReplayer(nil, &Router{})
	if err == nil {
		t.Fatal("expected error for nil store")
	}
	if r != nil {
		t.Fatal("expected nil replayer")
	}
}

func TestNewReplayer_NilRouter(t *testing.T) {
	s := newReplayStore(t)
	_, err := NewReplayer(s, nil)
	if err == nil {
		t.Fatal("expected error for nil router")
	}
}

func TestReplayer_Replay_SkipsTerminalByDefault(t *testing.T) {
	ctx := context.Background()
	s := newReplayStore(t)

	j := NewJob("q1", "task", []byte(`{}`))
	j.Status = StatusTerminal
	_ = s.Save(ctx, j)

	router, _ := NewRouter()
	replayer, _ := NewReplayer(s, router)

	res, err := replayer.Replay(ctx, Filter{}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", res.Skipped)
	}
	if res.Replayed != 0 {
		t.Errorf("expected 0 replayed, got %d", res.Replayed)
	}
}

func TestReplayer_Replay_Success(t *testing.T) {
	ctx := context.Background()
	s := newReplayStore(t)

	dispatched := make(map[string]bool)
	router, _ := NewRouter()
	_ = router.Register("q1", QueueHandlerFunc(func(_ context.Context, j *Job) error {
		dispatched[j.ID] = true
		return nil
	}))

	j := NewJob("q1", "task", []byte(`{}`))
	j.Status = StatusFailed
	j.Retries = 3
	j.LastError = "boom"
	j.UpdatedAt = time.Now().UTC()
	_ = s.Save(ctx, j)

	replayer, _ := NewReplayer(s, router)
	res, err := replayer.Replay(ctx, Filter{}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Replayed != 1 {
		t.Errorf("expected 1 replayed, got %d", res.Replayed)
	}
	if !dispatched[j.ID] {
		t.Error("job was not dispatched")
	}
	// Job should be removed from DLQ.
	_, getErr := s.Get(ctx, j.ID)
	if !errors.Is(getErr, ErrNotFound) {
		t.Errorf("expected job to be deleted from store, got: %v", getErr)
	}
}

func TestReplayer_Replay_ForceTerminal(t *testing.T) {
	ctx := context.Background()
	s := newReplayStore(t)

	router, _ := NewRouter()
	_ = router.Register("q1", QueueHandlerFunc(func(_ context.Context, _ *Job) error { return nil }))

	j := NewJob("q1", "task", []byte(`{}`))
	j.Status = StatusTerminal
	_ = s.Save(ctx, j)

	replayer, _ := NewReplayer(s, router)
	res, err := replayer.Replay(ctx, Filter{}, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Replayed != 1 {
		t.Errorf("expected 1 replayed, got %d", res.Replayed)
	}
}
