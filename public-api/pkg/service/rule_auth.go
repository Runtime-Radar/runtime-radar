package service

import (
	"context"

	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	enf_api "github.com/runtime-radar/runtime-radar/policy-enforcer/api"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RuleAuth struct {
	enf_api.RuleControllerClient

	Verifier jwt.Verifier
}

func (ra *RuleAuth) Create(ctx context.Context, req *enf_api.Rule, opts ...grpc.CallOption) (*enf_api.CreateRuleResp, error) {
	if err := ra.Verifier.VerifyPermission(ctx, jwt.PermissionRules, jwt.ActionCreate); err != nil {
		return nil, auth.PermissionErrorToStatus(err)
	}

	return ra.RuleControllerClient.Create(ctx, req, opts...)
}

func (ra *RuleAuth) Read(ctx context.Context, req *enf_api.ReadRuleReq, opts ...grpc.CallOption) (*enf_api.ReadRuleResp, error) {
	if err := ra.Verifier.VerifyPermission(ctx, jwt.PermissionRules, jwt.ActionRead); err != nil {
		return nil, auth.PermissionErrorToStatus(err)
	}

	return ra.RuleControllerClient.Read(ctx, req, opts...)
}

func (ra *RuleAuth) Update(ctx context.Context, req *enf_api.Rule, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	if err := ra.Verifier.VerifyPermission(ctx, jwt.PermissionRules, jwt.ActionUpdate); err != nil {
		return nil, auth.PermissionErrorToStatus(err)
	}

	return ra.RuleControllerClient.Update(ctx, req, opts...)
}

func (ra *RuleAuth) Delete(ctx context.Context, req *enf_api.DeleteRuleReq, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	if err := ra.Verifier.VerifyPermission(ctx, jwt.PermissionRules, jwt.ActionDelete); err != nil {
		return nil, auth.PermissionErrorToStatus(err)
	}

	return ra.RuleControllerClient.Delete(ctx, req, opts...)
}

func (ra *RuleAuth) ListPage(ctx context.Context, req *enf_api.ListRulePageReq, opts ...grpc.CallOption) (*enf_api.ListRulePageResp, error) {
	if err := ra.Verifier.VerifyPermission(ctx, jwt.PermissionRules, jwt.ActionRead); err != nil {
		return nil, auth.PermissionErrorToStatus(err)
	}

	return ra.RuleControllerClient.ListPage(ctx, req, opts...)
}

func (ra RuleAuth) NotifyTargetsInUse(ctx context.Context, req *enf_api.NotifyTargetsInUseReq, opts ...grpc.CallOption) (resp *enf_api.NotifyTargetsInUseResp, err error) {
	if err := ra.Verifier.VerifyPermission(ctx, jwt.PermissionRules, jwt.ActionRead); err != nil {
		return nil, auth.PermissionErrorToStatus(err)
	}

	return ra.RuleControllerClient.NotifyTargetsInUse(ctx, req, opts...)
}
