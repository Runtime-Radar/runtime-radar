package main

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"math"
	"net"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/cluster-manager/api"
	"github.com/runtime-radar/runtime-radar/cluster-manager/pkg/config"
	"github.com/runtime-radar/runtime-radar/cluster-manager/pkg/database"
	"github.com/runtime-radar/runtime-radar/cluster-manager/pkg/model"
	"github.com/runtime-radar/runtime-radar/lib/logger"
	"github.com/runtime-radar/runtime-radar/lib/security"
	"github.com/runtime-radar/runtime-radar/lib/security/cipher"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

const (
	listenGRPCAddr = "127.0.0.1:8000"
)

var (
	clusterController       api.ClusterControllerClient
	deniedClusterController api.ClusterControllerClient
	unauthClusterController api.ClusterControllerClient

	cfg *config.Config
)

var (
	clusterCompareOpts = []cmp.Option{
		protocmp.Transform(),
		protocmp.IgnoreFields(&api.Cluster_Config_Postgres{}, "password"),
		protocmp.IgnoreFields(&api.Cluster_Config_Clickhouse{}, "password"),
		protocmp.IgnoreFields(&api.Cluster_Config_Redis{}, "password"),
		protocmp.IgnoreFields(&api.Cluster_Config_Rabbit{}, "password"),
		protocmp.IgnoreFields(&api.Cluster_Config_Registry{}, "password"),
	}
)

func TestMain(m *testing.M) {
	cfg = config.New()
	cfg.EncryptionKey = hex.EncodeToString(security.Rand(32))

	if testing.Verbose() {
		logger.Init("", "DEBUG")
	} else {
		logger.Init("", "INFO")
	}

	crypter, err := cipher.NewCrypt(cfg.EncryptionKey)
	if err != nil {
		log.Fatal().Msgf("### Failed to parse encryption key: %v", err)
	}

	lis, err := net.Listen("tcp", listenGRPCAddr)
	if err != nil {
		log.Fatal().Msgf("### Failed to listen: %v", err)
	}

	var verifier jwt.Verifier
	var tokenKey []byte
	if cfg.Auth {
		// cfg.TokenKey can be set to certain value to test some edge cases.
		// In other cases random value should be used.
		if cfg.TokenKey == "" {
			tokenBytes := security.Rand(32)
			cfg.TokenKey = hex.EncodeToString(tokenBytes)
		}

		if cfg.PublicAccessTokenSaltKey == "" {
			saltBytes := security.Rand(64)
			cfg.PublicAccessTokenSaltKey = hex.EncodeToString(saltBytes)
		}

		verifier, tokenKey, err = jwt.NewKeyVerifier(cfg.TokenKey)
		if err != nil {
			log.Fatal().Msgf("### Failed to instantiate key verifier: %v", err)
		}
	}

	db, closeDB, err := database.New(
		cfg.PostgresAddr,
		cfg.PostgresDB+"_test", // <-- use test DB
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresSSLMode,
		cfg.PostgresSSLCheckCert,
	)
	if err != nil {
		log.Fatal().Msgf("### Failed to open DB: %v", err)
	}
	if err := database.Migrate(
		db,
		true, // <-- recreate test DB from scratch
	); err != nil {
		log.Fatal().Msgf("### Failed to migrate DB: %v", err)
	}

	opts := []grpc.ServerOption{grpc.ChainUnaryInterceptor(interceptor.Recovery, interceptor.Correlation)}

	var tlsConfig *tls.Config
	if cfg.TLS {
		// Load TLS config
		tlsConfig, err = security.LoadTLS(caFile, certFile, keyFile)
		if err != nil {
			log.Fatal().Msgf("### Failed to load TLS config: %v", err)
		}
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}

	grpcSrv := grpc.NewServer(opts...)
	clusterSvc := composeServices(db, verifier, crypter, cfg.TokenKey, cfg.EncryptionKey, cfg.PublicAccessTokenSaltKey, cfg.CSVersion, cfg.TLS, cfg.Auth, cfg.AdministratorUsername, cfg.AdministratorPassword)

	api.RegisterClusterControllerServer(grpcSrv, clusterSvc)

	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatal().Msgf("### Can't serve gRPC requests: %v", err)
		}
	}()

	closeClients, err := initClients(listenGRPCAddr, tlsConfig, tokenKey)
	if err != nil {
		log.Fatal().Msgf("### Can't init gRPC clients")
	}

	res := m.Run() // <-- run tests

	// This kind of tier down is not required in tests, but we want to keep everything as clean as possible
	closeClients()
	grpcSrv.GracefulStop()
	closeDB()

	os.Exit(res)
}

