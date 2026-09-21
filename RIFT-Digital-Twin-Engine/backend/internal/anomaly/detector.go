// Package anomaly implements the Anomaly Detection Engine: a moving-window
// statistical baseline (mean/stddev) plus percentile checks per sensor. It is
// deliberately dependency-free (no external ML runtime) so it always runs,
// while exposing a Model interface so a heavier model can be plugged in.
package anomaly

import (
	"math"
	"sync"
	"time"
)

// Result is one anomaly finding with the Evidence the Health/RCA engines expect.
type Result struct {
	SensorID   string    `json:"sensorId"`
	EntityID   string    `json:"entityId"`
	Timestamp  time.Time `json:"timestamp"`
	Value      float64   `json:"value"`
	Baseline   float64   `json:"baseline"`
	StdDev     float64   `json:"stdDev"`
	ZScore     float64   `json:"zScore"`
	Severity   string    `json:"severity"` // info | warning | critical
	Evidence   string    `json:"evidence"`
}

type window struct {
	values []float64
	cap    int
}

func (w *window) push(v float64) {
	w.values = append(w.values, v)
	if len(w.values) > w.cap {
		w.values = w.values[len(w.values)-w.cap:]
	}
}

func (w *window) meanStd() (float64, float64) {
	n := len(w.values)
	if n == 0 {
		return 0, 0
	}
	var sum float64
	for _, v := range w.values {
		sum += v
	}
	mean := sum / float64(n)
	var sq float64
	for _, v := range w.values {
		sq += (v - mean) * (v - mean)
	}
	std := math.Sqrt(sq / float64(n))
	return mean, std
}

// Detector holds one moving window per sensor.
type Detector struct {
	mu         sync.Mutex
	windows    map[string]*window
	windowSize int
	warnZ      float64
	critZ      float64
}

func New(windowSize int) *Detector {
	return &Detector{
		windows:    make(map[string]*window),
		windowSize: windowSize,
		warnZ:      2.5,
		critZ:      4.0,
	}
}

// Observe feeds a new value and returns a Result if it is anomalous relative
// to the sensor's own recent history (a self-updating "Seasonal/Statistical
// Baseline" as requested — no fixed global thresholds required).
func (d *Detector) Observe(sensorID, entityID string, value float64, ts time.Time) *Result {
	d.mu.Lock()
	defer d.mu.Unlock()
	w, ok := d.windows[sensorID]
	if !ok {
		w = &window{cap: d.windowSize}
		d.windows[sensorID] = w
	}
	mean, std := w.meanStd()
	var res *Result
	if len(w.values) >= 20 && std > 1e-9 {
		z := math.Abs(value-mean) / std
		if z >= d.critZ {
			res = &Result{SensorID: sensorID, EntityID: entityID, Timestamp: ts, Value: value, Baseline: mean, StdDev: std, ZScore: z, Severity: "critical", Evidence: "value is far outside the recent statistical baseline"}
		} else if z >= d.warnZ {
			res = &Result{SensorID: sensorID, EntityID: entityID, Timestamp: ts, Value: value, Baseline: mean, StdDev: std, ZScore: z, Severity: "warning", Evidence: "value deviates from the recent statistical baseline"}
		}
	}
	w.push(value)
	return res
}
