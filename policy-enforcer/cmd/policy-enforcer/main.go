package main

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/gops/agent"
	"github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/lib/logger"
	"github.com/runtime-radar/runtime-radar/lib/security"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	"github.com/runtime-radar/runtime-radar/lib/server/healthcheck"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	"github.com/runtime-radar/runtime-radar/policy-enforcer/api"
	"github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/build"
	"github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/cache"
	"github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/config"
	"github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/database"
	"github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/metrics"
	"github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/server"
	"github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/service"
	"go.uber.org/automaxprocs/maxprocs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"gorm.io/gorm"
)

const (
	// TLS cert file name.
	certFile = "cert.pem"
	// TLS key file name.
	keyFile = "key.pem"
	// CA cert file name.
	caFile = "ca.pem"

	// Timeout on graceful shutdown.
	gracefulTimeout = 15 * time.Second
)

var (
	// Channel for stopping the program.
	shutdown = make(chan struct{})
)

func main() {
	cfg := config.New()
	logger.Init(cfg.LogFile, cfg.LogLevel)

	log.Info().Str("build_release", build.Release).Str("build_branch", build.Branch).Str("build_commit", build.Commit).Str("build_date", build.Date).Msgf("-> %s started", build.AppName)
	defer log.Info().Msgf("<- %s exited", build.AppName)

	if _, err := maxprocs.Set(maxprocs.Logger(log.Debug().Msgf)); err != nil {
		log.Warn().Msgf("Can't set maxprocs: %v", err)
	}

	if err := agent.Listen(agent.Options{
		Addr: cfg.GopsAddr,
	}); err != nil {
		log.Fatal().Msgf("### Failed to start gops agent: %v", err)
	}
	defer agent.Close()

	go signalListener()

	lis, err := net.Listen("tcp", cfg.ListenGRPCAddr)
	if err != nil {
		log.Fatal().Msgf("### Failed to listen: %v", err)
	}

	var verifier jwt.Verifier
	if cfg.Auth {
		verifier, _, err = jwt.NewKeyVerifier(cfg.TokenKey)
		if err != nil {
			log.Fatal().Msgf("### Failed to instantiate key verifier: %v", err)
		}
	}

	// Connect to DB
	db, closeDB, err := database.New(cfg.PostgresAddr, cfg.PostgresDB, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresSSLMode, cfg.PostgresSSLCheckCert)
	if err != nil {
		log.Fatal().Msgf("### Failed to open DB: %v", err)
	}
	defer closeDB()

	// Recreate DB from scratch, or migrate automatically when needed
	if err := database.Migrate(db, cfg.NewDB); err != nil {
		log.Fatal().Msgf("### Failed to migrate DB: %v", err)
	}

	remoteCache, closeCache, err := cache.NewRemote(cfg.RedisAddr, cfg.RedisUser, cfg.RedisPassword, cfg.RedisTLSMode, cfg.RedisTLSCheckCert, "")
	if err != nil {
		log.Fatal().Msgf("### Failed to open cache: %v", err)
	}
	defer closeCache()

	grpcMetrics := prometheus.NewServerMetrics(prometheus.WithServerHandlingTimeHistogram())

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptor.Recovery,
			interceptor.Correlation,
			grpcMetrics.UnaryServerInterceptor(),
		),
		grpc.MaxRecvMsgSize(server.MaxRecvMsgSize),
	}

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

	// Register reflection service on gRPC server
	reflection.Register(grpcSrv)

	// Initialize metrics
	m, err := metrics.PrepareRegistry(build.AppName, cfg.OwnCSURL, grpcMetrics)
	if err != nil {
		log.Fatal().Msgf("### Failed to initialize metrics: %v", err)
	}

	iSrv := server.NewInstrumentation(cfg.InstrumentationAddr, m)

	// Run the instrumentation HTTP server for metrics, probes, etc.
	go func() {
		if err := iSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Msgf("### Can't serve instrumentation HTTP requests: %v", err)
		}
	}()
	log.Info().Msgf("Instrumentation HTTP server listening at %v", cfg.InstrumentationAddr)

	// Run gRPC server
	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatal().Msgf("### Can't serve gRPC requests: %v", err)
		}
	}()
	log.Info().Msgf("gRPC server listening at %v", lis.Addr())

	httpSrv, err := server.New(cfg.ListenHTTPAddr, cfg.ListenGRPCAddr, tlsConfig)
	if err != nil {
		log.Fatal().Msgf("### Can't setup HTTP server: %v", err)
	}

	// Run HTTP server
	go func() {
		if cfg.TLS {
			if err := httpSrv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				log.Fatal().Msgf("### Can't serve HTTP requests: %v", err)
			}
		} else {
			if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal().Msgf("### Can't serve HTTP requests: %v", err)
			}
		}
	}()
	log.Info().Msgf("HTTP server listening at %v", httpSrv.Addr)

	healthcheck.SetReady() // <-- turn on ready status for k8s

	<-shutdown

	log.Info().Msg("gRPC server stopping gracefully")
	grpcSrv.GracefulStop()

	ctx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
	defer cancel()

	log.Info().Msg("HTTP server stopping gracefully")
	httpSrv.Shutdown(ctx) // we don't care about errors here

	log.Info().Msg("Instrumentation HTTP server stopping gracefully")
	_ = iSrv.Shutdown(ctx)
}

func composeServices(
	db *gorm.DB,
	remoteCache cache.Cache,
	verifier jwt.Verifier,
	isAuth bool,
) (ruleSvc api.RuleControllerServer, enforcerSvc api.EnforcerServer) {
	ruleSvc = &service.RuleGeneric{
		RuleRepository: &database.RuleDatabase{db},
		RuleMatcher: &cache.RuleCache{
			remoteCache,
			&database.RuleDatabase{db},
		},
	}

	enforcerSvc = &service.EnforcerGeneric{
		RuleMatcher: &cache.RuleCache{
			remoteCache,
			&database.RuleDatabase{db},
		},
		RuleRepository: &database.RuleDatabase{db},
	}

	if isAuth {
		ruleSvc = &service.RuleAuth{
			RuleControllerServer: ruleSvc,
			Verifier:             verifier,
		}

		enforcerSvc = &service.EnforcerAuth{
			EnforcerServer: enforcerSvc,
			Verifier:       verifier,
		}
	}

	ruleSvc = &service.RuleLogging{ruleSvc}
	enforcerSvc = &service.EnforcerLogging{enforcerSvc}

	return
}

func signalListener() {
	defer close(shutdown)

	sigTerm := make(chan os.Signal, 10)
	sigIgnore := make(chan os.Signal, 10)

	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM)
	signal.Notify(sigIgnore, syscall.SIGHUP)

	// Wait for signals
	for {
		select {
		case s := <-sigTerm:
			log.Info().Str("signal", s.String()).Msg("Signal caught, terminating")
			return
		case s := <-sigIgnore:
			// Ignoring, like with "nohup"
			log.Info().Str("signal", s.String()).Msg("Signal caught, ignoring")
		}
	}
}
