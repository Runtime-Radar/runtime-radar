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

type UserLogging struct {
	api.UserControllerServer
}

func (ul *UserLogging) Read(ctx context.Context, req *api.ReadUserReq) (resp *api.UserResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("args", req).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called UserServer.Read")
	}(time.Now())

	resp, err = ul.UserControllerServer.Read(ctx, req)
	return
}

func (ul *UserLogging) ReadList(ctx context.Context, empty *emptypb.Empty) (resp *api.UserListResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("args", empty).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called UserServer.ReadList")
	}(time.Now())

	resp, err = ul.UserControllerServer.ReadList(ctx, empty)
	return
}

func (ul *UserLogging) Create(ctx context.Context, req *api.CreateUserReq) (resp *api.UserResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		masked := proto.Clone(req).(*api.CreateUserReq)
		masked.Password = maskedPassword

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("args", masked).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called UserServer.Create")
	}(time.Now())

	resp, err = ul.UserControllerServer.Create(ctx, req)
	return
}

func (ul *UserLogging) Update(ctx context.Context, req *api.UpdateUserReq) (resp *api.UserResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("args", req).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called UserServer.Update")
	}(time.Now())

	resp, err = ul.UserControllerServer.Update(ctx, req)
	return
}

func (ul *UserLogging) Delete(ctx context.Context, req *api.DeleteUserReq) (resp *api.DeleteUserResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("args", req).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called UserServer.Delete")
	}(time.Now())

	resp, err = ul.UserControllerServer.Delete(ctx, req)
	return
}

func (ul *UserLogging) ChangePassword(ctx context.Context, req *api.ChangePasswordReq) (resp *api.SignInResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		masked := proto.Clone(req).(*api.ChangePasswordReq)
		masked.Password = maskedPassword

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("args", masked).
			Interface("result", maskTokens(resp)).
			Stringer("correlation_id", corrID).
			Msg("Called UserServer.ChangePassword")
	}(time.Now())

	resp, err = ul.UserControllerServer.ChangePassword(ctx, req)
	return
}
