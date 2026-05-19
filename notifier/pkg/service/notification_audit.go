package service

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	lib_context "github.com/runtime-radar/runtime-radar/lib/security/jwt/context"
	"github.com/runtime-radar/runtime-radar/notifier/api"
	"google.golang.org/protobuf/types/known/emptypb"
)

type NotificationAudit struct {
	api.NotificationControllerServer
}

func (na *NotificationAudit) Create(ctx context.Context, req *api.Notification) (resp *api.CreateNotificationResp, err error) {
	ctx = lib_context.WithEmptyUserID(ctx)

	defer func(t0 time.Time) {
		corrID, userID, authorized := extractIDs(ctx)

		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Bool("audit", true).
			Bool("authorized", authorized).
			Str("user_id", userID).
			Str("correlation_id", corrID).
			Str("object", "notification").
			Str("operation", "create").
			Interface("args", req).
			Interface("result", resp).
			Send()
	}(time.Now())

	resp, err = na.NotificationControllerServer.Create(ctx, req)
	return
}

func (na *NotificationAudit) Update(ctx context.Context, req *api.Notification) (resp *emptypb.Empty, err error) {
	ctx = lib_context.WithEmptyUserID(ctx)

	defer func(t0 time.Time) {
		corrID, userID, authorized := extractIDs(ctx)

		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Bool("audit", true).
			Bool("authorized", authorized).
			Str("user_id", userID).
			Str("correlation_id", corrID).
			Str("object", "notification").
			Str("operation", "update").
			Interface("args", req).
			Interface("result", resp).
			Send()
	}(time.Now())

	resp, err = na.NotificationControllerServer.Update(ctx, req)
	return
}

func (na *NotificationAudit) Delete(ctx context.Context, req *api.DeleteNotificationReq) (resp *emptypb.Empty, err error) {
	ctx = lib_context.WithEmptyUserID(ctx)

	defer func(t0 time.Time) {
		corrID, userID, authorized := extractIDs(ctx)

		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Bool("audit", true).
			Bool("authorized", authorized).
			Str("user_id", userID).
			Str("correlation_id", corrID).
			Str("object", "notification").
			Str("operation", "delete").
			Interface("args", req).
			Interface("result", resp).
			Send()
	}(time.Now())

	resp, err = na.NotificationControllerServer.Delete(ctx, req)
	return
}
