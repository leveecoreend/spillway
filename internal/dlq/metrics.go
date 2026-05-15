package dlq

import "sync"

// MetricsSnapshot is an immutable point-in-time view of DLQ metrics.
type MetricsSnapshot struct {
	TotalEnqueued int
	TotalRetried  int
	TotalTerminal int
	TotalDeleted  int
}

// Metrics tracks operational counters for the DLQ system.
type Metrics struct {
	mu            sync.Mutex
	totalEnqueued int
	totalRetried  int
	totalTerminal int
	totalDeleted  int
}

// NewMetrics creates a zeroed Metrics instance.
func NewMetrics() *Metrics {
	return &Metrics{}
}

// RecordEnqueued increments the enqueued counter.
func (m *Metrics) RecordEnqueued() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalEnqueued++
}

// RecordRetried increments the retried counter.
func (m *Metrics) RecordRetried() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalRetried++
}

// RecordTerminal increments the terminal (retry-exhausted) counter.
func (m *Metrics) RecordTerminal() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalTerminal++
}

// RecordDeleted increments the deleted counter.
func (m *Metrics) RecordDeleted() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalDeleted++
}

// Snapshot returns an immutable copy of the current metrics.
func (m *Metrics) Snapshot() MetricsSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	return MetricsSnapshot{
		TotalEnqueued: m.totalEnqueued,
		TotalRetried:  m.totalRetried,
		TotalTerminal: m.totalTerminal,
		TotalDeleted:  m.totalDeleted,
	}
}
