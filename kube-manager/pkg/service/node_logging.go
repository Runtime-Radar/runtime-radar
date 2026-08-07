package service

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/kube-manager/api"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
)

type NodeLogging struct {
	api.NodeControllerServer
}

func (nl *NodeLogging) Get(ctx context.Context, req *api.GetNodeReq) (resp *api.GetNodeResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("args", req).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called NodeController.Get")
	}(time.Now())

	resp, err = nl.NodeControllerServer.Get(ctx, req)
	return
}

func (nl *NodeLogging) List(ctx context.Context, req *api.ListNodeMetaReq) (resp *api.ListNodeMetaResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called NodeController.ListMeta")
	}(time.Now())

	resp, err = nl.NodeControllerServer.ListMeta(ctx, req)
	return
}

func (nl *NodeLogging) ListPage(ctx context.Context, req *api.ListNodePageReq) (resp *api.ListNodePageResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called NodeController.ListPage")
	}(time.Now())

	resp, err = nl.NodeControllerServer.ListPage(ctx, req)
	return
}
