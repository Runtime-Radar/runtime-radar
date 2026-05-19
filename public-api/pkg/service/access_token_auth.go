package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	lib_context "github.com/runtime-radar/runtime-radar/lib/security/jwt/context"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errTokenInaccessible error = status.Error(codes.PermissionDenied, "token is not accessible")

// AccessTokenAuth is a layer for jwt-based authentication.
// Base server interface should not be embedded here unlike
// in implementations of other layers.
// All required methods should be explicitly implemented to ensure
// that new methods of the basic server are implemented for auth layer.
type AccessTokenAuth struct {
	AccessToken AccessToken
	SecretKey   []byte
	Verifier    jwt.Verifier
}

func (a *AccessTokenAuth) Create(ctx context.Context, req *model.CreateAccessTokenReq) (id uuid.UUID, token string, err error) {
	if err := a.Verifier.VerifyPermission(ctx, jwt.PermissionPublicAccessTokens, jwt.ActionCreate); err != nil {
		return uuid.Nil, "", errcommon.PermissionErrorToStatus(err)
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return uuid.Nil, "", status.Errorf(codes.Unauthenticated, "can't get user id: %v", err)
	}

	if userID != req.UserID {
		msg := "current user != target user"
		return uuid.Nil, "", errcommon.StatusWithReason(codes.PermissionDenied, errcommon.PermissionDenied, msg).Err()
	}

	lib_context.SetUserID(ctx)

	return a.AccessToken.Create(ctx, req)
}

func (a *AccessTokenAuth) ListPage(ctx context.Context, pageNum, pageSize int, order string) ([]*model.AccessTokenResp, int, error) {
	if err := a.Verifier.VerifyPermission(ctx, jwt.PermissionPublicAccessTokens, jwt.ActionRead); err != nil {
		return nil, 0, errcommon.PermissionErrorToStatus(err)
	}

	lib_context.SetUserID(ctx)

	return a.AccessToken.ListPage(ctx, pageNum, pageSize, order)
}

func (a *AccessTokenAuth) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := a.verifyAccessToken(ctx, id, jwt.ActionDelete); err != nil {
		return err
	}

	lib_context.SetUserID(ctx)

	return a.AccessToken.Delete(ctx, id)
}

func (a *AccessTokenAuth) InvalidateAll(ctx context.Context) error {
	if err := a.Verifier.VerifyPermission(ctx, jwt.PermissionInvalidatePublicAccessTokens, jwt.ActionExecute); err != nil {
		return errcommon.PermissionErrorToStatus(err)
	}

	lib_context.SetUserID(ctx)

	return a.AccessToken.InvalidateAll(ctx)
}

func (a *AccessTokenAuth) GetByID(ctx context.Context, id uuid.UUID) (at *model.AccessTokenResp, err error) {
	if at, err = a.verifyAccessToken(ctx, id, jwt.ActionRead); err != nil {
		return nil, err
	}

	lib_context.SetUserID(ctx)

	return at, nil
}

func (a *AccessTokenAuth) verifyAccessToken(ctx context.Context, id uuid.UUID, as ...jwt.Action) (*model.AccessTokenResp, error) {
	if err := a.Verifier.VerifyPermission(ctx, jwt.PermissionPublicAccessTokens, as...); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	at, err := a.AccessToken.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "can't get user id: %v", err)
	}

	if at.UserID != userID {
		return nil, errTokenInaccessible
	}

	return at, nil
}
