package dlq

import (
	"testing"
	"time"
)

// TestAlerter_Integration verifies that the Alerter fires correctly when
// wired to real Metrics and a Manager that enqueues jobs.
func TestAlerter_Integration(t *testing.T) {
	policy := DefaultRetryPolicy()
	policy.MaxRetries = 1
	policy.InitialInterval = 10 * time.Millisecond

	manager, err := NewManager(policy)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	var alerts []Alert
	alerter := NewAlerter(
		AlertRule{MaxFailures: 2, MaxRetryExhausted: 0},
		func(a Alert) { alerts = append(alerts, a) },
	)

	for i := 0; i < 3; i++ {
		j := NewJob("queue-a", []byte(`{}`))
		if enqErr := manager.Enqueue(j); enqErr != nil {
			t.Fatalf("Enqueue: %v", enqErr)
		}
	}

	snap := manager.Metrics().Snapshot()
	alerter.Evaluate(snap)

	if len(alerts) == 0 {
		t.Fatal("expected at least one alert after exceeding MaxFailures threshold")
	}
	if alerts[0].Level != AlertLevelWarning {
		t.Errorf("expected warning level, got %s", alerts[0].Level)
	}
}
