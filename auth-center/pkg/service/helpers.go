package service

import (
	"unicode"
	"unicode/utf8"

	"github.com/runtime-radar/runtime-radar/auth-center/api"
	"github.com/runtime-radar/runtime-radar/auth-center/pkg/model"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	"golang.org/x/crypto/bcrypt"
)

const maskedPassword = "******"

func haveUpper(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) && unicode.IsLetter(r) {
			return true
		}
	}

	return false
}

func haveLower(s string) bool {
	for _, r := range s {
		if unicode.IsLower(r) && unicode.IsLetter(r) {
			return true
		}
	}

	return false
}

func haveDigit(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}

	return false
}

func newPasswordCheck(pass string, passwordCheck map[string]struct{}) string {
	if pass == "" {
		return "PASSWORD_IS_EMPTY"
	}

	if _, ok := passwordCheck[pass]; ok {
		return "PASSWORD_FOUND_IN_PASS_LIST"
	}

	if utf8.RuneCountInString(pass) < 8 {
		return "PASSWORD_IS_SHORT"
	}

	if utf8.RuneCountInString(pass) > 16 {
		return "PASSWORD_IS_LONG"
	}

	if !haveUpper(pass) {
		return "PASSWORD_UPPERCASE_LETTER_IS_MISSED"
	}

	if !haveLower(pass) {
		return "PASSWORD_LOWERCASE_LETTER_IS_MISSED"
	}

	if !haveDigit(pass) {
		return "PASSWORD_DIGIT_IS_MISSED"
	}

	return ""
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func convertActionToStringArray(actions jwt.Actions) []string {
	strArrayActions := make([]string, 0, len(actions))

	for _, action := range actions {
		strArrayActions = append(strArrayActions, string(action))
	}
	return strArrayActions
}

func fillPermission(p *jwt.Permission) *api.ReadRoleResp_Permission {
	return &api.ReadRoleResp_Permission{
		Actions:     convertActionToStringArray(p.Actions),
		Description: p.Description,
	}
}

func fillRoleResp(role *model.Role) *api.ReadRoleResp {
	return &api.ReadRoleResp{
		RoleName: role.RoleName,
		Id:       role.ID.String(),
		RolePermissions: &api.ReadRoleResp_RolePermissions{
			Users:                        fillPermission(role.RolePermissions.Users),
			Roles:                        fillPermission(role.RolePermissions.Roles),
			Rules:                        fillPermission(role.RolePermissions.Rules),
			Scanning:                     fillPermission(role.RolePermissions.Scanning),
			Events:                       fillPermission(role.RolePermissions.Events),
			Registries:                   fillPermission(role.RolePermissions.Registries),
			Images:                       fillPermission(role.RolePermissions.Images),
			Integrations:                 fillPermission(role.RolePermissions.Integrations),
			Notifications:                fillPermission(role.RolePermissions.Notifications),
			SystemSettings:               fillPermission(role.RolePermissions.SystemSettings),
			Clusters:                     fillPermission(role.RolePermissions.Clusters),
			InvalidatePublicAccessTokens: fillPermission(role.RolePermissions.InvalidatePublicAccessTokens),
			PublicAccessTokens:           fillPermission(role.RolePermissions.PublicAccessTokens),
		},
		Description: role.Description,
	}
}
