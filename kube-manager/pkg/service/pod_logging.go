package service

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/kube-manager/api"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PodLogging struct {
	api.PodControllerServer
}

func (pl *PodLogging) Get(ctx context.Context, req *api.GetPodReq) (resp *api.GetPodResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("args", req).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called PodControllerServer.Get")
	}(time.Now())

	resp, err = pl.PodControllerServer.Get(ctx, req)
	return
}

func (pl *PodLogging) ListMeta(ctx context.Context, req *api.ListPodMetaReq) (resp *api.ListPodMetaResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called PodControllerServer.ListMeta")
	}(time.Now())

	resp, err = pl.PodControllerServer.ListMeta(ctx, req)
	return
}

func (pl *PodLogging) ListPage(ctx context.Context, req *api.ListPodPageReq) (resp *api.ListPodPageResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called PodControllerServer.ListPage")
	}(time.Now())

	resp, err = pl.PodControllerServer.ListPage(ctx, req)
	return
}

func (pl *PodLogging) Kill(ctx context.Context, req *api.KillPodReq) (resp *emptypb.Empty, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Any("args", req).
			Any("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called PodControllerServer.Kill")
	}(time.Now())

	resp, err = pl.PodControllerServer.Kill(ctx, req)
	return
}
