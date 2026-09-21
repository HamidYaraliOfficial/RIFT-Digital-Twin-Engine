package api

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"rift/internal/exportimport"
	"rift/internal/model"
	"rift/internal/operatinghours"
	"rift/internal/registry"
	"rift/internal/simulation"
)

// ---------- Twins ----------

func (s *Server) handleListTwins(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.Reg.ListTwins())
}

type createTwinRequest struct {
	Name     string `json:"name"`
	Template string `json:"template"`
}

func (s *Server) handleCreateTwin(w http.ResponseWriter, r *http.Request) {
	var req createTwinRequest
	if err := decodeJSON(r, &req); err != nil || req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	twin := s.Reg.CreateTwin(&registry.Twin{OrgID: "org_default", WorkspaceID: "ws_default", Name: req.Name, Template: req.Template})
	s.Clocks[twin.ID] = simulation.NewClock(time.Now().UTC())
	s.Clocks[twin.ID].Resume()
	s.Audit.Record(userIDFrom(r), "twin_created", "twin:"+twin.ID, twin.ID, map[string]interface{}{"name": req.Name})
	writeJSON(w, http.StatusCreated, twin)
}

func (s *Server) handleGetTwin(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	twin, ok := s.Reg.GetTwin(id)
	if !ok {
		writeError(w, http.StatusNotFound, "twin not found")
		return
	}
	writeJSON(w, http.StatusOK, twin)
}

func (s *Server) handleCloneTwin(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Name string `json:"name"`
	}
	_ = decodeJSON(r, &req)
	if req.Name == "" {
		req.Name = "Clone"
	}
	clone, err := s.Reg.CloneTwin(id, req.Name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	s.Clocks[clone.ID] = simulation.NewClock(time.Now().UTC())
	s.Clocks[clone.ID].Resume()
	s.Audit.Record(userIDFrom(r), "twin_cloned", "twin:"+clone.ID, clone.ID, map[string]interface{}{"from": id})
	writeJSON(w, http.StatusCreated, clone)
}

