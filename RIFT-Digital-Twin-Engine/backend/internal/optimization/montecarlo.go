// Package optimization implements the Monte Carlo Engine, the Parameter
// Sweep Engine and a general-purpose Optimization Engine (weighted,
// combinable objectives via randomized local search) used for maintenance
// scheduling, energy, production and routing configuration problems.
package optimization

import (
	"math"
	"math/rand"
	"sort"

	"rift/internal/model"
)

// MetricFunc runs one experiment for a given seed/magnitude and returns the
// scalar outcome to aggregate (e.g. "% throughput lost", "entities affected").
type MetricFunc func(seed int64, magnitude float64) float64

// RunMonteCarlo executes `runs` independent randomized trials (random seed
// each time, magnitude drawn uniformly from [magMin, magMax]) and returns the
// full statistical summary: mean, p10/p50/p90, best/worst case.
func RunMonteCarlo(metric string, runs int, magMin, magMax float64, fn MetricFunc) model.MonteCarloResult {
	samples := make([]float64, 0, runs)
	rng := rand.New(rand.NewSource(42)) // fixed master seed so the *set* of trial seeds is reproducible
	for i := 0; i < runs; i++ {
		seed := rng.Int63()
		mag := magMin + rng.Float64()*(magMax-magMin)
		samples = append(samples, fn(seed, mag))
	}
	sorted := append([]float64(nil), samples...)
	sort.Float64s(sorted)

	return model.MonteCarloResult{
		Runs:    runs,
		Metric:  metric,
		Mean:    mean(sorted),
		P10:     percentile(sorted, 10),
		P50:     percentile(sorted, 50),
		P90:     percentile(sorted, 90),
		Best:    sorted[0],
		Worst:   sorted[len(sorted)-1],
		Samples: samples,
	}
}

func mean(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	var s float64
	for _, x := range v {
		s += x
	}
	return round2(s / float64(len(v)))
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(p/100*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return round2(sorted[idx])
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }
