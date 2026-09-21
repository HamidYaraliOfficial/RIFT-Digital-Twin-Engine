package api

import (
	"net/http"
	"time"

	"rift/internal/model"
	"rift/internal/optimization"
	"rift/internal/registry"
	"rift/internal/simulation"
)

func (s *Server) handleListScenarios(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.Scenarios.List(r.PathValue("id")))
}

type createScenarioRequest struct {
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	Seed          int64              `json:"seed"`
	Deterministic bool               `json:"deterministic"`
	Faults        []model.FaultSpec  `json:"faults"`
	DurationTicks int                `json:"durationTicks"`
}

func (s *Server) handleCreateScenario(w http.ResponseWriter, r *http.Request) {
	twinID := r.PathValue("id")
	var req createScenarioRequest
	if err := decodeJSON(r, &req); err != nil || req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.DurationTicks <= 0 {
		req.DurationTicks = 50
	}
	sc := s.Scenarios.Create(&model.Scenario{
		ID: registry.NewID("scn"), TwinID: twinID, Name: req.Name, Description: req.Description,
		Seed: req.Seed, Deterministic: req.Deterministic, Faults: req.Faults, DurationTicks: req.DurationTicks,
	})
	s.Audit.Record(userIDFrom(r), "scenario_created", "scenario:"+sc.ID, twinID, map[string]interface{}{"name": req.Name})
	writeJSON(w, http.StatusCreated, sc)
}

func (s *Server) handleGetScenario(w http.ResponseWriter, r *http.Request) {
	sc, ok := s.Scenarios.Get(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "scenario not found")
		return
	}
	writeJSON(w, http.StatusOK, sc)
}

func (s *Server) handleRunScenario(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	sc, ok := s.Scenarios.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "scenario not found")
		return
	}
	baseline := s.Reg.Snapshot(sc.TwinID)
	result := s.Scenarios.Execute(sc, baseline)
	s.Alerts.Raise(model.Alert{
		TwinID: sc.TwinID, Type: model.AlertScenarioResult, Severity: model.SeverityInfo, EntityID: "",
		Message: result.Summary, DedupKey: "scenario:" + sc.ID,
	})
	s.Audit.Record(userIDFrom(r), "scenario_run", "scenario:"+id, sc.TwinID, map[string]interface{}{"affected": len(result.AffectedEntities)})
	writeJSON(w, http.StatusOK, result)
}

// ---------- Monte Carlo ----------

type monteCarloRequest struct {
	Faults        []model.FaultSpec `json:"faults"`
	DurationTicks int               `json:"durationTicks"`
	Runs          int               `json:"runs"`
	MagMin        float64           `json:"magMin"`
	MagMax        float64           `json:"magMax"`
}

func (s *Server) handleMonteCarlo(w http.ResponseWriter, r *http.Request) {
	twinID := r.PathValue("id")
	var req monteCarloRequest
	if err := decodeJSON(r, &req); err != nil || len(req.Faults) == 0 {
		writeError(w, http.StatusBadRequest, "at least one fault is required")
		return
	}
	if req.Runs <= 0 {
		req.Runs = 100
	}
	if req.DurationTicks <= 0 {
		req.DurationTicks = 30
	}
	if req.MagMax == 0 {
		req.MagMax = 1
	}
	baseline := s.Reg.Snapshot(twinID)
	result := optimization.RunMonteCarlo("affected_entities", req.Runs, req.MagMin, req.MagMax, func(seed int64, magnitude float64) float64 {
		faults := scaleFaultMagnitudes(req.Faults, magnitude)
		branch := simulation.NewBranch(baseline, seed)
		events := simulation.Run(branch, faults, req.DurationTicks, []simulation.BehaviorModel{simulation.PropagationModel{}})
		cmp := simulation.Compare(baseline, branch, req.DurationTicks, events)
		return float64(len(cmp.AffectedEntities))
	})
	writeJSON(w, http.StatusOK, result)
}

// ---------- Parameter Sweep ----------

type sweepRequest struct {
	Parameter     string             `json:"parameter"`
	Values        []float64          `json:"values"`
	Faults        []model.FaultSpec  `json:"faults"`
	DurationTicks int                `json:"durationTicks"`
}

func (s *Server) handleSweep(w http.ResponseWriter, r *http.Request) {
	twinID := r.PathValue("id")
	var req sweepRequest
	if err := decodeJSON(r, &req); err != nil || len(req.Values) == 0 || len(req.Faults) == 0 {
		writeError(w, http.StatusBadRequest, "values and faults are required")
		return
	}
	if req.DurationTicks <= 0 {
		req.DurationTicks = 30
	}
	baseline := s.Reg.Snapshot(twinID)
	results := optimization.RunSweep(req.Parameter, req.Values, func(v float64) float64 {
		faults := scaleFaultMagnitudes(req.Faults, v)
		branch := simulation.NewBranch(baseline, 1)
		events := simulation.Run(branch, faults, req.DurationTicks, []simulation.BehaviorModel{simulation.PropagationModel{}})
		cmp := simulation.Compare(baseline, branch, req.DurationTicks, events)
		return float64(len(cmp.AffectedEntities))
	})
	writeJSON(w, http.StatusOK, results)
}

