package service

import (
	"context"

	"github.com/runtime-radar/runtime-radar/auth-center/api"
	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserAuth struct {
	api.UnsafeUserControllerServer
	UserControllerServer api.UserControllerServer
	Verifier             jwt.Verifier
}

func (ua *UserAuth) Read(ctx context.Context, req *api.ReadUserReq) (resp *api.UserResp, err error) {
	if err := ua.Verifier.VerifyPermission(ctx, jwt.PermissionUsers, jwt.ActionRead); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}
	resp, err = ua.UserControllerServer.Read(ctx, req)
	return
}

func (ua *UserAuth) ReadList(ctx context.Context, empty *emptypb.Empty) (resp *api.UserListResp, err error) {
	if err := ua.Verifier.VerifyPermission(ctx, jwt.PermissionUsers, jwt.ActionRead); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}
	resp, err = ua.UserControllerServer.ReadList(ctx, empty)
	return
}

func (ua *UserAuth) Create(ctx context.Context, req *api.CreateUserReq) (resp *api.UserResp, err error) {
	if err := ua.Verifier.VerifyPermission(ctx, jwt.PermissionUsers, jwt.ActionCreate); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}
	resp, err = ua.UserControllerServer.Create(ctx, req)
	return
}

func (ua *UserAuth) Update(ctx context.Context, req *api.UpdateUserReq) (resp *api.UserResp, err error) {
	if err := ua.Verifier.VerifyPermission(ctx, jwt.PermissionUsers, jwt.ActionUpdate); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}
	resp, err = ua.UserControllerServer.Update(ctx, req)
	return
}

func (ua *UserAuth) Delete(ctx context.Context, req *api.DeleteUserReq) (resp *api.DeleteUserResp, err error) {
	if err := ua.Verifier.VerifyPermission(ctx, jwt.PermissionUsers, jwt.ActionDelete); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}
	resp, err = ua.UserControllerServer.Delete(ctx, req)
	return
}

func (ua *UserAuth) ChangePassword(ctx context.Context, req *api.ChangePasswordReq) (resp *api.SignInResp, err error) {
	if err := ua.Verifier.VerifyPermission(ctx, jwt.PermissionUsers, jwt.ActionExecute); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}
	resp, err = ua.UserControllerServer.ChangePassword(ctx, req)
	return
}
