// Package health implements the Health Score Engine: combines reliability,
// sensor quality, maintenance state, resource usage and recent anomalies
// into Health / Risk / Criticality scores for one entity or a whole twin.
package health

import (
	"time"

	"rift/internal/model"
)

// Inputs is the evidence the score is computed from; every field is
// evidence-backed rather than invented so a caller can show its work.
type Inputs struct {
	Status              model.Status
	RecentAnomalyCount  int // in the observation window
	RecentCriticalCount int
	SensorOfflineRatio  float64 // 0..1 of this entity's sensors currently offline
	Age                 time.Duration
	MaintenanceOverdue  bool
	DependencyCount     int // number of entities that depend_on / powered_by / feeds this one
}

// Compute derives a HealthScore. The formula is intentionally simple and
// documented so operators can trust and tune it, rather than a black box.
func Compute(in Inputs) model.HealthScore {
	health := 100.0

	switch in.Status {
	case model.StatusWarning:
		health -= 15
	case model.StatusDegraded:
		health -= 35
	case model.StatusMaintenance:
		health -= 20
	case model.StatusFailed:
		health -= 70
	}

	health -= float64(in.RecentAnomalyCount) * 3
	health -= float64(in.RecentCriticalCount) * 10
	health -= in.SensorOfflineRatio * 20
	if in.MaintenanceOverdue {
		health -= 15
	}
	if health < 0 {
		health = 0
	}
	if health > 100 {
		health = 100
	}

	risk := 100 - health
	// Criticality grows with how many other entities depend on this one:
	// a failing leaf sensor matters less than a failing power feed.
	criticality := clamp(float64(in.DependencyCount)*8, 0, 100)

	return model.HealthScore{
		Health:      round1(health),
		Risk:        round1(risk),
		Criticality: round1(criticality),
		UpdatedAt:   time.Now().UTC(),
	}
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func round1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}