type clients struct {
	clusterController api.ClusterControllerClient
	closer            func() error
}

func newClients(address string, opts ...grpc.DialOption) (*clients, error) {
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		return nil, err
	}

	return &clients{
		api.NewClusterControllerClient(conn),
		conn.Close,
	}, nil
}

func generateServiceCredentials(key []byte, actions []jwt.Action) (credentials.PerRPCCredentials, error) {
	rp := &jwt.RolePermissions{
		Clusters: &jwt.Permission{
			Actions: actions,
		},
		Scanning: &jwt.Permission{
			Actions: actions,
		},
	}
	return jwt.GeneratePerRPCCredentials(key, "test", rp)
}

func initClients(address string, tlsConfig *tls.Config, tokenKey []byte) (func() error, error) {
	creds := insecure.NewCredentials()
	if tlsConfig != nil {
		creds = credentials.NewTLS(tlsConfig)
	}

	if len(tokenKey) != 0 {
		return initClientsWithAuth(address, creds, tokenKey)
	}
	return initClientsWithoutAuth(address, creds)
}

func initClientsWithoutAuth(address string, creds credentials.TransportCredentials) (func() error, error) {
	goodClients, err := newClients(address, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}

	clusterController = goodClients.clusterController

	return goodClients.closer, nil
}

func initClientsWithAuth(address string, transportCreds credentials.TransportCredentials, tokenKey []byte) (func() error, error) {
	creds, err := generateServiceCredentials(tokenKey, []jwt.Action{jwt.ActionCreate, jwt.ActionRead, jwt.ActionUpdate, jwt.ActionDelete, jwt.ActionExecute})
	if err != nil {
		return nil, err
	}

	unauthorizedCreds, err := generateServiceCredentials(tokenKey, []jwt.Action{})
	if err != nil {
		return nil, err
	}

	goodClients, err := newClients(address, grpc.WithTransportCredentials(transportCreds), grpc.WithPerRPCCredentials(creds))
	if err != nil {
		return nil, err
	}

	clusterController = goodClients.clusterController

	deniedClients, err := newClients(address, grpc.WithTransportCredentials(transportCreds), grpc.WithPerRPCCredentials(unauthorizedCreds))
	if err != nil {
		return nil, err
	}

	deniedClusterController = deniedClients.clusterController

	unauthClients, err := newClients(address, grpc.WithTransportCredentials(transportCreds))
	if err != nil {
		return nil, err
	}

	unauthClusterController = unauthClients.clusterController

	closeClients := func() error {
		goodClients.closer()
		deniedClients.closer()
		unauthClients.closer()
		return nil
	}

	return closeClients, nil
}

func TestClusterCreateAndReadE2E(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Usecase: Create and Read cluster with pre-declared ID", func(t *testing.T) {
		t.Parallel()

		toCreate := newCluster("New cluster with pre-declared ID")
		toCreate.Id = uuid.NewString()

		// ---> Send Create request <---
		createResp, err := clusterController.Create(ctx, toCreate)
		if err != nil {
			t.Fatalf("Can't perform Create operation: %v", err)
		}

		// ---> Send Read request <---
		readResp, err := clusterController.Read(ctx, &api.ReadClusterReq{Id: createResp.Id})
		if err != nil {
			t.Fatalf("Can't perform Read operation: %v", err)
		}
		toCompare := readResp.Cluster

		if diff := cmp.Diff(toCreate, toCompare, append(clusterCompareOpts, protocmp.IgnoreFields(&api.Cluster{}, "created_at"))...); diff != "" {
			t.Fatalf("Clusters are not equal: %s", diff)
		}
	})

	t.Run("Usecase: Create and Read cluster without ID", func(t *testing.T) {
		t.Parallel()

		toCreate := newCluster("New cluster without ID")

		// ---> Send Create request <---
		createResp, err := clusterController.Create(ctx, toCreate)
		if err != nil {
			t.Fatalf("Can't perform Create operation: %v", err)
		}
		toCreate.Id = createResp.Id

		// ---> Send Read request <---
		readResp, err := clusterController.Read(ctx, &api.ReadClusterReq{Id: createResp.Id})
		if err != nil {
			t.Fatalf("Can't perform Read operation: %v", err)
		}
		toCompare := readResp.Cluster

		if diff := cmp.Diff(toCreate, toCompare, append(clusterCompareOpts, protocmp.IgnoreFields(&api.Cluster{}, "created_at"))...); diff != "" {
			t.Fatalf("Clusters are not equal: %s", diff)
		}
	})
}

