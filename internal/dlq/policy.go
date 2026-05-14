package dlq

import (
	"errors"
	"time"
)

// BackoffStrategy defines how retry delays are calculated.
type BackoffStrategy string

const (
	BackoffFixed       BackoffStrategy = "fixed"
	BackoffExponential BackoffStrategy = "exponential"
	BackoffLinear      BackoffStrategy = "linear"
)

// RetryPolicy defines the retry behaviour for failed jobs.
type RetryPolicy struct {
	MaxAttempts     int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	Strategy        BackoffStrategy
}

// DefaultRetryPolicy returns a sensible default retry policy.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:     5,
		InitialInterval: 5 * time.Second,
		MaxInterval:     5 * time.Minute,
		Multiplier:      2.0,
		Strategy:        BackoffExponential,
	}
}

// Validate checks that the policy has valid values.
func (p RetryPolicy) Validate() error {
	if p.MaxAttempts < 1 {
		return errors.New("max_attempts must be at least 1")
	}
	if p.InitialInterval <= 0 {
		return errors.New("initial_interval must be positive")
	}
	if p.MaxInterval < p.InitialInterval {
		return errors.New("max_interval must be >= initial_interval")
	}
	if p.Strategy == BackoffExponential && p.Multiplier <= 1.0 {
		return errors.New("multiplier must be > 1.0 for exponential backoff")
	}
	return nil
}

// NextInterval calculates the delay before the next retry attempt.
func (p RetryPolicy) NextInterval(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	var interval time.Duration
	switch p.Strategy {
	case BackoffExponential:
		interval = p.InitialInterval
		for i := 1; i < attempt; i++ {
			interval = time.Duration(float64(interval) * p.Multiplier)
		}
	case BackoffLinear:
		interval = p.InitialInterval * time.Duration(attempt)
	default: // fixed
		interval = p.InitialInterval
	}
	if interval > p.MaxInterval {
		return p.MaxInterval
	}
	return interval
}
