package jwt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "accessToken"
	TokenTypeRefresh TokenType = "refreshToken"
)

type JSONTime time.Time

func (jt JSONTime) MarshalJSON() ([]byte, error) {
	tm := time.Time(jt)
	if tm.IsZero() {
		return []byte("0"), nil
	}
	return json.Marshal(tm.Unix())
}

func (jt *JSONTime) UnmarshalJSON(b []byte) error {
	if jt == nil {
		return errors.New("JSONTime is nil")
	}

	var unixTime float64
	if err := json.Unmarshal(b, &unixTime); err != nil {
		return err
	}

	*jt = JSONTime(time.Unix(int64(unixTime), 0))
	return nil
}

func (jt JSONTime) String() string {
	return time.Time(jt).String()
}

type Token struct {
	Username              string    `json:"username"`
	UserID                string    `json:"user_id"`
	TokenType             TokenType `json:"token_type"`
	Email                 string    `json:"email,omitempty"`
	Role                  *Role     `json:"role,omitempty"`
	LastPasswordChangedAt JSONTime  `json:"last_password_changed_at"`

	jwt.RegisteredClaims
}

func (t *Token) GetUserID() string {
	if t == nil {
		return ""
	}

	return t.UserID
}

type Role struct {
	Description     string           `json:"description,omitempty"`
	RoleName        string           `json:"role_name"`
	RolePermissions *RolePermissions `json:"role_permissions,omitempty"`
}

type RolePermissions struct {
	Users                        *Permission `json:"users,omitempty"`
	Roles                        *Permission `json:"roles,omitempty"`
	Rules                        *Permission `json:"rules,omitempty"`
	Scopes                       *Permission `json:"scopes,omitempty"`
	Scanning                     *Permission `json:"scanning,omitempty"`
	Events                       *Permission `json:"events,omitempty"`
	Registries                   *Permission `json:"registries,omitempty"`
	Images                       *Permission `json:"images,omitempty"`
	Integrations                 *Permission `json:"integrations,omitempty"`
	Notifications                *Permission `json:"notifications,omitempty"`
	SystemSettings               *Permission `json:"system_settings,omitempty"`
	Clusters                     *Permission `json:"clusters,omitempty"`
	InvalidatePublicAccessTokens *Permission `json:"invalidate_public_access_tokens,omitempty"`
	PublicAccessTokens           *Permission `json:"public_access_tokens,omitempty"`
}

type PermissionType uint8

const (
	PermissionUsers PermissionType = iota
	PermissionRoles
	PermissionRules
	PermissionScanning
	PermissionEvents
	PermissionRegistries
	PermissionImages
	PermissionIntegrations
	PermissionNotifications
	PermissionSystemSettings
	PermissionClusters
	PermissionInvalidatePublicAccessTokens
	PermissionPublicAccessTokens
)

func (rp *RolePermissions) GetPermission(pt PermissionType) *Permission {
	if rp == nil {
		return nil
	}
	switch pt {
	case PermissionUsers:
		return rp.Users
	case PermissionRoles:
		return rp.Roles
	case PermissionRules:
		return rp.Rules
	case PermissionScanning:
		return rp.Scanning
	case PermissionEvents:
		return rp.Events
	case PermissionRegistries:
		return rp.Registries
	case PermissionImages:
		return rp.Images
	case PermissionIntegrations:
		return rp.Integrations
	case PermissionNotifications:
		return rp.Notifications
	case PermissionSystemSettings:
		return rp.SystemSettings
	case PermissionClusters:
		return rp.Clusters
	case PermissionInvalidatePublicAccessTokens:
		return rp.InvalidatePublicAccessTokens
	case PermissionPublicAccessTokens:
		return rp.PublicAccessTokens
	}

	return nil
}

type Permission struct {
	Actions     Actions `json:"actions"`
	Description string  `json:"description,omitempty"`
}

func (p *Permission) containsAll(as ...Action) bool {
	if p == nil {
		return false
	}

	return p.Actions.containsAll(as...)
}

func UnverifiedTokenFromContext(ctx context.Context) (*Token, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	auth := md.Get("authorization")
	if !ok || len(auth) == 0 {
		return nil, errors.New("no authorization token provided")
	}

	tokenStr, ok := strings.CutPrefix(auth[0], "Bearer ")
	if !ok {
		return nil, errors.New("unknown authorization format")
	}

	token, _, err := jwt.NewParser().ParseUnverified(tokenStr, &Token{})
	if err != nil {
		return nil, fmt.Errorf("token was incorrectly parsed: %w", err)
	}

	t, ok := token.Claims.(*Token)
	if !ok {
		return nil, errors.New("can't get token from claims")
	}

	return t, nil
}

func TokenFromContext(ctx context.Context, key []byte) (*Token, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	auth := md.Get("authorization")
	if !ok || len(auth) == 0 {
		return nil, errors.New("no authorization token provided")
	}

	tokenStr, ok := strings.CutPrefix(auth[0], "Bearer ")
	if !ok {
		return nil, errors.New("unknown authorization format")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &Token{}, func(*jwt.Token) (any, error) {
		return key, nil
	})
	if err != nil {
		return nil, fmt.Errorf("token was incorrectly parsed: %w", err)
	}

	t, ok := token.Claims.(*Token)
	if !ok {
		return nil, errors.New("can't get token from claims")
	}

	return t, nil
}

func SignHS256(token *Token, key []byte) (string, error) {
	builder := jwt.NewWithClaims(jwt.SigningMethodHS256, token)
	signed, err := builder.SignedString(key)
	if err != nil {
		return "", err
	}
	return signed, nil
}
