package service

import (
	"context"

	"github.com/runtime-radar/runtime-radar/kube-manager/api"
	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
)

// NodeAuth is a layer for jwt-based authentication.
// Base server interface should not be embedded here unlike
// in implementations of other layers.
// All required methods should be explicitly implemented to ensure
// that new methods of the basic server are implemented for auth layer.
type NodeAuth struct {
	// UnsafeNodeControllerServer is embedded to opt out of forward
	// compatibility promised by protobuf library.
	// It merely contains an empty `mustEmbedUnimplementedNodeControllerServer()`
	// method.
	api.UnsafeNodeControllerServer

	// NodeControllerServer is a base server interface to pass
	// response to the next layer.

	NodeControllerServer api.NodeControllerServer
	Verifier             jwt.Verifier
}

// Get requires permission to execute get node.
func (na *NodeAuth) Get(ctx context.Context, req *api.GetNodeReq) (resp *api.GetNodeResp, err error) {
	if err := na.Verifier.VerifyPermission(ctx, jwt.PermissionClusters, jwt.ActionRead); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	resp, err = na.NodeControllerServer.Get(ctx, req)
	return
}

func (na *NodeAuth) ListMeta(ctx context.Context, req *api.ListNodeMetaReq) (resp *api.ListNodeMetaResp, err error) {
	if err := na.Verifier.VerifyPermission(ctx, jwt.PermissionClusters, jwt.ActionRead); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	resp, err = na.NodeControllerServer.ListMeta(ctx, req)
	return
}

func (na *NodeAuth) ListPage(ctx context.Context, req *api.ListNodePageReq) (resp *api.ListNodePageResp, err error) {
	if err := na.Verifier.VerifyPermission(ctx, jwt.PermissionClusters, jwt.ActionRead); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	resp, err = na.NodeControllerServer.ListPage(ctx, req)
	return
}