func scaleFaultMagnitudes(faults []model.FaultSpec, magnitude float64) []model.FaultSpec {
	out := make([]model.FaultSpec, len(faults))
	for i, f := range faults {
		out[i] = f
		if out[i].Magnitude == 0 {
			out[i].Magnitude = magnitude
		} else {
			out[i].Magnitude = out[i].Magnitude * magnitude
		}
	}
	return out
}

// ---------- Optimization (combinable objectives) ----------

type optimizeRequest struct {
	ParameterName string            `json:"parameterName"`
	Min           float64           `json:"min"`
	Max           float64           `json:"max"`
	Faults        []model.FaultSpec `json:"faults"`
	DurationTicks int               `json:"durationTicks"`
	Trials        int               `json:"trials"`
	CostWeight    float64           `json:"costWeight"`
	RiskWeight    float64           `json:"riskWeight"`
}

// handleOptimize implements the Optimization Engine for combinable
// objectives: it minimizes (CostWeight * parameter value) + (RiskWeight *
// simulated impact) by actually running the What-If simulation at each
// candidate parameter value — the risk term is never guessed.
func (s *Server) handleOptimize(w http.ResponseWriter, r *http.Request) {
	twinID := r.PathValue("id")
	var req optimizeRequest
	if err := decodeJSON(r, &req); err != nil || req.ParameterName == "" || len(req.Faults) == 0 {
		writeError(w, http.StatusBadRequest, "parameterName and faults are required")
		return
	}
	if req.Trials <= 0 {
		req.Trials = 60
	}
	if req.DurationTicks <= 0 {
		req.DurationTicks = 30
	}
	if req.CostWeight == 0 {
		req.CostWeight = 1
	}
	if req.RiskWeight == 0 {
		req.RiskWeight = 1
	}
	baseline := s.Reg.Snapshot(twinID)
	riskOf := func(params map[string]float64) float64 {
		mag := params[req.ParameterName]
		faults := scaleFaultMagnitudes(req.Faults, mag)
		branch := simulation.NewBranch(baseline, 7)
		events := simulation.Run(branch, faults, req.DurationTicks, []simulation.BehaviorModel{simulation.PropagationModel{}})
		cmp := simulation.Compare(baseline, branch, req.DurationTicks, events)
		return float64(len(cmp.AffectedEntities))
	}
	objectives := []optimization.Objective{
		{Name: "cost", Weight: req.CostWeight, Direction: -1, Evaluate: func(p map[string]float64) float64 { return p[req.ParameterName] }},
		{Name: "risk", Weight: req.RiskWeight, Direction: -1, Evaluate: riskOf},
	}
	bounds := map[string]optimization.Bounds{req.ParameterName: {Min: req.Min, Max: req.Max}}
	result := optimization.Optimize(objectives, bounds, req.Trials, time.Now().UnixNano())
	writeJSON(w, http.StatusOK, result)
}

// ---------- Time Engine / Clock ----------

func (s *Server) handleGetClock(w http.ResponseWriter, r *http.Request) {
	twinID := r.PathValue("id")
	clock, ok := s.Clocks[twinID]
	if !ok {
		writeError(w, http.StatusNotFound, "no clock for this twin")
		return
	}
	writeJSON(w, http.StatusOK, clock.State())
}

func (s *Server) handleClockAction(w http.ResponseWriter, r *http.Request) {
	twinID := r.PathValue("id")
	action := r.PathValue("action")
	clock, ok := s.Clocks[twinID]
	if !ok {
		writeError(w, http.StatusNotFound, "no clock for this twin")
		return
	}
	var body struct {
		Speed   float64 `json:"speed"`
		Seconds int     `json:"seconds"`
		Seek    string  `json:"seek"`
	}
	_ = decodeJSON(r, &body)
	switch action {
	case "pause":
		clock.Pause()
	case "resume":
		clock.Resume()
	case "speed":
		clock.SetSpeed(body.Speed)
	case "step":
		clock.Step(time.Duration(body.Seconds) * time.Second)
	case "seek":
		if t, err := time.Parse(time.RFC3339, body.Seek); err == nil {
			clock.Seek(t)
		}
	default:
		writeError(w, http.StatusBadRequest, "unknown clock action")
		return
	}
	writeJSON(w, http.StatusOK, clock.State())
}
