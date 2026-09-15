package model

import (
	"testing"

	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
)

// Creating and deleting users are administrator-only actions. Any other predeclared role holding
// them would let its holders manage accounts other than their own, which is what the check in the
// service layer exists to prevent.
func TestOnlyAdministratorManagesUsers(t *testing.T) {
	restricted := []jwt.Action{"create", "delete"}

	for _, action := range restricted {
		var adminHolds bool

		for _, role := range PredeclaredRoles {
			users := role.RolePermissions.Users
			if users == nil {
				continue
			}

			var holds bool
			for _, granted := range users.Actions {
				if granted == action {
					holds = true
				}
			}

			switch {
			case role.ID == AdminRoleID:
				adminHolds = holds
			case holds:
				t.Errorf("Role %q must not hold users:%s", role.RoleName, action)
			}
		}

		if !adminHolds {
			t.Errorf("Expected the administrator role to hold users:%s", action)
		}
	}
}

func TestOnlyCICDLacksUserExecute(t *testing.T) {
	for _, role := range PredeclaredRoles {
		users := role.RolePermissions.Users
		if users == nil {
			t.Errorf("Role %q must declare users permission", role.RoleName)
			continue
		}

		holdsExecute := false
		for _, granted := range users.Actions {
			if granted == jwt.ActionExecute {
				holdsExecute = true
			}
		}

		if role.RoleName == "CI/CD" {
			if holdsExecute {
				t.Errorf("Role %q must not hold users:%s", role.RoleName, jwt.ActionExecute)
			}
			continue
		}

		if !holdsExecute {
			t.Errorf("Role %q must hold users:%s", role.RoleName, jwt.ActionExecute)
		}
	}
}