func TestClusterCreateUpdateAndReadE2E(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Usecase: Create, Update and Read cluster", func(t *testing.T) {
		t.Parallel()

		toCreate := newCluster("New cluster for update")

		// ---> Send Create request <---
		createResp, err := clusterController.Create(ctx, toCreate)
		if err != nil {
			t.Fatalf("Can't perform Create operation: %v", err)
		}

		// ---> Send Read request <---
		readResp, err := clusterController.Read(ctx, &api.ReadClusterReq{Id: createResp.Id})
		if err != nil {
			t.Fatalf("Can't perform Read operation: %v", err)
		}
		toUpdate := readResp.Cluster

		toUpdate.Name = "Updated name"
		toUpdate.Config.Ingress = nil
		toUpdate.Config.Clickhouse = &api.Cluster_Config_Clickhouse{
			Address:     "new-clickhouse:9000",
			User:        "default",
			Password:    "default",
			UseTls:      false,
			Persistence: false,
			Database:    "default",
		}

		// ---> Send Update request <---
		_, err = clusterController.Update(ctx, toUpdate)
		if err != nil {
			t.Fatalf("Can't perform Update operation: %v", err)
		}

		// ---> Send Read request <---
		readResp, err = clusterController.Read(ctx, &api.ReadClusterReq{Id: createResp.Id})
		if err != nil {
			t.Fatalf("Can't perform Read operation: %v", err)
		}
		toCompare := readResp.Cluster

		if diff := cmp.Diff(toUpdate, toCompare, clusterCompareOpts...); diff != "" {
			t.Fatalf("Clusters are not equal: %s", diff)
		}
	})

	t.Run("Usecase: Update cluster with empty ID", func(t *testing.T) {
		t.Parallel()

		toUpdate := newCluster("New cluster for update with empty ID")

		// ---> Send Update request <---
		if _, err := clusterController.Update(ctx, toUpdate); err == nil {
			t.Fatalf("Update with empty ID was allowed: %+v", toUpdate)
		} else if st, ok := status.FromError(err); !ok {
			t.Fatalf("Non-status error: %v", err)
		} else if st.Code() != codes.InvalidArgument {
			t.Fatalf("Incorrect status code: 'codes.InvalidArgument' != '%v'", st.Code())
		}
	})

}

func TestClusterCreateDeleteAndReadE2E(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Usecase: Create, Delete and try to Read cluster", func(t *testing.T) {
		t.Parallel()

		toDelete := newCluster("New cluster for Delete")

		// ---> Send Create request <---
		createResp, err := clusterController.Create(ctx, toDelete)
		if err != nil {
			t.Fatalf("Can't perform Create operation: %v", err)
		}
		toDelete.Id = createResp.Id

		// ---> Send Delete request <---
		if _, err := clusterController.Delete(ctx, &api.DeleteClusterReq{Id: createResp.Id}); err != nil {
			t.Fatalf("Can't perform Delete operation: %v", err)
		}

		// ---> Send Read request <---
		readResp, err := clusterController.Read(ctx, &api.ReadClusterReq{Id: createResp.Id}) // this operation fetches soft-deleted entries as well
		if err != nil {
			t.Fatalf("Can't perform Read operation: %v", err)
		} else if !readResp.Deleted {
			t.Fatalf("Cluster wasn't deleted: %+v", readResp.Cluster)
		}
	})
}

