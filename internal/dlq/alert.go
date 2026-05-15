package dlq

import (
	"fmt"
	"time"
)

// AlertLevel represents the severity of an alert.
type AlertLevel string

const (
	AlertLevelInfo    AlertLevel = "info"
	AlertLevelWarning AlertLevel = "warning"
	AlertLevelCritical AlertLevel = "critical"
)

// Alert represents a notification event triggered by DLQ activity.
type Alert struct {
	Level     AlertLevel
	Message   string
	JobID     string
	Queue     string
	OccurredAt time.Time
}

// AlertHandler is a function that receives alerts.
type AlertHandler func(alert Alert)

// AlertRule defines conditions under which an alert is triggered.
type AlertRule struct {
	// MaxFailures triggers an alert when a queue exceeds this many failed jobs.
	MaxFailures int
	// MaxRetryExhausted triggers a critical alert when terminal jobs exceed this count.
	MaxRetryExhausted int
}

// Alerter evaluates metrics and dispatches alerts via registered handlers.
type Alerter struct {
	rule     AlertRule
	handlers []AlertHandler
}

// NewAlerter creates an Alerter with the given rule and handlers.
func NewAlerter(rule AlertRule, handlers ...AlertHandler) *Alerter {
	return &Alerter{
		rule:     rule,
		handlers: handlers,
	}
}

// Evaluate checks the given metrics snapshot and fires alerts as needed.
func (a *Alerter) Evaluate(snap MetricsSnapshot) {
	if a.rule.MaxFailures > 0 && snap.TotalEnqueued > a.rule.MaxFailures {
		a.dispatch(Alert{
			Level:      AlertLevelWarning,
			Message:    fmt.Sprintf("enqueued jobs (%d) exceeded threshold (%d)", snap.TotalEnqueued, a.rule.MaxFailures),
			OccurredAt: time.Now(),
		})
	}
	if a.rule.MaxRetryExhausted > 0 && snap.TotalTerminal > a.rule.MaxRetryExhausted {
		a.dispatch(Alert{
			Level:      AlertLevelCritical,
			Message:    fmt.Sprintf("terminal jobs (%d) exceeded threshold (%d)", snap.TotalTerminal, a.rule.MaxRetryExhausted),
			OccurredAt: time.Now(),
		})
	}
}

func (a *Alerter) dispatch(alert Alert) {
	for _, h := range a.handlers {
		h(alert)
	}
}
