package dlq

import (
	"strings"
	"time"
)

// FilterOptions defines criteria for filtering jobs from the store.
type FilterOptions struct {
	Queue      string
	Status     JobStatus
	MinRetries int
	CreatedBefore time.Time
	CreatedAfter  time.Time
}

// Filter returns jobs from the store matching all non-zero filter criteria.
func Filter(jobs []*Job, opts FilterOptions) []*Job {
	var result []*Job
	for _, j := range jobs {
		if !matchesFilter(j, opts) {
			continue
		}
		result = append(result, j)
	}
	return result
}

func matchesFilter(j *Job, opts FilterOptions) bool {
	if opts.Queue != "" && !strings.EqualFold(j.Queue, opts.Queue) {
		return false
	}
	if opts.Status != "" && j.Status != opts.Status {
		return false
	}
	if opts.MinRetries > 0 && j.Retries < opts.MinRetries {
		return false
	}
	if !opts.CreatedBefore.IsZero() && j.CreatedAt.After(opts.CreatedBefore) {
		return false
	}
	if !opts.CreatedAfter.IsZero() && j.CreatedAt.Before(opts.CreatedAfter) {
		return false
	}
	return true
}
