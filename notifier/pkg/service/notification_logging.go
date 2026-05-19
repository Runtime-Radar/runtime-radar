package service

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	"github.com/runtime-radar/runtime-radar/notifier/api"
	"google.golang.org/protobuf/types/known/emptypb"
)

type NotificationLogging struct {
	api.NotificationControllerServer
}

func (nl *NotificationLogging) Create(ctx context.Context, req *api.Notification) (resp *api.CreateNotificationResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("args", req).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called NotificationControllerServer.Create")
	}(time.Now())

	resp, err = nl.NotificationControllerServer.Create(ctx, req)
	return
}

func (nl *NotificationLogging) Read(ctx context.Context, req *api.ReadNotificationReq) (resp *api.ReadNotificationResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("args", req).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called NotificationControllerServer.Read")
	}(time.Now())

	resp, err = nl.NotificationControllerServer.Read(ctx, req)
	return
}

func (nl *NotificationLogging) Update(ctx context.Context, req *api.Notification) (resp *emptypb.Empty, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Bool("audit", true).
			Interface("args", req).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called NotificationControllerServer.Update")
	}(time.Now())

	resp, err = nl.NotificationControllerServer.Update(ctx, req)
	return
}

func (nl *NotificationLogging) Delete(ctx context.Context, req *api.DeleteNotificationReq) (resp *emptypb.Empty, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Bool("audit", true).
			Interface("args", req).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called NotificationControllerServer.Delete")
	}(time.Now())

	resp, err = nl.NotificationControllerServer.Delete(ctx, req)
	return
}

func (nl *NotificationLogging) List(ctx context.Context, req *api.ListNotificationReq) (resp *api.ListNotificationResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("args", req).
			Int("result", len(resp.GetNotifications())).
			Stringer("correlation_id", corrID).
			Msg("Called NotificationControllerServer.List")
	}(time.Now())

	resp, err = nl.NotificationControllerServer.List(ctx, req)
	return
}

func (nl *NotificationLogging) DefaultTemplate(ctx context.Context, req *api.DefaultTemplateReq) (resp *api.DefaultTemplateResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).Str("delay", time.Since(t0).String()).
			Interface("args", req).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called NotificationControllerServer.DefaultTemplate")
	}(time.Now())

	resp, err = nl.NotificationControllerServer.DefaultTemplate(ctx, req)
	return
}
