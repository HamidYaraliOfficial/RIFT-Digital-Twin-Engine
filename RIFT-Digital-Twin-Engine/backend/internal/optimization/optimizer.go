package optimization

import "math/rand"

// Objective is one term of a combinable multi-objective function: Minimize
// Cost + Minimize Failure Risk + Maximize Throughput + Minimize Energy, etc.
// Direction is +1 to maximize, -1 to minimize; Weight lets objectives be
// combined with different importance.
type Objective struct {
	Name      string
	Weight    float64
	Direction float64 // +1 maximize, -1 minimize
	Evaluate  func(params map[string]float64) float64
}

// Bounds constrains one tunable parameter (Maintenance Schedule offset,
// Cooling Setpoint, Resource Allocation share, Vehicle Routing weight, ...).
type Bounds struct {
	Min float64
	Max float64
}

// Result is the best configuration the Optimization Engine found.
type Result struct {
	Params map[string]float64 `json:"params"`
	Score  float64            `json:"score"`
	Trials int                `json:"trials"`
}

// Optimize performs randomized local search (random restarts + greedy hill
// climbing): a dependency-free, always-available optimizer suitable for the
// noisy, black-box objectives simulation-derived metrics produce. `trials`
// controls the search budget.
func Optimize(objectives []Objective, bounds map[string]Bounds, trials int, seed int64) Result {
	rng := rand.New(rand.NewSource(seed))
	score := func(p map[string]float64) float64 {
		total := 0.0
		for _, o := range objectives {
			total += o.Weight * o.Direction * o.Evaluate(p)
		}
		return total
	}

	randomPoint := func() map[string]float64 {
		p := make(map[string]float64, len(bounds))
		for name, b := range bounds {
			p[name] = b.Min + rng.Float64()*(b.Max-b.Min)
		}
		return p
	}

	best := randomPoint()
	bestScore := score(best)
	for t := 1; t < trials; t++ {
		var candidate map[string]float64
		if t%5 == 0 {
			candidate = randomPoint() // random restart to escape local optima
		} else {
			candidate = perturb(best, bounds, rng)
		}
		s := score(candidate)
		if s > bestScore {
			best, bestScore = candidate, s
		}
	}
	return Result{Params: best, Score: round2(bestScore), Trials: trials}
}

func perturb(p map[string]float64, bounds map[string]Bounds, rng *rand.Rand) map[string]float64 {
	out := make(map[string]float64, len(p))
	for name, v := range p {
		b := bounds[name]
		step := (b.Max - b.Min) * 0.1
		nv := v + (rng.Float64()*2-1)*step
		if nv < b.Min {
			nv = b.Min
		}
		if nv > b.Max {
			nv = b.Max
		}
		out[name] = nv
	}
	return out
}
