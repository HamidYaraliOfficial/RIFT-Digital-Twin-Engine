package model

import "time"

// SensorType enumerates the built-in sensor kinds. Custom sensors are
// supported by leaving Type set to SensorCustom and describing the shape in
// Metadata.
type SensorType string

const (
	SensorTemperature SensorType = "temperature"
	SensorHumidity    SensorType = "humidity"
	SensorPressure    SensorType = "pressure"
	SensorVibration   SensorType = "vibration"
	SensorEnergy      SensorType = "energy"
	SensorVoltage     SensorType = "voltage"
	SensorCurrent     SensorType = "current"
	SensorAirQuality  SensorType = "air_quality"
	SensorLight       SensorType = "light"
	SensorMotion      SensorType = "motion"
	SensorSpeed       SensorType = "speed"
	SensorLocation    SensorType = "location"
	SensorCPU         SensorType = "cpu"
	SensorMemory      SensorType = "memory"
	SensorDisk        SensorType = "disk"
	SensorNetwork     SensorType = "network"
	SensorCustom      SensorType = "custom"
)

// SensorStatus reflects link/health of the physical or simulated device.
type SensorStatus string

const (
	SensorOnline      SensorStatus = "online"
	SensorOffline     SensorStatus = "offline"
	SensorDegraded    SensorStatus = "degraded"
	SensorCalibrating SensorStatus = "calibrating"
)

// Sensor describes a single measurement channel attached to an Entity.
type Sensor struct {
	ID             string       `json:"id"`
	EntityID       string       `json:"entityId"`
	TwinID         string       `json:"twinId"`
	Name           string       `json:"name"`
	Type           SensorType   `json:"type"`
	Unit           string       `json:"unit"`
	SamplingRateMs int64        `json:"samplingRateMs"`
	Accuracy       float64      `json:"accuracy"` // +/- in Unit
	RangeMin       float64      `json:"rangeMin"`
	RangeMax       float64      `json:"rangeMax"`
	Status         SensorStatus `json:"status"`
	Calibration    time.Time    `json:"calibration"`
	Source         string       `json:"source"` // adapter name: simulated, http, mqtt, opcua...
	ChannelID      string       `json:"channelId"`
	LastSeen       time.Time    `json:"lastSeen"`
	CreatedAt      time.Time    `json:"createdAt"`
}

// Quality mirrors OPC-UA-style data quality flags for a telemetry reading.
type Quality string

const (
	QualityGood      Quality = "good"
	QualityUncertain Quality = "uncertain"
	QualityBad       Quality = "bad"
)

// Reading is a single telemetry sample flowing through the Ingestion Pipeline.
type Reading struct {
	SensorID  string    `json:"sensorId"`
	EntityID  string    `json:"entityId"`
	TwinID    string    `json:"twinId"`
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	Quality   Quality   `json:"quality"`
	Sequence  uint64    `json:"sequence"`
	Source    string    `json:"source"`
}
