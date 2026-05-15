package dlq

import (
	"testing"
	"time"
)

func TestAlert_Fields(t *testing.T) {
	now := time.Now()
	a := Alert{
		Level:      AlertLevelInfo,
		Message:    "test message",
		JobID:      "job-123",
		Queue:      "emails",
		OccurredAt: now,
	}
	if a.Level != AlertLevelInfo {
		t.Errorf("unexpected level: %s", a.Level)
	}
	if a.JobID != "job-123" {
		t.Errorf("unexpected job id: %s", a.JobID)
	}
	if a.Queue != "emails" {
		t.Errorf("unexpected queue: %s", a.Queue)
	}
	if !a.OccurredAt.Equal(now) {
		t.Errorf("unexpected time: %v", a.OccurredAt)
	}
}

func TestAlertRule_Defaults(t *testing.T) {
	rule := AlertRule{}
	if rule.MaxFailures != 0 {
		t.Errorf("expected zero MaxFailures, got %d", rule.MaxFailures)
	}
	if rule.MaxRetryExhausted != 0 {
		t.Errorf("expected zero MaxRetryExhausted, got %d", rule.MaxRetryExhausted)
	}
}

func TestAlertLevel_Constants(t *testing.T) {
	levels := []AlertLevel{AlertLevelInfo, AlertLevelWarning, AlertLevelCritical}
	expected := []string{"info", "warning", "critical"}
	for i, l := range levels {
		if string(l) != expected[i] {
			t.Errorf("level[%d]: got %q, want %q", i, l, expected[i])
		}
	}
}

func TestAlerter_OccurredAtIsSet(t *testing.T) {
	before := time.Now()
	var fired []Alert
	a := NewAlerter(
		AlertRule{MaxFailures: 1},
		func(alert Alert) { fired = append(fired, alert) },
	)
	a.Evaluate(MetricsSnapshot{TotalEnqueued: 5})
	after := time.Now()

	if len(fired) == 0 {
		t.Fatal("expected alert to fire")
	}
	if fired[0].OccurredAt.Before(before) || fired[0].OccurredAt.After(after) {
		t.Errorf("OccurredAt %v not in expected range [%v, %v]", fired[0].OccurredAt, before, after)
	}
}