func (s *Server) handleArchiveTwin(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.Reg.ArchiveTwin(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	s.Audit.Record(userIDFrom(r), "twin_archived", "twin:"+id, id, nil)
	writeJSON(w, http.StatusOK, map[string]bool{"archived": true})
}

func (s *Server) handleExportTwinJSON(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	data, err := exportimport.ExportJSON(s.Reg, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.json", id))
	w.Write(data)
}

func (s *Server) handleExportTwinCSV(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	csvText, err := exportimport.ExportCSV(s.Reg, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", id))
	io.WriteString(w, csvText)
}

func (s *Server) handleImportTwin(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	contentType := r.Header.Get("Content-Type")
	if contentType == "text/csv" {
		body, _ := io.ReadAll(r.Body)
		created, errs := exportimport.ImportCSV(s.Reg, id, string(body))
		writeJSON(w, http.StatusOK, map[string]interface{}{"created": len(created), "entities": created, "errors": errs})
		return
	}
	var req struct {
		Rows []exportimport.EntityImportRow `json:"rows"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	created, errs := exportimport.ImportJSON(s.Reg, id, req.Rows)
	s.Audit.Record(userIDFrom(r), "twin_imported", "twin:"+id, id, map[string]interface{}{"created": len(created)})
	writeJSON(w, http.StatusOK, map[string]interface{}{"created": len(created), "entities": created, "errors": errs})
}

func (s *Server) handleAuditTrail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	writeJSON(w, http.StatusOK, s.Audit.Recent(id, 200))
}

// ---------- Entities ----------

func (s *Server) handleListEntities(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	writeJSON(w, http.StatusOK, s.Reg.ListEntities(id))
}

func (s *Server) handleCreateEntity(w http.ResponseWriter, r *http.Request) {
	twinID := r.PathValue("id")
	var e model.Entity
	if err := decodeJSON(r, &e); err != nil || e.Name == "" || e.Type == "" {
		writeError(w, http.StatusBadRequest, "name and type are required")
		return
	}
	e.TwinID = twinID
	created := s.Reg.CreateEntity(&e)
	s.Audit.Record(userIDFrom(r), "entity_created", "entity:"+created.ID, twinID, map[string]interface{}{"name": e.Name, "type": e.Type})
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleGetEntity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	e, ok := s.Reg.GetEntity(id)
	if !ok {
		writeError(w, http.StatusNotFound, "entity not found")
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (s *Server) handleUpdateEntity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var patch map[string]interface{}
	if err := decodeJSON(r, &patch); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	e, err := s.Reg.UpdateEntity(id, func(e *model.Entity) { applyEntityPatch(e, patch) })
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	s.Audit.Record(userIDFrom(r), "entity_updated", "entity:"+id, e.TwinID, patch)
	writeJSON(w, http.StatusOK, e)
}

func applyEntityPatch(e *model.Entity, patch map[string]interface{}) {
	if v, ok := patch["name"].(string); ok {
		e.Name = v
	}
	if v, ok := patch["status"].(string); ok {
		e.Status = model.Status(v)
	}
	if v, ok := patch["position"].(map[string]interface{}); ok {
		if x, ok := v["x"].(float64); ok {
			e.Position.X = x
		}
		if y, ok := v["y"].(float64); ok {
			e.Position.Y = y
		}
		if z, ok := v["z"].(float64); ok {
			e.Position.Z = z
		}
	}
	if v, ok := patch["properties"].(map[string]interface{}); ok {
		if e.Properties == nil {
			e.Properties = map[string]interface{}{}
		}
		for k, val := range v {
			e.Properties[k] = val
		}
	}
}

func (s *Server) handleDeleteEntity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.Reg.DeleteEntity(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	s.Audit.Record(userIDFrom(r), "entity_deleted", "entity:"+id, "", nil)
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

// ---------- Relationships ----------

func (s *Server) handleListRelationships(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.Reg.ListRelationships(r.PathValue("id")))
}

func (s *Server) handleCreateRelationship(w http.ResponseWriter, r *http.Request) {
	var rel model.Relationship
	if err := decodeJSON(r, &rel); err != nil || rel.SourceID == "" || rel.TargetID == "" || rel.Type == "" {
		writeError(w, http.StatusBadRequest, "sourceId, targetId and type are required")
		return
	}
	if rel.Confidence == 0 {
		rel.Confidence = 1.0
	}
	created := s.Reg.CreateRelationship(&rel)
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleDeleteRelationship(w http.ResponseWriter, r *http.Request) {
	if err := s.Reg.DeleteRelationship(r.PathValue("id")); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

// ---------- Sensors ----------

func (s *Server) handleListSensors(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.Reg.ListSensors(r.PathValue("id")))
}

func (s *Server) handleCreateSensor(w http.ResponseWriter, r *http.Request) {
	twinID := r.PathValue("id")
	var sn model.Sensor
	if err := decodeJSON(r, &sn); err != nil || sn.EntityID == "" || sn.Type == "" {
		writeError(w, http.StatusBadRequest, "entityId and type are required")
		return
	}
	sn.TwinID = twinID
	created := s.Reg.CreateSensor(&sn)
	writeJSON(w, http.StatusCreated, created)
}

// ---------- Operating Hours ----------

func (s *Server) handleListOperatingHours(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.Reg.ListOperatingHours(r.PathValue("id")))
}

func (s *Server) handleSetOperatingHours(w http.ResponseWriter, r *http.Request) {
	twinID := r.PathValue("id")
	var h model.OperatingHours
	if err := decodeJSON(r, &h); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	h.TwinID = twinID
	saved := s.Reg.SetOperatingHours(&h)
	s.Audit.Record(userIDFrom(r), "operating_hours_set", "hours:"+saved.ID, twinID, nil)
	writeJSON(w, http.StatusOK, saved)
}

func (s *Server) handleOperatingHoursStatus(w http.ResponseWriter, r *http.Request) {
	twinID := r.PathValue("id")
	all := s.Reg.ListOperatingHours(twinID)
	type statusEntry struct {
		Schedule model.OperatingHours  `json:"schedule"`
		Status   model.OperatingStatus `json:"status"`
	}
	out := make([]statusEntry, 0, len(all))
	for _, h := range all {
		st, err := operatinghours.Compute(*h, time.Now())
		if err != nil {
			continue
		}
		out = append(out, statusEntry{Schedule: *h, Status: st})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleDeleteOperatingHours(w http.ResponseWriter, r *http.Request) {
	if err := s.Reg.DeleteOperatingHours(r.PathValue("id")); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}
