package api

import (
	"net/http"

	"rift/internal/model"
	"rift/internal/registry"
)

func (s *Server) handleListRules(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.Rules.ListRules(r.PathValue("id")))
}

func (s *Server) handleCreateRule(w http.ResponseWriter, r *http.Request) {
	twinID := r.PathValue("id")
	var rule model.Rule
	if err := decodeJSON(r, &rule); err != nil || rule.Name == "" || rule.Trigger == "" {
		writeError(w, http.StatusBadRequest, "name and trigger are required")
		return
	}
	rule.ID = registry.NewID("rule")
	rule.TwinID = twinID
	rule.Version = 1
	if rule.CooldownMs == 0 {
		rule.CooldownMs = 30000
	}
	s.Rules.AddRule(&rule)
	s.Audit.Record(userIDFrom(r), "rule_created", "rule:"+rule.ID, twinID, map[string]interface{}{"name": rule.Name})
	writeJSON(w, http.StatusCreated, rule)
}

func (s *Server) handleDeleteRule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.Rules.RemoveRule(id)
	s.Audit.Record(userIDFrom(r), "rule_deleted", "rule:"+id, "", nil)
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

// ---------- Alerts ----------

func (s *Server) handleListAlerts(w http.ResponseWriter, r *http.Request) {
	includeResolved := r.URL.Query().Get("includeResolved") == "true"
	writeJSON(w, http.StatusOK, s.Alerts.List(r.PathValue("id"), includeResolved))
}

func (s *Server) handleAckAlert(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.Alerts.Acknowledge(id, userIDFrom(r)); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"acknowledged": true})
}

func (s *Server) handleResolveAlert(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.Alerts.Resolve(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"resolved": true})
}
