package service

import (
	"context"

	"github.com/runtime-radar/runtime-radar/kube-manager/api"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/auth"
	"google.golang.org/grpc"
)

type PodAuth struct {
	Pod

	Verifier jwt.Verifier
}

func (p *PodAuth) Get(ctx context.Context, req *api.GetPodReq, opts ...grpc.CallOption) (*api.GetPodResp, error) {
	if err := p.Verifier.VerifyPermission(ctx, jwt.PermissionClusters, jwt.ActionRead); err != nil {
		return nil, auth.PermissionErrorToStatus(err)
	}

	return p.Pod.Get(ctx, req, opts...)
}

func (p *PodAuth) ListMeta(ctx context.Context, req *api.ListPodMetaReq, opts ...grpc.CallOption) (*api.ListPodMetaResp, error) {
	if err := p.Verifier.VerifyPermission(ctx, jwt.PermissionClusters, jwt.ActionRead); err != nil {
		return nil, auth.PermissionErrorToStatus(err)
	}

	return p.Pod.ListMeta(ctx, req, opts...)
}

func (p *PodAuth) ListPage(ctx context.Context, req *api.ListPodPageReq, opts ...grpc.CallOption) (*api.ListPodPageResp, error) {
	if err := p.Verifier.VerifyPermission(ctx, jwt.PermissionClusters, jwt.ActionRead); err != nil {
		return nil, auth.PermissionErrorToStatus(err)
	}

	return p.Pod.ListPage(ctx, req, opts...)
}
