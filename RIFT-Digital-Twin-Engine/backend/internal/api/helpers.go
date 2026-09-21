package api

import "rift/internal/model"

func roleOrDefault(r string) model.Role {
	switch model.Role(r) {
	case model.RoleAdmin, model.RoleOperator, model.RoleEngineer, model.RoleViewer, model.RoleMaintainer:
		return model.Role(r)
	default:
		return model.RoleViewer
	}
}
