package client

//go:generate minimock -i github.com/runtime-radar/runtime-radar/kube-manager/api.PodControllerClient -o ./kube_manager_mock.go -n PodControllerClientMock -g

import (
	"crypto/tls"

	"github.com/runtime-radar/runtime-radar/event-processor/pkg/build"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/server"
	"github.com/runtime-radar/runtime-radar/kube-manager/api"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func NewPodController(address string, tlsConfig *tls.Config, tokenKey []byte) (api.PodControllerClient, func() error, error) {
	var creds credentials.TransportCredentials
	if tlsConfig != nil {
		creds = credentials.NewTLS(tlsConfig)
	} else {
		creds = insecure.NewCredentials()
	}

	opts := []grpc.DialOption{grpc.WithTransportCredentials(creds), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(server.MaxRecvMsgSize))}

	if len(tokenKey) > 0 {
		rp := &jwt.RolePermissions{
			KillPods: &jwt.Permission{
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

	client := api.NewPodControllerClient(conn)

	return client, conn.Close, nil
}
