package service

import (
	"context"

	"github.com/runtime-radar/runtime-radar/kube-manager/api"
	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	"google.golang.org/protobuf/types/known/emptypb"
)

// PodAuth is a layer for jwt-based authentication.
// Base server interface should not be embedded here unlike
// in implementations of other layers.
// All required methods should be explicitly implemented to ensure
// that new methods of the basic server are implemented for auth layer.
type PodAuth struct {
	// UnsafePodControllerServer is embedded to opt out of forward
	// compatibility promised by protobuf library.
	// It merely contains an empty `mustEmbedUnimplementedPodControllerServer()`
	// method.
	api.UnsafePodControllerServer

	// PodControllerServer is a base server interface to pass
	// response to the next layer.

	PodControllerServer api.PodControllerServer
	Verifier            jwt.Verifier
}

// Get requires permission to execute get pod.
func (pa *PodAuth) Get(ctx context.Context, req *api.GetPodReq) (resp *api.GetPodResp, err error) {
	if err := pa.Verifier.VerifyPermission(ctx, jwt.PermissionClusters, jwt.ActionRead); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	resp, err = pa.PodControllerServer.Get(ctx, req)
	return
}

func (pa *PodAuth) ListMeta(ctx context.Context, req *api.ListPodMetaReq) (resp *api.ListPodMetaResp, err error) {
	if err := pa.Verifier.VerifyPermission(ctx, jwt.PermissionClusters, jwt.ActionRead); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	resp, err = pa.PodControllerServer.ListMeta(ctx, req)
	return
}

func (pa *PodAuth) ListPage(ctx context.Context, req *api.ListPodPageReq) (resp *api.ListPodPageResp, err error) {
	if err := pa.Verifier.VerifyPermission(ctx, jwt.PermissionClusters, jwt.ActionRead); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	resp, err = pa.PodControllerServer.ListPage(ctx, req)
	return
}

func (pa *PodAuth) Kill(ctx context.Context, req *api.KillPodReq) (resp *emptypb.Empty, err error) {
	if err := pa.Verifier.VerifyPermission(ctx, jwt.PermissionKillPods, jwt.ActionExecute); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	resp, err = pa.PodControllerServer.Kill(ctx, req)
	return
}
