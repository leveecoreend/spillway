package dlq

import (
	"time"
)

// JobStatus represents the current state of a dead-letter job.
type JobStatus string

const (
	StatusPending   JobStatus = "pending"
	StatusRetrying  JobStatus = "retrying"
	StatusSucceeded JobStatus = "succeeded"
	StatusFailed    JobStatus = "failed"
	StatusDiscarded JobStatus = "discarded"
)

// Job represents a failed job stored in the dead-letter queue.
type Job struct {
	ID          string            `json:"id"`
	Queue       string            `json:"queue"`
	Backend     string            `json:"backend"`
	Payload     []byte            `json:"payload"`
	Error       string            `json:"error"`
	Status      JobStatus         `json:"status"`
	Attempts    int               `json:"attempts"`
	MaxAttempts int               `json:"max_attempts"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	NextRetryAt *time.Time        `json:"next_retry_at,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// NewJob creates a new Job with default values.
func NewJob(id, queue, backend string, payload []byte, errMsg string, maxAttempts int) *Job {
	now := time.Now().UTC()
	return &Job{
		ID:          id,
		Queue:       queue,
		Backend:     backend,
		Payload:     payload,
		Error:       errMsg,
		Status:      StatusPending,
		Attempts:    0,
		MaxAttempts: maxAttempts,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    make(map[string]string),
	}
}

// CanRetry returns true if the job has remaining retry attempts.
func (j *Job) CanRetry() bool {
	return j.Attempts < j.MaxAttempts
}

// IsTerminal returns true if the job is in a final state.
func (j *Job) IsTerminal() bool {
	return j.Status == StatusSucceeded ||
		j.Status == StatusFailed ||
		j.Status == StatusDiscarded
}
