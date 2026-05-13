package main

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"maps"
	"math"
	"net"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/lib/logger"
	"github.com/runtime-radar/runtime-radar/lib/security"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	"github.com/runtime-radar/runtime-radar/policy-enforcer/api"
	"github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/cache"
	"github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/config"
	"github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/database"
	"github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/model"
	"github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/model/convert"
	"github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/server"
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
	ruleController api.RuleControllerClient
	enforcer       api.EnforcerClient

	deniedRuleController api.RuleControllerClient
	deniedEnforcer       api.EnforcerClient

	unauthRuleController api.RuleControllerClient
	unauthEnforcer       api.EnforcerClient

	cfg *config.Config

	// remoteCache is exposed for some gray-box test cases.
	remoteCache *cache.Remote
)

func TestMain(m *testing.M) {
	cfg = config.New()

	if testing.Verbose() {
		logger.Init("", "DEBUG")
	} else {
		logger.Init("", "INFO")
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

	var closeCache func() error
	remoteCache, closeCache, err = cache.NewRemote(cfg.RedisAddr, cfg.RedisUser, cfg.RedisPassword, cfg.RedisTLSMode, cfg.RedisTLSCheckCert, "test_")
	if err != nil {
		log.Fatal().Msgf("### Failed to open cache: %v", err)
	}

	opts := []grpc.ServerOption{grpc.ChainUnaryInterceptor(interceptor.Recovery, interceptor.Correlation), grpc.MaxRecvMsgSize(server.MaxRecvMsgSize)}

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
	ruleSvc, enforcerSvc := composeServices(db, remoteCache, verifier, cfg.Auth)

	api.RegisterRuleControllerServer(grpcSrv, ruleSvc)
	api.RegisterEnforcerServer(grpcSrv, enforcerSvc)

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
	closeCache()

	os.Exit(res)
}

type clients struct {
	ruleController api.RuleControllerClient
	enforcer       api.EnforcerClient
	closer         func() error
}

func newClients(address string, opts ...grpc.DialOption) (*clients, error) {
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		return nil, err
	}

	return &clients{
		api.NewRuleControllerClient(conn),
		api.NewEnforcerClient(conn),
		conn.Close,
	}, nil
}

func generateServiceCredentials(key []byte, actions []jwt.Action) (credentials.PerRPCCredentials, error) {
	rp := &jwt.RolePermissions{
		Rules: &jwt.Permission{
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

	ruleController = goodClients.ruleController
	enforcer = goodClients.enforcer

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

	ruleController = goodClients.ruleController
	enforcer = goodClients.enforcer

	deniedClients, err := newClients(address, grpc.WithTransportCredentials(transportCreds), grpc.WithPerRPCCredentials(unauthorizedCreds))
	if err != nil {
		return nil, err
	}

	deniedRuleController = deniedClients.ruleController
	deniedEnforcer = deniedClients.enforcer

	unauthClients, err := newClients(address, grpc.WithTransportCredentials(transportCreds))
	if err != nil {
		return nil, err
	}

	unauthRuleController = unauthClients.ruleController
	unauthEnforcer = unauthClients.enforcer

	closeClients := func() error {
		goodClients.closer()
		deniedClients.closer()
		unauthClients.closer()
		return nil
	}

	return closeClients, nil
}

func TestEnforcerWithNoAuthE2E(t *testing.T) {
	t.Parallel()

	if !cfg.Auth {
		t.Skip("Auth is not enabled")
	}

	ctx := context.Background()

	t.Run("Usecase: EvaluatePolicyRuntime without token", func(t *testing.T) {
		t.Parallel()

		_, err := unauthEnforcer.EvaluatePolicyRuntimeEvent(ctx, &api.EvaluatePolicyRuntimeEventReq{})
		if err == nil {
			t.Fatal("Request without token was allowed")
		} else if st, ok := status.FromError(err); !ok {
			t.Fatalf("Non-status error: %v", err)
		} else if st.Code() != codes.Unauthenticated {
			t.Fatalf("Incorrect status code: 'codes.Unauthenticated' != '%v'", st.Code())
		}
	})

	t.Run("Usecase: EvaluatePolicyRuntime with incomplete token", func(t *testing.T) {
		t.Parallel()

		_, err := deniedEnforcer.EvaluatePolicyRuntimeEvent(ctx, &api.EvaluatePolicyRuntimeEventReq{})
		if err == nil {
			t.Fatal("Request without token was allowed")
		} else if st, ok := status.FromError(err); !ok {
			t.Fatalf("Non-status error: %v", err)
		} else if st.Code() != codes.PermissionDenied {
			t.Fatalf("Incorrect status code: 'codes.PermissionDenied' != '%v'", st.Code())
		}
	})
}

func TestRuleCreateAndReadE2E(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Usecase: Create and Read rule with pre-declared ID", func(t *testing.T) {
		t.Parallel()

		toCreate := newRule("New rule with pre-declared ID")
		toCreate.Id = uuid.NewString()

		// ---> Send Create request <---
		createResp, err := ruleController.Create(ctx, toCreate)
		if err != nil {
			t.Fatalf("Can't perform Create operation: %v", err)
		}

		// ---> Send Read request <---
		readResp, err := ruleController.Read(ctx, &api.ReadRuleReq{Id: createResp.Id})
		if err != nil {
			t.Fatalf("Can't perform Read operation: %v", err)
		}
		toCompare := readResp.Rule

		if diff := cmp.Diff(toCreate, toCompare, protocmp.Transform()); diff != "" {
			t.Fatalf("Rules are not equal: %s", diff)
		}
	})

	t.Run("Usecase: Create and Read rule without ID", func(t *testing.T) {
		t.Parallel()

		toCreate := newRule("New rule without ID")

		// ---> Send Create request <---
		createResp, err := ruleController.Create(ctx, toCreate)
		if err != nil {
			t.Fatalf("Can't perform Create operation: %v", err)
		}
		toCreate.Id = createResp.Id

		// ---> Send Read request <---
		readResp, err := ruleController.Read(ctx, &api.ReadRuleReq{Id: createResp.Id})
		if err != nil {
			t.Fatalf("Can't perform Read operation: %v", err)
		}
		toCompare := readResp.Rule

		if diff := cmp.Diff(toCreate, toCompare, protocmp.Transform()); diff != "" {
			t.Fatalf("Rules are not equal: %s", diff)
		}
	})

	t.Run("Usecase: Create rule with invalid severity", func(t *testing.T) {
		t.Parallel()

		toCreate := newRule("New rule without ID")
		toCreate.Rule.Block.Severity = "block-severity-invalid"

		// ---> Send Create request <---
		_, err := ruleController.Create(ctx, toCreate)
		if err == nil {
			t.Fatal("Request with unknown severity was allowed")
		} else if st, ok := status.FromError(err); !ok {
			t.Fatalf("Non-status error: %v", err)
		} else if st.Code() != codes.InvalidArgument {
			t.Fatalf("Incorrect status code: 'codes.InvalidArgument' != '%v'", st.Code())
		}
	})
}

