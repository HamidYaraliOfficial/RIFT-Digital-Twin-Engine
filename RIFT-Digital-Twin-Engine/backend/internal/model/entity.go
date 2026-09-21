// Package model defines the Universal Twin Model shared by every engine in RIFT.
package model

import "time"

// EntityType is an open string type: RIFT does not restrict what kinds of
// physical or cyber-physical things can be modeled. A set of well known
// constants is provided for convenience and for the bundled templates.
type EntityType string

const (
	EntityCampus     EntityType = "campus"
	EntityBuilding   EntityType = "building"
	EntityFloor      EntityType = "floor"
	EntityRoom       EntityType = "room"
	EntityMachine    EntityType = "machine"
	EntityLine       EntityType = "production_line"
	EntityServer     EntityType = "server"
	EntityRack       EntityType = "rack"
	EntityVehicle    EntityType = "vehicle"
	EntityWarehouse  EntityType = "warehouse"
	EntityShelf      EntityType = "shelf"
	EntityRoad       EntityType = "road"
	EntityGrid       EntityType = "power_grid"
	EntityHVAC       EntityType = "hvac"
	EntityPipeline   EntityType = "pipeline"
	EntityWaterSys   EntityType = "water_system"
	EntityEmployee   EntityType = "employee"
	EntityRobot      EntityType = "robot"
	EntityDevice     EntityType = "device"
	EntityNetwork    EntityType = "network"
	EntityCity       EntityType = "city"
	EntityDistrict   EntityType = "district"
	EntityDataCenter EntityType = "data_center"
	EntityUPS        EntityType = "ups"
	EntityPDU        EntityType = "pdu"
	EntityCustom     EntityType = "custom"
)

// Status is the operational state used by the State Machine Engine.
type Status string

const (
	StatusOperational Status = "operational"
	StatusWarning     Status = "warning"
	StatusDegraded    Status = "degraded"
	StatusMaintenance Status = "maintenance"
	StatusFailed      Status = "failed"
	StatusRecovered   Status = "recovered"
	StatusUnknown     Status = "unknown"
)

// Vector3 is used for Position, Orientation (Euler degrees) and Dimensions.
type Vector3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// HealthScore is produced by the Health Score Engine.
type HealthScore struct {
	Health      float64   `json:"health"`      // 0-100, higher is better
	Risk        float64   `json:"risk"`        // 0-100, higher is worse
	Criticality float64   `json:"criticality"` // 0-100, business impact weight
	UpdatedAt   time.Time `json:"updatedAt"`
}

// RelationshipType enumerates the standard edge semantics of the Dependency Graph.
type RelationshipType string

const (
	RelContains        RelationshipType = "contains"
	RelConnectedTo     RelationshipType = "connected_to"
	RelPoweredBy       RelationshipType = "powered_by"
	RelDependsOn       RelationshipType = "depends_on"
	RelLocatedIn       RelationshipType = "located_in"
	RelControlledBy    RelationshipType = "controlled_by"
	RelCommunicatesWith RelationshipType = "communicates_with"
	RelFeeds           RelationshipType = "feeds"
	RelProduces        RelationshipType = "produces"
	RelConsumes        RelationshipType = "consumes"
	RelTransports      RelationshipType = "transports"
	RelMonitors        RelationshipType = "monitors"
)

// RelationshipChange records a mutation to a Relationship for history/audit.
type RelationshipChange struct {
	Timestamp time.Time              `json:"timestamp"`
	Field     string                 `json:"field"`
	OldValue  interface{}            `json:"oldValue,omitempty"`
	NewValue  interface{}            `json:"newValue,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Relationship is a typed, weighted, auditable edge between two Entities.
type Relationship struct {
	ID         string                 `json:"id"`
	SourceID   string                 `json:"sourceId"`
	TargetID   string                 `json:"targetId"`
	Type       RelationshipType       `json:"type"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Confidence float64                `json:"confidence"`
	Timestamp  time.Time              `json:"timestamp"`
	History    []RelationshipChange   `json:"history,omitempty"`
}

// Entity is the single Universal Twin Model node used for every physical or
// cyber-physical object RIFT can represent: sensors, actuators, machines,
// rooms, buildings, floors, factory lines, servers, racks, vehicles,
// warehouses, roads, power grid segments, HVAC units, pipelines, water
// systems, employees, robots, devices, network nodes and custom resources.
type Entity struct {
	ID          string                 `json:"id"`
	TwinID      string                 `json:"twinId"`
	Type        EntityType             `json:"type"`
	Name        string                 `json:"name"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	State       map[string]interface{} `json:"state,omitempty"`
	Position    Vector3                `json:"position"`
	Orientation Vector3                `json:"orientation"`
	Dimensions  Vector3                `json:"dimensions"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Owner       string                 `json:"owner"` // organization/workspace id
	ParentID    string                 `json:"parentId,omitempty"`
	Children    []string               `json:"children,omitempty"`
	Health      HealthScore            `json:"health"`
	Status      Status                 `json:"status"`
	Version     int                    `json:"version"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

// Clone returns a deep-enough copy of the entity suitable for simulation
// branches: maps and slices are copied so mutating the clone never touches
// the live twin.
func (e *Entity) Clone() *Entity {
	c := *e
	c.Properties = cloneMap(e.Properties)
	c.State = cloneMap(e.State)
	c.Metadata = cloneMap(e.Metadata)
	c.Children = append([]string(nil), e.Children...)
	return &c
}

func cloneMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
