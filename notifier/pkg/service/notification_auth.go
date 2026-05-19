package service

import (
	"context"

	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	lib_context "github.com/runtime-radar/runtime-radar/lib/security/jwt/context"
	"github.com/runtime-radar/runtime-radar/notifier/api"
	"google.golang.org/protobuf/types/known/emptypb"
)

// NotificationAuth is a layer for jwt-based authentication.
// Base server interface should not be embedded here unlike
// in implementations of other layers.
// All required methods should be explicitly implemented to ensure
// that new methods of the basic server are implemented for auth layer.
type NotificationAuth struct {
	api.UnsafeNotificationControllerServer

	NotificationControllerServer api.NotificationControllerServer
	Verifier                     jwt.Verifier
}

func (na *NotificationAuth) Create(ctx context.Context, req *api.Notification) (resp *api.CreateNotificationResp, err error) {
	if err := na.Verifier.VerifyPermission(ctx, jwt.PermissionNotifications, jwt.ActionCreate); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	lib_context.SetUserID(ctx)

	resp, err = na.NotificationControllerServer.Create(ctx, req)
	return
}

func (na *NotificationAuth) Read(ctx context.Context, req *api.ReadNotificationReq) (resp *api.ReadNotificationResp, err error) {
	if err := na.Verifier.VerifyPermission(ctx, jwt.PermissionNotifications, jwt.ActionRead); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	lib_context.SetUserID(ctx)

	resp, err = na.NotificationControllerServer.Read(ctx, req)
	return
}

func (na *NotificationAuth) Update(ctx context.Context, req *api.Notification) (resp *emptypb.Empty, err error) {
	if err := na.Verifier.VerifyPermission(ctx, jwt.PermissionNotifications, jwt.ActionUpdate); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	lib_context.SetUserID(ctx)

	resp, err = na.NotificationControllerServer.Update(ctx, req)
	return
}

func (na *NotificationAuth) Delete(ctx context.Context, req *api.DeleteNotificationReq) (resp *emptypb.Empty, err error) {
	if err := na.Verifier.VerifyPermission(ctx, jwt.PermissionNotifications, jwt.ActionDelete); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	lib_context.SetUserID(ctx)

	resp, err = na.NotificationControllerServer.Delete(ctx, req)
	return
}

func (na *NotificationAuth) List(ctx context.Context, req *api.ListNotificationReq) (resp *api.ListNotificationResp, err error) {
	if err := na.Verifier.VerifyPermission(ctx, jwt.PermissionNotifications, jwt.ActionRead); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	lib_context.SetUserID(ctx)

	resp, err = na.NotificationControllerServer.List(ctx, req)
	return
}

func (na *NotificationAuth) DefaultTemplate(ctx context.Context, req *api.DefaultTemplateReq) (resp *api.DefaultTemplateResp, err error) {
	if err := na.Verifier.VerifyPermission(ctx, jwt.PermissionNotifications, jwt.ActionRead); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	lib_context.SetUserID(ctx)

	resp, err = na.NotificationControllerServer.DefaultTemplate(ctx, req)
	return
}
