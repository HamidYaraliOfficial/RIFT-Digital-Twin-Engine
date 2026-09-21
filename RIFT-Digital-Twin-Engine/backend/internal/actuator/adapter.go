package actuator

// Adapter is the extension point for real actuator connectivity
// (OPC-UA, Modbus/TCP, MQTT, gRPC, HTTP). RIFT ships SimulatedAdapter so
// every command has a safe destination out of the box; production adapters
// implement this same interface against real hardware, always sitting
// behind the SafetyEngine gate above — a Rule or the AI Assistant can never
// reach an Adapter directly.
type Adapter interface {
	Execute(cmd Command) error
}

// SimulatedAdapter records commands via a callback (used to update the
// entity's simulated/expected state in the State Engine) instead of talking
// to real equipment.
type SimulatedAdapter struct {
	OnExecute func(cmd Command) error
}

func (s SimulatedAdapter) Execute(cmd Command) error {
	if s.OnExecute != nil {
		return s.OnExecute(cmd)
	}
	return nil
}
