// Package audit implements the Audit Trail: an append-only record of every
// change to a Twin, Rule, Sensor, Scenario, Configuration, Permission or
// Actuator Command. Entries are kept in memory and appended to a JSON-lines
// file on disk so the trail survives restarts without requiring a database.
package audit

import (
	"bytes"
	"encoding/json"
	"os"
	"sync"
	"time"

	"rift/internal/registry"
)

type Entry struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Actor     string                 `json:"actor"` // user id or "system"
	Action    string                 `json:"action"`
	Target    string                 `json:"target"` // e.g. "entity:ent_123"
	TwinID    string                 `json:"twinId,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

type Log struct {
	mu      sync.Mutex
	path    string
	entries []Entry
	cap     int
}

func New(path string, capacity int) *Log {
	l := &Log{path: path, cap: capacity}
	l.load()
	return l
}

func (l *Log) load() {
	data, err := os.ReadFile(l.path)
	if err != nil {
		return
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	for {
		var e Entry
		if err := dec.Decode(&e); err != nil {
			break
		}
		l.entries = append(l.entries, e)
	}
	if len(l.entries) > l.cap {
		l.entries = l.entries[len(l.entries)-l.cap:]
	}
}

func (l *Log) Record(actor, action, target, twinID string, details map[string]interface{}) Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	e := Entry{ID: registry.NewID("audit"), Timestamp: time.Now().UTC(), Actor: actor, Action: action, Target: target, TwinID: twinID, Details: details}
	l.entries = append(l.entries, e)
	if len(l.entries) > l.cap {
		l.entries = l.entries[len(l.entries)-l.cap:]
	}
	l.appendToDisk(e)
	return e
}

func (l *Log) appendToDisk(e Entry) {
	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	data, _ := json.Marshal(e)
	f.Write(append(data, '\n'))
}

func (l *Log) Recent(twinID string, n int) []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Entry, 0, n)
	for i := len(l.entries) - 1; i >= 0 && len(out) < n; i-- {
		if twinID == "" || l.entries[i].TwinID == twinID {
			out = append(out, l.entries[i])
		}
	}
	return out
}
