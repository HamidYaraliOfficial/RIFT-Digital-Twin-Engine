package model

import "time"

// ConditionOperator enumerates comparison operators available to Rule conditions.
type ConditionOperator string

const (
	OpGreaterThan ConditionOperator = "gt"
	OpLessThan    ConditionOperator = "lt"
	OpGreaterEq   ConditionOperator = "gte"
	OpLessEq      ConditionOperator = "lte"
	OpEquals      ConditionOperator = "eq"
	OpNotEquals   ConditionOperator = "neq"
)

// Condition tests a single field (a sensor type, a state key, or "event.type")
// against a value using Operator. Conditions in a Rule are AND-combined.
type Condition struct {
	Field    string            `json:"field"`
	Operator ConditionOperator `json:"operator"`
	Value    float64           `json:"value"`
	// StringValue is used when Field refers to a textual state/event field
	// (e.g. "event.type" eq "power_failure"). Operator is limited to eq/neq.
	StringValue string `json:"stringValue,omitempty"`
}

// ActionType enumerates what a Rule (or a Scenario step) may trigger.
type ActionType string

const (
	ActionEmitEvent       ActionType = "emit_event"
	ActionSetState        ActionType = "set_state"
	ActionSetStatus       ActionType = "set_status"
	ActionCreateAlert     ActionType = "create_alert"
	ActionActuatorCommand ActionType = "actuator_command"
	ActionRunScenario     ActionType = "run_scenario"
)

// Action is one effect a fired Rule produces.
type Action struct {
	Type   ActionType             `json:"type"`
	Params map[string]interface{} `json:"params"`
}

// Rule is a user-authored automation: "if <trigger> and <conditions> then <actions>".
type Rule struct {
	ID             string      `json:"id"`
	TwinID         string      `json:"twinId"`
	Name           string      `json:"name"`
	EntityID       string      `json:"entityId"` // entity (or sensor owner) this rule watches
	Trigger        string      `json:"trigger"`  // sensor type, "event:<type>", or "*"
	Conditions     []Condition `json:"conditions"`
	Actions        []Action    `json:"actions"`
	Priority       int         `json:"priority"`
	CooldownMs     int64       `json:"cooldownMs"`
	Enabled        bool        `json:"enabled"`
	Version        int         `json:"version"`
	LastFired      time.Time   `json:"lastFired"`
	RequirePermit  bool        `json:"requirePermit"` // safety guard for sensitive actions
	CreatedAt      time.Time   `json:"createdAt"`
	UpdatedAt      time.Time   `json:"updatedAt"`
}