func TestRuleCreateUpdateAndReadE2E(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Usecase: Create, Update and Read rule", func(t *testing.T) {
		t.Parallel()

		toCreate := newRule("New rule for Update")

		// ---> Send Create request <---
		createResp, err := ruleController.Create(ctx, toCreate)
		if err != nil {
			t.Fatalf("Can't perform Create operation: %v", err)
		}

		// ---> Send Read request <---
		readResp, err := ruleController.Read(ctx, &api.ReadRuleReq{Id: createResp.Id})
		if err != nil {
			t.Fatalf("Can't perform Read operation: %v", err)
		}
		toUpdate := readResp.Rule

		toUpdate.Name = "Updated name"
		toUpdate.Rule.Block = &api.Rule_RuleJSON_Block{Severity: model.NoneSeverity.String()}
		toUpdate.Rule.Notify = &api.Rule_RuleJSON_Notify{Severity: model.HighSeverity.String()}
		toUpdate.Scope.Clusters = []string{"updated-cluster-*"}
		toUpdate.Scope.Namespaces = []string{"updated-namespace-*"}

		// ---> Send Update request <---
		_, err = ruleController.Update(ctx, toUpdate)
		if err != nil {
			t.Fatalf("Can't perform Update operation: %v", err)
		}

		// ---> Send Read request <---
		readResp, err = ruleController.Read(ctx, &api.ReadRuleReq{Id: createResp.Id})
		if err != nil {
			t.Fatalf("Can't perform Read operation: %v", err)
		}
		toCompare := readResp.Rule

		if diff := cmp.Diff(toUpdate, toCompare, protocmp.Transform()); diff != "" {
			t.Fatalf("Rules are not equal: %s", diff)
		}
	})

	t.Run("Usecase: Update rule with empty ID", func(t *testing.T) {
		t.Parallel()

		toUpdate := newRule("New rule for Update with empty ID")

		// ---> Send Update request <---
		if _, err := ruleController.Update(ctx, toUpdate); err == nil {
			t.Fatalf("Update with empty ID was allowed: %+v", toUpdate)
		} else if st, ok := status.FromError(err); !ok {
			t.Fatalf("Non-status error: %v", err)
		} else if st.Code() != codes.InvalidArgument {
			t.Fatalf("Incorrect status code: 'codes.InvalidArgument' != '%v'", st.Code())
		}
	})

	t.Run("Usecase: Update rule with invalid severity", func(t *testing.T) {
		t.Parallel()

		toCreate := newRule("New rule for Update with invalid severity")

		// ---> Send Create request <---
		createResp, err := ruleController.Create(ctx, toCreate)
		if err != nil {
			t.Fatalf("Can't perform Create operation: %v", err)
		}

		// ---> Send Read request <---
		readResp, err := ruleController.Read(ctx, &api.ReadRuleReq{Id: createResp.Id})
		if err != nil {
			t.Fatalf("Can't perform Read operation: %v", err)
		}
		toUpdate := readResp.Rule
		toUpdate.Rule.Notify.Severity = "notify-severity-invalid"

		// ---> Send Update request <---
		_, err = ruleController.Update(ctx, toUpdate)
		if err == nil {
			t.Fatal("Request with unknown severity was allowed")
		} else if st, ok := status.FromError(err); !ok {
			t.Fatalf("Non-status error: %v", err)
		} else if st.Code() != codes.InvalidArgument {
			t.Fatalf("Incorrect status code: 'codes.InvalidArgument' != '%v'", st.Code())
		}
	})
}

