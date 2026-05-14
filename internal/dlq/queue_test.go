package dlq

import (
	"context"
	"testing"
)

// mockBackend is a test implementation of Backend.
type mockBackend struct {
	name    string
	enqueued []*Job
}

func (m *mockBackend) Name() string { return m.name }
func (m *mockBackend) Enqueue(_ context.Context, job *Job) error {
	m.enqueued = append(m.enqueued, job)
	return nil
}

func TestNewRouter(t *testing.T) {
	r := NewRouter()
	if r == nil {
		t.Fatal("expected non-nil router")
	}
	if len(r.Names()) != 0 {
		t.Errorf("expected empty router, got %d backends", len(r.Names()))
	}
}

func TestRouter_Register(t *testing.T) {
	r := NewRouter()
	b := &mockBackend{name: "redis"}

	if err := r.Register(b); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Names()) != 1 {
		t.Errorf("expected 1 backend, got %d", len(r.Names()))
	}
}

func TestRouter_Register_Duplicate(t *testing.T) {
	r := NewRouter()
	b := &mockBackend{name: "redis"}

	_ = r.Register(b)
	if err := r.Register(b); err == nil {
		t.Fatal("expected error for duplicate backend, got nil")
	}
}

func TestRouter_Register_Nil(t *testing.T) {
	r := NewRouter()
	if err := r.Register(nil); err == nil {
		t.Fatal("expected error for nil backend, got nil")
	}
}

func TestRouter_Register_EmptyName(t *testing.T) {
	r := NewRouter()
	b := &mockBackend{name: ""}
	if err := r.Register(b); err == nil {
		t.Fatal("expected error for empty backend name, got nil")
	}
}

func TestRouter_Resolve(t *testing.T) {
	r := NewRouter()
	b := &mockBackend{name: "sqs"}
	_ = r.Register(b)

	got, err := r.Resolve("sqs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != b {
		t.Errorf("expected backend %v, got %v", b, got)
	}
}

func TestRouter_Resolve_NotFound(t *testing.T) {
	r := NewRouter()
	if _, err := r.Resolve("unknown"); err == nil {
		t.Fatal("expected error for unknown backend, got nil")
	}
}
