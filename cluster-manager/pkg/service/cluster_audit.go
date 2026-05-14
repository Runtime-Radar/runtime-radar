package service

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/cluster-manager/api"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	lib_context "github.com/runtime-radar/runtime-radar/lib/security/jwt/context"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ClusterAudit struct {
	api.ClusterControllerServer
}

func extractIDs(ctx context.Context) (string, string, bool) {
	corrID, _ := interceptor.CorrelationIDFromContext(ctx)
	token, _ := jwt.UnverifiedTokenFromContext(ctx)
	tokenUserID := token.GetUserID()
	userID, _ := lib_context.GetUserID(ctx)
	authorized := userID != "" && userID == tokenUserID

	return corrID.String(), tokenUserID, authorized
}

func (ca *ClusterAudit) Create(ctx context.Context, req *api.Cluster) (resp *api.CreateClusterResp, err error) {
	ctx = lib_context.WithEmptyUserID(ctx)

	defer func(t0 time.Time) {
		corrID, userID, authorized := extractIDs(ctx)
		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Bool("audit", true).
			Bool("authorized", authorized).
			Str("user_id", userID).
			Str("correlation_id", corrID).
			Str("object", "cluster").
			Str("operation", "create").
			Interface("args", req).
			Interface("result", resp).
			Send()
	}(time.Now())

	resp, err = ca.ClusterControllerServer.Create(ctx, req)
	return
}

func (ca *ClusterAudit) Delete(ctx context.Context, req *api.DeleteClusterReq) (resp *emptypb.Empty, err error) {
	ctx = lib_context.WithEmptyUserID(ctx)

	defer func(t0 time.Time) {
		corrID, userID, authorized := extractIDs(ctx)
		log.Err(err).Str("delay", time.Since(t0).String()).
		Bool("audit", true).Bool("authorized", authorized).
			Str("user_id", userID).
			Str("correlation_id", corrID).
			Str("object", "cluster").
			Str("operation", "delete").
			Interface("args", req).
			Interface("result", resp).
			Send()
	}(time.Now())

	resp, err = ca.ClusterControllerServer.Delete(ctx, req)
	return
}

func (ca *ClusterAudit) Register(ctx context.Context, req *api.RegisterClusterReq) (resp *emptypb.Empty, err error) {
	ctx = lib_context.WithEmptyUserID(ctx)

	defer func(t0 time.Time) {
		corrID, userID, authorized := extractIDs(ctx)
		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Bool("audit", true).
			Bool("authorized", authorized).
			Str("user_id", userID).
			Str("correlation_id", corrID).
			Str("object", "cluster").
			Str("operation", "register").
			Interface("args", req).
			Interface("result", resp).
			Send()
	}(time.Now())

	resp, err = ca.ClusterControllerServer.Register(ctx, req)
	return
}

func (ca *ClusterAudit) Unregister(ctx context.Context, req *api.UnregisterClusterReq) (resp *emptypb.Empty, err error) {
	ctx = lib_context.WithEmptyUserID(ctx)

	defer func(t0 time.Time) {
		corrID, userID, authorized := extractIDs(ctx)
		log.Err(err).
			Str("delay", time.Since(t0).String()).
			Bool("audit", true).
			Bool("authorized", authorized).
			Str("user_id", userID).
			Str("correlation_id", corrID).
			Str("object", "cluster").
			Str("operation", "unregister").
			Interface("args", req).
			Interface("result", resp).
			Send()
	}(time.Now())

	resp, err = ca.ClusterControllerServer.Unregister(ctx, req)
	return
}
