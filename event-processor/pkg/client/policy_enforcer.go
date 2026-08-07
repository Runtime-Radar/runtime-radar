package client

//go:generate minimock -i github.com/runtime-radar/runtime-radar/policy-enforcer/api.EnforcerClient -o ./policy_enforcer_mock.go -n EnforcerClientMock -g

import (
	"crypto/tls"

	"github.com/runtime-radar/runtime-radar/event-processor/pkg/build"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/server"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	"github.com/runtime-radar/runtime-radar/policy-enforcer/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func NewPolicyEnforcer(address string, tlsConfig *tls.Config, tokenKey []byte) (api.EnforcerClient, func() error, error) {
	var creds credentials.TransportCredentials
	if tlsConfig != nil {
		creds = credentials.NewTLS(tlsConfig)
	} else {
		creds = insecure.NewCredentials()
	}

	opts := []grpc.DialOption{grpc.WithTransportCredentials(creds), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(server.MaxRecvMsgSize))}

	if len(tokenKey) > 0 {
		rp := &jwt.RolePermissions{
			Scanning: &jwt.Permission{
				Actions: []jwt.Action{jwt.ActionExecute},
			},
		}

		creds, err := jwt.GeneratePerRPCCredentials(tokenKey, build.AppName, rp)
		if err != nil {
			return nil, nil, err
		}

		opts = append(opts, grpc.WithPerRPCCredentials(creds))
	}

	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		return nil, nil, err
	}

	client := api.NewEnforcerClient(conn)

	return client, conn.Close, nil
}
