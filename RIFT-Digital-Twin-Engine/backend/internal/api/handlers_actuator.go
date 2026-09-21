package api

import (
	"net/http"
	"time"

	"rift/internal/actuator"
	"rift/internal/model"
	"rift/internal/registry"
)

type commandRequest struct {
	Type         string  `json:"type"`
	Value        float64 `json:"value"`
	Reason       string  `json:"reason"`
	ConfirmToken string  `json:"confirmToken"`
}

// handleActuatorCommand is the single, safety-gated entry point every
// caller — the UI, a Rule, or the AI Assistant — must use to affect a real
// or simulated actuator. It always passes through actuator.SafetyEngine
// first: schema validation, permission check, range check, interlocks, rate
// limiting and (for RequireConfirm policies) an explicit confirmation step.
func (s *Server) handleActuatorCommand(w http.ResponseWriter, r *http.Request) {
	entityID := r.PathValue("id")
	entity, ok := s.Reg.GetEntity(entityID)
	if !ok {
		writeError(w, http.StatusNotFound, "entity not found")
		return
	}
	var req commandRequest
	if err := decodeJSON(r, &req); err != nil || req.Type == "" {
		writeError(w, http.StatusBadRequest, "command type is required")
		return
	}
	cmd := actuator.Command{
		ID: registry.NewID("cmd"), EntityID: entityID, Type: actuator.CommandType(req.Type),
		Value: req.Value, IssuedBy: userIDFrom(r), Reason: req.Reason, Simulated: true, RequestedAt: time.Now().UTC(),
	}
	result := s.Actuators.Submit(cmd, req.ConfirmToken)
	_ = entity
	status := http.StatusOK
	if !result.Accepted {
		status = http.StatusForbidden
	}
	writeJSON(w, status, result)
}

type policyRequest struct {
	Command           string   `json:"command"`
	Min               float64  `json:"min"`
	Max               float64  `json:"max"`
	RequiredScope     string   `json:"requiredScope"`
	MinIntervalMs     int64    `json:"minIntervalMs"`
	RequireConfirm    bool     `json:"requireConfirm"`
}

// handleSetActuatorPolicy lets an engineer configure the Command Safety
// Engine's allow-list, range, and confirmation requirements per entity —
// this is what makes "a rule/AI can never directly execute a dangerous
// command" an enforced property rather than a convention.
func (s *Server) handleSetActuatorPolicy(w http.ResponseWriter, r *http.Request) {
	entityID := r.PathValue("id")
	var req policyRequest
	if err := decodeJSON(r, &req); err != nil || req.Command == "" {
		writeError(w, http.StatusBadRequest, "command is required")
		return
	}
	interval := time.Duration(req.MinIntervalMs) * time.Millisecond
	if interval <= 0 {
		interval = time.Second
	}
	var vr *actuator.Range
	if req.Max != 0 || req.Min != 0 {
		vr = &actuator.Range{Min: req.Min, Max: req.Max}
	}
	scope := req.RequiredScope
	if scope == "" {
		scope = "actuator_control"
	}
	s.Actuators.SetPolicy(entityID, actuator.CommandType(req.Command), actuator.Policy{
		AllowedCommands:   map[actuator.CommandType]bool{actuator.CommandType(req.Command): true},
		ValueRange:        vr,
		RequiredScope:     scope,
		MinIntervalPerCmd: interval,
		RequireConfirm:    req.RequireConfirm,
	})
	s.Audit.Record(userIDFrom(r), "actuator_policy_set", "entity:"+entityID, "", map[string]interface{}{"command": req.Command})
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// handleGrantPermission implements the Twin Permission Model endpoint: an
// admin grants a User or Role a set of Scopes on a Twin (optionally scoped
// to one Entity) — e.g. Maintenance gets actuator_control on specific pumps
// only, while another team gets read-only across the whole factory.
func (s *Server) handleGrantPermission(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TwinID   string   `json:"twinId"`
		EntityID string   `json:"entityId"`
		UserID   string   `json:"userId"`
		Role     string   `json:"role"`
		Scopes   []string `json:"scopes"`
	}
	if err := decodeJSON(r, &req); err != nil || req.TwinID == "" || len(req.Scopes) == 0 {
		writeError(w, http.StatusBadRequest, "twinId and scopes are required")
		return
	}
	scopes := make([]model.ScopeAction, 0, len(req.Scopes))
	for _, sc := range req.Scopes {
		scopes = append(scopes, model.ScopeAction(sc))
	}
	var role model.Role
	if req.Role != "" {
		role = roleOrDefault(req.Role)
	}
	perm := s.AuthStore.Grant(model.TwinPermission{
		TwinID: req.TwinID, EntityID: req.EntityID, UserID: req.UserID, Role: role, Scopes: scopes,
	})
	s.Audit.Record(userIDFrom(r), "permission_granted", "twin:"+req.TwinID, req.TwinID, map[string]interface{}{"scopes": req.Scopes})
	writeJSON(w, http.StatusCreated, perm)
}
