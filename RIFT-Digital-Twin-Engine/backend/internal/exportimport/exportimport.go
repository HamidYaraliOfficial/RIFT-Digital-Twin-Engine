// Package exportimport implements the Twin Import Engine and the Report/
// Export side of the Twin Snapshot System: JSON and CSV in, JSON and CSV
// out, with schema validation on the way in.
package exportimport

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"rift/internal/model"
	"rift/internal/registry"
)

// EntityImportRow is the minimal schema the CSV/JSON importer accepts. Extra
// columns are preserved into Properties automatically.
type EntityImportRow struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	ParentName string                 `json:"parent,omitempty"`
	X          float64                `json:"x"`
	Y          float64                `json:"y"`
	Z          float64                `json:"z"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// ValidationError describes exactly one problem found in imported data so
// the Twin Builder UI can point the user at the offending row.
type ValidationError struct {
	Row     int    `json:"row"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ImportJSON validates and imports a slice of EntityImportRow into twinID,
// resolving ParentName references to already-imported (or existing) entities
// by name. Returns created entities and any validation errors found
// (partial success: valid rows are still imported).
func ImportJSON(reg *registry.Registry, twinID string, rows []EntityImportRow) ([]*model.Entity, []ValidationError) {
	var errs []ValidationError
	var created []*model.Entity
	nameToID := map[string]string{}
	for _, e := range reg.ListEntities(twinID) {
		nameToID[e.Name] = e.ID
	}
	for i, row := range rows {
		if strings.TrimSpace(row.Name) == "" {
			errs = append(errs, ValidationError{Row: i, Field: "name", Message: "name is required"})
			continue
		}
		if strings.TrimSpace(row.Type) == "" {
			errs = append(errs, ValidationError{Row: i, Field: "type", Message: "type is required"})
			continue
		}
		parentID := ""
		if row.ParentName != "" {
			pid, ok := nameToID[row.ParentName]
			if !ok {
				errs = append(errs, ValidationError{Row: i, Field: "parent", Message: fmt.Sprintf("parent %q not found (import parents before children)", row.ParentName)})
				continue
			}
			parentID = pid
		}
		e := reg.CreateEntity(&model.Entity{
			TwinID: twinID, Name: row.Name, Type: model.EntityType(row.Type), ParentID: parentID,
			Position: model.Vector3{X: row.X, Y: row.Y, Z: row.Z}, Properties: row.Properties,
		})
		nameToID[e.Name] = e.ID
		created = append(created, e)
	}
	return created, errs
}

// ImportCSV parses a CSV with header columns: name,type,parent,x,y,z and any
// number of extra numeric-or-text columns, which become Properties.
func ImportCSV(reg *registry.Registry, twinID string, csvText string) ([]*model.Entity, []ValidationError) {
	r := csv.NewReader(strings.NewReader(csvText))
	records, err := r.ReadAll()
	if err != nil || len(records) < 1 {
		return nil, []ValidationError{{Row: 0, Field: "file", Message: "could not parse CSV"}}
	}
	header := records[0]
	colIdx := map[string]int{}
	for i, h := range header {
		colIdx[strings.ToLower(strings.TrimSpace(h))] = i
	}
	rows := make([]EntityImportRow, 0, len(records)-1)
	for _, rec := range records[1:] {
		row := EntityImportRow{Properties: map[string]interface{}{}}
		for col, idx := range colIdx {
			if idx >= len(rec) {
				continue
			}
			val := rec[idx]
			switch col {
			case "name":
				row.Name = val
			case "type":
				row.Type = val
			case "parent":
				row.ParentName = val
			case "x":
				row.X, _ = strconv.ParseFloat(val, 64)
			case "y":
				row.Y, _ = strconv.ParseFloat(val, 64)
			case "z":
				row.Z, _ = strconv.ParseFloat(val, 64)
			default:
				if f, err := strconv.ParseFloat(val, 64); err == nil {
					row.Properties[col] = f
				} else if val != "" {
					row.Properties[col] = val
				}
			}
		}
		rows = append(rows, row)
	}
	return ImportJSON(reg, twinID, rows)
}

// ExportCSV writes every entity of a twin as CSV (id,name,type,parent,status,x,y,z).
func ExportCSV(reg *registry.Registry, twinID string) (string, error) {
	var sb strings.Builder
	w := csv.NewWriter(&sb)
	if err := w.Write([]string{"id", "name", "type", "parent", "status", "x", "y", "z", "updatedAt"}); err != nil {
		return "", err
	}
	for _, e := range reg.ListEntities(twinID) {
		row := []string{
			e.ID, e.Name, string(e.Type), e.ParentID, string(e.Status),
			fmt.Sprintf("%.3f", e.Position.X), fmt.Sprintf("%.3f", e.Position.Y), fmt.Sprintf("%.3f", e.Position.Z),
			e.UpdatedAt.Format(time.RFC3339),
		}
		if err := w.Write(row); err != nil {
			return "", err
		}
	}
	w.Flush()
	return sb.String(), w.Error()
}

// ExportJSON returns the full twin snapshot as pretty JSON.
func ExportJSON(reg *registry.Registry, twinID string) ([]byte, error) {
	return reg.ExportTwinJSON(twinID)
}

var _ = json.Marshal // keep encoding/json import used even if helpers above change
