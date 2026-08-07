package client

import (
	"crypto/tls"

	"github.com/runtime-radar/runtime-radar/kube-manager/api"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/build"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type KubeManagerClients struct {
	Config api.ConfigControllerClient
	Node   api.NodeControllerClient
	Pod    api.PodControllerClient

	close func() error
}

func (c *KubeManagerClients) Close() error {
	if c.close != nil {
		return c.close()
	}
	return nil
}

func NewKubeManager(address string, tlsConfig *tls.Config, tokenKey []byte) (*KubeManagerClients, error) {
	var creds credentials.TransportCredentials
	if tlsConfig != nil {
		creds = credentials.NewTLS(tlsConfig)
	} else {
		creds = insecure.NewCredentials()
	}

	opts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}

	if len(tokenKey) > 0 && tlsConfig != nil {
		rp := &jwt.RolePermissions{
			SystemSettings: &jwt.Permission{
				Actions: []jwt.Action{
					jwt.ActionCreate, jwt.ActionUpdate, jwt.ActionRead,
				},
			},
			Clusters: &jwt.Permission{
				Actions: []jwt.Action{
					jwt.ActionRead,
				},
			},
		}

		perRPCCreds, err := jwt.GeneratePerRPCCredentials(tokenKey, build.AppName, rp)
		if err != nil {
			return nil, err
		}

		opts = append(opts, grpc.WithPerRPCCredentials(perRPCCreds))
	}

	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, err
	}

	return &KubeManagerClients{
		Config: api.NewConfigControllerClient(conn),
		Node:   api.NewNodeControllerClient(conn),
		Pod:    api.NewPodControllerClient(conn),
		close:  conn.Close,
	}, nil
}
