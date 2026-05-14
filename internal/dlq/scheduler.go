package dlq

import (
	"context"
	"log"
	"sync"
	"time"
)

// Scheduler periodically scans the store for jobs ready to be retried
// and re-enqueues them via the Manager.
type Scheduler struct {
	manager  *Manager
	store    Store
	interval time.Duration

	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
}

// NewScheduler creates a Scheduler that polls at the given interval.
func NewScheduler(manager *Manager, store Store, interval time.Duration) *Scheduler {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &Scheduler{
		manager:  manager,
		store:    store,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start begins the background polling loop. It is safe to call once.
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return nil
	}
	s.running = true
	s.stopCh = make(chan struct{})
	go s.loop(ctx)
	return nil
}

// Stop signals the polling loop to exit.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return
	}
	s.running = false
	close(s.stopCh)
}

func (s *Scheduler) loop(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.tick(ctx)
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	jobs, err := s.store.List(ctx)
	if err != nil {
		log.Printf("scheduler: list error: %v", err)
		return
	}
	now := time.Now()
	for _, job := range jobs {
		if job.CanRetry() && !job.NextRetryAt.IsZero() && !job.NextRetryAt.After(now) {
			if err := s.manager.Retry(ctx, job.ID); err != nil {
				log.Printf("scheduler: retry job %s error: %v", job.ID, err)
			}
		}
	}
}
