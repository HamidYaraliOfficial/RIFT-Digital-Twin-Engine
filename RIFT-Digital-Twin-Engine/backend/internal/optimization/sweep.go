package optimization

import "rift/internal/model"

// RunSweep evaluates fn at each value in `values`, implementing the
// Parameter Sweep Engine (e.g. sweeping cooling capacity, production rate,
// traffic density, or failure rate across a defined range).
func RunSweep(parameter string, values []float64, fn func(v float64) float64) []model.SweepResult {
	out := make([]model.SweepResult, 0, len(values))
	for _, v := range values {
		out = append(out, model.SweepResult{Parameter: parameter, Value: v, Metric: round2(fn(v))})
	}
	return out
}
