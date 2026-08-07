package service

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/kube-manager/api"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	"google.golang.org/grpc"
)

type NodeLogging struct {
	api.NodeControllerClient
}

func (n *NodeLogging) Get(ctx context.Context, req *api.GetNodeReq, opts ...grpc.CallOption) (resp *api.GetNodeResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Stringer("correlation_id", corrID).
			Interface("args", req).
			Msg("Called Node.Get")
	}(time.Now())

	resp, err = n.NodeControllerClient.Get(ctx, req, opts...)
	return
}

func (n *NodeLogging) ListMeta(ctx context.Context, req *api.ListNodeMetaReq, opts ...grpc.CallOption) (resp *api.ListNodeMetaResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)
		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Stringer("correlation_id", corrID).
			Int("result_total", int(resp.GetTotal())).
			Interface("args", req).
			Msg("Called Node.ListMeta")
	}(time.Now())

	resp, err = n.NodeControllerClient.ListMeta(ctx, req, opts...)
	return
}

func (n *NodeLogging) ListPage(ctx context.Context, req *api.ListNodePageReq, opts ...grpc.CallOption) (resp *api.ListNodePageResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)
		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Stringer("correlation_id", corrID).
			Interface("args", req).
			Int("result_total", int(resp.GetTotal())).
			Msg("Called Node.ListPage")
	}(time.Now())

	resp, err = n.NodeControllerClient.ListPage(ctx, req, opts...)
	return
}
