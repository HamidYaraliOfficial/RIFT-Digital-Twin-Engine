// Package registry implements the Twin Registry and Twin Builder backing
// store: a thread-safe, in-memory graph of Entities, Relationships and
// Sensors, with JSON snapshot persistence to disk (see persistence.go).
//
// The store is intentionally storage-agnostic at the interface boundary
// (Save/Load work against any io target) so a production deployment can
// swap this package's internals for PostgreSQL/TimescaleDB without touching
// any engine that depends on Registry.
package registry

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"rift/internal/model"
)

// Twin is one Digital Twin instance: a named environment (a factory, a
// building, a data center...) that owns a set of Entities.
type Twin struct {
	ID          string    `json:"id"`
	OrgID       string    `json:"orgId"`
	WorkspaceID string    `json:"workspaceId"`
	Name        string    `json:"name"`
	Template    string    `json:"template"` // factory | building | data_center | warehouse | smart_city | custom
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Archived    bool      `json:"archived"`
}

// Registry is the concurrency-safe in-memory graph store.
type Registry struct {
	mu sync.RWMutex

	twins     map[string]*Twin
	entities  map[string]*model.Entity
	relations map[string]*model.Relationship
	sensors   map[string]*model.Sensor
	hours     map[string]*model.OperatingHours

	// byTwinEntities indexes entity IDs per twin for fast listing.
	byTwinEntities map[string]map[string]bool
}

func New() *Registry {
	return &Registry{
		twins:          make(map[string]*Twin),
		entities:       make(map[string]*model.Entity),
		relations:      make(map[string]*model.Relationship),
		sensors:        make(map[string]*model.Sensor),
		hours:          make(map[string]*model.OperatingHours),
		byTwinEntities: make(map[string]map[string]bool),
	}
}

// ---------- Twins ----------

func (r *Registry) CreateTwin(t *Twin) *Twin {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t.ID == "" {
		t.ID = NewID("twin")
	}
	t.CreatedAt = time.Now().UTC()
	t.UpdatedAt = t.CreatedAt
	r.twins[t.ID] = t
	r.byTwinEntities[t.ID] = make(map[string]bool)
	return t
}

func (r *Registry) GetTwin(id string) (*Twin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.twins[id]
	return t, ok
}

func (r *Registry) ListTwins() []*Twin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Twin, 0, len(r.twins))
	for _, t := range r.twins {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}

