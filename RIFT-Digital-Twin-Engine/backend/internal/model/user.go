package model

import "time"

// Role is a coarse-grained RBAC role. Fine-grained per-twin scopes are
// layered on top via Permission (see TwinPermission).
type Role string

const (
	RoleAdmin     Role = "admin"
	RoleOperator  Role = "operator"
	RoleEngineer  Role = "engineer"
	RoleViewer    Role = "viewer"
	RoleMaintainer Role = "maintainer"
)

// Organization is the top level of the multi-tenant hierarchy.
type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

// Workspace groups Twins, Users and Teams inside an Organization.
type Workspace struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"orgId"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

// User is an authenticated principal.
type User struct {
	ID           string    `json:"id"`
	OrgID        string    `json:"orgId"`
	Username     string    `json:"username"`
	DisplayName  string    `json:"displayName"`
	PasswordHash string    `json:"-"`
	PasswordSalt string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
}

// ScopeAction enumerates what a TwinPermission grants.
type ScopeAction string

const (
	ScopeRead           ScopeAction = "read"
	ScopeWrite          ScopeAction = "write"
	ScopeScenario       ScopeAction = "scenario"
	ScopeActuatorControl ScopeAction = "actuator_control"
	ScopeAdmin          ScopeAction = "admin"
)

// TwinPermission implements the Twin Permission Model: a User or Role gets a
// set of Scopes on a specific Twin (or specific EntityID for actuator scopes).
type TwinPermission struct {
	ID       string        `json:"id"`
	TwinID   string        `json:"twinId"`
	EntityID string        `json:"entityId,omitempty"` // empty = whole twin
	UserID   string        `json:"userId,omitempty"`
	Role     Role          `json:"role,omitempty"`
	Scopes   []ScopeAction `json:"scopes"`
}
