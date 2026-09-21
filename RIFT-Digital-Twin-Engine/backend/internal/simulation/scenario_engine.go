package simulation

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"rift/internal/model"
)

// Branch is an isolated, in-memory copy of a twin's graph that a What-If run
// mutates freely. It is never written back to the Registry: Simulations
// never change real state, exactly as required.
type Branch struct {
	Entities  map[string]*model.Entity
	Relations map[string]*model.Relationship
	Sensors   map[string]*model.Sensor
	events    []model.Event
	rng       *rand.Rand
}

func NewBranch(snap *model.Snapshot, seed int64) *Branch {
	b := &Branch{
		Entities:  make(map[string]*model.Entity, len(snap.Entities)),
		Relations: make(map[string]*model.Relationship, len(snap.Relations)),
		Sensors:   make(map[string]*model.Sensor, len(snap.Sensors)),
	}
	for id, e := range snap.Entities {
		b.Entities[id] = e.Clone()
	}
	for id, r := range snap.Relations {
		cp := *r
		b.Relations[id] = &cp
	}
	for id, s := range snap.Sensors {
		cp := *s
		b.Sensors[id] = &cp
	}
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	b.rng = rand.New(rand.NewSource(seed))
	return b
}

func (b *Branch) RelationSlice() []*model.Relationship {
	out := make([]*model.Relationship, 0, len(b.Relations))
	for _, r := range b.Relations {
		out = append(out, r)
	}
	return out
}

func (b *Branch) degrade(e *model.Entity, severity float64) {
	prev := e.Status
	if e.Status == model.StatusOperational || e.Status == model.StatusWarning {
		if severity >= 0.5 {
			e.Status = model.StatusDegraded
		} else {
			e.Status = model.StatusWarning
		}
	}
	if prev != e.Status {
		b.events = append(b.events, model.Event{
			ID: fmt.Sprintf("sim-evt-%d", len(b.events)), Type: "status_changed",
			SourceEntityID: e.ID, Severity: model.SeverityWarning, Timestamp: time.Now().UTC(),
			Payload: map[string]interface{}{"from": prev, "to": e.Status},
		})
	}
}

// applyFault mutates the branch to reflect one FaultSpec at the tick it fires.
func (b *Branch) applyFault(f model.FaultSpec) {
	e, ok := b.Entities[f.EntityID]
	if !ok {
		return
	}
	prevStatus := e.Status
	switch f.Type {
	case model.FaultSensorFailure:
		for _, s := range b.Sensors {
			if s.EntityID == f.EntityID {
				s.Status = model.SensorOffline
			}
		}
	case model.FaultEquipmentFailure, model.FaultPowerFailure:
		e.Status = model.StatusFailed
	case model.FaultNetworkFailure, model.FaultCommDelay, model.FaultPacketLoss:
		for _, s := range b.Sensors {
			if s.EntityID == f.EntityID {
				s.Status = model.SensorDegraded
			}
		}
		b.degrade(e, 0.4)
	case model.FaultResourceExhaust:
		if v, ok := numeric(e.Properties["capacity"]); ok {
			e.Properties["capacity"] = v * (1 - clamp01(f.Magnitude))
		}
		b.degrade(e, 0.3)
	}
	if prevStatus != e.Status {
		b.events = append(b.events, model.Event{
			ID: fmt.Sprintf("sim-evt-fault-%s-%d", e.ID, len(b.events)), Type: string(f.Type),
			SourceEntityID: e.ID, Severity: model.SeverityCritical, Timestamp: time.Now().UTC(),
			Payload: map[string]interface{}{"from": prevStatus, "to": e.Status},
		})
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// Run executes DurationTicks of simulation on the branch, applying faults at
// their scheduled tick and stepping every registered BehaviorModel each
// tick. It returns the branch (post-run state) and the events it generated.
func Run(branch *Branch, faults []model.FaultSpec, ticks int, models []BehaviorModel) []model.Event {
	pending := make(map[int][]model.FaultSpec)
	for _, f := range faults {
		pending[f.AtTick] = append(pending[f.AtTick], f)
	}
	for t := 0; t < ticks; t++ {
		for _, f := range pending[t] {
			branch.applyFault(f)
		}
		for _, m := range models {
			m.Step(t, branch)
		}
	}
	return branch.events
}

// Compare builds a ScenarioResult from a baseline snapshot and a post-run branch.
func Compare(baseline *model.Snapshot, branch *Branch, ticks int, events []model.Event) *model.ScenarioResult {
	var deltas []model.EntityDelta
	affected := map[string]bool{}
	for id, be := range branch.Entities {
		orig, ok := baseline.Entities[id]
		if !ok {
			continue
		}
		if orig.Status != be.Status {
			affected[id] = true
		}
		for _, key := range []string{"throughput", "capacity", "output_rate", "energy_consumption", "load"} {
			ov, okO := numeric(orig.Properties[key])
			nv, okN := numeric(be.Properties[key])
			if okO && okN && ov != 0 {
				pct := (nv - ov) / ov * 100
				if pct != 0 {
					affected[id] = true
					deltas = append(deltas, model.EntityDelta{EntityID: id, Field: key, Baseline: round2(ov), Simulated: round2(nv), DeltaPct: round2(pct)})
				}
			}
		}
	}
	affectedList := make([]string, 0, len(affected))
	for id := range affected {
		affectedList = append(affectedList, id)
	}
	sort.Strings(affectedList)

	summary := fmt.Sprintf("Simulated %d ticks: %d entities affected, %d events generated.", ticks, len(affectedList), len(events))
	return &model.ScenarioResult{
		DurationTicks:    ticks,
		AffectedEntities: affectedList,
		Deltas:           deltas,
		EventsGenerated:  len(events),
		Summary:          summary,
	}
}

func round2(v float64) float64 { return float64(int(v*100+0.5)) / 100 }

// ---------- Scenario store / orchestration ----------

// Engine owns the set of Scenarios for a process and runs them. It is what
// the API layer talks to; ScenarioEngine never touches the live Registry.
type Engine struct {
	mu        sync.RWMutex
	scenarios map[string]*model.Scenario
	models    []BehaviorModel
}

func NewEngine(models ...BehaviorModel) *Engine {
	if len(models) == 0 {
		models = []BehaviorModel{PropagationModel{}}
	}
	return &Engine{scenarios: make(map[string]*model.Scenario), models: models}
}

func (e *Engine) Create(s *model.Scenario) *model.Scenario {
	e.mu.Lock()
	defer e.mu.Unlock()
	s.Status = model.ScenarioPending
	s.CreatedAt = time.Now().UTC()
	e.scenarios[s.ID] = s
	return s
}

func (e *Engine) Get(id string) (*model.Scenario, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	s, ok := e.scenarios[id]
	return s, ok
}

func (e *Engine) List(twinID string) []*model.Scenario {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]*model.Scenario, 0)
	for _, s := range e.scenarios {
		if s.TwinID == twinID {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

// Execute runs a scenario synchronously against the given baseline snapshot
// and stores the result on the Scenario record.
func (e *Engine) Execute(s *model.Scenario, baseline *model.Snapshot) *model.ScenarioResult {
	e.mu.Lock()
	s.Status = model.ScenarioRunning
	e.mu.Unlock()

	branch := NewBranch(baseline, s.Seed)
	events := Run(branch, s.Faults, s.DurationTicks, e.models)
	result := Compare(baseline, branch, s.DurationTicks, events)

	e.mu.Lock()
	s.Status = model.ScenarioDone
	s.Result = result
	now := time.Now().UTC()
	s.FinishedAt = &now
	e.mu.Unlock()
	return result
}
