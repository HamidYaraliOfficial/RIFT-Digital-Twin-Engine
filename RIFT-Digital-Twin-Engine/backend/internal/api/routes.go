package api

import "net/http"

// routes registers every REST + SSE endpoint. Go 1.22's enhanced ServeMux
// gives method+pattern routing with path parameters natively, so the whole
// API Layer needs no external router dependency.
//
// Authorization policy for this reference build: every read (GET) endpoint
// is open so the Control Center can browse freely; every mutating endpoint
// requires a valid session (s.requireAuth) except /api/auth/login itself and
// /api/telemetry, which is the machine-to-machine sensor ingestion endpoint
// (production deployments should protect it with a separate per-device API
// key rather than a user session token — see the README's security notes).
func (s *Server) routes() {
	mux := s.mux
	auth := s.requireAuth

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "rift-digital-twin-engine"})
	})
	mux.HandleFunc("GET /api/stats", s.handlePipelineStats)

	// Auth
	mux.HandleFunc("POST /api/auth/login", s.handleLogin)
	mux.HandleFunc("POST /api/auth/register", auth(s.handleRegister))

	// Twins
	mux.HandleFunc("GET /api/twins", s.handleListTwins)
	mux.HandleFunc("POST /api/twins", auth(s.handleCreateTwin))
	mux.HandleFunc("GET /api/twins/{id}", s.handleGetTwin)
	mux.HandleFunc("POST /api/twins/{id}/clone", auth(s.handleCloneTwin))
	mux.HandleFunc("POST /api/twins/{id}/archive", auth(s.handleArchiveTwin))
	mux.HandleFunc("GET /api/twins/{id}/export.json", s.handleExportTwinJSON)
	mux.HandleFunc("GET /api/twins/{id}/export.csv", s.handleExportTwinCSV)
	mux.HandleFunc("POST /api/twins/{id}/import", auth(s.handleImportTwin))
	mux.HandleFunc("GET /api/twins/{id}/audit", s.handleAuditTrail)
	mux.HandleFunc("GET /api/twins/{id}/stream", s.handleStream)
	mux.HandleFunc("GET /api/twins/{id}/health", s.handleTwinHealth)
	mux.HandleFunc("GET /api/twins/{id}/events", s.handleRecentEvents)

	// Entities
	mux.HandleFunc("GET /api/twins/{id}/entities", s.handleListEntities)
	mux.HandleFunc("POST /api/twins/{id}/entities", auth(s.handleCreateEntity))
	mux.HandleFunc("GET /api/entities/{id}", s.handleGetEntity)
	mux.HandleFunc("PATCH /api/entities/{id}", auth(s.handleUpdateEntity))
	mux.HandleFunc("DELETE /api/entities/{id}", auth(s.handleDeleteEntity))
	mux.HandleFunc("GET /api/entities/{id}/impact", s.handleImpactAnalysis)
	mux.HandleFunc("GET /api/entities/{id}/rootcause", s.handleRootCause)
	mux.HandleFunc("POST /api/entities/{id}/command", auth(s.handleActuatorCommand))
	mux.HandleFunc("POST /api/entities/{id}/actuator-policy", auth(s.handleSetActuatorPolicy))

	// Relationships
	mux.HandleFunc("GET /api/twins/{id}/relationships", s.handleListRelationships)
	mux.HandleFunc("POST /api/twins/{id}/relationships", auth(s.handleCreateRelationship))
	mux.HandleFunc("DELETE /api/relationships/{id}", auth(s.handleDeleteRelationship))

	// Sensors + telemetry
	mux.HandleFunc("GET /api/twins/{id}/sensors", s.handleListSensors)
	mux.HandleFunc("POST /api/twins/{id}/sensors", auth(s.handleCreateSensor))
	mux.HandleFunc("POST /api/telemetry", s.handleIngestTelemetry)
	mux.HandleFunc("GET /api/sensors/{id}/telemetry", s.handleSensorTelemetry)
	mux.HandleFunc("GET /api/sensors/{id}/downsample", s.handleSensorDownsample)

	// Operating hours (user-entered schedule -> live open/closed + countdown)
	mux.HandleFunc("GET /api/twins/{id}/operating-hours", s.handleListOperatingHours)
	mux.HandleFunc("POST /api/twins/{id}/operating-hours", auth(s.handleSetOperatingHours))
	mux.HandleFunc("GET /api/twins/{id}/operating-hours/status", s.handleOperatingHoursStatus)
	mux.HandleFunc("DELETE /api/operating-hours/{id}", auth(s.handleDeleteOperatingHours))

	// Rules + alerts
	mux.HandleFunc("GET /api/twins/{id}/rules", s.handleListRules)
	mux.HandleFunc("POST /api/twins/{id}/rules", auth(s.handleCreateRule))
	mux.HandleFunc("DELETE /api/rules/{id}", auth(s.handleDeleteRule))
	mux.HandleFunc("GET /api/twins/{id}/alerts", s.handleListAlerts)
	mux.HandleFunc("POST /api/alerts/{id}/ack", auth(s.handleAckAlert))
	mux.HandleFunc("POST /api/alerts/{id}/resolve", auth(s.handleResolveAlert))

	// Scenarios / What-If / Monte Carlo / Sweep / Optimization / Clock
	mux.HandleFunc("GET /api/twins/{id}/scenarios", s.handleListScenarios)
	mux.HandleFunc("POST /api/twins/{id}/scenarios", auth(s.handleCreateScenario))
	mux.HandleFunc("GET /api/scenarios/{id}", s.handleGetScenario)
	mux.HandleFunc("POST /api/scenarios/{id}/run", auth(s.handleRunScenario))
	mux.HandleFunc("POST /api/twins/{id}/montecarlo", auth(s.handleMonteCarlo))
	mux.HandleFunc("POST /api/twins/{id}/sweep", auth(s.handleSweep))
	mux.HandleFunc("POST /api/twins/{id}/optimize", auth(s.handleOptimize))
	mux.HandleFunc("GET /api/twins/{id}/clock", s.handleGetClock)
	mux.HandleFunc("POST /api/twins/{id}/clock/{action}", auth(s.handleClockAction))

	// Permissions
	mux.HandleFunc("POST /api/permissions", auth(s.handleGrantPermission))

	// Synthetic Twin Generator (templates + benchmark seeding)
	mux.HandleFunc("POST /api/twins/{id}/synthetic/{template}", auth(s.handleGenerateTemplate))
	mux.HandleFunc("POST /api/twins/{id}/synthetic/benchmark", auth(s.handleGenerateBenchmark))
}
