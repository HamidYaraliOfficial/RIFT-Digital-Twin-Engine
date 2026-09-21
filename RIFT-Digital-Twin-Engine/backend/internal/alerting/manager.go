// Package alerting implements the Alerting Engine: severity, deduplication,
// cooldown, escalation, acknowledgement, resolution, and outbound
// notification via webhooks (and any other adapter implementing Notifier).
package alerting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"rift/internal/model"
	"rift/internal/registry"
)

// Notifier delivers an already-created Alert to an external channel.
// WebhookNotifier below is the reference implementation; Email-compatible or
// chat-ops adapters implement the same one-method interface.
type Notifier interface {
	Notify(alert model.Alert) error
}

type WebhookNotifier struct {
	URL    string
	Client *http.Client
}

func (w WebhookNotifier) Notify(alert model.Alert) error {
	if w.URL == "" {
		return nil
	}
	body, _ := json.Marshal(alert)
	client := w.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	resp, err := client.Post(w.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}

type Manager struct {
	mu        sync.RWMutex
	alerts    map[string]*model.Alert
	cooldown  time.Duration
	notifiers []Notifier
}

func NewManager(cooldown time.Duration, notifiers ...Notifier) *Manager {
	return &Manager{alerts: make(map[string]*model.Alert), cooldown: cooldown, notifiers: notifiers}
}

// Raise creates a new alert or, if an open alert with the same DedupKey
// exists within the cooldown window, increments its occurrence count instead
// of spamming a duplicate — this is the Deduplication + Cooldown requirement.
func (m *Manager) Raise(a model.Alert) *model.Alert {
	m.mu.Lock()
	now := time.Now().UTC()
	for _, existing := range m.alerts {
		if existing.DedupKey == a.DedupKey && existing.Status != model.AlertResolved {
			existing.OccurrenceCount++
			existing.LastOccurredAt = now
			if now.Sub(existing.CreatedAt) > m.cooldown {
				existing.EscalationLevel++
				existing.CreatedAt = now // restart cooldown window after escalating
			}
			m.mu.Unlock()
			return existing
		}
	}
	a.ID = registry.NewID("alert")
	a.CreatedAt, a.LastOccurredAt = now, now
	a.OccurrenceCount = 1
	a.Status = model.AlertOpen
	m.alerts[a.ID] = &a
	m.mu.Unlock()

	for _, n := range m.notifiers {
		go func(n Notifier, alert model.Alert) { _ = n.Notify(alert) }(n, a)
	}
	return &a
}

func (m *Manager) Acknowledge(id, by string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.alerts[id]
	if !ok {
		return fmt.Errorf("alert %s not found", id)
	}
	now := time.Now().UTC()
	a.Status = model.AlertAcknowledged
	a.AckBy = by
	a.AckAt = &now
	return nil
}

func (m *Manager) Resolve(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.alerts[id]
	if !ok {
		return fmt.Errorf("alert %s not found", id)
	}
	now := time.Now().UTC()
	a.Status = model.AlertResolved
	a.ResolvedAt = &now
	return nil
}

func (m *Manager) List(twinID string, includeResolved bool) []*model.Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*model.Alert, 0)
	for _, a := range m.alerts {
		if a.TwinID != twinID {
			continue
		}
		if !includeResolved && a.Status == model.AlertResolved {
			continue
		}
		out = append(out, a)
	}
	return out
}
