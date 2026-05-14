package dlq

import (
	"sync"
	"testing"
)

func TestNewMetrics(t *testing.T) {
	m := NewMetrics()
	if m == nil {
		t.Fatal("expected non-nil Metrics")
	}
	snap := m.Snapshot()
	if snap.Enqueued != 0 || snap.Retried != 0 || snap.Dead != 0 || snap.Deleted != 0 {
		t.Errorf("expected all counters to be 0, got %+v", snap)
	}
}

func TestMetrics_RecordOperations(t *testing.T) {
	m := NewMetrics()

	m.RecordEnqueue()
	m.RecordEnqueue()
	m.RecordRetry()
	m.RecordDead()
	m.RecordDead()
	m.RecordDead()
	m.RecordDelete()

	snap := m.Snapshot()

	if snap.Enqueued != 2 {
		t.Errorf("expected Enqueued=2, got %d", snap.Enqueued)
	}
	if snap.Retried != 1 {
		t.Errorf("expected Retried=1, got %d", snap.Retried)
	}
	if snap.Dead != 3 {
		t.Errorf("expected Dead=3, got %d", snap.Dead)
	}
	if snap.Deleted != 1 {
		t.Errorf("expected Deleted=1, got %d", snap.Deleted)
	}
}

func TestMetrics_SnapshotIsImmutable(t *testing.T) {
	m := NewMetrics()
	m.RecordEnqueue()

	snap1 := m.Snapshot()
	m.RecordEnqueue()
	snap2 := m.Snapshot()

	if snap1.Enqueued != 1 {
		t.Errorf("expected snap1.Enqueued=1, got %d", snap1.Enqueued)
	}
	if snap2.Enqueued != 2 {
		t.Errorf("expected snap2.Enqueued=2, got %d", snap2.Enqueued)
	}
}

func TestMetrics_ConcurrentAccess(t *testing.T) {
	m := NewMetrics()
	const goroutines = 50
	const opsEach = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < opsEach; j++ {
				m.RecordEnqueue()
				m.RecordRetry()
			}
		}()
	}
	wg.Wait()

	snap := m.Snapshot()
	expected := int64(goroutines * opsEach)
	if snap.Enqueued != expected {
		t.Errorf("expected Enqueued=%d, got %d", expected, snap.Enqueued)
	}
	if snap.Retried != expected {
		t.Errorf("expected Retried=%d, got %d", expected, snap.Retried)
	}
}
