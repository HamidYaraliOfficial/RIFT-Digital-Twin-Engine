package model

import "time"

// Weekday indices follow time.Weekday (Sunday = 0).

// DayHours describes the open/close window for a single day of the week.
// Closed=true means the facility/entity does not open at all that day.
type DayHours struct {
	Closed bool   `json:"closed"`
	Open   string `json:"open"`  // "HH:MM", 24h, local to TimeZone
	Close  string `json:"close"` // "HH:MM", 24h, local to TimeZone. May be < Open to mean "past midnight".
}

// OperatingHours is a fully user-configurable weekly schedule attached to a
// Twin or a specific Entity (e.g. a warehouse dock, a building, a factory
// line shift). It is entered entirely by the user through the Control
// Center; RIFT then computes live open/closed status and countdowns from it.
type OperatingHours struct {
	ID        string              `json:"id"`
	TwinID    string              `json:"twinId"`
	EntityID  string              `json:"entityId,omitempty"` // empty = applies to whole twin
	Label     string              `json:"label"`
	TimeZone  string              `json:"timeZone"` // IANA name, e.g. "Asia/Baku"
	Schedule  map[string]DayHours `json:"schedule"` // key: "0".."6" (time.Weekday as string)
	UpdatedAt time.Time           `json:"updatedAt"`
}

// OperatingStatus is the computed, live answer for "are we open right now".
type OperatingStatus struct {
	IsOpen            bool      `json:"isOpen"`
	Now               time.Time `json:"now"`
	NextChangeAt      time.Time `json:"nextChangeAt"`
	NextChangeIsOpen  bool      `json:"nextChangeIsOpen"` // what the state becomes at NextChangeAt
	TimeUntilNext     string    `json:"timeUntilNext"`    // human readable, e.g. "3h12m"
	TimeUntilNextSecs int64     `json:"timeUntilNextSecs"`
}
