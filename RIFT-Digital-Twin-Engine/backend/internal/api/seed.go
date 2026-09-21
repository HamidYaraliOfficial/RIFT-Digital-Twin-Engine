package api

import (
	"time"

	"rift/internal/model"
	"rift/internal/registry"
	"rift/internal/synthetic"
)

// seedDemoData runs once on a completely fresh install so the Control Center
// is never an empty screen: it creates one Twin per bundled Template, a
// default admin account, and a starter Operating Hours schedule. Every
// value it creates is a real, queryable, simulate-able object — nothing here
// is a UI mock; it is the exact same CreateTwin/CreateEntity path a user
// clicking through the Twin Builder would exercise.
func (s *Server) seedDemoData() {
	admin, err := s.AuthStore.CreateUser("org_default", "admin", "Administrator", "admin123", model.RoleAdmin)
	if err == nil {
		s.Audit.Record("system", "user_created", "user:"+admin.ID, "", nil)
	}

	templates := []struct {
		name, template string
		build          func(*synthetic.Builder)
	}{
		{"Riftford Factory", "factory", func(b *synthetic.Builder) { b.Factory() }},
		{"HQ Tower (Smart Building)", "building", func(b *synthetic.Builder) { b.Building() }},
		{"DC-1 (Data Center)", "data_center", func(b *synthetic.Builder) { b.DataCenter() }},
		{"Distribution Center 1 (Warehouse)", "warehouse", func(b *synthetic.Builder) { b.Warehouse() }},
		{"Riftford (Smart City)", "smart_city", func(b *synthetic.Builder) { b.SmartCity() }},
	}
	for _, tpl := range templates {
		twin := s.Reg.CreateTwin(&registry.Twin{OrgID: "org_default", WorkspaceID: "ws_default", Name: tpl.name, Template: tpl.template})
		builder := synthetic.NewBuilder(s.Reg, twin.ID, 7)
		tpl.build(builder)

		s.Reg.SetOperatingHours(&model.OperatingHours{
			TwinID: twin.ID, Label: tpl.name + " — default hours", TimeZone: "UTC",
			Schedule: map[string]model.DayHours{
				"0": {Closed: true},
				"1": {Open: "08:00", Close: "18:00"},
				"2": {Open: "08:00", Close: "18:00"},
				"3": {Open: "08:00", Close: "18:00"},
				"4": {Open: "08:00", Close: "18:00"},
				"5": {Open: "08:00", Close: "18:00"},
				"6": {Closed: true},
			},
		})
	}
	_ = time.Now()
}
