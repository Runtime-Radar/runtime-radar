package service

import (
	"context"

	"github.com/runtime-radar/runtime-radar/event-processor/api"
	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	lib_context "github.com/runtime-radar/runtime-radar/lib/security/jwt/context"
	"google.golang.org/protobuf/types/known/emptypb"
)

// DetectorAuth is a layer for jwt-based authentication.
// Base server interface should not be embedded here unlike
// in implementations of other layers.
// All required methods should be explicitly implemented to ensure
// that new methods of the basic server are implemented for auth layer.
type DetectorAuth struct {
	// UnsafeEmailControllerServer is embedded to opt out of forward
	// compatibility promised by protobuf library.
	// It merely contains an empty `mustEmbedUnimplementedEmailControllerServer()`
	// method.
	api.UnsafeDetectorControllerServer

	// DetectorControllerServer is a base server interface to pass
	// response to the next layer.
	DetectorControllerServer api.DetectorControllerServer
	Verifier                 jwt.Verifier
}

func (da *DetectorAuth) Create(ctx context.Context, req *api.CreateDetectorReq) (resp *api.CreateDetectorResp, err error) {
	if err := da.Verifier.VerifyPermission(ctx, jwt.PermissionSystemSettings, jwt.ActionCreate); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	lib_context.SetUserID(ctx)

	resp, err = da.DetectorControllerServer.Create(ctx, req)
	return
}

func (da *DetectorAuth) Delete(ctx context.Context, req *api.DeleteDetectorReq) (resp *emptypb.Empty, err error) {
	if err := da.Verifier.VerifyPermission(ctx, jwt.PermissionSystemSettings, jwt.ActionDelete); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	lib_context.SetUserID(ctx)

	resp, err = da.DetectorControllerServer.Delete(ctx, req)
	return
}

func (da *DetectorAuth) ListPage(ctx context.Context, req *api.ListDetectorPageReq) (resp *api.ListDetectorPageResp, err error) {
	if err := da.Verifier.VerifyPermission(ctx, jwt.PermissionSystemSettings, jwt.ActionRead); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	lib_context.SetUserID(ctx)

	resp, err = da.DetectorControllerServer.ListPage(ctx, req)
	return
}
