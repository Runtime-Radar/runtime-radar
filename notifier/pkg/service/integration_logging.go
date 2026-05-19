package service

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	"github.com/runtime-radar/runtime-radar/notifier/api"
	"google.golang.org/protobuf/types/known/emptypb"
)

type IntegrationLogging struct {
	api.IntegrationControllerServer
}

func (il *IntegrationLogging) Create(ctx context.Context, req *api.Integration) (resp *api.CreateIntegrationResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("args", hidePassword(req)).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called IntegrationControllerServer.Create")
	}(time.Now())

	resp, err = il.IntegrationControllerServer.Create(ctx, req)
	return
}

func (il *IntegrationLogging) Read(ctx context.Context, req *api.ReadIntegrationReq) (resp *api.Integration, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("args", req).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called IntegrationControllerServer.Read")
	}(time.Now())

	resp, err = il.IntegrationControllerServer.Read(ctx, req)
	return
}

func (il *IntegrationLogging) Update(ctx context.Context, req *api.Integration) (resp *emptypb.Empty, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("args", hidePassword(req)).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called IntegrationControllerServer.Update")
	}(time.Now())

	resp, err = il.IntegrationControllerServer.Update(ctx, req)
	return
}

func (il *IntegrationLogging) Delete(ctx context.Context, req *api.DeleteIntegrationReq) (resp *emptypb.Empty, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("args", req).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called IntegrationControllerServer.Delete")
	}(time.Now())

	resp, err = il.IntegrationControllerServer.Delete(ctx, req)
	return
}

func (il *IntegrationLogging) List(ctx context.Context, req *api.ListIntegrationReq) (resp *api.ListIntegrationResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("args", req).
			Int("result", len(resp.GetIntegrations())).
			Stringer("correlation_id", corrID).
			Msg("Called IntegrationControllerServer.List")
	}(time.Now())

	resp, err = il.IntegrationControllerServer.List(ctx, req)
	return
}
