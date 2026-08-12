package service

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/kube-manager/api"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ConfigLogging struct {
	api.ConfigControllerClient
}

func (c *ConfigLogging) Add(ctx context.Context, req *api.Config, opts ...grpc.CallOption) (resp *emptypb.Empty, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Stringer("correlation_id", corrID).
			Interface("args", req).
			Msg("Called Config.Add")
	}(time.Now())

	resp, err = c.ConfigControllerClient.Add(ctx, req, opts...)
	return
}

func (c *ConfigLogging) Read(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (resp *api.Config, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Stringer("correlation_id", corrID).
			Msg("Called Config.Read")
	}(time.Now())

	resp, err = c.ConfigControllerClient.Read(ctx, req, opts...)
	return
}