func TestClusterCreateAndListPageE2E(t *testing.T) {
	t.Parallel()

	var oneOfClusters *api.Cluster
	ctx := context.Background()
	total := 15

	for i := 0; i < total; i++ {
		toCreate := newCluster("000 New cluster for ListPage") // "000" is for ordering
		// ---> Send Create request <---
		if _, err := clusterController.Create(ctx, toCreate); err != nil {
			t.Fatalf("Can't perform Create operation: %v", err)
		}

		if i == 0 { // doesn't matter which one
			oneOfClusters = toCreate
		}
	}

	t.Run("Usecase: Create and ListPage on many clusters", func(t *testing.T) {
		pageSize := 10

		// ---> SendListPage request <---
		listPageResp, err := clusterController.ListPage(ctx, &api.ListClusterPageReq{
			PageNum:  1,
			PageSize: uint32(pageSize),
		})
		if err != nil {
			t.Fatalf("Cant perform ListPage operation: %v", err)
		} else if listPageResp.Total < uint32(total) { // can be more than total if other tests run in parallel
			t.Fatalf("Incorrect number of total items: %d", listPageResp.Total)
		} else if len(listPageResp.Clusters) != pageSize {
			t.Fatalf("Incorrect number of items per page: %d", len(listPageResp.Clusters))
		}

		// Do not compare items with default ordering, because some other items can be created by tests running in parallel
	})

	t.Run("Usecase: Create and ListPage on many clusters with different page size and ordering", func(t *testing.T) {
		pageSize := 5

		// ---> SendListPage request <---
		listPageResp, err := clusterController.ListPage(ctx, &api.ListClusterPageReq{
			PageNum:  uint32(math.Ceil(float64(total / pageSize))),
			PageSize: uint32(pageSize),
			Order:    "name asc", // <-- different order
		})

		if err != nil {
			t.Fatalf("Cant perform ListPage operation: %v", err)
		} else if listPageResp.Total < uint32(total) { // can be more than total if other tests run in parallel
			t.Fatalf("Incorrect number of total items: %d", listPageResp.Total)
		} else if len(listPageResp.Clusters) != pageSize {
			t.Fatalf("Incorrect number of items per page: %d", len(listPageResp.Clusters))
		}

		// Compare last element
		toCompare := listPageResp.Clusters[len(listPageResp.Clusters)-1]

		// Both "id" and "name" has to be ignored, because both are randomly generated.
		// Name has some random addition because it needs to be unique.
		if diff := cmp.Diff(oneOfClusters, toCompare, append(clusterCompareOpts, protocmp.IgnoreFields(&api.Cluster{}, "id", "name", "created_at"))...); diff != "" {
			t.Fatalf("Clusters are not equal: %s", diff)
		}
	})

	t.Run("Usecase: ListPage with invalid order", func(t *testing.T) {
		// ---> SendListPage request <---
		_, err := clusterController.ListPage(ctx, &api.ListClusterPageReq{
			Order: "unknown asc", // <-- different order
		})

		if err == nil {
			t.Fatal("Request with unknown order was allowed")
		} else if st, ok := status.FromError(err); !ok {
			t.Fatalf("Non-status error: %v", err)
		} else if st.Code() != codes.InvalidArgument {
			t.Fatalf("Incorrect status code: 'codes.InvalidArgument' != '%v'", st.Code())
		}
	})
}

func newCluster(name string) *api.Cluster {
	name = name + " " + security.RandAlphaNum(5) // name should be unique

	return &api.Cluster{
		Name: name,
		Config: &api.Cluster_Config{
			Version:      string(model.ClusterConfigVersion),
			OwnCsUrl:     "https://cs2.example.com",
			CentralCsUrl: "https://cs.example.com",
			Postgres: &api.Cluster_Config_Postgres{
				Address:      "postgres:5432",
				Database:     "cs",
				User:         "cs",
				Password:     "cs",
				UseTls:       true,
				CheckCert:    false,
				Ca:           "",
				Persistence:  true,
				StorageClass: "",
			},
			Clickhouse: &api.Cluster_Config_Clickhouse{
				Address:      "clickhouse:9000",
				Database:     "cs",
				User:         "clickhouse",
				Password:     "clickhouse",
				UseTls:       true,
				CheckCert:    false,
				Ca:           "",
				Persistence:  true,
				StorageClass: "",
			},
			Redis: &api.Cluster_Config_Redis{
				Address:      "redis:6379",
				User:         "redis",
				Password:     "test",
				UseTls:       true,
				CheckCert:    false,
				Ca:           "",
				Persistence:  true,
				StorageClass: "",
			},
			Rabbit: &api.Cluster_Config_Rabbit{
				Address:      "rabbit:5672",
				User:         "rabbit",
				Password:     "test",
				Persistence:  true,
				StorageClass: "",
			},
			Registry: &api.Cluster_Config_Registry{
				Address:         "registry:5000",
				User:            "registry",
				Password:        "test",
				ImageShortNames: true,
			},
			Ingress: &api.Cluster_Config_Ingress{
				IngressClass: "nginx",
				Hostname:     "https://cs2.example.com",
				Cert:         "blabla",
				CertKey:      "blabla",
			},
			NodePort: &api.Cluster_Config_NodePort{
				Port: "8080",
			},
			Namespace: "cs",
			ProxyUrl:  "https://proxy.example.com",
		},
	}
}
