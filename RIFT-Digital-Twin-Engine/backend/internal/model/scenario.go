package model

import "time"

// FaultType enumerates the controlled failures the Fault Injection Engine can apply.
type FaultType string

const (
	FaultSensorFailure    FaultType = "sensor_failure"
	FaultNetworkFailure   FaultType = "network_failure"
	FaultPowerFailure     FaultType = "power_failure"
	FaultEquipmentFailure FaultType = "equipment_failure"
	FaultCommDelay        FaultType = "communication_delay"
	FaultPacketLoss       FaultType = "packet_loss"
	FaultResourceExhaust  FaultType = "resource_exhaustion"
)

// FaultSpec is one fault to inject at a given simulated tick.
type FaultSpec struct {
	Type     FaultType `json:"type"`
	EntityID string    `json:"entityId"`
	AtTick   int       `json:"atTick"`
	Magnitude float64  `json:"magnitude"` // meaning depends on Type, e.g. % capacity lost
	DurationTicks int  `json:"durationTicks"`
}

// ScenarioStatus tracks a What-If run.
type ScenarioStatus string

const (
	ScenarioPending ScenarioStatus = "pending"
	ScenarioRunning ScenarioStatus = "running"
	ScenarioDone    ScenarioStatus = "done"
	ScenarioFailed  ScenarioStatus = "failed"
)

// EntityDelta captures how a single entity differed between baseline and scenario branch.
type EntityDelta struct {
	EntityID   string  `json:"entityId"`
	Field      string  `json:"field"`
	Baseline   float64 `json:"baseline"`
	Simulated  float64 `json:"simulated"`
	DeltaPct   float64 `json:"deltaPct"`
}

// ScenarioResult is the comparison produced once a What-If run finishes.
type ScenarioResult struct {
	DurationTicks     int           `json:"durationTicks"`
	AffectedEntities  []string      `json:"affectedEntities"`
	Deltas            []EntityDelta `json:"deltas"`
	EventsGenerated   int           `json:"eventsGenerated"`
	Summary           string        `json:"summary"`
}

// Scenario is a What-If experiment definition and its outcome.
type Scenario struct {
	ID             string         `json:"id"`
	TwinID         string         `json:"twinId"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	Seed           int64          `json:"seed"`
	Deterministic  bool           `json:"deterministic"`
	Faults         []FaultSpec    `json:"faults"`
	DurationTicks  int            `json:"durationTicks"`
	Status         ScenarioStatus `json:"status"`
	Result         *ScenarioResult `json:"result,omitempty"`
	CreatedAt      time.Time      `json:"createdAt"`
	FinishedAt     *time.Time     `json:"finishedAt,omitempty"`
}

// Snapshot is a full or incremental capture of a Twin usable for restore,
// clone, fork, compare and branch-simulation.
type Snapshot struct {
	ID        string                    `json:"id"`
	TwinID    string                    `json:"twinId"`
	Label     string                    `json:"label"`
	CreatedAt time.Time                 `json:"createdAt"`
	Entities  map[string]*Entity        `json:"entities"`
	Relations map[string]*Relationship  `json:"relations"`
	Sensors   map[string]*Sensor        `json:"sensors"`
}

// MonteCarloResult aggregates many randomized Scenario runs.
type MonteCarloResult struct {
	Runs        int                `json:"runs"`
	Metric      string             `json:"metric"`
	Mean        float64            `json:"mean"`
	P10         float64            `json:"p10"`
	P50         float64            `json:"p50"`
	P90         float64            `json:"p90"`
	Best        float64            `json:"best"`
	Worst       float64            `json:"worst"`
	Samples     []float64          `json:"samples"`
}

// SweepResult is one point of a Parameter Sweep.
type SweepResult struct {
	Parameter string  `json:"parameter"`
	Value     float64 `json:"value"`
	Metric    float64 `json:"metric"`
}
