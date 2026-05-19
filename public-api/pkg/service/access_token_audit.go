package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	lib_context "github.com/runtime-radar/runtime-radar/lib/security/jwt/context"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/model"
)

type AccessTokenAudit struct {
	AccessToken
}

func extractIDs(ctx context.Context) (string, string, bool) {
	corrID, _ := interceptor.CorrelationIDFromContext(ctx)
	token, _ := jwt.UnverifiedTokenFromContext(ctx)
	tokenUserID := token.GetUserID()
	userID, _ := lib_context.GetUserID(ctx)
	authorized := userID != "" && userID == tokenUserID

	return corrID.String(), tokenUserID, authorized
}

func (a *AccessTokenAudit) Create(ctx context.Context, req *model.CreateAccessTokenReq) (id uuid.UUID, token string, err error) {
	ctx = lib_context.WithEmptyUserID(ctx)

	defer func(t0 time.Time) {
		corrID, userID, authorized := extractIDs(ctx)

		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Bool("audit", true).
			Bool("authorized", authorized).
			Str("user_id", userID).
			Str("correlation_id", corrID).
			Str("object", "token").
			Str("operation", "create").
			Interface("args", req).
			Stringer("token_id", id).
			Send()
	}(time.Now())

	id, token, err = a.AccessToken.Create(ctx, req)
	return
}

func (a *AccessTokenAudit) Delete(ctx context.Context, id uuid.UUID) (err error) {
	ctx = lib_context.WithEmptyUserID(ctx)

	defer func(t0 time.Time) {
		corrID, userID, authorized := extractIDs(ctx)

		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Bool("audit", true).
			Bool("authorized", authorized).
			Str("user_id", userID).
			Str("correlation_id", corrID).
			Stringer("token_id", id).
			Str("object", "token").
			Str("operation", "revoke").
			Send()
	}(time.Now())

	err = a.AccessToken.Delete(ctx, id)
	return
}

func (a *AccessTokenAudit) InvalidateAll(ctx context.Context) (err error) {
	ctx = lib_context.WithEmptyUserID(ctx)

	defer func(t0 time.Time) {
		corrID, userID, authorized := extractIDs(ctx)

		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Bool("audit", true).
			Bool("authorized", authorized).
			Str("user_id", userID).
			Str("correlation_id", corrID).
			Str("object", "token").
			Str("operation", "revokeAll").
			Send()
	}(time.Now())

	err = a.AccessToken.InvalidateAll(ctx)
	return
}
