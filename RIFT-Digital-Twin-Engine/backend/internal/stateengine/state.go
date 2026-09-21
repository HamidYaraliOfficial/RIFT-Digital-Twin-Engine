// Package stateengine tracks Current/Previous/Expected/Simulated state per
// Entity and enforces valid Status transitions (the State Machine Engine).
package stateengine

import (
	"fmt"
	"sync"
	"time"

	"rift/internal/model"
)

// Snapshot4 separates the four state views RIFT keeps for every entity.
type Snapshot4 struct {
	Current   map[string]interface{} `json:"current"`
	Previous  map[string]interface{} `json:"previous"`
	Expected  map[string]interface{} `json:"expected"`
	Simulated map[string]interface{} `json:"simulated"`
	UpdatedAt time.Time               `json:"updatedAt"`
}

type Engine struct {
	mu     sync.RWMutex
	states map[string]*Snapshot4 // entityID -> snapshot
}

func New() *Engine {
	return &Engine{states: make(map[string]*Snapshot4)}
}

func (e *Engine) get(entityID string) *Snapshot4 {
	s, ok := e.states[entityID]
	if !ok {
		s = &Snapshot4{
			Current:   map[string]interface{}{},
			Previous:  map[string]interface{}{},
			Expected:  map[string]interface{}{},
			Simulated: map[string]interface{}{},
		}
		e.states[entityID] = s
	}
	return s
}

// ApplyCurrent merges fields into the Current state, moving the prior values
// into Previous first so callers can always diff "what changed".
func (e *Engine) ApplyCurrent(entityID string, fields map[string]interface{}) *Snapshot4 {
	e.mu.Lock()
	defer e.mu.Unlock()
	s := e.get(entityID)
	for k, v := range s.Current {
		s.Previous[k] = v
	}
	for k, v := range fields {
		s.Current[k] = v
	}
	s.UpdatedAt = time.Now().UTC()
	return s
}

func (e *Engine) SetExpected(entityID string, fields map[string]interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()
	s := e.get(entityID)
	for k, v := range fields {
		s.Expected[k] = v
	}
}

func (e *Engine) SetSimulated(entityID string, fields map[string]interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()
	s := e.get(entityID)
	for k, v := range fields {
		s.Simulated[k] = v
	}
}

func (e *Engine) Get(entityID string) Snapshot4 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	s := e.get(entityID)
	return *s
}

// ---------- State Machine (Status transitions) ----------

// allowedTransitions encodes which Status -> Status moves are valid. Any
// status may move to Failed or Maintenance (real systems can always break or
// be pulled offline), but recovery must go through a defined path.
var allowedTransitions = map[model.Status][]model.Status{
	model.StatusOperational: {model.StatusWarning, model.StatusDegraded, model.StatusMaintenance, model.StatusFailed},
	model.StatusWarning:     {model.StatusOperational, model.StatusDegraded, model.StatusMaintenance, model.StatusFailed},
	model.StatusDegraded:    {model.StatusWarning, model.StatusOperational, model.StatusMaintenance, model.StatusFailed},
	model.StatusMaintenance: {model.StatusOperational, model.StatusRecovered, model.StatusFailed},
	model.StatusFailed:      {model.StatusMaintenance, model.StatusRecovered},
	model.StatusRecovered:   {model.StatusOperational, model.StatusWarning},
	model.StatusUnknown:     {model.StatusOperational, model.StatusWarning, model.StatusDegraded, model.StatusMaintenance, model.StatusFailed},
}

// ValidateTransition returns an error if `from -> to` is not an allowed move.
func ValidateTransition(from, to model.Status) error {
	if from == to {
		return nil
	}
	for _, allowed := range allowedTransitions[from] {
		if allowed == to {
			return nil
		}
	}
	return fmt.Errorf("invalid status transition: %s -> %s", from, to)
}
