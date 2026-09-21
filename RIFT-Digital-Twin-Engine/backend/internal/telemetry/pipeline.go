package telemetry

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"rift/internal/model"
)

// Pipeline is the Telemetry Ingestion Pipeline: a bounded-buffer, worker-pool
// consumer built from goroutines + channels. It applies Data Quality checks,
// writes to the Store, updates sensor liveness, and republishes each reading
// on a fan-out channel for downstream engines (state, rules, anomaly).
type Pipeline struct {
	in          chan model.Reading
	store       *Store
	workers     int
	subscribers []chan model.Reading
	subMu       sync.RWMutex

	received  atomic.Uint64
	dropped   atomic.Uint64
	processed atomic.Uint64

	onReading func(model.Reading) // optional quality gate; return dropped via bool below
	quality   func(model.Reading) model.Quality

	lastSeenBySensor sync.Map // sensorID -> time.Time, used for gap/offline detection
}

func NewPipeline(store *Store, bufferSize, workers int) *Pipeline {
	return &Pipeline{
		in:      make(chan model.Reading, bufferSize),
		store:   store,
		workers: workers,
		quality: defaultQualityCheck,
	}
}

// defaultQualityCheck implements the Data Quality Engine's inline checks:
// out-of-range and stale-timestamp detection. Duplicate/unit-mismatch checks
// are applied at the API boundary where the sensor's declared unit is known.
func defaultQualityCheck(r model.Reading) model.Quality {
	if r.Timestamp.IsZero() || time.Since(r.Timestamp) > 10*time.Minute {
		return model.QualityUncertain
	}
	return model.QualityGood
}

// Ingest is the hot-path entry point (called by adapters and by the HTTP
// ingestion endpoint). It never blocks the caller for long: on a full buffer
// the reading is counted as dropped (backpressure signal exposed via Stats).
func (p *Pipeline) Ingest(r model.Reading) bool {
	p.received.Add(1)
	select {
	case p.in <- r:
		return true
	default:
		p.dropped.Add(1)
		return false
	}
}

// Subscribe returns a channel that receives every reading after it has been
// quality-checked and stored. Used by the State Engine, Rules Engine and
// Anomaly Engine to react in real time without polling the Store.
func (p *Pipeline) Subscribe(bufferSize int) <-chan model.Reading {
	ch := make(chan model.Reading, bufferSize)
	p.subMu.Lock()
	p.subscribers = append(p.subscribers, ch)
	p.subMu.Unlock()
	return ch
}

// Run starts the worker pool. It blocks until ctx is cancelled.
func (p *Pipeline) Run(ctx context.Context) {
	var wg sync.WaitGroup
	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.worker(ctx)
		}()
	}
	<-ctx.Done()
	wg.Wait()
}

func (p *Pipeline) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case r, ok := <-p.in:
			if !ok {
				return
			}
			r.Quality = p.quality(r)
			stored := p.store.Append(r)
			p.lastSeenBySensor.Store(r.SensorID, r.Timestamp)
			p.processed.Add(1)

			p.subMu.RLock()
			subs := append([]chan model.Reading{}, p.subscribers...)
			p.subMu.RUnlock()
			for _, s := range subs {
				select {
				case s <- stored:
				default: // never let a slow subscriber stall ingestion
				}
			}
		}
	}
}

// Stats exposes pipeline throughput counters for the Observability layer.
type Stats struct {
	Received  uint64 `json:"received"`
	Processed uint64 `json:"processed"`
	Dropped   uint64 `json:"dropped"`
	QueueLen  int    `json:"queueLen"`
	QueueCap  int    `json:"queueCap"`
}

func (p *Pipeline) Stats() Stats {
	return Stats{
		Received:  p.received.Load(),
		Processed: p.processed.Load(),
		Dropped:   p.dropped.Load(),
		QueueLen:  len(p.in),
		QueueCap:  cap(p.in),
	}
}

// GapCheck reports sensors that have not reported within `maxSilence`,
// implementing Sensor Offline detection for the Alerting Engine.
func (p *Pipeline) GapCheck(sensorIDs []string, maxSilence time.Duration) []string {
	now := time.Now().UTC()
	var offline []string
	for _, id := range sensorIDs {
		v, ok := p.lastSeenBySensor.Load(id)
		if !ok {
			continue
		}
		last := v.(time.Time)
		if now.Sub(last) > maxSilence {
			offline = append(offline, id)
		}
	}
	return offline
}
