package telemetry

import (
	"context"
	"math"
	"math/rand"
	"time"

	"rift/internal/model"
)

// Adapter is the extension point of the Connector Framework for inbound
// telemetry. RIFT ships a fully working SimulatedAdapter (so every demo/
// template twin has live data with zero external setup) and an HTTPAdapter
// reference implementation. Real deployments add MQTT/gRPC/Kafka/NATS/OPC-UA/
// Modbus adapters behind this exact same interface; the Pipeline, Rules
// Engine, Anomaly Engine etc. never need to know which transport produced a
// Reading.
type Adapter interface {
	Name() string
	// Start begins pushing Readings into out until ctx is cancelled.
	Start(ctx context.Context, out chan<- model.Reading)
}

// SimulatedAdapter generates physically-plausible synthetic telemetry for a
// fixed set of sensors: a random walk around a per-sensor baseline, clamped
// to the sensor's declared range, with occasional noise spikes. It is
// deterministic when Seed is non-zero, enabling reproducible demos/tests.
type SimulatedAdapter struct {
	Sensors  []*model.Sensor
	Interval time.Duration
	Seed     int64

	baselines map[string]float64
}

func NewSimulatedAdapter(sensors []*model.Sensor, interval time.Duration, seed int64) *SimulatedAdapter {
	return &SimulatedAdapter{Sensors: sensors, Interval: interval, Seed: seed, baselines: map[string]float64{}}
}

func (a *SimulatedAdapter) Name() string { return "simulated" }

func (a *SimulatedAdapter) Start(ctx context.Context, out chan<- model.Reading) {
	rng := rand.New(rand.NewSource(a.Seed))
	if a.Seed == 0 {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	for _, s := range a.Sensors {
		mid := (s.RangeMin + s.RangeMax) / 2
		a.baselines[s.ID] = mid
	}
	ticker := time.NewTicker(a.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, s := range a.Sensors {
				base := a.baselines[s.ID]
				spread := (s.RangeMax - s.RangeMin) * 0.02
				if spread == 0 {
					spread = 0.5
				}
				delta := rng.NormFloat64() * spread
				// 1.5% chance of a noise spike to give the Anomaly Engine something real to catch.
				if rng.Float64() < 0.015 {
					delta *= 8
				}
				next := base + delta
				next = math.Max(s.RangeMin, math.Min(s.RangeMax, next))
				a.baselines[s.ID] = next
				select {
				case out <- model.Reading{
					SensorID:  s.ID,
					EntityID:  s.EntityID,
					TwinID:    s.TwinID,
					Timestamp: time.Now().UTC(),
					Value:     round2(next),
					Unit:      s.Unit,
					Quality:   model.QualityGood,
					Source:    "simulated",
				}:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
