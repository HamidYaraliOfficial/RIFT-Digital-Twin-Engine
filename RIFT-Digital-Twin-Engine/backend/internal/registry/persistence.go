package registry

import (
	"encoding/json"
	"os"
	"path/filepath"

	"rift/internal/model"
)

// diskImage is the full serialized state of the Registry, used for the
// on-disk persistence layer (data/rift-state.json by default) and for
// full Export/Import.
type diskImage struct {
	Twins     map[string]*Twin                    `json:"twins"`
	Entities  map[string]*model.Entity             `json:"entities"`
	Relations map[string]*model.Relationship       `json:"relations"`
	Sensors   map[string]*model.Sensor             `json:"sensors"`
	Hours     map[string]*model.OperatingHours     `json:"hours"`
}

// SaveToFile persists the entire registry to a JSON file (atomic via temp+rename).
func (r *Registry) SaveToFile(path string) error {
	r.mu.RLock()
	img := diskImage{
		Twins:     r.twins,
		Entities:  r.entities,
		Relations: r.relations,
		Sensors:   r.sensors,
		Hours:     r.hours,
	}
	data, err := json.MarshalIndent(img, "", "  ")
	r.mu.RUnlock()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// LoadFromFile replaces the in-memory graph with what is stored on disk.
// It is safe to call on an empty/missing file (a fresh install).
func (r *Registry) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var img diskImage
	if err := json.Unmarshal(data, &img); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if img.Twins != nil {
		r.twins = img.Twins
	}
	if img.Entities != nil {
		r.entities = img.Entities
	}
	if img.Relations != nil {
		r.relations = img.Relations
	}
	if img.Sensors != nil {
		r.sensors = img.Sensors
	}
	if img.Hours != nil {
		r.hours = img.Hours
	}
	r.byTwinEntities = make(map[string]map[string]bool)
	for id, e := range r.entities {
		if r.byTwinEntities[e.TwinID] == nil {
			r.byTwinEntities[e.TwinID] = make(map[string]bool)
		}
		r.byTwinEntities[e.TwinID][id] = true
	}
	for tid := range r.twins {
		if r.byTwinEntities[tid] == nil {
			r.byTwinEntities[tid] = make(map[string]bool)
		}
	}
	return nil
}

// ExportTwinJSON returns a standalone JSON document for a single twin,
// suitable for the Twin Import Engine on another RIFT instance.
func (r *Registry) ExportTwinJSON(twinID string) ([]byte, error) {
	snap := r.Snapshot(twinID)
	r.mu.RLock()
	twin := r.twins[twinID]
	r.mu.RUnlock()
	out := struct {
		Twin     *Twin           `json:"twin"`
		Snapshot *model.Snapshot `json:"snapshot"`
	}{Twin: twin, Snapshot: snap}
	return json.MarshalIndent(out, "", "  ")
}
