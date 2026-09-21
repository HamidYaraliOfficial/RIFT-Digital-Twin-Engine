package api

import (
	"net/http"
	"strconv"
	"time"

	"rift/internal/graph"
	"rift/internal/health"
	"rift/internal/model"
	"rift/internal/rootcause"
)

// handleIngestTelemetry is the HTTP transport implementation of the Sensor
// Abstraction Layer's inbound side (the "HTTP adapter" from the connector
// framework): any real sensor, gateway, or SCADA bridge can push readings
// here with a simple POST instead of implementing a pull-based Adapter.
func (s *Server) handleIngestTelemetry(w http.ResponseWriter, r *http.Request) {
	var reading model.Reading
	if err := decodeJSON(r, &reading); err != nil {
		writeError(w, http.StatusBadRequest, "invalid reading payload")
		return
	}
	sensor, ok := s.Reg.GetSensor(reading.SensorID)
	if !ok {
		writeError(w, http.StatusNotFound, "unknown sensorId")
		return
	}
	reading.EntityID = sensor.EntityID
	reading.TwinID = sensor.TwinID
	if reading.Timestamp.IsZero() {
		reading.Timestamp = time.Now().UTC()
	}
	if reading.Source == "" {
		reading.Source = "http"
	}
	accepted := s.Pipeline.Ingest(reading)
	writeJSON(w, http.StatusAccepted, map[string]interface{}{"accepted": accepted})
}

func (s *Server) handleSensorTelemetry(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	n := 200
	if v := r.URL.Query().Get("window"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			n = parsed
		}
	}
	writeJSON(w, http.StatusOK, s.Store.Window(id, n))
}

func (s *Server) handleSensorDownsample(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	bucketSecs := 60
	if v := r.URL.Query().Get("bucketSeconds"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			bucketSecs = parsed
		}
	}
	writeJSON(w, http.StatusOK, s.Store.Downsample(id, time.Duration(bucketSecs)*time.Second))
}

func (s *Server) handlePipelineStats(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.Pipeline.Stats())
}

func (s *Server) handleRecentEvents(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	n := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			n = parsed
		}
	}
	writeJSON(w, http.StatusOK, s.Bus.Recent(id, n))
}

// ---------- Health / Impact / Root Cause (Digital Twin analytics engines) ----------

func (s *Server) handleTwinHealth(w http.ResponseWriter, r *http.Request) {
	twinID := r.PathValue("id")
	entities := s.Reg.ListEntities(twinID)
	total, sum := 0.0, 0.0
	for _, e := range entities {
		sum += e.Health.Health
		total++
	}
	overall := 100.0
	if total > 0 {
		overall = sum / total
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"overallHealth": overall, "entities": entities})
}

func (s *Server) handleImpactAnalysis(w http.ResponseWriter, r *http.Request) {
	entityID := r.PathValue("id")
	entity, ok := s.Reg.GetEntity(entityID)
	if !ok {
		writeError(w, http.StatusNotFound, "entity not found")
		return
	}
	g := graph.Build(s.Reg.ListRelationships(entity.TwinID))
	depth := 10
	if v := r.URL.Query().Get("depth"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			depth = parsed
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"entityId":     entityID,
		"impact":       g.ImpactAnalysis(entityID, depth),
		"dependents":   g.Dependents(entityID),
		"reachable":    g.Reachable(entityID),
		"dependencyDepth": g.Depth(entityID),
	})
}

func (s *Server) handleRootCause(w http.ResponseWriter, r *http.Request) {
	entityID := r.PathValue("id")
	entity, ok := s.Reg.GetEntity(entityID)
	if !ok {
		writeError(w, http.StatusNotFound, "entity not found")
		return
	}
	windowMin := 30
	if v := r.URL.Query().Get("windowMinutes"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			windowMin = parsed
		}
	}
	g := graph.Build(s.Reg.ListRelationships(entity.TwinID))
	end := time.Now().UTC()
	start := end.Add(-time.Duration(windowMin) * time.Minute)
	events := s.Bus.Recent(entity.TwinID, 1000)
	candidates := rootcause.Analyze(g, events, entityID, start, end)
	writeJSON(w, http.StatusOK, map[string]interface{}{"entityId": entityID, "candidates": candidates, "windowStart": start, "windowEnd": end})
}

var _ = health.Inputs{} // package referenced for entity health computation elsewhere in this file's callers
