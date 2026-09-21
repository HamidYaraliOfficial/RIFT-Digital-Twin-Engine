package simulation

import (
	"rift/internal/graph"
	"rift/internal/model"
)

// BehaviorModel is the Physics/Behavior Abstraction Layer: RIFT's simulation
// core has zero built-in assumptions about thermodynamics, fluid flow, or
// traffic — it only knows how to call Step() once per tick and read back
// whatever numeric Properties a specific model chooses to mutate. Domain
// plugins (temperature transfer, energy consumption, queueing, traffic
// movement, ...) implement this interface; RIFT ships PropagationModel as the
// default, generically-applicable behavior so every twin works out of the
// box even before a specialized plugin is installed.
type BehaviorModel interface {
	Name() string
	Step(tick int, b *Branch)
}

// PropagationModel is the default, domain-agnostic behavior: it propagates
// failure/degradation along the dependency graph (a failed power feed
// degrades what it feeds; a failed producer reduces a consumer's throughput)
// and only ever touches numeric Properties the entity itself already
// declares, so it never invents data for a twin.
type PropagationModel struct{}

func (PropagationModel) Name() string { return "propagation" }

func (PropagationModel) Step(tick int, b *Branch) {
	g := graph.Build(b.RelationSlice())
	for id, e := range b.Entities {
		if e.Status != model.StatusFailed && e.Status != model.StatusDegraded {
			continue
		}
		severity := 0.6
		if e.Status == model.StatusDegraded {
			severity = 0.3
		}
		for _, depID := range g.Dependents(id) {
			dep, ok := b.Entities[depID]
			if !ok {
				continue
			}
			b.degrade(dep, severity)
			// Generic numeric propagation: only touch properties the entity
			// itself already defines, scaled by how directly it depends on
			// the failed upstream entity.
			for _, key := range []string{"throughput", "capacity", "output_rate"} {
				if v, ok := numeric(dep.Properties[key]); ok {
					dep.Properties[key] = v * (1 - severity*0.5)
				}
			}
			for _, key := range []string{"energy_consumption", "load"} {
				if v, ok := numeric(dep.Properties[key]); ok {
					dep.Properties[key] = v * (1 + severity*0.3)
				}
			}
		}
	}
}

func numeric(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	default:
		return 0, false
	}
}
