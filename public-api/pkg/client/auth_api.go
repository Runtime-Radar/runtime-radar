package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/build"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/model"
)

const (
	AuthAPITimeout = 10 * time.Second
)

var (
	ErrUserNotFound = errors.New("user not found")
)

var authAPIPermissions = &jwt.RolePermissions{
	Users: &jwt.Permission{
		Actions: []jwt.Action{jwt.ActionRead},
	},
}

type UsersGetter interface {
	GetUser(ctx context.Context, userID uuid.UUID) (*AuthAPIUser, error)
}

type AuthAPI struct {
	*http.Client
	baseURL string
	token   string
}

type AuthAPIUser struct {
	Email                 *string     `json:"email"`
	RoleID                string      `json:"role_id"`
	MappingRoleID         *string     `json:"mapping_role_id"`
	ID                    string      `json:"id"`
	Username              string      `json:"username"`
	Role                  AuthAPIRole `json:"role"`
	AuthType              string      `json:"auth_type"`
	LastPasswordChangedAt string      `json:"last_password_changed_at"`
}

type AuthAPIRole struct {
	RoleName        string             `json:"role_name"`
	RolePermissions *model.Permissions `json:"role_permissions"`
	Description     string             `json:"description"`
	ID              uuid.UUID          `json:"id"`
}

func NewAuthAPI(baseURL string, tlsConfig *tls.Config, tokenKey []byte) (authAPI *AuthAPI, closeFn func(), err error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsConfig

	client := &AuthAPI{
		Client: &http.Client{
			Transport: transport,
			Timeout:   AuthAPITimeout,
		},
		baseURL: strings.TrimSuffix(baseURL, "/"),
	}
	if len(tokenKey) > 0 {
		var err error
		client.token, err = jwt.GenerateServiceToken(tokenKey, build.AppName, authAPIPermissions)
		if err != nil {
			return nil, nil, err
		}
	}

	return client, client.CloseIdleConnections, nil
}

func (c *AuthAPI) GetUser(ctx context.Context, userID uuid.UUID) (*AuthAPIUser, error) {
	path := fmt.Sprintf("/api/v1/user/%s", userID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("can't build http request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("auth.GetUser: status code is not 200: %d", resp.StatusCode)
	}

	var user AuthAPIUser
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("can't decode response: %w", err)
	}

	return &user, err
}
