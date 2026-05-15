package service

import (
	"context"

	"github.com/runtime-radar/runtime-radar/event-processor/api"
	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	lib_context "github.com/runtime-radar/runtime-radar/lib/security/jwt/context"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ConfigAuth is a layer for jwt-based authentication.
// Base server interface should not be embedded here unlike
// in implementations of other layers.
// All required methods should be explicitly implemented to ensure
// that new methods of the basic server are implemented for auth layer.
type ConfigAuth struct {
	api.UnsafeConfigControllerServer

	ConfigControllerServer api.ConfigControllerServer
	Verifier               jwt.Verifier
}

func (ca *ConfigAuth) Add(ctx context.Context, req *api.Config) (resp *emptypb.Empty, err error) {
	if err := ca.Verifier.VerifyPermission(ctx, jwt.PermissionSystemSettings, jwt.ActionCreate, jwt.ActionUpdate); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	lib_context.SetUserID(ctx)

	resp, err = ca.ConfigControllerServer.Add(ctx, req)
	return
}

func (ca *ConfigAuth) Read(ctx context.Context, req *emptypb.Empty) (resp *api.Config, err error) {
	if err := ca.Verifier.VerifyPermission(ctx, jwt.PermissionSystemSettings, jwt.ActionRead); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	lib_context.SetUserID(ctx)

	resp, err = ca.ConfigControllerServer.Read(ctx, req)
	return
}
