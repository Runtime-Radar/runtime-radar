package service

import (
	"context"

	history_api "github.com/runtime-radar/runtime-radar/history-api/api"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/auth"
	"google.golang.org/grpc"
)

type RuntimeHistoryAuth struct {
	history_api.RuntimeHistoryClient

	Verifier jwt.Verifier
}

func (rha *RuntimeHistoryAuth) ListRuntimeEventSlice(ctx context.Context, req *history_api.ListRuntimeEventSliceReq, opts ...grpc.CallOption) (*history_api.ListRuntimeEventSliceResp, error) {
	if err := rha.Verifier.VerifyPermission(ctx, jwt.PermissionEvents, jwt.ActionRead); err != nil {
		return nil, auth.PermissionErrorToStatus(err)
	}

	return rha.RuntimeHistoryClient.ListRuntimeEventSlice(ctx, req, opts...)
}
