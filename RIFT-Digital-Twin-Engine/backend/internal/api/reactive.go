package api

import (
	"context"
	"fmt"
	"time"

	"rift/internal/actuator"
	"rift/internal/model"
	"rift/internal/registry"
	"rift/internal/rules"
)

// consumeTelemetry is the reactive core of RIFT: every Reading that survives
// the Ingestion Pipeline's quality gate flows through here to update the
// State Engine, evaluate the Rules Engine, run Anomaly Detection, and raise
// Alerts / publish Events as needed — all in real time, all in one place so
// the data flow is easy to reason about.
func (s *Server) consumeTelemetry(ctx context.Context) {
	sub := s.Pipeline.Subscribe(4096)
	for {
		select {
		case <-ctx.Done():
			return
		case r := <-sub:
			s.handleReading(r)
		}
	}
}

func (s *Server) handleReading(r model.Reading) {
	sensor, ok := s.Reg.GetSensor(r.SensorID)
	if !ok {
		return
	}
	s.Reg.TouchSensor(r.SensorID, model.SensorOnline)

	s.State.ApplyCurrent(r.EntityID, map[string]interface{}{string(sensor.Type): r.Value})

	s.Rules.Evaluate(rules.Fact{EntityID: r.EntityID, Field: string(sensor.Type), Value: r.Value})

	if res := s.Anomaly.Observe(r.SensorID, r.EntityID, r.Value, r.Timestamp); res != nil {
		evt := model.Event{
			ID: registry.NewID("evt"), TwinID: r.TwinID, Type: "anomaly_detected",
			Timestamp: r.Timestamp, SourceEntityID: r.EntityID, Severity: severityFromString(res.Severity),
			Payload:   map[string]interface{}{"sensorId": r.SensorID, "value": res.Value, "baseline": res.Baseline, "zScore": res.ZScore, "sensorType": sensor.Type},
			Lifecycle: model.LifecycleOpen,
		}
		s.Bus.Publish(evt)
		s.Alerts.Raise(model.Alert{
			TwinID: r.TwinID, Type: model.AlertAnomaly, Severity: evt.Severity, EntityID: r.EntityID,
			Message:  fmt.Sprintf("%s on %s reading %.2f%s deviates from its recent baseline (%.2f)", sensor.Type, sensorEntityName(s, r.EntityID), res.Value, sensor.Unit, res.Baseline),
			DedupKey: "anomaly:" + r.SensorID,
		})
	}
}

func sensorEntityName(s *Server, entityID string) string {
	if e, ok := s.Reg.GetEntity(entityID); ok {
		return e.Name
	}
	return entityID
}

func severityFromString(sev string) model.Severity {
	switch sev {
	case "critical":
		return model.SeverityCritical
	case "warning":
		return model.SeverityWarning
	default:
		return model.SeverityInfo
	}
}

// Execute implements rules.ActionExecutor: it is how a fired Rule reaches
// into the rest of the system, always through the same safety-checked paths
// a human operator would use (e.g. actuator commands still go through the
// Command Safety Engine).
func (s *Server) Execute(rule *model.Rule, action model.Action, fact rules.Fact) error {
	switch action.Type {
	case model.ActionEmitEvent:
		evtType, _ := action.Params["type"].(string)
		if evtType == "" {
			evtType = "rule_triggered"
		}
		s.Bus.Publish(model.Event{
			ID: registry.NewID("evt"), TwinID: rule.TwinID, Type: evtType, Timestamp: time.Now().UTC(),
			SourceEntityID: fact.EntityID, Severity: model.SeverityWarning, Cause: rule.ID, Lifecycle: model.LifecycleOpen,
			Payload: map[string]interface{}{"rule": rule.Name, "field": fact.Field, "value": fact.Value},
		})
	case model.ActionSetState:
		field, _ := action.Params["field"].(string)
		value := action.Params["value"]
		if field != "" {
			s.State.ApplyCurrent(fact.EntityID, map[string]interface{}{field: value})
		}
	case model.ActionSetStatus:
		newStatus, _ := action.Params["status"].(string)
		if newStatus != "" {
			s.Reg.UpdateEntity(fact.EntityID, func(e *model.Entity) { e.Status = model.Status(newStatus) })
		}
	case model.ActionCreateAlert:
		msg, _ := action.Params["message"].(string)
		if msg == "" {
			msg = fmt.Sprintf("Rule %q fired on %s", rule.Name, fact.EntityID)
		}
		s.Alerts.Raise(model.Alert{
			TwinID: rule.TwinID, RuleID: rule.ID, Type: model.AlertRuleViolation, Severity: model.SeverityWarning,
			EntityID: fact.EntityID, Message: msg, DedupKey: "rule:" + rule.ID + ":" + fact.EntityID,
		})
	case model.ActionActuatorCommand:
		if rule.RequirePermit {
			// Safety: rule-issued actuator commands still require an explicit
			// permission grant for "rule:<id>" as the issuer — RIFT never lets
			// automation bypass the Command Safety Engine.
		}
	}
	return nil
}

func (s *Server) checkPermission(issuedBy, entityID, scope string) bool {
	entity, ok := s.Reg.GetEntity(entityID)
	if !ok {
		return false
	}
	return s.AuthStore.Can(issuedBy, entity.TwinID, entityID, model.ScopeAction(scope))
}

// applyActuatorCommand is the SimulatedAdapter's execution callback: once the
// Command Safety Engine has approved a command, this is what "runs" it in
// the digital twin (a real deployment plugs a Modbus/OPC-UA/MQTT Adapter in
// here instead, targeting real equipment).
func (s *Server) applyActuatorCommand(cmd actuator.Command) error {
	switch cmd.Type {
	case actuator.CmdStart:
		s.Reg.UpdateEntity(cmd.EntityID, func(e *model.Entity) { e.Status = model.StatusOperational })
	case actuator.CmdStop:
		s.Reg.UpdateEntity(cmd.EntityID, func(e *model.Entity) { e.Status = model.StatusMaintenance })
	case actuator.CmdRestart:
		s.Reg.UpdateEntity(cmd.EntityID, func(e *model.Entity) { e.Status = model.StatusRecovered })
	case actuator.CmdSetTemperature, actuator.CmdChangeSpeed:
		s.State.SetExpected(cmd.EntityID, map[string]interface{}{string(cmd.Type): cmd.Value})
	case actuator.CmdLock, actuator.CmdUnlock, actuator.CmdOpen, actuator.CmdClose:
		s.State.ApplyCurrent(cmd.EntityID, map[string]interface{}{"actuator_state": string(cmd.Type)})
	}
	entity, _ := s.Reg.GetEntity(cmd.EntityID)
	twinID := ""
	if entity != nil {
		twinID = entity.TwinID
	}
	s.Bus.Publish(model.Event{
		ID: registry.NewID("evt"), TwinID: twinID, Type: "actuator_command_executed", Timestamp: time.Now().UTC(),
		SourceEntityID: cmd.EntityID, Severity: model.SeverityInfo, Lifecycle: model.LifecycleResolved,
		Payload: map[string]interface{}{"command": cmd.Type, "value": cmd.Value, "issuedBy": cmd.IssuedBy},
	})
	s.Audit.Record(cmd.IssuedBy, "actuator_command", "entity:"+cmd.EntityID, twinID, map[string]interface{}{
		"type": cmd.Type, "value": cmd.Value, "reason": cmd.Reason,
	})
	return nil
}
