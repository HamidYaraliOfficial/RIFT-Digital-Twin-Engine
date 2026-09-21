// Package actuator implements the Actuator Layer and the Command Safety
// Engine: every command — whether issued by a human, a Rule, or the AI
// Assistant — passes schema validation, permission check, range check,
// interlock, rate limiting and (for sensitive actions) an explicit
// confirmation policy before anything is allowed to execute. Real hardware
// adapters (OPC-UA, Modbus/TCP, MQTT, gRPC) implement the Adapter interface;
// RIFT ships a SimulatedAdapter so commands always have somewhere safe to go.
package actuator

import (
	"fmt"
	"sync"
	"time"
)

type CommandType string

const (
	CmdStart          CommandType = "start"
	CmdStop           CommandType = "stop"
	CmdOpen           CommandType = "open"
	CmdClose          CommandType = "close"
	CmdSetTemperature CommandType = "set_temperature"
	CmdChangeSpeed    CommandType = "change_speed"
	CmdRestart        CommandType = "restart"
	CmdLock           CommandType = "lock"
	CmdUnlock         CommandType = "unlock"
	CmdCustom         CommandType = "custom"
)

// Command is one actuator instruction targeting an entity.
type Command struct {
	ID         string      `json:"id"`
	EntityID   string      `json:"entityId"`
	Type       CommandType `json:"type"`
	Value      float64     `json:"value,omitempty"`
	IssuedBy   string      `json:"issuedBy"`   // user id, "rule:<id>", or "ai_assistant"
	Reason     string      `json:"reason,omitempty"`
	Simulated  bool        `json:"simulated"`
	RequestedAt time.Time  `json:"requestedAt"`
}

// Result records what actually happened to a Command after the safety gate.
type Result struct {
	Command  Command   `json:"command"`
	Accepted bool      `json:"accepted"`
	Reason   string    `json:"reason"`
	AppliedAt time.Time `json:"appliedAt"`
}

// Range constrains CmdSetTemperature/CmdChangeSpeed-style numeric commands.
type Range struct{ Min, Max float64 }

// Policy is the Command Safety Engine configuration for one entity/command type.
type Policy struct {
	AllowedCommands   map[CommandType]bool
	ValueRange        *Range          // nil = no numeric range check
	RequiredScope     string          // permission scope the issuer must hold, e.g. "actuator_control"
	Interlocks        []func() error  // must all return nil, e.g. "door must be closed before Lock"
	MinIntervalPerCmd time.Duration   // rate limit
	RequireConfirm    bool            // sensitive commands need an explicit second confirmation token
}

// PermissionChecker is injected so the actuator package never depends on the
// auth package directly (keeps the module dependency graph acyclic).
type PermissionChecker func(issuedBy, entityID, scope string) bool

type SafetyEngine struct {
	mu           sync.Mutex
	policies     map[string]Policy // key: entityID+"|"+commandType, or "*|"+commandType for defaults
	lastCommand  map[string]time.Time
	pendingConfirm map[string]Command
	checkPermit  PermissionChecker
	executor     Adapter
}

func NewSafetyEngine(checkPermit PermissionChecker, executor Adapter) *SafetyEngine {
	return &SafetyEngine{
		policies:       make(map[string]Policy),
		lastCommand:    make(map[string]time.Time),
		pendingConfirm: make(map[string]Command),
		checkPermit:    checkPermit,
		executor:       executor,
	}
}

func (s *SafetyEngine) SetPolicy(entityID string, cmd CommandType, p Policy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policies[entityID+"|"+string(cmd)] = p
}

func (s *SafetyEngine) policyFor(entityID string, cmd CommandType) Policy {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p, ok := s.policies[entityID+"|"+string(cmd)]; ok {
		return p
	}
	// Conservative default: command must be explicitly allow-listed per entity;
	// nothing sensitive is permitted "by accident".
	return Policy{AllowedCommands: map[CommandType]bool{}, RequiredScope: "actuator_control", MinIntervalPerCmd: time.Second}
}

// Submit runs the full Command Safety Engine gate and, only if every check
// passes, forwards the command to the configured Adapter.
func (s *SafetyEngine) Submit(cmd Command, confirmToken string) Result {
	cmd.RequestedAt = time.Now().UTC()
	pol := s.policyFor(cmd.EntityID, cmd.Type)

	if !pol.AllowedCommands[cmd.Type] {
		return Result{Command: cmd, Accepted: false, Reason: "command type is not allow-listed for this entity"}
	}
	if s.checkPermit != nil && !s.checkPermit(cmd.IssuedBy, cmd.EntityID, pol.RequiredScope) {
		return Result{Command: cmd, Accepted: false, Reason: "issuer lacks required permission scope"}
	}
	if pol.ValueRange != nil && (cmd.Value < pol.ValueRange.Min || cmd.Value > pol.ValueRange.Max) {
		return Result{Command: cmd, Accepted: false, Reason: fmt.Sprintf("value %.2f outside allowed range [%.2f, %.2f]", cmd.Value, pol.ValueRange.Min, pol.ValueRange.Max)}
	}
	for _, guard := range pol.Interlocks {
		if err := guard(); err != nil {
			return Result{Command: cmd, Accepted: false, Reason: "interlock failed: " + err.Error()}
		}
	}
	s.mu.Lock()
	last, ok := s.lastCommand[cmd.EntityID+"|"+string(cmd.Type)]
	s.mu.Unlock()
	if ok && time.Since(last) < pol.MinIntervalPerCmd {
		return Result{Command: cmd, Accepted: false, Reason: "rate limited: command issued too recently"}
	}
	if pol.RequireConfirm && confirmToken == "" {
		s.mu.Lock()
		s.pendingConfirm[cmd.ID] = cmd
		s.mu.Unlock()
		return Result{Command: cmd, Accepted: false, Reason: "confirmation required: resubmit with the confirmation token for command id " + cmd.ID}
	}
	if pol.RequireConfirm {
		s.mu.Lock()
		pending, ok := s.pendingConfirm[cmd.ID]
		s.mu.Unlock()
		if !ok || confirmToken != pending.ID {
			return Result{Command: cmd, Accepted: false, Reason: "invalid or expired confirmation token"}
		}
	}

	s.mu.Lock()
	s.lastCommand[cmd.EntityID+"|"+string(cmd.Type)] = time.Now().UTC()
	delete(s.pendingConfirm, cmd.ID)
	s.mu.Unlock()

	if s.executor != nil {
		if err := s.executor.Execute(cmd); err != nil {
			return Result{Command: cmd, Accepted: false, Reason: "adapter error: " + err.Error()}
		}
	}
	return Result{Command: cmd, Accepted: true, Reason: "executed", AppliedAt: time.Now().UTC()}
}
