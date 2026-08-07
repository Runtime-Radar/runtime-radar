package service

import (
	"context"

	"github.com/runtime-radar/runtime-radar/kube-manager/api"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ConfigAuth struct {
	api.ConfigControllerClient

	Verifier jwt.Verifier
}

func (c *ConfigAuth) Add(ctx context.Context, req *api.Config, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	if err := c.Verifier.VerifyPermission(ctx, jwt.PermissionSystemSettings, jwt.ActionCreate, jwt.ActionUpdate); err != nil {
		return nil, auth.PermissionErrorToStatus(err)
	}

	return c.ConfigControllerClient.Add(ctx, req, opts...)
}

func (c *ConfigAuth) Read(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*api.Config, error) {
	if err := c.Verifier.VerifyPermission(ctx, jwt.PermissionSystemSettings, jwt.ActionRead); err != nil {
		return nil, auth.PermissionErrorToStatus(err)
	}

	return c.ConfigControllerClient.Read(ctx, req, opts...)
}
