package dlq

import (
	"testing"
)

func TestNewAlerter(t *testing.T) {
	a := NewAlerter(AlertRule{MaxFailures: 5})
	if a == nil {
		t.Fatal("expected non-nil Alerter")
	}
}

func TestAlerter_NoAlertBelowThreshold(t *testing.T) {
	var fired []Alert
	a := NewAlerter(
		AlertRule{MaxFailures: 10, MaxRetryExhausted: 5},
		func(alert Alert) { fired = append(fired, alert) },
	)
	a.Evaluate(MetricsSnapshot{TotalEnqueued: 3, TotalTerminal: 2})
	if len(fired) != 0 {
		t.Fatalf("expected no alerts, got %d", len(fired))
	}
}

func TestAlerter_WarningOnMaxFailures(t *testing.T) {
	var fired []Alert
	a := NewAlerter(
		AlertRule{MaxFailures: 5},
		func(alert Alert) { fired = append(fired, alert) },
	)
	a.Evaluate(MetricsSnapshot{TotalEnqueued: 6})
	if len(fired) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(fired))
	}
	if fired[0].Level != AlertLevelWarning {
		t.Errorf("expected warning, got %s", fired[0].Level)
	}
}

func TestAlerter_CriticalOnMaxTerminal(t *testing.T) {
	var fired []Alert
	a := NewAlerter(
		AlertRule{MaxRetryExhausted: 3},
		func(alert Alert) { fired = append(fired, alert) },
	)
	a.Evaluate(MetricsSnapshot{TotalTerminal: 4})
	if len(fired) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(fired))
	}
	if fired[0].Level != AlertLevelCritical {
		t.Errorf("expected critical, got %s", fired[0].Level)
	}
}

func TestAlerter_BothThresholdsExceeded(t *testing.T) {
	var fired []Alert
	a := NewAlerter(
		AlertRule{MaxFailures: 2, MaxRetryExhausted: 1},
		func(alert Alert) { fired = append(fired, alert) },
	)
	a.Evaluate(MetricsSnapshot{TotalEnqueued: 5, TotalTerminal: 3})
	if len(fired) != 2 {
		t.Fatalf("expected 2 alerts, got %d", len(fired))
	}
}

func TestAlerter_MultipleHandlers(t *testing.T) {
	count := 0
	inc := func(_ Alert) { count++ }
	a := NewAlerter(
		AlertRule{MaxFailures: 1},
		inc, inc, inc,
	)
	a.Evaluate(MetricsSnapshot{TotalEnqueued: 2})
	if count != 3 {
		t.Errorf("expected 3 handler calls, got %d", count)
	}
}

func TestAlerter_ZeroRuleDoesNotFire(t *testing.T) {
	var fired []Alert
	a := NewAlerter(
		AlertRule{},
		func(alert Alert) { fired = append(fired, alert) },
	)
	a.Evaluate(MetricsSnapshot{TotalEnqueued: 1000, TotalTerminal: 1000})
	if len(fired) != 0 {
		t.Fatalf("expected no alerts with zero rule, got %d", len(fired))
	}
}
