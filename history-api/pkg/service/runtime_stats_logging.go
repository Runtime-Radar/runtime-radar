package service

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/history-api/api"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
)

type RuntimeStatsLogging struct {
	api.RuntimeStatsServer
}

func (sl *RuntimeStatsLogging) CountEvents(ctx context.Context, req *api.RuntimeEventsCounterReq) (resp *api.Counter, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Interface("args", req).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called RuntimeStatsServer.CountEvents")
	}(time.Now())

	resp, err = sl.RuntimeStatsServer.CountEvents(ctx, req)
	return
}

func (sl *RuntimeStatsLogging) DetectorRating(ctx context.Context, req *api.DetectorRatingReq) (resp *api.DetectorRatingResp, err error) {
	defer func(t0 time.Time) {
		corrID, _ := interceptor.CorrelationIDFromContext(ctx)

		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Interface("args", req).
			Interface("result", resp).
			Stringer("correlation_id", corrID).
			Msg("Called RuntimeStatsServer.GetDetectorRating")
	}(time.Now())

	resp, err = sl.RuntimeStatsServer.DetectorRating(ctx, req)
	return
}
