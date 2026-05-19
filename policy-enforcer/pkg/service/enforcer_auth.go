package service

import (
	"context"

	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	lib_context "github.com/runtime-radar/runtime-radar/lib/security/jwt/context"
	"github.com/runtime-radar/runtime-radar/policy-enforcer/api"
)

// EnforcerAuth is a layer for jwt-based authentication.
// Base server interface should not be embedded here unlike
// in implementations of other layers.
// All required methods should be explicitly implemented to ensure
// that new methods of the basic server are implemented for auth layer.
type EnforcerAuth struct {
	// UnsafeEmailControllerServer is embedded to opt out of forward
	// compatibility promised by protobuf library.
	// It merely contains an empty `mustEmbedUnimplementedEmailControllerServer()`
	// method.
	api.UnsafeEnforcerServer

	// EnforcerServer is a base server interface to pass
	// response to the next layer.
	EnforcerServer api.EnforcerServer
	Verifier       jwt.Verifier
}

func (ea *EnforcerAuth) EvaluatePolicyRuntimeEvent(ctx context.Context, req *api.EvaluatePolicyRuntimeEventReq) (resp *api.EvaluatePolicyRuntimeEventReq, err error) {
	if err := ea.Verifier.VerifyPermission(ctx, jwt.PermissionScanning, jwt.ActionExecute); err != nil {
		return nil, errcommon.PermissionErrorToStatus(err)
	}

	lib_context.SetUserID(ctx)

	resp, err = ea.EnforcerServer.EvaluatePolicyRuntimeEvent(ctx, req)
	return
}
