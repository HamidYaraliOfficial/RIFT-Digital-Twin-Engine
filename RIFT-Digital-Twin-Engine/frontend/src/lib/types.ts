// Mirrors the Go models in backend/internal/model exactly enough for the
// Control Center to consume the REST API without any implicit "any" data
// flowing through the UI.

export interface Vector3 { x: number; y: number; z: number; }

export interface HealthScore {
  health: number;
  risk: number;
  criticality: number;
  updatedAt: string;
}

export type EntityStatus =
  | "operational" | "warning" | "degraded" | "maintenance" | "failed" | "recovered" | "unknown";

export interface Entity {
  id: string;
  twinId: string;
  type: string;
  name: string;
  properties?: Record<string, unknown>;
  state?: Record<string, unknown>;
  position: Vector3;
  orientation: Vector3;
  dimensions: Vector3;
  metadata?: Record<string, unknown>;
  owner: string;
  parentId?: string;
  children?: string[];
  health: HealthScore;
  status: EntityStatus;
  version: number;
  createdAt: string;
  updatedAt: string;
}

export interface Relationship {
  id: string;
  sourceId: string;
  targetId: string;
  type: string;
  confidence: number;
  timestamp: string;
  metadata?: Record<string, unknown>;
}

export interface Sensor {
  id: string;
  entityId: string;
  twinId: string;
  name: string;
  type: string;
  unit: string;
  samplingRateMs: number;
  accuracy: number;
  rangeMin: number;
  rangeMax: number;
  status: "online" | "offline" | "degraded" | "calibrating";
  source: string;
  lastSeen: string;
  createdAt: string;
}

export interface Reading {
  sensorId: string;
  entityId: string;
  twinId: string;
  timestamp: string;
  value: number;
  unit: string;
  quality: "good" | "uncertain" | "bad";
  sequence: number;
  source: string;
}

export interface RiftEvent {
  id: string;
  twinId: string;
  type: string;
  timestamp: string;
  sourceEntityId?: string;
  targetEntityId?: string;
  payload?: Record<string, unknown>;
  severity: "info" | "warning" | "critical";
  lifecycle: "open" | "acknowledged" | "resolved";
}

export interface Alert {
  id: string;
  twinId: string;
  ruleId?: string;
  type: string;
  severity: "info" | "warning" | "critical";
  entityId: string;
  message: string;
  createdAt: string;
  lastOccurredAt: string;
  occurrenceCount: number;
  status: "open" | "acknowledged" | "resolved";
  escalationLevel: number;
}

export interface Rule {
  id: string;
  twinId: string;
  name: string;
  entityId: string;
  trigger: string;
  conditions: { field: string; operator: string; value: number; stringValue?: string }[];
  actions: { type: string; params: Record<string, unknown> }[];
  priority: number;
  cooldownMs: number;
  enabled: boolean;
  version: number;
}

export interface Twin {
  id: string;
  orgId: string;
  workspaceId: string;
  name: string;
  template: string;
  createdAt: string;
  updatedAt: string;
  archived: boolean;
}

export interface FaultSpec {
  type: string;
  entityId: string;
  atTick: number;
  magnitude: number;
  durationTicks?: number;
}

export interface ScenarioResult {
  durationTicks: number;
  affectedEntities: string[];
  deltas: { entityId: string; field: string; baseline: number; simulated: number; deltaPct: number }[];
  eventsGenerated: number;
  summary: string;
}

export interface Scenario {
  id: string;
  twinId: string;
  name: string;
  description: string;
  seed: number;
  deterministic: boolean;
  faults: FaultSpec[];
  durationTicks: number;
  status: "pending" | "running" | "done" | "failed";
  result?: ScenarioResult;
  createdAt: string;
}

export interface DayHours { closed: boolean; open: string; close: string; }

export interface OperatingHours {
  id: string;
  twinId: string;
  entityId?: string;
  label: string;
  timeZone: string;
  schedule: Record<string, DayHours>;
  updatedAt: string;
}

export interface OperatingStatus {
  isOpen: boolean;
  now: string;
  nextChangeAt: string;
  nextChangeIsOpen: boolean;
  timeUntilNext: string;
  timeUntilNextSecs: number;
}

export interface OperatingHoursEntry { schedule: OperatingHours; status: OperatingStatus; }

export interface ClockState { running: boolean; speed: number; simTime: string; ticks: number; }

export interface PipelineStats { received: number; processed: number; dropped: number; queueLen: number; queueCap: number; }

export interface MonteCarloResult {
  runs: number; metric: string; mean: number; p10: number; p50: number; p90: number; best: number; worst: number; samples: number[];
}

export interface OptimizeResult { params: Record<string, number>; score: number; trials: number; }

export interface ImpactResult { entityId: string; depth: number; path: string[]; }

export interface RootCause { entityId: string; reason: string; confidence: number; evidence: string[]; }

export interface AuditEntry {
  id: string; timestamp: string; actor: string; action: string; target: string; twinId?: string; details?: Record<string, unknown>;
}

export interface LoginResponse { token: string; userId: string; displayName: string; role: string; expiresAt: string; }
