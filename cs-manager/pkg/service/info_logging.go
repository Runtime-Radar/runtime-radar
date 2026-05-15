package service

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/cs-manager/api"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	"google.golang.org/protobuf/types/known/emptypb"
)

type InfoLogging struct {
	api.InfoControllerServer
}

func (il *InfoLogging) GetVersion(ctx context.Context, req *emptypb.Empty) (resp *api.Version, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called InfoControllerServer.GetVersion")
	}(time.Now())

	resp, err = il.InfoControllerServer.GetVersion(ctx, req)
	return
}

func (il *InfoLogging) GetCentralCSURL(ctx context.Context, req *emptypb.Empty) (resp *api.URL, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called InfoControllerServer.GetCentralCSURL")
	}(time.Now())

	resp, err = il.InfoControllerServer.GetCentralCSURL(ctx, req)
	return
}

func (il *InfoLogging) GetGrafanaURL(ctx context.Context, req *emptypb.Empty) (resp *api.URL, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called InfoControllerServer.GetGrafanaURL")
	}(time.Now())

	resp, err = il.InfoControllerServer.GetGrafanaURL(ctx, req)
	return
}
