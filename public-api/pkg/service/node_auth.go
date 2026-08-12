package service

import (
	"context"

	"github.com/runtime-radar/runtime-radar/kube-manager/api"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/auth"
	"google.golang.org/grpc"
)

type NodeAuth struct {
	api.NodeControllerClient

	Verifier jwt.Verifier
}

func (n *NodeAuth) Get(ctx context.Context, req *api.GetNodeReq, opts ...grpc.CallOption) (*api.GetNodeResp, error) {
	if err := n.Verifier.VerifyPermission(ctx, jwt.PermissionClusters, jwt.ActionRead); err != nil {
		return nil, auth.PermissionErrorToStatus(err)
	}

	return n.NodeControllerClient.Get(ctx, req, opts...)
}

func (n *NodeAuth) ListMeta(ctx context.Context, req *api.ListNodeMetaReq, opts ...grpc.CallOption) (*api.ListNodeMetaResp, error) {
	if err := n.Verifier.VerifyPermission(ctx, jwt.PermissionClusters, jwt.ActionRead); err != nil {
		return nil, auth.PermissionErrorToStatus(err)
	}

	return n.NodeControllerClient.ListMeta(ctx, req, opts...)
}

func (n *NodeAuth) ListPage(ctx context.Context, req *api.ListNodePageReq, opts ...grpc.CallOption) (*api.ListNodePageResp, error) {
	if err := n.Verifier.VerifyPermission(ctx, jwt.PermissionClusters, jwt.ActionRead); err != nil {
		return nil, auth.PermissionErrorToStatus(err)
	}

	return n.NodeControllerClient.ListPage(ctx, req, opts...)
}
