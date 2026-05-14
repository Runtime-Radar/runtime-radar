package service

import (
	"context"

	"github.com/runtime-radar/runtime-radar/cluster-manager/api"
	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	lib_context "github.com/runtime-radar/runtime-radar/lib/security/jwt/context"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ClusterAuth is a layer for jwt-based authentication.
// Base server interface should not be embedded here unlike
// in implementations of other layers.
// All required methods should be explicitly implemented to ensure
// that new methods of the basic server are implemented for auth layer.
type ClusterAuth struct {
	// UnsafeEmailControllerServer is embedded to opt out of forward
	// compatibility promised by protobuf library.
	// It merely contains an empty `mustEmbedUnimplementedEmailControllerServer()`
	// method.
	api.UnsafeClusterControllerServer

	// ClusterControllerServer is a base server interface to pass
	// response to the next layer.
	ClusterControllerServer api.ClusterControllerServer
	Verifier                jwt.Verifier
}

func (ca *ClusterAuth) Create(ctx context.Context, req *api.Cluster) (resp *api.CreateClusterResp, err error) {
	if err := ca.Verifier.VerifyPermission(ctx, jwt.PermissionClusters, jwt.ActionCreate); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}
	lib_context.SetUserID(ctx)
	resp, err = ca.ClusterControllerServer.Create(ctx, req)
	return
}

func (ca *ClusterAuth) Read(ctx context.Context, req *api.ReadClusterReq) (resp *api.ReadClusterResp, err error) {
	if err := ca.Verifier.VerifyPermission(ctx, jwt.PermissionClusters, jwt.ActionRead); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}
	lib_context.SetUserID(ctx)
	resp, err = ca.ClusterControllerServer.Read(ctx, req)
	return
}

func (ca *ClusterAuth) Update(ctx context.Context, req *api.Cluster) (resp *emptypb.Empty, err error) {
	if err := ca.Verifier.VerifyPermission(ctx, jwt.PermissionClusters, jwt.ActionUpdate); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}
	lib_context.SetUserID(ctx)
	resp, err = ca.ClusterControllerServer.Update(ctx, req)
	return
}

func (ca *ClusterAuth) Delete(ctx context.Context, req *api.DeleteClusterReq) (resp *emptypb.Empty, err error) {
	if err := ca.Verifier.VerifyPermission(ctx, jwt.PermissionClusters, jwt.ActionDelete); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}
	lib_context.SetUserID(ctx)
	resp, err = ca.ClusterControllerServer.Delete(ctx, req)
	return
}

func (ca *ClusterAuth) ListPage(ctx context.Context, req *api.ListClusterPageReq) (resp *api.ListClusterPageResp, err error) {
	if err := ca.Verifier.VerifyPermission(ctx, jwt.PermissionClusters, jwt.ActionRead); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}
	lib_context.SetUserID(ctx)
	resp, err = ca.ClusterControllerServer.ListPage(ctx, req)
	return
}

func (ca *ClusterAuth) Register(ctx context.Context, req *api.RegisterClusterReq) (resp *emptypb.Empty, err error) {
	if err := ca.Verifier.VerifyPermission(ctx, jwt.PermissionClusters, jwt.ActionUpdate); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}
	lib_context.SetUserID(ctx)
	resp, err = ca.ClusterControllerServer.Register(ctx, req)
	return
}

func (ca *ClusterAuth) Unregister(ctx context.Context, req *api.UnregisterClusterReq) (resp *emptypb.Empty, err error) {
	if err := ca.Verifier.VerifyPermission(ctx, jwt.PermissionClusters, jwt.ActionDelete); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}
	lib_context.SetUserID(ctx)
	resp, err = ca.ClusterControllerServer.Unregister(ctx, req)
	return
}

func (ca *ClusterAuth) GenerateUninstallCmd(ctx context.Context, req *api.GenerateUninstallCmdReq) (resp *api.GenerateUninstallCmdResp, err error) {
	if err := ca.Verifier.VerifyPermission(ctx, jwt.PermissionClusters, jwt.ActionExecute); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}
	lib_context.SetUserID(ctx)
	resp, err = ca.ClusterControllerServer.GenerateUninstallCmd(ctx, req)
	return
}

func (ca *ClusterAuth) GenerateInstallCmd(ctx context.Context, req *api.GenerateInstallCmdReq) (resp *api.GenerateInstallCmdResp, err error) {
	if err := ca.Verifier.VerifyPermission(ctx, jwt.PermissionClusters, jwt.ActionExecute); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}
	lib_context.SetUserID(ctx)
	resp, err = ca.ClusterControllerServer.GenerateInstallCmd(ctx, req)
	return
}

func (ca *ClusterAuth) GenerateValuesYAML(ctx context.Context, req *api.GenerateValuesReq) (resp *api.GenerateValuesResp, err error) {
	if err := ca.Verifier.VerifyPermission(ctx, jwt.PermissionClusters, jwt.ActionExecute); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}
	lib_context.SetUserID(ctx)
	resp, err = ca.ClusterControllerServer.GenerateValuesYAML(ctx, req)
	return
}

func (ca *ClusterAuth) ListRegistered(ctx context.Context, req *emptypb.Empty) (resp *api.ListRegisteredResp, err error) {
	if err := ca.Verifier.VerifyPermission(ctx, jwt.PermissionClusters, jwt.ActionRead); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}
	lib_context.SetUserID(ctx)
	resp, err = ca.ClusterControllerServer.ListRegistered(ctx, req)
	return
}
