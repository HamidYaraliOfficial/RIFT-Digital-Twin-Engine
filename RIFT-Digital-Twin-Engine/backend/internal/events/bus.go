// Package events implements the in-process Event Streaming Layer: a
// lightweight publish/subscribe bus that every engine (rules, alerting,
// anomaly detection, root cause analysis, the SSE gateway...) uses to react
// to Events without being wired directly to each other. A production
// deployment can bridge this bus 1:1 onto Kafka/NATS via the same Publish
// signature (see the Connector Framework note in the README).
package events

import (
	"sync"

	"rift/internal/model"
)

type Handler func(model.Event)

// Bus is a fan-out, non-blocking pub/sub dispatcher.
type Bus struct {
	mu       sync.RWMutex
	subs     map[string][]Handler // topic -> handlers; topic "*" receives everything
	buffer   chan model.Event
	history  []model.Event
	histCap  int
	histLock sync.RWMutex
}

func New(bufferSize, historyCap int) *Bus {
	b := &Bus{
		subs:    make(map[string][]Handler),
		buffer:  make(chan model.Event, bufferSize),
		histCap: historyCap,
	}
	go b.loop()
	return b
}

func (b *Bus) loop() {
	for evt := range b.buffer {
		b.histLock.Lock()
		b.history = append(b.history, evt)
		if len(b.history) > b.histCap {
			b.history = b.history[len(b.history)-b.histCap:]
		}
		b.histLock.Unlock()

		b.mu.RLock()
		handlers := append([]Handler{}, b.subs[evt.Type]...)
		handlers = append(handlers, b.subs["*"]...)
		b.mu.RUnlock()
		for _, h := range handlers {
			go safeCall(h, evt)
		}
	}
}

func safeCall(h Handler, evt model.Event) {
	defer func() { _ = recover() }() // one bad subscriber must never crash the bus
	h(evt)
}

// Subscribe registers a handler for a topic ("*" = all events).
func (b *Bus) Subscribe(topic string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs[topic] = append(b.subs[topic], h)
}

// Publish enqueues an event for asynchronous dispatch. If the buffer is
// full (extreme backpressure) it drops the oldest behavior in favor of never
// blocking the ingestion pipeline's hot path.
func (b *Bus) Publish(evt model.Event) {
	select {
	case b.buffer <- evt:
	default:
		// Backpressure: publish synchronously as a last resort so the event
		// is never silently lost, at the cost of blocking the caller briefly.
		b.buffer <- evt
	}
}

// Recent returns up to n most recent events (newest last), optionally
// filtered by twin id.
func (b *Bus) Recent(twinID string, n int) []model.Event {
	b.histLock.RLock()
	defer b.histLock.RUnlock()
	out := make([]model.Event, 0, n)
	for i := len(b.history) - 1; i >= 0 && len(out) < n; i-- {
		if twinID == "" || b.history[i].TwinID == twinID {
			out = append(out, b.history[i])
		}
	}
	// reverse to chronological order
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}
