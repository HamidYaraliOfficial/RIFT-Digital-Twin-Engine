package model

import "time"

// AlertStatus tracks acknowledgement/resolution lifecycle.
type AlertStatus string

const (
	AlertOpen         AlertStatus = "open"
	AlertAcknowledged AlertStatus = "acknowledged"
	AlertResolved     AlertStatus = "resolved"
)

// AlertType classifies the alert source.
type AlertType string

const (
	AlertThreshold      AlertType = "threshold"
	AlertAnomaly        AlertType = "anomaly"
	AlertPredictive     AlertType = "predictive"
	AlertRuleViolation  AlertType = "rule_violation"
	AlertSensorOffline  AlertType = "sensor_offline"
	AlertEquipmentFail  AlertType = "equipment_failure"
	AlertSLOBreach      AlertType = "slo_breach"
	AlertScenarioResult AlertType = "scenario_result"
)

// Alert is a deduplicated, escalatable notification produced by the Alerting Engine.
type Alert struct {
	ID              string      `json:"id"`
	TwinID          string      `json:"twinId"`
	RuleID          string      `json:"ruleId,omitempty"`
	Type            AlertType   `json:"type"`
	Severity        Severity    `json:"severity"`
	EntityID        string      `json:"entityId"`
	Message         string      `json:"message"`
	DedupKey        string      `json:"dedupKey"`
	CreatedAt       time.Time   `json:"createdAt"`
	LastOccurredAt  time.Time   `json:"lastOccurredAt"`
	OccurrenceCount int         `json:"occurrenceCount"`
	Status          AlertStatus `json:"status"`
	EscalationLevel int         `json:"escalationLevel"`
	AckBy           string      `json:"ackBy,omitempty"`
	AckAt           *time.Time  `json:"ackAt,omitempty"`
	ResolvedAt      *time.Time  `json:"resolvedAt,omitempty"`
}
