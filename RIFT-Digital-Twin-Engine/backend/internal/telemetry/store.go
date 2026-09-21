// Package telemetry implements the Sensor Abstraction Layer, the Telemetry
// Ingestion Pipeline and the Time-Series Data Layer's in-process tier: a
// bounded, per-sensor ring buffer with retention and simple downsampling.
//
// The RingStore interface is intentionally the same shape a TimescaleDB-backed
// store would expose, so production deployments can swap this file's
// implementation without touching the pipeline or the API layer.
package telemetry

import (
	"sort"
	"sync"
	"time"

	"rift/internal/model"
)

// Store keeps a bounded history of readings per sensor in memory. Capacity is
// per-sensor; oldest samples are evicted first (a real deployment points this
// at TimescaleDB/InfluxDB instead for unbounded retention + compression).
type Store struct {
	mu       sync.RWMutex
	capacity int
	bySensor map[string][]model.Reading
	lastSeq  map[string]uint64
}

func NewStore(capacityPerSensor int) *Store {
	return &Store{
		capacity: capacityPerSensor,
		bySensor: make(map[string][]model.Reading),
		lastSeq:  make(map[string]uint64),
	}
}

// Append inserts a reading, assigning a monotonically increasing per-sensor
// Sequence number if the caller did not already set one.
func (s *Store) Append(r model.Reading) model.Reading {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastSeq[r.SensorID]++
	if r.Sequence == 0 {
		r.Sequence = s.lastSeq[r.SensorID]
	}
	buf := s.bySensor[r.SensorID]
	buf = append(buf, r)
	if len(buf) > s.capacity {
		buf = buf[len(buf)-s.capacity:]
	}
	s.bySensor[r.SensorID] = buf
	return r
}

// Latest returns the most recent reading for a sensor, if any.
func (s *Store) Latest(sensorID string) (model.Reading, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	buf := s.bySensor[sensorID]
	if len(buf) == 0 {
		return model.Reading{}, false
	}
	return buf[len(buf)-1], true
}

// Range returns readings for a sensor within [from, to], oldest first.
func (s *Store) Range(sensorID string, from, to time.Time) []model.Reading {
	s.mu.RLock()
	defer s.mu.RUnlock()
	buf := s.bySensor[sensorID]
	out := make([]model.Reading, 0, len(buf))
	for _, r := range buf {
		if (r.Timestamp.Equal(from) || r.Timestamp.After(from)) && (r.Timestamp.Equal(to) || r.Timestamp.Before(to)) {
			out = append(out, r)
		}
	}
	return out
}

// Window returns the last `n` readings for a sensor, oldest first.
func (s *Store) Window(sensorID string, n int) []model.Reading {
	s.mu.RLock()
	defer s.mu.RUnlock()
	buf := s.bySensor[sensorID]
	if len(buf) <= n {
		out := make([]model.Reading, len(buf))
		copy(out, buf)
		return out
	}
	out := make([]model.Reading, n)
	copy(out, buf[len(buf)-n:])
	return out
}

// Downsample performs simple time-bucket averaging (mean aggregation),
// implementing the Time-Series Data Layer's downsampling/aggregation
// requirement without a heavyweight TSDB dependency.
func (s *Store) Downsample(sensorID string, bucket time.Duration) []model.Reading {
	s.mu.RLock()
	buf := append([]model.Reading(nil), s.bySensor[sensorID]...)
	s.mu.RUnlock()
	if len(buf) == 0 || bucket <= 0 {
		return buf
	}
	sort.Slice(buf, func(i, j int) bool { return buf[i].Timestamp.Before(buf[j].Timestamp) })

	type acc struct {
		sum   float64
		count int
		last  model.Reading
	}
	buckets := make(map[int64]*acc)
	var order []int64
	for _, r := range buf {
		key := r.Timestamp.Truncate(bucket).Unix()
		a, ok := buckets[key]
		if !ok {
			a = &acc{}
			buckets[key] = a
			order = append(order, key)
		}
		a.sum += r.Value
		a.count++
		a.last = r
	}
	sort.Slice(order, func(i, j int) bool { return order[i] < order[j] })
	out := make([]model.Reading, 0, len(order))
	for _, key := range order {
		a := buckets[key]
		rr := a.last
		rr.Value = a.sum / float64(a.count)
		rr.Timestamp = time.Unix(key, 0).UTC()
		out = append(out, rr)
	}
	return out
}
