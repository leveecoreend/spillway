package dlq

import (
	"testing"
	"time"
)

func TestDefaultRetryPolicy(t *testing.T) {
	p := DefaultRetryPolicy()
	if err := p.Validate(); err != nil {
		t.Fatalf("default policy should be valid, got: %v", err)
	}
}

func TestRetryPolicy_Validate(t *testing.T) {
	tests := []struct {
		name    string
		policy  RetryPolicy
		wantErr bool
	}{
		{"valid fixed", RetryPolicy{MaxAttempts: 3, InitialInterval: time.Second, MaxInterval: time.Minute, Strategy: BackoffFixed}, false},
		{"zero max attempts", RetryPolicy{MaxAttempts: 0, InitialInterval: time.Second, MaxInterval: time.Minute}, true},
		{"zero initial interval", RetryPolicy{MaxAttempts: 3, InitialInterval: 0, MaxInterval: time.Minute}, true},
		{"max < initial", RetryPolicy{MaxAttempts: 3, InitialInterval: time.Minute, MaxInterval: time.Second}, true},
		{"exponential multiplier <= 1", RetryPolicy{MaxAttempts: 3, InitialInterval: time.Second, MaxInterval: time.Minute, Strategy: BackoffExponential, Multiplier: 1.0}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.policy.Validate()
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestRetryPolicy_NextInterval_Fixed(t *testing.T) {
	p := RetryPolicy{InitialInterval: 10 * time.Second, MaxInterval: time.Minute, Strategy: BackoffFixed}
	for _, attempt := range []int{1, 2, 5} {
		if got := p.NextInterval(attempt); got != 10*time.Second {
			t.Errorf("attempt %d: expected 10s, got %v", attempt, got)
		}
	}
}

func TestRetryPolicy_NextInterval_Exponential(t *testing.T) {
	p := RetryPolicy{InitialInterval: 5 * time.Second, MaxInterval: time.Hour, Strategy: BackoffExponential, Multiplier: 2.0}
	expected := []time.Duration{5 * time.Second, 10 * time.Second, 20 * time.Second}
	for i, want := range expected {
		if got := p.NextInterval(i + 1); got != want {
			t.Errorf("attempt %d: expected %v, got %v", i+1, want, got)
		}
	}
}

func TestRetryPolicy_NextInterval_CapsAtMax(t *testing.T) {
	p := RetryPolicy{InitialInterval: time.Second, MaxInterval: 5 * time.Second, Strategy: BackoffExponential, Multiplier: 10.0}
	if got := p.NextInterval(5); got != 5*time.Second {
		t.Errorf("expected max interval 5s, got %v", got)
	}
}

func TestRetryPolicy_NextInterval_Linear(t *testing.T) {
	p := RetryPolicy{InitialInterval: 3 * time.Second, MaxInterval: time.Hour, Strategy: BackoffLinear}
	if got := p.NextInterval(3); got != 9*time.Second {
		t.Errorf("expected 9s, got %v", got)
	}
}
