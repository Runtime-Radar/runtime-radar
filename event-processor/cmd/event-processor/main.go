//go:build !tinygo.wasm

package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/google/gops/agent"
	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/event-processor/api"
	detector_api "github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/build"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/client"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/config"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/consumer"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/database"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/model"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/processor"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/processor/detector"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/processor/updater"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/server"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/service"
	"github.com/runtime-radar/runtime-radar/lib/rabbit"
	"github.com/runtime-radar/runtime-radar/lib/security"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	"github.com/runtime-radar/runtime-radar/lib/server/healthcheck"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	notifier_api "github.com/runtime-radar/runtime-radar/notifier/api"
	enforcer_api "github.com/runtime-radar/runtime-radar/policy-enforcer/api"
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
	// Detector binaries glob pattern.
	wasmPattern = "*.wasm"

	// Timeout on graceful shutdown.
	gracefulTimeout = 15 * time.Second
)

var (
	// Channel for stopping the program.
	shutdown = make(chan struct{})
)

func main() {
	cfg := config.New()
	initLogger(cfg.LogFile, cfg.LogLevel)

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
	var tokenKey []byte
	if cfg.Auth {
		verifier, tokenKey, err = jwt.NewKeyVerifier(cfg.TokenKey)
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

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(interceptor.Recovery, interceptor.Correlation),
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

	enforcer, closePE, err := client.NewPolicyEnforcer(cfg.PolicyEnforcerGRPCAddr, tlsConfig, tokenKey)
	if err != nil {
		log.Fatal().Msgf("### Failed to connect to Policy Enforcer: %v", err)
	}
	defer closePE()

	notifier, closeNotifier, err := client.NewNotifier(cfg.NotifierGRPCAddr, tlsConfig, tokenKey)
	if err != nil {
		log.Fatal().Msgf("### Failed to connect to Notifier: %v", err)
	}
	defer closeNotifier()

	runtimeMB, err := rabbit.NewMessageBroker(cfg.RabbitAddr, cfg.RabbitUser, cfg.RabbitPassword, cfg.RabbitRuntimeEventsQueue, rabbit.WithConsumer(build.AppName, cfg.RabbitRuntimeEventsQueuePrefetchCount))
	if err != nil {
		log.Fatal().Msgf("### Failed to initialize Message Broker: %v", err)
	}
	defer runtimeMB.Close()

	historyMB, err := rabbit.NewMessageBroker(cfg.RabbitAddr, cfg.RabbitUser, cfg.RabbitPassword, cfg.RabbitHistoryEventsQueue)
	if err != nil {
		log.Fatal().Msgf("### Failed to initialize Message Broker: %v", err)
	}
	defer historyMB.Close()

	plugin, err := detector.NewPlugin(context.Background())
	if err != nil {
		log.Fatal().Msgf("### Failed to instantiate detector: %v", err)
	}

	if err := ensureDetectors(db, plugin, cfg.DeployDir); err != nil {
		log.Fatal().Msgf("### Failed to add default detectors: %v", err)
	}

	pool, err := getPool(cfg.WorkersPoolSize, cfg.JobsBufferSize, db, historyMB, plugin, enforcer, notifier)
	if err != nil {
		log.Fatal().Msgf("### Failed to initialize workers pool: %v", err)
	}
	defer pool.Close()

	grpcSrv := grpc.NewServer(opts...)
	configSvc, detectorSvc := composeServices(db, pool, plugin, verifier, cfg.Auth)

	api.RegisterConfigControllerServer(grpcSrv, configSvc)
	api.RegisterDetectorControllerServer(grpcSrv, detectorSvc)

	// Register reflection service on gRPC server
	reflection.Register(grpcSrv)

	// Create and Run the instrumentation HTTP server for probes, etc.
	iSrv := server.NewInstrumentation(cfg.InstrumentationAddr)
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

	// Run events consumer
	go eventsConsumer(runtimeMB, pool)

	// Check pool config and detectors for updates periodically
	go poolUpdater(cfg.ConfigUpdateInterval, pool, db)

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
	pool *processor.WorkersPool,
	plugin *detector_api.DetectorPlugin,
	verifier jwt.Verifier,
	isAuth bool,
) (configSvc api.ConfigControllerServer, detectorSvc api.DetectorControllerServer) {
	configSvc = &service.ConfigGeneric{
		Processor:        pool,
		ConfigRepository: &database.ConfigDatabase{db},
	}

	detectorSvc = &service.DetectorGeneric{
		Processor:          pool,
		DetectorRepository: &database.DetectorDatabase{db},
		DetectorPlugin:     plugin,
	}

	if isAuth {
		configSvc = &service.ConfigAuth{
			ConfigControllerServer: configSvc,
			Verifier:               verifier,
		}

		detectorSvc = &service.DetectorAuth{
			DetectorControllerServer: detectorSvc,
			Verifier:                 verifier,
		}
	}

	configSvc = &service.ConfigLogging{configSvc}
	detectorSvc = &service.DetectorLogging{detectorSvc}

	return
}

func eventsConsumer(mb *rabbit.MessageBroker, wp *processor.WorkersPool) {
	c := &consumer.Consumer{
		PublishConsumer: mb,
		Processor:       wp,
	}

	c.Run(shutdown)
}

func ensureDetectors(db *gorm.DB, plugin *detector_api.DetectorPlugin, deployDir string) error {
	ctx := context.Background()
	repo := &database.DetectorDatabase{db}

	count, err := repo.GetCount(ctx, nil)
	if err != nil {
		return fmt.Errorf("can't get detectors count: %w", err)
	}

	// There might be better criteria
	if count > 0 {
		return nil
	}

	wasmFiles, err := filepath.Glob(filepath.Clean(deployDir + "/" + wasmPattern))
	if err != nil {
		return fmt.Errorf("can't get detector names by pattern: %w", err)
	}

	ds := []*model.Detector{}

	for _, file := range wasmFiles {
		b, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("can't read wasm file '%s': %w", file, err)
		}

		d, err := detector.BinToModel(ctx, plugin, b)
		if err != nil {
			return fmt.Errorf("can't get detector info from wasm file '%s': %w", file, err)
		}

		ds = append(ds, d)
	}

	if err := repo.Add(ctx, ds...); err != nil {
		return fmt.Errorf("can't add detectors to DB: %w", err)
	}

	return nil
}

func getPool(
	poolSize int,
	bufferSize int,
	db *gorm.DB,
	mb *rabbit.MessageBroker,
	plugin *detector_api.DetectorPlugin,
	enforcer enforcer_api.EnforcerClient,
	notifier notifier_api.NotifierClient,
) (*processor.WorkersPool, error) {
	ctx := context.Background()
	configRepo := &database.ConfigDatabase{db}
	detectorRepo := &database.DetectorDatabase{db}

	bins, err := detectorRepo.GetAllBins(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("can't get detector binaries from DB: %w", err)
	}

	cfg, err := configRepo.GetLast(ctx, true) // preload is on
	if err != nil {
		return nil, fmt.Errorf("can't get last config from DB: %w", err)
	}

	pool, err := processor.NewWorkersPool(poolSize, bufferSize, mb, plugin, enforcer, notifier, bins, cfg)
	if err != nil {
		return nil, fmt.Errorf("can't initialize workers pool: %w", err)
	}

	return pool, nil
}

func poolUpdater(interval time.Duration, pool *processor.WorkersPool, db *gorm.DB) {
	u := updater.Updater{
		interval,
		pool,
		&database.ConfigDatabase{db},
		&database.DetectorDatabase{db},
	}

	u.Run(shutdown)
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
