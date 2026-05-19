package auth

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"github.com/runtime-radar/runtime-radar/lib/security"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/client"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/database"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/model"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/server/middleware"
	"google.golang.org/grpc/codes"
)

var (
	ErrTokenInvalidated = errors.New("token was invalidated")
)

// Verifier checks whether public access token is provided, valid
// and has sufficient privileges to perform given actions.
type Verifier struct {
	UsersGetter           client.UsersGetter
	AccessTokenRepository database.AccessTokenRepository

	AccessTokenSalt []byte
}

// VerifyPermission tries to extract public access token from context
// and check whether it has enough privileges to perform actions on given pt.
func (v *Verifier) VerifyPermission(ctx context.Context, pt jwt.PermissionType, as ...jwt.Action) error {
	if len(as) == 0 {
		panic("empty Actions is not allowed here")
	}

	token, ok := middleware.AccessTokenFromContext(ctx)
	if !ok {
		return fmt.Errorf("%w: no access token in context", jwt.ErrUnauthenticated)
	}

	hashed := security.HashSaltedSHA512AsHex([]byte(token), v.AccessTokenSalt)

	at, err := v.AccessTokenRepository.GetByTokenHash(ctx, hashed)
	if err != nil {
		return fmt.Errorf("%w: %w", jwt.ErrUnauthenticated, err)
	}

	if at.ExpiresAt != nil && at.ExpiresAt.Before(time.Now()) {
		return jwt.ErrTokenExpired
	}

	if at.InvalidatedAt != nil {
		return ErrTokenInvalidated
	}

	user, err := v.UsersGetter.GetUser(ctx, at.UserID)
	if err != nil {
		return fmt.Errorf("%w: %w", jwt.ErrUnauthenticated, err)
	}

	perms := maxCommonPermissions(at.Permissions, user.Role.RolePermissions)

	p := perms.GetPermission(pt)
	if p == nil || len(p.Actions) == 0 {
		return jwt.ErrPermissionDenied
	}

	for _, a := range as {
		if !slices.Contains(p.Actions, a) {
			return jwt.ErrPermissionDenied
		}
	}

	return nil
}

// maxCommonPermissions computes the intersection of permissions between
// two sets of token and user permissions, returning only the permissions
// that exist in both.
func maxCommonPermissions(first, second *model.Permissions) *jwt.RolePermissions {
	return &jwt.RolePermissions{
		Users:                        maxCommonPermission(first.Users, second.Users),
		Roles:                        maxCommonPermission(first.Roles, second.Roles),
		Rules:                        maxCommonPermission(first.Rules, second.Rules),
		Scopes:                       maxCommonPermission(first.Scopes, second.Scopes),
		Scanning:                     maxCommonPermission(first.Scanning, second.Scanning),
		Events:                       maxCommonPermission(first.Events, second.Events),
		Registries:                   maxCommonPermission(first.Registries, second.Registries),
		Images:                       maxCommonPermission(first.Images, second.Images),
		Integrations:                 maxCommonPermission(first.Integrations, second.Integrations),
		Notifications:                maxCommonPermission(first.Notifications, second.Notifications),
		SystemSettings:               maxCommonPermission(first.SystemSettings, second.SystemSettings),
		Clusters:                     maxCommonPermission(first.Clusters, second.Clusters),
		InvalidatePublicAccessTokens: maxCommonPermission(first.InvalidatePublicAccessTokens, second.InvalidatePublicAccessTokens),
		PublicAccessTokens:           maxCommonPermission(first.PublicAccessTokens, second.PublicAccessTokens),
	}
}

func maxCommonPermission(first, second *jwt.Permission) *jwt.Permission {
	if first == nil || second == nil {
		return nil
	}
	return &jwt.Permission{
		Actions: maxCommonActions(first.Actions, second.Actions),
	}
}

func maxCommonActions(first, second jwt.Actions) jwt.Actions {
	res := []jwt.Action{}

	for _, action := range first {
		if slices.Contains(second, action) {
			res = append(res, action)
		}
	}

	return res
}

func PermissionErrorToStatus(err error) error {
	switch {
	case errors.Is(err, jwt.ErrUnauthenticated), errors.Is(err, ErrTokenInvalidated), errors.Is(err, jwt.ErrTokenExpired):
		return errcommon.StatusWithReason(codes.Unauthenticated, errcommon.Unauthenticated, "unauthenticated").Err()

	case errors.Is(err, jwt.ErrPermissionDenied):
		return errcommon.StatusWithReason(codes.PermissionDenied, errcommon.PermissionDenied, "permission denied").Err()

	default:
		return errcommon.StatusWithReason(codes.Internal, errcommon.Internal, "internal error during auth").Err()
	}
}
