package service

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	lib_context "github.com/runtime-radar/runtime-radar/lib/security/jwt/context"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	"github.com/runtime-radar/runtime-radar/policy-enforcer/api"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RuleAudit struct {
	api.RuleControllerServer
}

func extractIDs(ctx context.Context) (string, string, bool) {
	corrID, _ := interceptor.CorrelationIDFromContext(ctx)
	token, _ := jwt.UnverifiedTokenFromContext(ctx)
	tokenUserID := token.GetUserID()
	userID, _ := lib_context.GetUserID(ctx)
	authorized := userID != "" && userID == tokenUserID

	return corrID.String(), tokenUserID, authorized
}

func (ra *RuleAudit) Create(ctx context.Context, req *api.Rule) (resp *api.CreateRuleResp, err error) {
	ctx = lib_context.WithEmptyUserID(ctx)

	defer func(t0 time.Time) {
		corrID, userID, authorized := extractIDs(ctx)

		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Bool("audit", true).
			Bool("authorized", authorized).
			Str("user_id", userID).
			Str("correlation_id", corrID).
			Str("object", "rule").
			Str("operation", "create").
			Interface("args", req).
			Interface("result", resp).
			Send()
	}(time.Now())

	resp, err = ra.RuleControllerServer.Create(ctx, req)
	return
}

func (ra *RuleAudit) Update(ctx context.Context, req *api.Rule) (resp *emptypb.Empty, err error) {
	ctx = lib_context.WithEmptyUserID(ctx)

	defer func(t0 time.Time) {
		corrID, userID, authorized := extractIDs(ctx)

		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Bool("audit", true).
			Bool("authorized", authorized).
			Str("user_id", userID).
			Str("correlation_id", corrID).
			Str("object", "rule").
			Str("operation", "update").
			Interface("args", req).
			Interface("result", resp).
			Send()
	}(time.Now())

	resp, err = ra.RuleControllerServer.Update(ctx, req)
	return
}

func (ra *RuleAudit) Delete(ctx context.Context, req *api.DeleteRuleReq) (resp *emptypb.Empty, err error) {
	ctx = lib_context.WithEmptyUserID(ctx)

	defer func(t0 time.Time) {
		corrID, userID, authorized := extractIDs(ctx)

		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Bool("audit", true).
			Bool("authorized", authorized).
			Str("user_id", userID).
			Str("correlation_id", corrID).
			Str("object", "rule").
			Str("operation", "delete").
			Interface("args", req).
			Interface("result", resp).
			Send()
	}(time.Now())

	resp, err = ra.RuleControllerServer.Delete(ctx, req)
	return
}
