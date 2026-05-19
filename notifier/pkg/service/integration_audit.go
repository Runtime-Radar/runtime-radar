package service

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	lib_context "github.com/runtime-radar/runtime-radar/lib/security/jwt/context"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	"github.com/runtime-radar/runtime-radar/notifier/api"
	"google.golang.org/protobuf/types/known/emptypb"
)

type IntegrationAudit struct {
	api.IntegrationControllerServer
}

func extractIDs(ctx context.Context) (string, string, bool) {
	corrID, _ := interceptor.CorrelationIDFromContext(ctx)
	token, _ := jwt.UnverifiedTokenFromContext(ctx)
	tokenUserID := token.GetUserID()
	userID, _ := lib_context.GetUserID(ctx)
	authorized := userID != "" && userID == tokenUserID

	return corrID.String(), tokenUserID, authorized
}

func (ia *IntegrationAudit) Create(ctx context.Context, req *api.Integration) (resp *api.CreateIntegrationResp, err error) {
	ctx = lib_context.WithEmptyUserID(ctx)

	defer func(t0 time.Time) {
		corrID, userID, authorized := extractIDs(ctx)

		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Bool("audit", true).
			Bool("authorized", authorized).
			Str("user_id", userID).
			Str("correlation_id", corrID).
			Str("object", "integration").
			Str("operation", "create").
			Interface("args", hidePassword(req)).
			Interface("result", resp).
			Send()
	}(time.Now())

	resp, err = ia.IntegrationControllerServer.Create(ctx, req)
	return
}

func (ia *IntegrationAudit) Update(ctx context.Context, req *api.Integration) (resp *emptypb.Empty, err error) {
	ctx = lib_context.WithEmptyUserID(ctx)

	defer func(t0 time.Time) {
		corrID, userID, authorized := extractIDs(ctx)

		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Bool("audit", true).
			Bool("authorized", authorized).
			Str("user_id", userID).
			Str("correlation_id", corrID).
			Str("object", "integration").
			Str("operation", "update").
			Interface("args", hidePassword(req)).
			Interface("result", resp).
			Send()
	}(time.Now())

	resp, err = ia.IntegrationControllerServer.Update(ctx, req)
	return
}

func (ia *IntegrationAudit) Delete(ctx context.Context, req *api.DeleteIntegrationReq) (resp *emptypb.Empty, err error) {
	ctx = lib_context.WithEmptyUserID(ctx)

	defer func(t0 time.Time) {
		corrID, userID, authorized := extractIDs(ctx)

		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Bool("audit", true).
			Bool("authorized", authorized).
			Str("user_id", userID).
			Str("correlation_id", corrID).
			Str("object", "integration").
			Str("operation", "delete").
			Interface("args", req).
			Interface("result", resp).
			Send()
	}(time.Now())

	resp, err = ia.IntegrationControllerServer.Delete(ctx, req)
	return
}
