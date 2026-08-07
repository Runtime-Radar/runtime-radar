package service

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/kube-manager/api"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	"google.golang.org/grpc"
)

type PodLogging struct {
	Pod
}

func (p *PodLogging) Get(ctx context.Context, req *api.GetPodReq, opts ...grpc.CallOption) (resp *api.GetPodResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Stringer("correlation_id", corrID).
			Interface("args", req).
			Msg("Called Pod.Get")
	}(time.Now())

	resp, err = p.Pod.Get(ctx, req, opts...)
	return
}

func (p *PodLogging) ListMeta(ctx context.Context, req *api.ListPodMetaReq, opts ...grpc.CallOption) (resp *api.ListPodMetaResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)
		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Stringer("correlation_id", corrID).
			Int("result_total", int(resp.GetTotal())).
			Interface("args", req).
			Msg("Called Pod.ListMeta")
	}(time.Now())

	resp, err = p.Pod.ListMeta(ctx, req, opts...)
	return
}

func (p *PodLogging) ListPage(ctx context.Context, req *api.ListPodPageReq, opts ...grpc.CallOption) (resp *api.ListPodPageResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)
		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Stringer("correlation_id", corrID).
			Int("result_total", int(resp.GetTotal())).
			Interface("args", req).
			Msg("Called Pod.ListPage")
	}(time.Now())

	resp, err = p.Pod.ListPage(ctx, req, opts...)
	return
}