func TestRuleCreateDeleteAndReadE2E(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Usecase: Create, Delete and try to Read rule", func(t *testing.T) {
		t.Parallel()

		toDelete := newRule("New rule for Delete")

		// ---> Send Create request <---
		createResp, err := ruleController.Create(ctx, toDelete)
		if err != nil {
			t.Fatalf("Can't perform Create operation: %v", err)
		}
		toDelete.Id = createResp.Id

		// ---> Send Delete request <---
		if _, err := ruleController.Delete(ctx, &api.DeleteRuleReq{Id: createResp.Id}); err != nil {
			t.Fatalf("Can't perform Delete operation: %v", err)
		}

		// ---> Send Read request <---
		readResp, err := ruleController.Read(ctx, &api.ReadRuleReq{Id: createResp.Id}) // this operation fetches soft-deleted entries as well
		if err != nil {
			t.Fatalf("Can't perform Read operation: %v", err)
		} else if !readResp.Deleted {
			t.Fatalf("Rule wasn't deleted: %+v", readResp.Rule)
		}
	})
}

func TestRuleCreateAndListPageE2E(t *testing.T) {
	t.Parallel()

	var oneOfRules *api.Rule
	ctx := context.Background()
	total := 15

	for i := 0; i < total; i++ {
		toCreate := newRule("000 New rule for ListPage") // "000" is for ordering
		// ---> Send Create request <---
		if _, err := ruleController.Create(ctx, toCreate); err != nil {
			t.Fatalf("Can't perform Create operation: %v", err)
		}

		if i == 0 { // doesn't matter which one
			oneOfRules = toCreate
		}
	}

	t.Run("Usecase: Create and ListPage on many rules", func(t *testing.T) {
		pageSize := 10

		// ---> SendListPage request <---
		listPageResp, err := ruleController.ListPage(ctx, &api.ListRulePageReq{
			PageNum:  1,
			PageSize: uint32(pageSize),
		})
		if err != nil {
			t.Fatalf("Cant perform ListPage operation: %v", err)
		} else if listPageResp.Total < uint32(total) { // can be more than total if other tests run in parallel
			t.Fatalf("Incorrect number of total items: %d", listPageResp.Total)
		} else if len(listPageResp.Rules) != pageSize {
			t.Fatalf("Incorrect number of items per page: %d", len(listPageResp.Rules))
		}

		// Do not compare items with default ordering, because some other items can be created by tests running in parallel
	})

	t.Run("Usecase: Create and ListPage on many rules with different page size and ordering", func(t *testing.T) {
		pageSize := 5

		// ---> SendListPage request <---
		listPageResp, err := ruleController.ListPage(ctx, &api.ListRulePageReq{
			PageNum:  uint32(math.Ceil(float64(total / pageSize))),
			PageSize: uint32(pageSize),
			Order:    "name asc", // <-- different order
		})

		if err != nil {
			t.Fatalf("Cant perform ListPage operation: %v", err)
		} else if listPageResp.Total < uint32(total) { // can be more than total if other tests run in parallel
			t.Fatalf("Incorrect number of total items: %d", listPageResp.Total)
		} else if len(listPageResp.Rules) != pageSize {
			t.Fatalf("Incorrect number of items per page: %d", len(listPageResp.Rules))
		}

		// Compare last element
		toCompare := listPageResp.Rules[len(listPageResp.Rules)-1]

		// Both "id" and "name" has to be ignored, because both are randomly generated.
		// Name has some random addition because it needs to be unique.
		if diff := cmp.Diff(oneOfRules, toCompare,
			protocmp.Transform(),
			protocmp.IgnoreFields(&api.Rule{}, "id", "name"),
		); diff != "" {
			t.Fatalf("Rules are not equal: %s", diff)
		}
	})

	t.Run("Usecase: ListPage with invalid order", func(t *testing.T) {
		// ---> SendListPage request <---
		_, err := ruleController.ListPage(ctx, &api.ListRulePageReq{
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

func TestEnforcerCreateRuleAndEvaluateE2E(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Usecase: Add rules and try to EvaluatePolicyRuntimeEvent with relevant request", func(t *testing.T) {
		t.Parallel()

		shouldWork := []*api.Rule{
			newRuleEvaluatePolicyRuntimeEvent(
				"New rule for EvaluatePolicyRuntimeEvent 1",
				[]string{"default"}, []string{"my-image:0.3.*"}, []string{"registry-?.com"}, []string{"web-server-*"}, []string{"*"}, []string{"dev-?"},
				model.HighSeverity, model.LowSeverity,
				nil, nil,
			),
			newRuleEvaluatePolicyRuntimeEvent(
				"New rule for EvaluatePolicyRuntimeEvent 2",
				[]string{"default*"}, []string{"my-image:*"}, []string{"registry*.com"}, []string{"web-server-*"}, []string{"server"}, []string{"dev*"},
				model.HighSeverity, model.LowSeverity,
				nil, nil,
			),
		}

		shouldNotWork := []*api.Rule{
			newRuleEvaluatePolicyRuntimeEvent(
				"New rule for EvaluatePolicyRuntimeEvent 3",
				[]string{"non-default"}, []string{"not-my-image:*"}, []string{"registry-?.com"}, []string{"web-server-*"}, []string{"server"}, []string{"prod-?"},
				model.HighSeverity, model.LowSeverity,
				nil, nil,
			),
			newRuleEvaluatePolicyRuntimeEvent(
				"New rule for EvaluatePolicyRuntimeEvent 4",
				[]string{"another-non-default"}, []string{"*"}, []string{"registry*"}, []string{"web-server-*"}, []string{"*"}, []string{"prod*"},
				model.HighSeverity, model.LowSeverity,
				nil, nil,
			),
		}

		rules := append(shouldWork, shouldNotWork...)

		for _, r := range rules {
			// ---> Send Create request <---
			createResp, err := ruleController.Create(ctx, r)
			if err != nil {
				t.Fatalf("Can't perform Create operation: %v", err)
			}
			r.Id = createResp.Id
		}

		req := newEvaluatePolicyRuntimeEventReq("default", "my-image:0.3.2", "registry-1.com", "web-server-123", "server", "dev-1", "/bin/sh")

		evaluateResp, err := enforcer.EvaluatePolicyRuntimeEvent(ctx, req)
		if err != nil {
			t.Fatalf("Can't perform EvaluatePolicyRuntimeEvent: %v", err)
		}

		workedRules := []*api.Rule{}
		for _, event := range evaluateResp.Result.GetEvents() {
			for _, r := range event.Policy.GetBlockBy() {
				workedRules = append(workedRules, r)
			}

			for _, r := range event.Policy.GetNotifyBy() {
				workedRules = append(workedRules, r)
			}
		}

		uniqueWorkedRules := makeRulesUnique(workedRules)

		if diff := cmp.Diff(uniqueWorkedRules, shouldWork,
			cmpopts.SortSlices(func(a, b *api.Rule) bool {
				return a.GetId() < b.GetId() // particular way of sorting does not matter
			}),
			protocmp.Transform(),
		); diff != "" {
			t.Fatalf("Expected rules != actual: %s", diff)
		}
	})

	t.Run("Usecase: Add rules and try to EvaluatePolicyRuntimeEvent with whitelist", func(t *testing.T) {
		t.Parallel()

		shouldWork := []*api.Rule{
			newRuleEvaluatePolicyRuntimeEvent(
				"New (non-)whitelisted rule for EvaluatePolicyRuntimeEvent 1",
				[]string{"my-namespace-*"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"prod-?"},
				model.HighSeverity, model.LowSeverity,
				nil, nil,
			),
			newRuleEvaluatePolicyRuntimeEvent(
				"New (non-)whitelisted rule for EvaluatePolicyRuntimeEvent 2",
				[]string{"my-namespace-?"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"prod-?"},
				model.HighSeverity, model.LowSeverity,
				[]string{"CS_RUNTIME_99", "CS_RUNTIME_100"}, []string{"/usr/bin/*"},
			),
		}

		shouldNotWork := []*api.Rule{
			newRuleEvaluatePolicyRuntimeEvent(
				"New whitelisted rule for EvaluatePolicyRuntimeEvent 3",
				[]string{"my-namespace-*"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"prod-?"},
				model.HighSeverity, model.LowSeverity,
				[]string{"CS_RUNTIME_1", "CS_RUNTIME_2", "CS_RUNTIME_3"}, nil,
			),
			newRuleEvaluatePolicyRuntimeEvent(
				"New whitelisted rule for EvaluatePolicyRuntimeEvent 4",
				[]string{"my-namespace-?"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"prod-?"},
				model.HighSeverity, model.LowSeverity,
				nil, []string{"*/bash"},
			),
		}

		rules := append(shouldWork, shouldNotWork...)
		for _, r := range rules {
			// ---> Send Create request <---
			createResp, err := ruleController.Create(ctx, r)
			if err != nil {
				t.Fatalf("Can't perform Create operation: %v", err)
			}
			r.Id = createResp.Id
		}

		req := newEvaluatePolicyRuntimeEventReq("my-namespace-1", "my-image:latest", "registry.com", "web-server-123", "server", "prod-1", "/bin/bash")

		evaluateResp, err := enforcer.EvaluatePolicyRuntimeEvent(ctx, req)
		if err != nil {
			t.Fatalf("Can't perform EvaluatePolicyRuntimeEvent: %v", err)
		}

		workedRules := []*api.Rule{}
		for _, event := range evaluateResp.Result.GetEvents() {
			for _, r := range event.Policy.GetBlockBy() {
				workedRules = append(workedRules, r)
			}

			for _, r := range event.Policy.GetNotifyBy() {
				workedRules = append(workedRules, r)
			}
		}

		uniqueWorkedRules := makeRulesUnique(workedRules)

		if diff := cmp.Diff(uniqueWorkedRules, shouldWork,
			cmpopts.SortSlices(func(a, b *api.Rule) bool {
				return a.GetId() < b.GetId() // particular way of sorting does not matter
			}),
			protocmp.Transform(),
		); diff != "" {
			t.Fatalf("Expected rules != actual: %s", diff)
		}
	})
}

// TestRuleCacheUsageE2E is a set of test cases that check usage of rule cache.
// These make sure that cache is populated properly and used when needed.
// Different types of rules are used to ensure that caching is working for all of them.
//
// These cases are run without t.Parralel() so that they're executed before other cases are tested.
// Each case invalidates cache after executing so that other cases are not affected by those values.
func TestRuleCacheUsageE2E(t *testing.T) {
	ctx := context.Background()

	t.Run("Usecase: Add rules, try to EvaluatePolicyRuntimeEvent and check that rule cache is populated", func(t *testing.T) {
		key := cache.RuleCacheKey(model.RuleTypeRuntime)

		t.Cleanup(func() {
			if err := remoteCache.Del(ctx, key); err != nil {
				t.Fatalf("Can't invalidate cache: %v", err)
			}
		})

		rules := []*api.Rule{
			newRuleEvaluatePolicyRuntimeEvent(
				"New rule for EvaluatePolicyRuntimeEvent 1",
				[]string{"first-namespace-*"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"*"},
				model.HighSeverity, model.LowSeverity,
				nil, nil,
			),
			newRuleEvaluatePolicyRuntimeEvent(
				"New rule for EvaluatePolicyRuntimeEvent 2",
				[]string{"first-namespace-?"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"*"},
				model.HighSeverity, model.LowSeverity,
				nil, nil,
			),
		}

		for _, r := range rules {
			// ---> Send Create request <---
			resp, err := ruleController.Create(ctx, r)
			if err != nil {
				t.Fatalf("Can't perform Create operation: %v", err)
			}
			r.Id = resp.Id
		}

		req := newEvaluatePolicyRuntimeEventReq("first-namespace", "some", "random", "values", "in", "the", "request")

		// call to populate cache, we're not interested in response
		_, err := enforcer.EvaluatePolicyRuntimeEvent(ctx, req)
		if err != nil {
			t.Fatalf("Can't perform EvaluatePolicyRuntimeEvent: %v", err)
		}

		var matcherData *cache.RuleMatcherData
		ok, err := remoteCache.Get(ctx, key, &matcherData)
		if err != nil {
			t.Fatalf("Can't get value from cache: %v", err)
		}
		if !ok {
			t.Fatal("Expected cache to contain rules")
		}

		// check if all created rules are in the cache after calling EvaluatePolicyRuntimeEvent
		for _, r := range rules {
			id := uuid.MustParse(r.Id) // normally should never panic

			if _, ok := matcherData.RuleData[id]; !ok {
				t.Fatalf("Expected rule %v to be in cache", r)
			}
		}
	})

	t.Run("Usecase: Populate rule cache manually, try to EvaluatePolicyRuntimeEvent and check that cache values are returned", func(t *testing.T) {
		key := cache.RuleCacheKey(model.RuleTypeRuntime)

		t.Cleanup(func() {
			if err := remoteCache.Del(ctx, key); err != nil {
				t.Fatalf("Can't invalidate cache: %v", err)
			}
		})

		rules := []*api.Rule{
			newRuleEvaluatePolicyRuntimeEvent(
				"New rule for EvaluatePolicyRuntimeEvent 1",
				[]string{"second-namespace-*"}, []string{"*"}, []string{"*"}, []string{"cache-pod*"}, []string{"server"}, []string{"prod-?"},
				model.HighSeverity, model.LowSeverity,
				nil, nil,
			),
			newRuleEvaluatePolicyRuntimeEvent(
				"New rule for EvaluatePolicyRuntimeEvent 2",
				[]string{"second-namespace-?"}, []string{"*"}, []string{"*"}, []string{"cache-pod"}, []string{"server"}, []string{"prod-?"},
				model.HighSeverity, model.LowSeverity,
				nil, nil,
			),
		}

		matcherData := cache.NewRuleMatcherData()

		// Prepare rules to be saved at the cache
		for _, r := range rules {
			r.Id = uuid.NewString() // set id explicitly as we are not saving rules using api
			rule, err := convert.RuleFromProto(r)
			if err != nil {
				t.Fatalf("Can't convert rule from proto: %v", err)
			}

			sp, err := cache.NewScopePatterns(rule.Scope)
			if err != nil {
				t.Fatalf("Can't create ScopePatterns: %v", err)
			}

			matcherData.RuleData[rule.ID] = rule
			matcherData.MatchData[rule.ID] = sp
		}

		// Add rules to the cache directly without adding them to database
		if err := remoteCache.Set(ctx, key, matcherData, time.Hour*24); err != nil {
			t.Fatalf("Can't save RuleMatcherData to cache: %v", err)
		}

		// Make sure namespace matches cached rules' scopes
		req := newEvaluatePolicyRuntimeEventReq("second-namespace-1", "some", "random", "cache-pod", "server", "prod-1", "request")

		evaluateResp, err := enforcer.EvaluatePolicyRuntimeEvent(ctx, req)
		if err != nil {
			t.Fatalf("Can't perform EvaluatePolicyRuntimeEvent: %v", err)
		}

		workedRules := []*api.Rule{}
		for _, event := range evaluateResp.Result.GetEvents() {
			for _, r := range event.Policy.GetBlockBy() {
				workedRules = append(workedRules, r)
			}

			for _, r := range event.Policy.GetNotifyBy() {
				workedRules = append(workedRules, r)
			}
		}

		uniqueWorkedRules := makeRulesUnique(workedRules)
		if diff := cmp.Diff(uniqueWorkedRules, rules,
			cmpopts.SortSlices(func(a, b *api.Rule) bool {
				return a.GetId() < b.GetId() // particular way of sorting does not matter
			}),
			protocmp.Transform(),
		); diff != "" {
			t.Fatalf("Expected rules != actual: %s", diff)
		}
	})
}

// TestRuleCacheInvalidationE2E is a set of test cases that check invalidation of rule cache.
// These make sure that cache is invalidated when certain actions towards rules are performed.
// Different types of rules are used to ensure that caching is working for all of them.
//
// These cases are run without t.Parralel() so that they're executed before other cases are tested.
// Each case invalidates cache after executing so that other cases are not affected by those values.
func TestRuleCacheInvalidationE2E(t *testing.T) {
	ctx := context.Background()

	t.Run("Usecase: Create rules, call EvaluatePolicyRuntimeEvent, add new rule, call EvaluatePolicyRuntimeEvent and check if cache is invalidated", func(t *testing.T) {
		t.Cleanup(func() {
			if err := remoteCache.Del(ctx, cache.RuleCacheKey(model.RuleTypeRuntime)); err != nil {
				t.Fatalf("Can't invalidate cache: %v", err)
			}
		})

		rules := []*api.Rule{
			newRuleEvaluatePolicyRuntimeEvent(
				"New rule for EvaluatePolicyRuntimeEvent 1",
				[]string{"first-namespace*"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"*"},
				model.HighSeverity, model.LowSeverity,
				nil, nil,
			),
			newRuleEvaluatePolicyRuntimeEvent(
				"New rule for EvaluatePolicyRuntimeEvent 2",
				[]string{"first-namespace?"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"*"},
				model.HighSeverity, model.LowSeverity,
				nil, nil,
			),
		}

		for _, r := range rules {
			// ---> Send Create request <---
			resp, err := ruleController.Create(ctx, r)
			if err != nil {
				t.Fatalf("Can't perform Create operation: %v", err)
			}
			r.Id = resp.Id
		}

		req := newEvaluatePolicyRuntimeEventReq("first-namespace1", "some", "random", "values", "in", "the", "request")

		// call to populate cache, we're not interested in response
		_, err := enforcer.EvaluatePolicyRuntimeEvent(ctx, req)
		if err != nil {
			t.Fatalf("Can't perform EvaluatePolicyRuntimeEvent: %v", err)
		}

		anotherRule := newRuleEvaluatePolicyRuntimeEvent(
			"New rule for EvaluatePolicyRuntimeEvent 3",
			[]string{"first-namespace1"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"*"},
			model.HighSeverity, model.LowSeverity,
			nil, nil,
		)
		createResp, err := ruleController.Create(ctx, anotherRule)
		if err != nil {
			t.Fatalf("Can't perform Create operation: %v", err)
		}
		anotherRule.Id = createResp.Id
		rules = append(rules, anotherRule)

		// make sure that all rules expected to be in cache match given args
		req = newEvaluatePolicyRuntimeEventReq("first-namespace1", "some", "random", "values", "in", "the", "request")
		evaluateResp, err := enforcer.EvaluatePolicyRuntimeEvent(ctx, req)
		if err != nil {
			t.Fatalf("Can't perform EvaluatePolicyRuntimeEvent: %v", err)
		}

		workedRules := []*api.Rule{}
		for _, event := range evaluateResp.Result.GetEvents() {
			for _, r := range event.Policy.GetBlockBy() {
				workedRules = append(workedRules, r)
			}

			for _, r := range event.Policy.GetNotifyBy() {
				workedRules = append(workedRules, r)
			}
		}

		uniqueWorkedRules := makeRulesUnique(workedRules)

		if diff := cmp.Diff(uniqueWorkedRules, rules,
			cmpopts.SortSlices(func(a, b *api.Rule) bool {
				return a.GetId() < b.GetId() // particular way of sorting does not matter
			}),
			protocmp.Transform(),
		); diff != "" {
			t.Fatalf("Expected rules != actual: %s", diff)
		}
	})
	t.Run("Usecase: Create rules, call EvaluatePolicyRuntimeEvent, update existing rule, call EvaluatePolicyRuntimeEvent and check if cache is invalidated", func(t *testing.T) {
		t.Cleanup(func() {
			if err := remoteCache.Del(ctx, cache.RuleCacheKey(model.RuleTypeRuntime)); err != nil {
				t.Fatalf("Can't invalidate cache: %v", err)
			}
		})

		rules := []*api.Rule{
			newRuleEvaluatePolicyRuntimeEvent(
				"New rule for EvaluatePolicyRuntimeEvent 1",
				[]string{"second-namespace1-*"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"*"},
				model.HighSeverity, model.LowSeverity,
				nil, nil,
			),
			newRuleEvaluatePolicyRuntimeEvent(
				"New rule for EvaluatePolicyRuntimeEvent 2",
				[]string{"second-namespace1-?"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"*"},
				model.HighSeverity, model.LowSeverity,
				nil, nil,
			),
		}

		for _, r := range rules {
			// ---> Send Create request <---
			resp, err := ruleController.Create(ctx, r)
			if err != nil {
				t.Fatalf("Can't perform Create operation: %v", err)
			}
			r.Id = resp.Id
		}

		// call to populate cache, we're not interested in response
		_, err := enforcer.EvaluatePolicyRuntimeEvent(ctx, newEvaluatePolicyRuntimeEventReq("second-namespace1-1", "some", "random", "cache-pod1", "server", "prod-1", "request"))
		if err != nil {
			t.Fatalf("Can't perform EvaluateRuntimeEvent %v", err)
		}

		rules[0].Name = "Updated " + rules[0].Name
		_, err = ruleController.Update(ctx, rules[0])
		if err != nil {
			t.Fatalf("Can't perform Update operation: %v", err)
		}

		// make sure that all rules expected to be in cache match given namespace
		evaluateResp, err := enforcer.EvaluatePolicyRuntimeEvent(ctx, newEvaluatePolicyRuntimeEventReq("second-namespace1-2", "some", "random", "cache-pod1", "server", "prod-1", "request"))
		if err != nil {
			t.Fatalf("Can't perform EvaluateRuntimeEvent %v", err)
		}

		workedRules := []*api.Rule{}
		for _, event := range evaluateResp.Result.GetEvents() {
			for _, r := range event.Policy.GetBlockBy() {
				workedRules = append(workedRules, r)
			}

			for _, r := range event.Policy.GetNotifyBy() {
				workedRules = append(workedRules, r)
			}
		}

		uniqueWorkedRules := makeRulesUnique(workedRules)

		if diff := cmp.Diff(uniqueWorkedRules, rules,
			cmpopts.SortSlices(func(a, b *api.Rule) bool {
				return a.GetId() < b.GetId() // particular way of sorting does not matter
			}),
			protocmp.Transform(),
		); diff != "" {
			t.Fatalf("Expected rules != actual: %s", diff)
		}
	})

	t.Run("Usecase: Create rules, call EvaluatePolicyRuntimeEvent, delete 1 rule, call EvaluatePolicyRuntimeEvent and check if cache is invalidated", func(t *testing.T) {
		t.Cleanup(func() {
			if err := remoteCache.Del(ctx, cache.RuleCacheKey(model.RuleTypeRuntime)); err != nil {
				t.Fatalf("Can't invalidate cache: %v", err)
			}
		})

		rules := []*api.Rule{
			newRuleEvaluatePolicyRuntimeEvent(
				"New rule for EvaluatePolicyRuntimeEvent 1",
				[]string{"third-namespace-2-*"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"*"},
				model.HighSeverity, model.LowSeverity,
				nil, nil,
			),
			newRuleEvaluatePolicyRuntimeEvent(
				"New rule for EvaluatePolicyRuntimeEvent 2",
				[]string{"third-namespace-2-*"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"*"}, []string{"*"},
				model.HighSeverity, model.LowSeverity,
				nil, nil,
			),
		}

		for _, r := range rules {
			// ---> Send Create request <---
			resp, err := ruleController.Create(ctx, r)
			if err != nil {
				t.Fatalf("Can't perform Create operation: %v", err)
			}
			r.Id = resp.Id
		}

		// call to populate cache, we're not interested in response
		_, err := enforcer.EvaluatePolicyRuntimeEvent(ctx, newEvaluatePolicyRuntimeEventReq("third-namespace-2-1", "some", "random", "yet-another-cache-pod", "in", "the", "request"))
		if err != nil {
			t.Fatalf("Can't perform EvaluatePolicyRuntimeEvent: %v", err)
		}

		toDelete := rules[0]
		if _, err := ruleController.Delete(ctx, &api.DeleteRuleReq{Id: toDelete.Id}); err != nil {
			t.Fatalf("Can't perform Delete operation: %v", err)
		}
		rules = rules[1:]

		// make sure that all rules expected to be in cache match given namespace
		evaluateResp, err := enforcer.EvaluatePolicyRuntimeEvent(ctx, newEvaluatePolicyRuntimeEventReq("third-namespace-2-1", "some", "random", "yet-another-cache-pod", "in", "the", "request"))

		if err != nil {
			t.Fatalf("Can't perform EvaluatePolicyRuntime %v", err)
		}

		workedRules := []*api.Rule{}
		for _, event := range evaluateResp.Result.GetEvents() {
			for _, r := range event.Policy.GetBlockBy() {
				workedRules = append(workedRules, r)
			}

			for _, r := range event.Policy.GetNotifyBy() {
				workedRules = append(workedRules, r)
			}
		}
		uniqueWorkedRules := makeRulesUnique(workedRules)

		if diff := cmp.Diff(uniqueWorkedRules, rules,
			cmpopts.SortSlices(func(a, b *api.Rule) bool {
				return a.GetId() < b.GetId() // particular way of sorting does not matter
			}),
			protocmp.Transform(),
		); diff != "" {
			t.Fatalf("Expected rules != actual: %s", diff)
		}
	})
}

func newEvaluatePolicyRuntimeEventReq(namespace, image, registry, pod, container, node, binary string) *api.EvaluatePolicyRuntimeEventReq {
	return &api.EvaluatePolicyRuntimeEventReq{
		Actor: "event-processor",
		Action: &api.EvaluatePolicyRuntimeEventReq_Action{
			Args: &api.EvaluatePolicyRuntimeEventReq_Action_Args{
				Namespace: namespace,
				ImageName: image,
				Registry:  registry,
				Pod:       pod,
				Container: container,
				Node:      node,
				Binary:    binary,
			},
		},
		Result: &api.EvaluatePolicyRuntimeEventReq_Result{
			Events: []*api.EvaluatePolicyRuntimeEventReq_Result_Event{
				{
					DetectorId: "CS_RUNTIME_1",
					Severity:   model.HighSeverity.String(),
				},
				{
					DetectorId: "CS_RUNTIME_2",
					Severity:   model.LowSeverity.String(),
				},
				{
					DetectorId: "CS_RUNTIME_3",
					Severity:   model.MediumSeverity.String(),
				},
			},
		},
	}
}

func newRule(name string) *api.Rule {
	name = name + " " + security.RandAlphaNum(5) // name should be unique

	r := &api.Rule{
		Name: name,
		Rule: &api.Rule_RuleJSON{
			Version: string(model.RuleVersion),
			Block: &api.Rule_RuleJSON_Block{
				Severity: model.HighSeverity.String(),
			},
			Notify: &api.Rule_RuleJSON_Notify{
				Severity: model.LowSeverity.String(),
			},
		},
		Scope: &api.Rule_Scope{
			Version:    string(model.ScopeVersion),
			ImageNames: []string{"mytestgroup/*", "mytestsubgroup/*"},
			Namespaces: []string{"*"},
			Clusters:   []string{"*"},
		},
	}

	return r
}

func newRuleEvaluatePolicyRuntimeEvent(
	name string,
	namespaces, images, registries, pods, containers, nodes []string,
	block, notify model.Severity,
	threatsWL, binsWL []string,
) *api.Rule {
	name = name + " " + security.RandAlphaNum(5) // name should be unique

	r := &api.Rule{
		Name: name,
		Type: api.Rule_TYPE_RUNTIME,
		Rule: &api.Rule_RuleJSON{
			Version: string(model.RuleVersion),
			Block: &api.Rule_RuleJSON_Block{
				Severity: block.String(),
			},
			Notify: &api.Rule_RuleJSON_Notify{
				Severity: notify.String(),
			},
		},
		Scope: &api.Rule_Scope{
			Version:    string(model.ScopeVersion),
			Namespaces: namespaces,
			ImageNames: images,
			Registries: registries,
			Pods:       pods,
			Containers: containers,
			Nodes:      nodes,
		},
	}

	if len(threatsWL) != 0 || len(binsWL) != 0 {
		r.Rule.Whitelist = &api.Rule_RuleJSON_Whitelist{
			Threats:  threatsWL,
			Binaries: binsWL,
		}
	}

	return r
}

func makeRulesUnique(rules []*api.Rule) []*api.Rule {
	uniques := make(map[string]*api.Rule)

	for _, r := range rules {
		uniques[r.Id] = r
	}

	return slices.Collect(maps.Values(uniques))
}
