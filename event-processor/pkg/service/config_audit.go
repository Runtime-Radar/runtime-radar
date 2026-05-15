package service

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/event-processor/api"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	lib_context "github.com/runtime-radar/runtime-radar/lib/security/jwt/context"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ConfigAudit struct {
	api.ConfigControllerServer
}

func extractIDs(ctx context.Context) (string, string, bool) {
	corrID, _ := interceptor.CorrelationIDFromContext(ctx)
	token, _ := jwt.UnverifiedTokenFromContext(ctx)
	tokenUserID := token.GetUserID()
	userID, _ := lib_context.GetUserID(ctx)
	authorized := userID != "" && userID == tokenUserID

	return corrID.String(), tokenUserID, authorized
}

func (cl *ConfigAudit) Add(ctx context.Context, req *api.Config) (resp *emptypb.Empty, err error) {
	ctx = lib_context.WithEmptyUserID(ctx)

	defer func(t0 time.Time) {
		corrID, userID, authorized := extractIDs(ctx)

		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Bool("audit", true).
			Bool("authorized", authorized).
			Str("user_id", userID).
			Str("correlation_id", corrID).
			Str("object", "event_processor_config").
			Str("operation", "update").
			Interface("args", req).
			Interface("result", resp).
			Send()
	}(time.Now())

	resp, err = cl.ConfigControllerServer.Add(ctx, req)
	return
}
