package api

import (
	"net/http"
	"strconv"
	"time"

	"rift/internal/synthetic"
)

// handleGenerateTemplate populates an existing (typically freshly created,
// empty) twin from one of the bundled Templates — used by the CLI's
// `rift twin create --template=...` convenience flag and by the Twin
// Builder's "start from a template" action.
func (s *Server) handleGenerateTemplate(w http.ResponseWriter, r *http.Request) {
	twinID := r.PathValue("id")
	template := r.PathValue("template")
	if _, ok := s.Reg.GetTwin(twinID); !ok {
		writeError(w, http.StatusNotFound, "twin not found")
		return
	}
	b := synthetic.NewBuilder(s.Reg, twinID, time.Now().UnixNano())
	switch template {
	case "factory":
		b.Factory()
	case "building":
		b.Building()
	case "data_center":
		b.DataCenter()
	case "warehouse":
		b.Warehouse()
	case "smart_city":
		b.SmartCity()
	default:
		writeError(w, http.StatusBadRequest, "unknown template: use factory, building, data_center, warehouse or smart_city")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"entities": len(s.Reg.ListEntities(twinID))})
}

// handleGenerateBenchmark implements the Synthetic Twin Generator for load
// testing: it creates roughly `count` entities under twinID with a fixed
// seed so results are reproducible, and reports how long generation took —
// the Entity/sec figure the Benchmark suite in the README refers to.
func (s *Server) handleGenerateBenchmark(w http.ResponseWriter, r *http.Request) {
	twinID := r.PathValue("id")
	if _, ok := s.Reg.GetTwin(twinID); !ok {
		writeError(w, http.StatusNotFound, "twin not found")
		return
	}
	count := 1000
	if v := r.URL.Query().Get("count"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			count = parsed
		}
	}
	seed := int64(42)
	if v := r.URL.Query().Get("seed"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			seed = parsed
		}
	}
	start := time.Now()
	b := synthetic.NewBuilder(s.Reg, twinID, seed)
	created := b.Benchmark(count)
	elapsed := time.Since(start)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"entitiesCreated": created,
		"elapsedMs":       elapsed.Milliseconds(),
		"entitiesPerSec":  float64(created) / elapsed.Seconds(),
	})
}
