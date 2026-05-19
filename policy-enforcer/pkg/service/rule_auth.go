package service

import (
	"context"

	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	lib_context "github.com/runtime-radar/runtime-radar/lib/security/jwt/context"
	"github.com/runtime-radar/runtime-radar/policy-enforcer/api"
	"google.golang.org/protobuf/types/known/emptypb"
)

// RuleAuth is a layer for jwt-based authentication.
// Base server interface should not be embedded here unlike
// in implementations of other layers.
// All required methods should be explicitly implemented to ensure
// that new methods of the basic server are implemented for auth layer.
type RuleAuth struct {
	// UnsafeEmailControllerServer is embedded to opt out of forward
	// compatibility promised by protobuf library.
	// It merely contains an empty `mustEmbedUnimplementedEmailControllerServer()`
	// method.
	api.UnsafeRuleControllerServer

	// RuleControllerServer is a base server interface to pass
	// response to the next layer.
	RuleControllerServer api.RuleControllerServer
	Verifier             jwt.Verifier
}

func (ra *RuleAuth) Create(ctx context.Context, req *api.Rule) (resp *api.CreateRuleResp, err error) {
	if err := ra.Verifier.VerifyPermission(ctx, jwt.PermissionRules, jwt.ActionCreate); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	lib_context.SetUserID(ctx)

	resp, err = ra.RuleControllerServer.Create(ctx, req)
	return
}

func (ra *RuleAuth) Read(ctx context.Context, req *api.ReadRuleReq) (resp *api.ReadRuleResp, err error) {
	if err := ra.Verifier.VerifyPermission(ctx, jwt.PermissionRules, jwt.ActionRead); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	lib_context.SetUserID(ctx)

	resp, err = ra.RuleControllerServer.Read(ctx, req)
	return
}

func (ra *RuleAuth) Update(ctx context.Context, req *api.Rule) (resp *emptypb.Empty, err error) {
	if err := ra.Verifier.VerifyPermission(ctx, jwt.PermissionRules, jwt.ActionUpdate); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	lib_context.SetUserID(ctx)

	resp, err = ra.RuleControllerServer.Update(ctx, req)
	return
}

func (ra *RuleAuth) Delete(ctx context.Context, req *api.DeleteRuleReq) (resp *emptypb.Empty, err error) {
	if err := ra.Verifier.VerifyPermission(ctx, jwt.PermissionRules, jwt.ActionDelete); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	lib_context.SetUserID(ctx)

	resp, err = ra.RuleControllerServer.Delete(ctx, req)
	return
}

func (ra *RuleAuth) ListPage(ctx context.Context, req *api.ListRulePageReq) (resp *api.ListRulePageResp, err error) {
	if err := ra.Verifier.VerifyPermission(ctx, jwt.PermissionRules, jwt.ActionRead); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	lib_context.SetUserID(ctx)

	resp, err = ra.RuleControllerServer.ListPage(ctx, req)
	return
}

func (ra *RuleAuth) NotifyTargetsInUse(ctx context.Context, req *api.NotifyTargetsInUseReq) (resp *api.NotifyTargetsInUseResp, err error) {
	if err := ra.Verifier.VerifyPermission(ctx, jwt.PermissionRules, jwt.ActionRead); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	lib_context.SetUserID(ctx)

	resp, err = ra.RuleControllerServer.NotifyTargetsInUse(ctx, req)
	return
}
