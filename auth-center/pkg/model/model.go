package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	"gorm.io/gorm"
)

var (
	PredeclaredRoles = []Role{
		{
			ID:       AdminRoleID,
			RoleName: "Administrator",
			RolePermissions: Permissions{
				Users: &jwt.Permission{
					Actions:     []jwt.Action{"read", "update", "create", "delete"},
					Description: "User management",
				},
				Roles: &jwt.Permission{
					Actions:     []jwt.Action{"read", "update", "create", "delete"},
					Description: "Role management",
				},
				Rules: &jwt.Permission{
					Actions:     []jwt.Action{"read", "update", "create", "delete"},
					Description: "Safety rules",
				},
				Scanning: &jwt.Permission{
					Actions:     []jwt.Action{"execute"},
					Description: "Scanning safety rules",
				},
				Events: &jwt.Permission{
					Actions:     []jwt.Action{"read"},
					Description: "View history event rules",
				},
				Registries: &jwt.Permission{
					Actions:     []jwt.Action{"read", "update", "create", "delete"},
					Description: "Registry management",
				},
				Images: &jwt.Permission{
					Actions:     []jwt.Action{"read", "execute"},
					Description: "Viewing images in repositories",
				},
				Integrations: &jwt.Permission{
					Actions:     []jwt.Action{"read", "update", "create", "delete"},
					Description: "Notification management",
				},
				Notifications: &jwt.Permission{
					Actions:     []jwt.Action{"read", "update", "create", "delete"},
					Description: "Notification lists",
				},
				SystemSettings: &jwt.Permission{
					Actions:     []jwt.Action{"read", "update", "create", "delete", "execute"},
					Description: "CS settings management",
				},
				Clusters: &jwt.Permission{
					Actions:     []jwt.Action{"read", "update", "create", "delete", "execute"},
					Description: "Cluster management",
				},
				InvalidatePublicAccessTokens: &jwt.Permission{
					Actions:     []jwt.Action{"execute"},
					Description: "Token verification",
				},
				PublicAccessTokens: &jwt.Permission{
					Actions:     []jwt.Action{"read", "update", "create", "delete"},
					Description: "Access token management",
				},
			},
			Description: "CS administrator role",
		},
		{
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			RoleName: "Security engineer",
			RolePermissions: Permissions{
				Users: &jwt.Permission{
					Actions:     []jwt.Action{"read", "update", "create", "delete"},
					Description: "User management",
				},
				Roles: &jwt.Permission{
					Actions:     []jwt.Action{"read", "update", "create", "delete"},
					Description: "Role management",
				},
				Rules: &jwt.Permission{
					Actions:     []jwt.Action{"read", "update", "create", "delete"},
					Description: "Safety rules",
				},
				Scanning: &jwt.Permission{
					Actions:     []jwt.Action{"execute"},
					Description: "Scanning safety rules",
				},
				Events: &jwt.Permission{
					Actions:     []jwt.Action{"read"},
					Description: "View history event rules",
				},
				Registries: &jwt.Permission{
					Actions:     []jwt.Action{"read"},
					Description: "Registry management",
				},
				Images: &jwt.Permission{
					Actions:     []jwt.Action{"read", "execute"},
					Description: "Viewing images in repositories",
				},
				Integrations: &jwt.Permission{
					Actions:     []jwt.Action{"read", "update", "create", "delete"},
					Description: "Notification management",
				},
				Notifications: &jwt.Permission{
					Actions:     []jwt.Action{"read", "update", "create", "delete"},
					Description: "Notification lists",
				},
				SystemSettings: &jwt.Permission{
					Actions:     []jwt.Action{"read"},
					Description: "CS settings management",
				},
				Clusters: &jwt.Permission{
					Actions:     []jwt.Action{"read"},
					Description: "Cluster management",
				},
				InvalidatePublicAccessTokens: &jwt.Permission{
					Actions:     []jwt.Action{},
					Description: "Token verification",
				},
				PublicAccessTokens: &jwt.Permission{
					Actions:     []jwt.Action{"read", "update", "create", "delete"},
					Description: "Access token management",
				},
			},
			Description: "Security Engineer role",
		},
		{
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000003"),
			RoleName: "CI/CD",
			RolePermissions: Permissions{
				Users: &jwt.Permission{
					Actions:     []jwt.Action{},
					Description: "User management",
				},
				Roles: &jwt.Permission{
					Actions:     []jwt.Action{},
					Description: "Role management",
				},
				Rules: &jwt.Permission{
					Actions:     []jwt.Action{},
					Description: "Safety rules",
				},
				Scanning: &jwt.Permission{
					Actions:     []jwt.Action{"execute"},
					Description: "Scanning safety rules",
				},
				Events: &jwt.Permission{
					Actions:     []jwt.Action{},
					Description: "View history event rules",
				},
				Registries: &jwt.Permission{
					Actions:     []jwt.Action{},
					Description: "Registry management",
				},
				Images: &jwt.Permission{
					Actions:     []jwt.Action{},
					Description: "Viewing images in repositories",
				},
				Integrations: &jwt.Permission{
					Actions:     []jwt.Action{},
					Description: "Notification management",
				},
				Notifications: &jwt.Permission{
					Actions:     []jwt.Action{},
					Description: "Notification lists",
				},
				SystemSettings: &jwt.Permission{
					Actions:     []jwt.Action{},
					Description: "CS settings management",
				},
				Clusters: &jwt.Permission{
					Actions:     []jwt.Action{"read"},
					Description: "Cluster management",
				},
				InvalidatePublicAccessTokens: &jwt.Permission{
					Actions:     []jwt.Action{},
					Description: "Token verification",
				},
				PublicAccessTokens: &jwt.Permission{
					Actions:     []jwt.Action{"read", "update", "create", "delete"},
					Description: "Access token management",
				},
			},
			Description: "Role for CI/CD integrations without access to the web interface",
		},
		{
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000004"),
			RoleName: "Developer",
			RolePermissions: Permissions{
				Users: &jwt.Permission{
					Actions:     []jwt.Action{},
					Description: "User management",
				},
				Roles: &jwt.Permission{
					Actions:     []jwt.Action{"read"},
					Description: "Role management",
				},
				Rules: &jwt.Permission{
					Actions:     []jwt.Action{"read"},
					Description: "Safety rules",
				},
				Scanning: &jwt.Permission{
					Actions:     []jwt.Action{"execute"},
					Description: "Scanning safety rules",
				},
				Events: &jwt.Permission{
					Actions:     []jwt.Action{"read"},
					Description: "View history event rules",
				},
				Registries: &jwt.Permission{
					Actions:     []jwt.Action{"read"},
					Description: "Registry management",
				},
				Images: &jwt.Permission{
					Actions:     []jwt.Action{"read", "execute"},
					Description: "Viewing images in repositories",
				},
				Integrations: &jwt.Permission{
					Actions:     []jwt.Action{"read"},
					Description: "Notification management",
				},
				Notifications: &jwt.Permission{
					Actions:     []jwt.Action{"read"},
					Description: "Notification lists",
				},
				SystemSettings: &jwt.Permission{
					Actions:     []jwt.Action{"read"},
					Description: "CS settings management",
				},
				Clusters: &jwt.Permission{
					Actions:     []jwt.Action{"read"},
					Description: "Cluster management",
				},
				InvalidatePublicAccessTokens: &jwt.Permission{
					Actions:     []jwt.Action{},
					Description: "Token verification",
				},
				PublicAccessTokens: &jwt.Permission{
					Actions:     []jwt.Action{"read", "update", "create", "delete"},
					Description: "Access token management",
				},
			},
			Description: "Developer role",
		},
		{
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000005"),
			RoleName: "Auditor",
			RolePermissions: Permissions{
				Users: &jwt.Permission{
					Actions:     []jwt.Action{},
					Description: "User management",
				},
				Roles: &jwt.Permission{
					Actions:     []jwt.Action{"read"},
					Description: "Role management",
				},
				Rules: &jwt.Permission{
					Actions:     []jwt.Action{"read"},
					Description: "Safety rules",
				},
				Scanning: &jwt.Permission{
					Actions:     []jwt.Action{},
					Description: "Scanning safety rules",
				},
				Events: &jwt.Permission{
					Actions:     []jwt.Action{"read"},
					Description: "View history event rules",
				},
				Registries: &jwt.Permission{
					Actions:     []jwt.Action{"read"},
					Description: "Registry management",
				},
				Images: &jwt.Permission{
					Actions:     []jwt.Action{"read"},
					Description: "Viewing images in repositories",
				},
				Integrations: &jwt.Permission{
					Actions:     []jwt.Action{"read"},
					Description: "Notification management",
				},
				Notifications: &jwt.Permission{
					Actions:     []jwt.Action{"read"},
					Description: "Notification lists",
				},
				SystemSettings: &jwt.Permission{
					Actions:     []jwt.Action{"read"},
					Description: "CS settings management",
				},
				Clusters: &jwt.Permission{
					Actions:     []jwt.Action{"read"},
					Description: "Cluster management",
				},
				InvalidatePublicAccessTokens: &jwt.Permission{
					Actions:     []jwt.Action{},
					Description: "Token verification",
				},
				PublicAccessTokens: &jwt.Permission{
					Actions:     []jwt.Action{"read", "update", "create", "delete"},
					Description: "Access token management",
				},
			},
			Description: "Read-only role for results analysis",
		},
	}
)

// Base struct represents basic model to be used by all data structs.
type Base struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// BeforeCreate callback sets ID as newly generated UUID. This makes it possible to not install 'uuid-ossp' extension
// with PostgreSQL. If ID was set already, just go ahead and do nothing.
func (b *Base) BeforeCreate(_ *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}

	return nil
}
