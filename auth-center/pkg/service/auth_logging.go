package service

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/auth-center/api"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthLogging struct {
	api.AuthControllerServer
}

func (al *AuthLogging) SignIn(ctx context.Context, req *api.SignInReq) (resp *api.SignInResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		masked := proto.Clone(req).(*api.SignInReq)
		masked.Password = maskedPassword

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("args", masked).
			Interface("result", maskTokens(resp)).
			Stringer("correlation_id", corrID).
			Msg("Called AuthServer.SignIn")
	}(time.Now())

	resp, err = al.AuthControllerServer.SignIn(ctx, req)
	return
}

func (al *AuthLogging) RefreshTokens(ctx context.Context, empty *emptypb.Empty) (resp *api.SignInResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("args", empty).
			Interface("result", maskTokens(resp)).
			Stringer("correlation_id", corrID).
			Msg("Called AuthServer.RefreshTokens")
	}(time.Now())

	resp, err = al.AuthControllerServer.RefreshTokens(ctx, empty)
	return
}
