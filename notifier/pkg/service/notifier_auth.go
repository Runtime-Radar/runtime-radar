package service

import (
	"context"

	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	lib_context "github.com/runtime-radar/runtime-radar/lib/security/jwt/context"
	"github.com/runtime-radar/runtime-radar/notifier/api"
	"google.golang.org/protobuf/types/known/emptypb"
)

// NotifierAuth is a layer for jwt-based authentication.
// Base server interface should not be embedded here unlike
// in implementations of other layers.
// All required methods should be explicitly implemented to ensure
// that new methods of the basic server are implemented for auth layer.
type NotifierAuth struct {
	api.UnsafeNotifierServer

	NotifierServer api.NotifierServer
	Verifier       jwt.Verifier
}

func (na *NotifierAuth) Notify(ctx context.Context, req *api.NotifyReq) (resp *emptypb.Empty, err error) {
	if err := na.Verifier.VerifyPermission(ctx, jwt.PermissionNotifications, jwt.ActionExecute); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	lib_context.SetUserID(ctx)

	resp, err = na.NotifierServer.Notify(ctx, req)
	return
}
