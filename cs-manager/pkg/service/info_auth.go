package service

import (
	"context"

	"github.com/runtime-radar/runtime-radar/cs-manager/api"
	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	lib_context "github.com/runtime-radar/runtime-radar/lib/security/jwt/context"
	"google.golang.org/protobuf/types/known/emptypb"
)

// InfoAuth is a layer for jwt-based authentication.
// Base server interface should not be embedded here unlike
// in implementations of other layers.
// All required methods should be explicitly implemented to ensure
// that new methods of the basic server are implemented for auth layer.
type InfoAuth struct {
	// UnsafeInfoControllerServer is embedded to opt out of forward
	// compatibility promised by protobuf library.
	// It merely contains an empty `mustEmbedUnimplementedInfoControllerServer()`
	// method.
	api.UnsafeInfoControllerServer

	// InfoControllerServer is a base server interface to pass
	// response to the next layer.
	InfoControllerServer api.InfoControllerServer
	Verifier             jwt.Verifier
}

func (ia *InfoAuth) GetVersion(ctx context.Context, req *emptypb.Empty) (resp *api.Version, err error) {
	if err := ia.Verifier.VerifyPermission(ctx, jwt.PermissionSystemSettings, jwt.ActionRead); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	lib_context.SetUserID(ctx)

	resp, err = ia.InfoControllerServer.GetVersion(ctx, req)
	return
}

func (ia *InfoAuth) GetCentralCSURL(ctx context.Context, req *emptypb.Empty) (resp *api.URL, err error) {
	resp, err = ia.InfoControllerServer.GetCentralCSURL(ctx, req)
	return
}

func (ia *InfoAuth) GetGrafanaURL(ctx context.Context, req *emptypb.Empty) (resp *api.URL, err error) {
	resp, err = ia.InfoControllerServer.GetGrafanaURL(ctx, req)
	return
}
