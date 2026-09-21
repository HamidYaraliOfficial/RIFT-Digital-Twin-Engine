// Package simulation implements the Time Engine, the What-If Scenario
// Engine, the Fault Injection Engine and the Distributed Simulation
// Coordinator's single-process reference implementation (goroutine-per-
// scenario, safe for many concurrent branches on one machine; the same
// RunScenario call is what a real distributed coordinator would dispatch to
// remote workers).
package simulation

import (
	"sync"
	"time"
)

// Clock decouples simulated time from wall-clock time for one twin. It
// supports Pause/Resume/Fast-Forward/Slow-Motion/Step/Seek exactly as
// requested: `Speed` is a multiplier applied to real elapsed time.
type Clock struct {
	mu         sync.Mutex
	running    bool
	speed      float64
	simTime    time.Time
	realAnchor time.Time
	ticks      uint64
}

func NewClock(start time.Time) *Clock {
	return &Clock{simTime: start, realAnchor: time.Now(), speed: 1.0}
}

func (c *Clock) advanceLocked() {
	if !c.running {
		return
	}
	now := time.Now()
	elapsed := now.Sub(c.realAnchor)
	c.simTime = c.simTime.Add(time.Duration(float64(elapsed) * c.speed))
	c.realAnchor = now
}

func (c *Clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.advanceLocked()
	return c.simTime
}

func (c *Clock) Resume() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.running = true
	c.realAnchor = time.Now()
}

func (c *Clock) Pause() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.advanceLocked()
	c.running = false
}

func (c *Clock) SetSpeed(speed float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.advanceLocked()
	c.speed = speed
}

// Step advances sim time by a fixed duration regardless of running state,
// implementing "Step-by-Step" execution.
func (c *Clock) Step(d time.Duration) time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ticks++
	c.simTime = c.simTime.Add(d)
	c.realAnchor = time.Now()
	return c.simTime
}

// Seek jumps sim time directly to t ("Replay"/"Seek").
func (c *Clock) Seek(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.simTime = t
	c.realAnchor = time.Now()
}

type ClockState struct {
	Running bool      `json:"running"`
	Speed   float64   `json:"speed"`
	SimTime time.Time `json:"simTime"`
	Ticks   uint64    `json:"ticks"`
}

func (c *Clock) State() ClockState {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.advanceLocked()
	return ClockState{Running: c.running, Speed: c.speed, SimTime: c.simTime, Ticks: c.ticks}
}
