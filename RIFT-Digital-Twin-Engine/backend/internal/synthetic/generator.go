// Package synthetic implements the Synthetic Twin Generator: it builds the
// five bundled Templates (Factory, Smart Building, Data Center, Warehouse,
// Smart City) and can also generate arbitrarily large benchmark twins with a
// fixed seed for reproducible load testing (see cmd/rift's `benchmark`
// command).
package synthetic

import (
	"fmt"
	"math/rand"

	"rift/internal/model"
	"rift/internal/registry"
)

type Builder struct {
	Reg    *registry.Registry
	TwinID string
	rng    *rand.Rand
}

func NewBuilder(reg *registry.Registry, twinID string, seed int64) *Builder {
	return &Builder{Reg: reg, TwinID: twinID, rng: rand.New(rand.NewSource(seed))}
}

func (b *Builder) entity(typ model.EntityType, name, parentID string, pos model.Vector3, props map[string]interface{}) *model.Entity {
	e := &model.Entity{
		TwinID: b.TwinID, Type: typ, Name: name, ParentID: parentID,
		Position: pos, Properties: props, Owner: b.TwinID, Status: model.StatusOperational,
		Metadata: map[string]interface{}{"generated": true},
	}
	return b.Reg.CreateEntity(e)
}

func (b *Builder) rel(t model.RelationshipType, src, dst string) {
	b.Reg.CreateRelationship(&model.Relationship{SourceID: src, TargetID: dst, Type: t, Confidence: 1.0})
}

func (b *Builder) sensor(entityID string, typ model.SensorType, unit string, min, max float64) *model.Sensor {
	return b.Reg.CreateSensor(&model.Sensor{
		TwinID: b.TwinID, EntityID: entityID, Name: fmt.Sprintf("%s sensor", typ),
		Type: typ, Unit: unit, RangeMin: min, RangeMax: max, SamplingRateMs: 2000,
		Accuracy: 0.5, Source: "simulated",
	})
}

// Factory builds Factory -> Production Line -> Machine, plus power feed and
// a warehouse-style raw-material buffer, wired with depends_on/powered_by/monitors.
func (b *Builder) Factory() {
	plant := b.entity(model.EntityBuilding, "Main Plant", "", model.Vector3{}, map[string]interface{}{"throughput": 100.0})
	grid := b.entity(model.EntityGrid, "Site Power Feed", plant.ID, model.Vector3{X: -10}, map[string]interface{}{"capacity": 500.0})
	b.rel(model.RelContains, plant.ID, grid.ID)

	for line := 1; line <= 3; line++ {
		l := b.entity(model.EntityLine, fmt.Sprintf("Production Line %d", line), plant.ID, model.Vector3{X: float64(line) * 5}, map[string]interface{}{"throughput": 30.0, "output_rate": 30.0})
		b.rel(model.RelContains, plant.ID, l.ID)
		b.rel(model.RelPoweredBy, l.ID, grid.ID)
		for m := 1; m <= 4; m++ {
			mach := b.entity(model.EntityMachine, fmt.Sprintf("Line %d Machine %d", line, m), l.ID, model.Vector3{X: float64(line) * 5, Y: float64(m) * 2}, map[string]interface{}{"capacity": 100.0})
			b.rel(model.RelContains, l.ID, mach.ID)
			b.rel(model.RelDependsOn, l.ID, mach.ID)
			b.sensor(mach.ID, model.SensorTemperature, "°C", 20, 90)
			b.sensor(mach.ID, model.SensorVibration, "mm/s", 0, 12)
			b.sensor(mach.ID, model.SensorEnergy, "kWh", 0, 50)
		}
	}
}

// Building builds a Smart Building: Building -> Floors -> Rooms with HVAC and occupancy.
func (b *Builder) Building() {
	bldg := b.entity(model.EntityBuilding, "HQ Tower", "", model.Vector3{}, nil)
	hvac := b.entity(model.EntityHVAC, "Central HVAC", bldg.ID, model.Vector3{Z: 1}, map[string]interface{}{"capacity": 100.0})
	b.rel(model.RelContains, bldg.ID, hvac.ID)
	for floor := 1; floor <= 4; floor++ {
		f := b.entity(model.EntityFloor, fmt.Sprintf("Floor %d", floor), bldg.ID, model.Vector3{Z: float64(floor) * 3}, nil)
		b.rel(model.RelContains, bldg.ID, f.ID)
		b.rel(model.RelDependsOn, f.ID, hvac.ID)
		for r := 1; r <= 3; r++ {
			room := b.entity(model.EntityRoom, fmt.Sprintf("F%d Room %d", floor, r), f.ID, model.Vector3{X: float64(r) * 4, Z: float64(floor) * 3}, nil)
			b.rel(model.RelContains, f.ID, room.ID)
			b.rel(model.RelMonitors, hvac.ID, room.ID)
			b.sensor(room.ID, model.SensorTemperature, "°C", 16, 30)
			b.sensor(room.ID, model.SensorHumidity, "%", 20, 70)
			b.sensor(room.ID, model.SensorMotion, "bool", 0, 1)
			b.sensor(room.ID, model.SensorAirQuality, "AQI", 0, 300)
		}
	}
}

