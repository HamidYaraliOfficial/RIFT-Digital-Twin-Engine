package model

import "time"

// Severity is shared by Events and Alerts.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// LifecycleState tracks an Event/Incident through its life.
type LifecycleState string

const (
	LifecycleOpen         LifecycleState = "open"
	LifecycleAcknowledged LifecycleState = "acknowledged"
	LifecycleResolved     LifecycleState = "resolved"
)

// Event is emitted by the Event Engine for every meaningful change in the
// twin: machine started/stopped, temperature spikes, power failures, sensor
// offline, alarms, arrivals, completions, maintenance requirements, and any
// number of custom event types.
type Event struct {
	ID             string                 `json:"id"`
	TwinID         string                 `json:"twinId"`
	Type           string                 `json:"type"`
	Timestamp      time.Time              `json:"timestamp"`
	SourceEntityID string                 `json:"sourceEntityId,omitempty"`
	TargetEntityID string                 `json:"targetEntityId,omitempty"`
	Payload        map[string]interface{} `json:"payload,omitempty"`
	Severity       Severity               `json:"severity"`
	CorrelationID  string                 `json:"correlationId,omitempty"`
	Cause          string                 `json:"cause,omitempty"` // id of the event/rule/reading that caused this
	Consequences   []string               `json:"consequences,omitempty"`
	Lifecycle      LifecycleState         `json:"lifecycle"`
}

// Incident groups correlated Events under a single investigable timeline.
type Incident struct {
	ID                   string         `json:"id"`
	TwinID               string         `json:"twinId"`
	Title                string         `json:"title"`
	Severity             Severity       `json:"severity"`
	AffectedEntityIDs    []string       `json:"affectedEntityIds"`
	EventIDs             []string       `json:"eventIds"`
	RootCauseCandidates  []RootCause    `json:"rootCauseCandidates,omitempty"`
	Lifecycle            LifecycleState `json:"lifecycle"`
	CreatedAt            time.Time      `json:"createdAt"`
	UpdatedAt            time.Time      `json:"updatedAt"`
	ResolvedAt           *time.Time     `json:"resolvedAt,omitempty"`
}

// RootCause is one candidate explanation produced by the Root Cause Analysis Engine.
type RootCause struct {
	EntityID   string  `json:"entityId"`
	Reason     string  `json:"reason"`
	Confidence float64 `json:"confidence"`
	Evidence   []string `json:"evidence"`
}