func (r *Registry) ArchiveTwin(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.twins[id]
	if !ok {
		return fmt.Errorf("twin %s not found", id)
	}
	t.Archived = true
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// CloneTwin performs a deep copy of a twin's entities, relations and sensors
// under a new twin ID. Used by Twin Registry "Clone" and as the basis for
// Scenario branching (see simulation.ScenarioEngine).
func (r *Registry) CloneTwin(sourceID, newName string) (*Twin, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	src, ok := r.twins[sourceID]
	if !ok {
		return nil, fmt.Errorf("twin %s not found", sourceID)
	}
	nt := &Twin{
		ID:          NewID("twin"),
		OrgID:       src.OrgID,
		WorkspaceID: src.WorkspaceID,
		Name:        newName,
		Template:    src.Template,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	r.twins[nt.ID] = nt
	r.byTwinEntities[nt.ID] = make(map[string]bool)

	idMap := make(map[string]string)
	for eid := range r.byTwinEntities[sourceID] {
		e := r.entities[eid].Clone()
		oldID := e.ID
		e.ID = NewID("ent")
		e.TwinID = nt.ID
		idMap[oldID] = e.ID
		r.entities[e.ID] = e
		r.byTwinEntities[nt.ID][e.ID] = true
	}
	// remap parent/children references
	for eid := range r.byTwinEntities[nt.ID] {
		e := r.entities[eid]
		if e.ParentID != "" {
			if np, ok := idMap[e.ParentID]; ok {
				e.ParentID = np
			}
		}
		newChildren := make([]string, 0, len(e.Children))
		for _, c := range e.Children {
			if nc, ok := idMap[c]; ok {
				newChildren = append(newChildren, nc)
			}
		}
		e.Children = newChildren
	}
	for _, rel := range r.relations {
		srcNew, okS := idMap[rel.SourceID]
		tgtNew, okT := idMap[rel.TargetID]
		if okS && okT {
			nr := *rel
			nr.ID = NewID("rel")
			nr.SourceID = srcNew
			nr.TargetID = tgtNew
			r.relations[nr.ID] = &nr
		}
	}
	for _, s := range r.sensors {
		if newEnt, ok := idMap[s.EntityID]; ok {
			ns := *s
			ns.ID = NewID("sen")
			ns.EntityID = newEnt
			ns.TwinID = nt.ID
			r.sensors[ns.ID] = &ns
		}
	}
	return nt, nil
}

// ---------- Entities ----------

func (r *Registry) CreateEntity(e *model.Entity) *model.Entity {
	r.mu.Lock()
	defer r.mu.Unlock()
	if e.ID == "" {
		e.ID = NewID("ent")
	}
	now := time.Now().UTC()
	e.CreatedAt, e.UpdatedAt = now, now
	if e.Status == "" {
		e.Status = model.StatusOperational
	}
	e.Version = 1
	r.entities[e.ID] = e
	if r.byTwinEntities[e.TwinID] == nil {
		r.byTwinEntities[e.TwinID] = make(map[string]bool)
	}
	r.byTwinEntities[e.TwinID][e.ID] = true
	if e.ParentID != "" {
		if p, ok := r.entities[e.ParentID]; ok {
			p.Children = append(p.Children, e.ID)
		}
	}
	return e
}

func (r *Registry) GetEntity(id string) (*model.Entity, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.entities[id]
	return e, ok
}

func (r *Registry) ListEntities(twinID string) []*model.Entity {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*model.Entity, 0)
	for id := range r.byTwinEntities[twinID] {
		out = append(out, r.entities[id])
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// UpdateEntity applies a partial patch and bumps Version/UpdatedAt.
func (r *Registry) UpdateEntity(id string, patch func(*model.Entity)) (*model.Entity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.entities[id]
	if !ok {
		return nil, fmt.Errorf("entity %s not found", id)
	}
	patch(e)
	e.Version++
	e.UpdatedAt = time.Now().UTC()
	return e, nil
}

func (r *Registry) DeleteEntity(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.entities[id]
	if !ok {
		return fmt.Errorf("entity %s not found", id)
	}
	delete(r.entities, id)
	delete(r.byTwinEntities[e.TwinID], id)
	return nil
}

// ---------- Relationships ----------

func (r *Registry) CreateRelationship(rel *model.Relationship) *model.Relationship {
	r.mu.Lock()
	defer r.mu.Unlock()
	if rel.ID == "" {
		rel.ID = NewID("rel")
	}
	rel.Timestamp = time.Now().UTC()
	r.relations[rel.ID] = rel
	return rel
}

func (r *Registry) ListRelationships(twinID string) []*model.Relationship {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := r.byTwinEntities[twinID]
	out := make([]*model.Relationship, 0)
	for _, rel := range r.relations {
		if ids[rel.SourceID] || ids[rel.TargetID] {
			out = append(out, rel)
		}
	}
	return out
}

func (r *Registry) DeleteRelationship(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.relations[id]; !ok {
		return fmt.Errorf("relationship %s not found", id)
	}
	delete(r.relations, id)
	return nil
}

// ---------- Sensors ----------

func (r *Registry) CreateSensor(s *model.Sensor) *model.Sensor {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s.ID == "" {
		s.ID = NewID("sen")
	}
	s.CreatedAt = time.Now().UTC()
	if s.Status == "" {
		s.Status = model.SensorOnline
	}
	r.sensors[s.ID] = s
	return s
}

func (r *Registry) GetSensor(id string) (*model.Sensor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sensors[id]
	return s, ok
}

func (r *Registry) ListSensors(twinID string) []*model.Sensor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*model.Sensor, 0)
	for _, s := range r.sensors {
		if s.TwinID == twinID {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (r *Registry) SensorsForEntity(entityID string) []*model.Sensor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*model.Sensor, 0)
	for _, s := range r.sensors {
		if s.EntityID == entityID {
			out = append(out, s)
		}
	}
	return out
}

func (r *Registry) TouchSensor(id string, status model.SensorStatus) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s, ok := r.sensors[id]; ok {
		s.LastSeen = time.Now().UTC()
		s.Status = status
	}
}

// ---------- Operating Hours ----------

func (r *Registry) SetOperatingHours(h *model.OperatingHours) *model.OperatingHours {
	r.mu.Lock()
	defer r.mu.Unlock()
	if h.ID == "" {
		h.ID = NewID("hrs")
	}
	h.UpdatedAt = time.Now().UTC()
	r.hours[h.ID] = h
	return h
}

func (r *Registry) ListOperatingHours(twinID string) []*model.OperatingHours {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*model.OperatingHours, 0)
	for _, h := range r.hours {
		if h.TwinID == twinID {
			out = append(out, h)
		}
	}
	return out
}

func (r *Registry) DeleteOperatingHours(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.hours[id]; !ok {
		return fmt.Errorf("operating hours %s not found", id)
	}
	delete(r.hours, id)
	return nil
}

// Snapshot captures the full graph for one twin (used by simulation branches
// and by the Twin Snapshot System).
func (r *Registry) Snapshot(twinID string) *model.Snapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()
	snap := &model.Snapshot{
		ID:        NewID("snap"),
		TwinID:    twinID,
		CreatedAt: time.Now().UTC(),
		Entities:  make(map[string]*model.Entity),
		Relations: make(map[string]*model.Relationship),
		Sensors:   make(map[string]*model.Sensor),
	}
	for id := range r.byTwinEntities[twinID] {
		snap.Entities[id] = r.entities[id].Clone()
	}
	for id, rel := range r.relations {
		if snap.Entities[rel.SourceID] != nil || snap.Entities[rel.TargetID] != nil {
			cp := *rel
			snap.Relations[id] = &cp
		}
	}
	for id, s := range r.sensors {
		if s.TwinID == twinID {
			cp := *s
			snap.Sensors[id] = &cp
		}
	}
	return snap
}
