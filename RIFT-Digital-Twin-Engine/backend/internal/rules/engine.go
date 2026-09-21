// Package rules implements the Rules Engine: user-authored
// trigger/condition/action automations evaluated against live telemetry and
// events, with priority ordering, cooldown, and a safety gate for sensitive
// actions (see RequirePermit on model.Rule and internal/actuator for the
// downstream Command Safety Engine).
package rules

import (
	"sync"
	"time"

	"rift/internal/model"
)

// Fact is the generic "current value" the engine conditions are evaluated
// against: either a telemetry reading (Field = sensor type string) or an
// event (Field = "event.type", StringVal = event type).
type Fact struct {
	EntityID  string
	Field     string
	Value     float64
	StringVal string
}

// ActionExecutor performs the side effect for a fired Action. Kept as an
// injected interface so the Rules package has zero dependency on the API,
// alerting, or actuator packages (Modular / Event-Driven architecture).
type ActionExecutor interface {
	Execute(rule *model.Rule, action model.Action, fact Fact) error
}

type Engine struct {
	mu       sync.RWMutex
	rules    map[string][]*model.Rule // trigger field -> rules
	byID     map[string]*model.Rule
	executor ActionExecutor
}

func New(executor ActionExecutor) *Engine {
	return &Engine{
		rules:    make(map[string][]*model.Rule),
		byID:     make(map[string]*model.Rule),
		executor: executor,
	}
}

func (e *Engine) AddRule(r *model.Rule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.byID[r.ID] = r
	e.rules[r.Trigger] = append(e.rules[r.Trigger], r)
	sortByPriority(e.rules[r.Trigger])
}

func (e *Engine) RemoveRule(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	r, ok := e.byID[id]
	if !ok {
		return
	}
	delete(e.byID, id)
	list := e.rules[r.Trigger]
	out := list[:0]
	for _, rr := range list {
		if rr.ID != id {
			out = append(out, rr)
		}
	}
	e.rules[r.Trigger] = out
}

func (e *Engine) ListRules(twinID string) []*model.Rule {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]*model.Rule, 0)
	for _, r := range e.byID {
		if r.TwinID == twinID {
			out = append(out, r)
		}
	}
	return out
}

func sortByPriority(rules []*model.Rule) {
	for i := 1; i < len(rules); i++ {
		for j := i; j > 0 && rules[j].Priority > rules[j-1].Priority; j-- {
			rules[j], rules[j-1] = rules[j-1], rules[j]
		}
	}
}

// Evaluate runs every rule registered for fact.Field (plus wildcard "*"
// rules) against the given Fact and fires matching, non-cooling-down rules
// in priority order.
func (e *Engine) Evaluate(fact Fact) {
	e.mu.RLock()
	candidates := append([]*model.Rule{}, e.rules[fact.Field]...)
	candidates = append(candidates, e.rules["*"]...)
	e.mu.RUnlock()

	for _, r := range candidates {
		if !r.Enabled || r.EntityID != "" && r.EntityID != fact.EntityID {
			continue
		}
		if r.CooldownMs > 0 && time.Since(r.LastFired) < time.Duration(r.CooldownMs)*time.Millisecond {
			continue
		}
		if !matchConditions(r.Conditions, fact) {
			continue
		}
		e.mu.Lock()
		r.LastFired = time.Now().UTC()
		e.mu.Unlock()
		for _, a := range r.Actions {
			if e.executor != nil {
				_ = e.executor.Execute(r, a, fact)
			}
		}
	}
}

func matchConditions(conds []model.Condition, fact Fact) bool {
	for _, c := range conds {
		if c.StringValue != "" {
			switch c.Operator {
			case model.OpEquals:
				if fact.StringVal != c.StringValue {
					return false
				}
			case model.OpNotEquals:
				if fact.StringVal == c.StringValue {
					return false
				}
			}
			continue
		}
		switch c.Operator {
		case model.OpGreaterThan:
			if !(fact.Value > c.Value) {
				return false
			}
		case model.OpLessThan:
			if !(fact.Value < c.Value) {
				return false
			}
		case model.OpGreaterEq:
			if !(fact.Value >= c.Value) {
				return false
			}
		case model.OpLessEq:
			if !(fact.Value <= c.Value) {
				return false
			}
		case model.OpEquals:
			if fact.Value != c.Value {
				return false
			}
		case model.OpNotEquals:
			if fact.Value == c.Value {
				return false
			}
		}
	}
	return true
}
