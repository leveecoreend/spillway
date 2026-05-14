package dlq

import (
	"context"
	"errors"
)

// ErrJobNotFound is returned when a job cannot be located in the store.
var ErrJobNotFound = errors.New("job not found")

// ListFilter defines optional filters when listing jobs.
type ListFilter struct {
	Queue   string
	Backend string
	Status  JobStatus
	Limit   int
	Offset  int
}

// Store defines the interface for persisting and retrieving dead-letter jobs.
type Store interface {
	// Save persists a new job or updates an existing one.
	Save(ctx context.Context, job *Job) error

	// Get retrieves a job by its ID.
	Get(ctx context.Context, id string) (*Job, error)

	// List returns jobs matching the provided filter.
	List(ctx context.Context, filter ListFilter) ([]*Job, error)

	// Delete removes a job from the store.
	Delete(ctx context.Context, id string) error

	// CountByStatus returns the number of jobs grouped by status.
	CountByStatus(ctx context.Context) (map[JobStatus]int64, error)
}

// InMemoryStore is a simple in-memory implementation of Store for testing.
type InMemoryStore struct {
	jobs map[string]*Job
}

// NewInMemoryStore creates a new InMemoryStore.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{jobs: make(map[string]*Job)}
}

func (s *InMemoryStore) Save(_ context.Context, job *Job) error {
	s.jobs[job.ID] = job
	return nil
}

func (s *InMemoryStore) Get(_ context.Context, id string) (*Job, error) {
	j, ok := s.jobs[id]
	if !ok {
		return nil, ErrJobNotFound
	}
	return j, nil
}

func (s *InMemoryStore) List(_ context.Context, filter ListFilter) ([]*Job, error) {
	var result []*Job
	for _, j := range s.jobs {
		if filter.Queue != "" && j.Queue != filter.Queue {
			continue
		}
		if filter.Backend != "" && j.Backend != filter.Backend {
			continue
		}
		if filter.Status != "" && j.Status != filter.Status {
			continue
		}
		result = append(result, j)
	}
	return result, nil
}

func (s *InMemoryStore) Delete(_ context.Context, id string) error {
	if _, ok := s.jobs[id]; !ok {
		return ErrJobNotFound
	}
	delete(s.jobs, id)
	return nil
}

func (s *InMemoryStore) CountByStatus(_ context.Context) (map[JobStatus]int64, error) {
	counts := make(map[JobStatus]int64)
	for _, j := range s.jobs {
		counts[j.Status]++
	}
	return counts, nil
}