// DataCenter builds Data Center -> Racks -> Servers with power/cooling dependencies.
func (b *Builder) DataCenter() {
	dc := b.entity(model.EntityDataCenter, "DC-1", "", model.Vector3{}, nil)
	ups := b.entity(model.EntityUPS, "UPS-A", dc.ID, model.Vector3{X: -8}, map[string]interface{}{"capacity": 200.0})
	pdu := b.entity(model.EntityPDU, "PDU-A", dc.ID, model.Vector3{X: -6}, nil)
	b.rel(model.RelContains, dc.ID, ups.ID)
	b.rel(model.RelContains, dc.ID, pdu.ID)
	b.rel(model.RelPoweredBy, pdu.ID, ups.ID)
	for r := 1; r <= 4; r++ {
		rack := b.entity(model.EntityRack, fmt.Sprintf("Rack-%02d", r), dc.ID, model.Vector3{X: float64(r) * 2}, map[string]interface{}{"capacity": 42.0})
		b.rel(model.RelContains, dc.ID, rack.ID)
		b.rel(model.RelPoweredBy, rack.ID, pdu.ID)
		for s := 1; s <= 6; s++ {
			srv := b.entity(model.EntityServer, fmt.Sprintf("Rack-%02d-Srv-%02d", r, s), rack.ID, model.Vector3{X: float64(r) * 2, Y: float64(s)}, map[string]interface{}{"load": 40.0})
			b.rel(model.RelContains, rack.ID, srv.ID)
			b.rel(model.RelPoweredBy, srv.ID, rack.ID)
			b.sensor(srv.ID, model.SensorCPU, "%", 0, 100)
			b.sensor(srv.ID, model.SensorMemory, "%", 0, 100)
			b.sensor(srv.ID, model.SensorNetwork, "Mbps", 0, 1000)
			b.sensor(srv.ID, model.SensorTemperature, "°C", 18, 45)
		}
	}
}

// Warehouse builds Warehouse -> Zones -> Shelves plus a robot fleet.
func (b *Builder) Warehouse() {
	wh := b.entity(model.EntityWarehouse, "Distribution Center 1", "", model.Vector3{}, map[string]interface{}{"throughput": 500.0})
	for z := 1; z <= 3; z++ {
		zone := b.entity(model.EntityCustom, fmt.Sprintf("Storage Zone %d", z), wh.ID, model.Vector3{X: float64(z) * 10}, nil)
		b.rel(model.RelContains, wh.ID, zone.ID)
		for s := 1; s <= 5; s++ {
			shelf := b.entity(model.EntityShelf, fmt.Sprintf("Z%d Shelf %d", z, s), zone.ID, model.Vector3{X: float64(z) * 10, Y: float64(s) * 2}, nil)
			b.rel(model.RelContains, zone.ID, shelf.ID)
		}
	}
	for r := 1; r <= 4; r++ {
		robot := b.entity(model.EntityRobot, fmt.Sprintf("AMR-%02d", r), wh.ID, model.Vector3{X: float64(r), Y: -2}, map[string]interface{}{"load": 0.0})
		b.rel(model.RelContains, wh.ID, robot.ID)
		b.sensor(robot.ID, model.SensorSpeed, "m/s", 0, 3)
		b.sensor(robot.ID, model.SensorLocation, "m", 0, 100)
	}
}

// SmartCity builds City -> Districts -> Roads/Vehicles/Utility feeds.
func (b *Builder) SmartCity() {
	city := b.entity(model.EntityCity, "Riftford", "", model.Vector3{}, nil)
	grid := b.entity(model.EntityGrid, "City Power Grid", city.ID, model.Vector3{}, map[string]interface{}{"capacity": 2000.0})
	b.rel(model.RelContains, city.ID, grid.ID)
	for d := 1; d <= 3; d++ {
		dist := b.entity(model.EntityDistrict, fmt.Sprintf("District %d", d), city.ID, model.Vector3{X: float64(d) * 20}, nil)
		b.rel(model.RelContains, city.ID, dist.ID)
		b.rel(model.RelPoweredBy, dist.ID, grid.ID)
		road := b.entity(model.EntityRoad, fmt.Sprintf("District %d Main Road", d), dist.ID, model.Vector3{X: float64(d) * 20}, nil)
		b.rel(model.RelContains, dist.ID, road.ID)
		for v := 1; v <= 3; v++ {
			veh := b.entity(model.EntityVehicle, fmt.Sprintf("D%d Vehicle %d", d, v), road.ID, model.Vector3{X: float64(d)*20 + float64(v)}, nil)
			b.rel(model.RelTransports, road.ID, veh.ID)
			b.sensor(veh.ID, model.SensorSpeed, "km/h", 0, 120)
			b.sensor(veh.ID, model.SensorLocation, "m", 0, 10000)
		}
		b.sensor(road.ID, model.SensorMotion, "vehicles/min", 0, 200)
	}
}

// Benchmark generates a wide-and-deep synthetic twin with roughly
// `entityCount` entities (grouped into buildings/floors/rooms/devices) to
// drive the throughput benchmarks described in the README.
func (b *Builder) Benchmark(entityCount int) int {
	created := 0
	root := b.entity(model.EntityCampus, "Benchmark Campus", "", model.Vector3{}, nil)
	created++
	buildingSize := 200
	for created < entityCount {
		bldg := b.entity(model.EntityBuilding, fmt.Sprintf("Bldg-%d", created), root.ID, model.Vector3{X: float64(created)}, nil)
		b.rel(model.RelContains, root.ID, bldg.ID)
		created++
		for i := 0; i < buildingSize && created < entityCount; i++ {
			dev := b.entity(model.EntityDevice, fmt.Sprintf("Device-%d", created), bldg.ID, model.Vector3{X: float64(created)}, map[string]interface{}{"load": b.rng.Float64() * 100})
			b.rel(model.RelContains, bldg.ID, dev.ID)
			b.sensor(dev.ID, model.SensorTemperature, "°C", 15, 60)
			created++
		}
	}
	return created
}
