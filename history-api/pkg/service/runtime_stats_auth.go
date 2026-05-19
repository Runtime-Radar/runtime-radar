package service

import (
	"context"

	"github.com/runtime-radar/runtime-radar/history-api/api"
	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
)

// RuntimeStatsAuth is a layer for jwt-based authentication.
// Base server interface should not be embedded here unlike
// in implementations of other layers.
// All required methods should be explicitly implemented to ensure
// that new methods of the basic server are implemented for auth layer.
type RuntimeStatsAuth struct {
	// UnsafeStatsServer is embedded to opt out of forward
	// compatibility promised by protobuf library.
	// It merely contains an empty `mustEmbedUnimplementedStatsServer()`
	// method.
	api.UnsafeRuntimeStatsServer

	// RuntimeStatsServer is a base server interface to pass
	// response to the next layer.
	RuntimeStatsServer api.RuntimeStatsServer
	Verifier           jwt.Verifier
}

func (sa *RuntimeStatsAuth) CountEvents(ctx context.Context, req *api.RuntimeEventsCounterReq) (resp *api.Counter, err error) {
	if err := sa.Verifier.VerifyPermission(ctx, jwt.PermissionEvents, jwt.ActionRead); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}
	resp, err = sa.RuntimeStatsServer.CountEvents(ctx, req)
	return
}

func (sa *RuntimeStatsAuth) DetectorRating(ctx context.Context, req *api.DetectorRatingReq) (resp *api.DetectorRatingResp, err error) {
	if err := sa.Verifier.VerifyPermission(ctx, jwt.PermissionEvents, jwt.ActionRead); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}
	resp, err = sa.RuntimeStatsServer.DetectorRating(ctx, req)
	return
}
