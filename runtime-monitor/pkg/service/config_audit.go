package service

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	lib_context "github.com/runtime-radar/runtime-radar/lib/security/jwt/context"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	"github.com/runtime-radar/runtime-radar/runtime-monitor/api"
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

		conf := req.GetConfig()
		tp := make(map[string]*api.TracingPolicy, len(conf.GetTracingPolicies()))
		for name, policy := range conf.GetTracingPolicies() {
			tp[name] = &api.TracingPolicy{
				Name:    policy.GetName(),
				Enabled: policy.GetEnabled(),
			}
		}

		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Bool("audit", true).
			Bool("authorized", authorized).
			Str("user_id", userID).
			Str("correlation_id", corrID).
			Str("object", "runtime_monitor_config").
			Str("operation", "update").
			Interface("args", &api.Config{
				Id: req.GetId(),
				Config: &api.Config_ConfigJSON{
					Version:         conf.GetVersion(),
					TracingPolicies: tp,
				},
			}).
			Interface("result", resp).
			Send()
	}(time.Now())

	resp, err = cl.ConfigControllerServer.Add(ctx, req)
	return
}
