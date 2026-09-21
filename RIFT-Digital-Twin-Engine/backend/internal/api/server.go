// Package api wires every engine together behind a REST + Server-Sent-Events
// HTTP surface (the API Layer). It is the only package allowed to import all
// the others; every internal engine package stays independent and
// unit-testable on its own, which is what "Digital Twin Core has no direct
// dependency on UI/DB/Provider" means in practice.
package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"rift/internal/actuator"
	"rift/internal/alerting"
	"rift/internal/anomaly"
	"rift/internal/audit"
	"rift/internal/auth"
	"rift/internal/events"
	"rift/internal/graph"
	"rift/internal/health"
	"rift/internal/model"
	"rift/internal/registry"
	"rift/internal/rules"
	"rift/internal/simulation"
	"rift/internal/stateengine"
	"rift/internal/telemetry"
)

type Server struct {
	Reg        *registry.Registry
	Pipeline   *telemetry.Pipeline
	Store      *telemetry.Store
	Bus        *events.Bus
	State      *stateengine.Engine
	Rules      *rules.Engine
	Anomaly    *anomaly.Detector
	Alerts     *alerting.Manager
	Scenarios  *simulation.Engine
	Actuators  *actuator.SafetyEngine
	AuthStore  *auth.Store
	Tokens     *auth.TokenIssuer
	Audit      *audit.Log
	Clocks     map[string]*simulation.Clock
	DataDir    string

	mux *http.ServeMux
}

func NewServer(dataDir string) *Server {
	reg := registry.New()
	store := telemetry.NewStore(2000)
	pipeline := telemetry.NewPipeline(store, 10000, 8)
	bus := events.New(4096, 5000)
	tokenIssuer := auth.NewTokenIssuer(randomSecret())

	s := &Server{
		Reg:       reg,
		Pipeline:  pipeline,
		Store:     store,
		Bus:       bus,
		State:     stateengine.New(),
		Anomaly:   anomaly.New(60),
		Alerts:    alerting.NewManager(2 * time.Minute),
		Scenarios: simulation.NewEngine(),
		AuthStore: auth.NewStore(),
		Tokens:    tokenIssuer,
		Audit:     audit.New(dataDir+"/audit.log", 20000),
		Clocks:    map[string]*simulation.Clock{},
		DataDir:   dataDir,
	}
	s.Actuators = actuator.NewSafetyEngine(s.checkPermission, actuator.SimulatedAdapter{OnExecute: s.applyActuatorCommand})
	s.Rules = rules.New(s)

	_ = reg.LoadFromFile(dataDir + "/rift-state.json")
	if len(reg.ListTwins()) == 0 {
		s.seedDemoData()
	}
	for _, t := range reg.ListTwins() {
		s.Clocks[t.ID] = simulation.NewClock(time.Now().UTC())
		s.Clocks[t.ID].Resume()
	}

	s.mux = http.NewServeMux()
	s.routes()
	return s
}

func randomSecret() string {
	// A fresh random signing secret per process start is fine for the
	// reference deployment (sessions simply expire on restart); production
	// operators should set RIFT_SESSION_SECRET explicitly so restarts don't
	// invalidate active sessions. See cmd/riftd/main.go.
	return registry.NewID("secret") + registry.NewID("secret")
}

func (s *Server) Handler() http.Handler {
	return withCORS(withLogging(s.mux))
}

// RunBackground starts the telemetry pipeline workers, the reactive
// telemetry consumer (state + rules + anomaly + alerts), periodic health
// scoring, sensor-offline detection, and periodic snapshot persistence. It
// blocks until ctx is cancelled.
func (s *Server) RunBackground(ctx context.Context) {
	go s.Pipeline.Run(ctx)
	go s.consumeTelemetry(ctx)
	go s.periodicHealthScoring(ctx)
	go s.periodicPersistence(ctx)
	go s.runSimulatedAdapters(ctx)
	<-ctx.Done()
}

func (s *Server) periodicPersistence(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			if err := s.Reg.SaveToFile(s.DataDir + "/rift-state.json"); err != nil {
				log.Printf("final snapshot save failed: %v", err)
			}
			return
		case <-ticker.C:
			if err := s.Reg.SaveToFile(s.DataDir + "/rift-state.json"); err != nil {
				log.Printf("snapshot save failed: %v", err)
			}
		}
	}
}

// runSimulatedAdapters starts one SimulatedAdapter per twin over its own
// sensors, so every seeded/created twin has live telemetry flowing through
// the exact same Ingestion Pipeline a real MQTT/OPC-UA feed would use.
func (s *Server) runSimulatedAdapters(ctx context.Context) {
	started := map[string]bool{}
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, t := range s.Reg.ListTwins() {
				if started[t.ID] || t.Archived {
					continue
				}
				sensors := s.Reg.ListSensors(t.ID)
				var toSimulate []*model.Sensor
				for _, sn := range sensors {
					if sn.Source == "simulated" {
						toSimulate = append(toSimulate, sn)
					}
				}
				if len(toSimulate) == 0 {
					continue
				}
				started[t.ID] = true
				adapter := telemetry.NewSimulatedAdapter(toSimulate, 2*time.Second, 0)
				out := make(chan model.Reading, 256)
				go adapter.Start(ctx, out)
				go func() {
					for r := range out {
						s.Pipeline.Ingest(r)
					}
				}()
			}
		}
	}
}

func (s *Server) periodicHealthScoring(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, t := range s.Reg.ListTwins() {
				g := graph.Build(s.Reg.ListRelationships(t.ID))
				for _, e := range s.Reg.ListEntities(t.ID) {
					sensors := s.Reg.SensorsForEntity(e.ID)
					offline := 0
					for _, sn := range sensors {
						if sn.Status == model.SensorOffline {
							offline++
						}
					}
					ratio := 0.0
					if len(sensors) > 0 {
						ratio = float64(offline) / float64(len(sensors))
					}
					score := health.Compute(health.Inputs{
						Status:             e.Status,
						SensorOfflineRatio: ratio,
						DependencyCount:    len(g.Dependents(e.ID)),
					})
					s.Reg.UpdateEntity(e.ID, func(ent *model.Entity) { ent.Health = score })
				}
			}
		}
	}
}
