package dlq

import "sync/atomic"

// Metrics tracks operational statistics for the DLQ manager.
type Metrics struct {
	enqueued  atomic.Int64
	retried   atomic.Int64
	dead      atomic.Int64
	deleted   atomic.Int64
}

// NewMetrics creates a new Metrics instance.
func NewMetrics() *Metrics {
	return &Metrics{}
}

// RecordEnqueue increments the enqueued job counter.
func (m *Metrics) RecordEnqueue() {
	m.enqueued.Add(1)
}

// RecordRetry increments the retried job counter.
func (m *Metrics) RecordRetry() {
	m.retried.Add(1)
}

// RecordDead increments the dead (terminal) job counter.
func (m *Metrics) RecordDead() {
	m.dead.Add(1)
}

// RecordDelete increments the deleted job counter.
func (m *Metrics) RecordDelete() {
	m.deleted.Add(1)
}

// Snapshot returns a point-in-time copy of the current metrics.
func (m *Metrics) Snapshot() MetricsSnapshot {
	return MetricsSnapshot{
		Enqueued: m.enqueued.Load(),
		Retried:  m.retried.Load(),
		Dead:     m.dead.Load(),
		Deleted:  m.deleted.Load(),
	}
}

// MetricsSnapshot is an immutable copy of metrics at a point in time.
type MetricsSnapshot struct {
	Enqueued int64
	Retried  int64
	Dead     int64
	Deleted  int64
}
