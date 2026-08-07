package service

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/kube-manager/api"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ConfigLogging struct {
	api.ConfigControllerServer
}

func (cl *ConfigLogging) Add(ctx context.Context, req *api.Config) (resp *emptypb.Empty, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Bool("audit", true).
			Interface("args", req).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("ConfigControllerServer.Add")
	}(time.Now())

	resp, err = cl.ConfigControllerServer.Add(ctx, req)
	return
}

func (cl *ConfigLogging) Read(ctx context.Context, req *emptypb.Empty) (resp *api.Config, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("args", req).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called ConfigControllerServer.Read")
	}(time.Now())

	resp, err = cl.ConfigControllerServer.Read(ctx, req)
	return
}
