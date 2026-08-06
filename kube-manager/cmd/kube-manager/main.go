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
	"github.com/runtime-radar/runtime-radar/kube-manager/api"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/build"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/config"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/database"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/informers"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/inventory"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/metrics"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/server"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/service"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/updater"
	"github.com/runtime-radar/runtime-radar/lib/logger"
	"github.com/runtime-radar/runtime-radar/lib/security"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	"github.com/runtime-radar/runtime-radar/lib/server/healthcheck"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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
	defer func() { _ = closeDB() }()

	// Recreate DB from scratch, or migrate automatically when needed
	if err := database.Migrate(db, cfg.NewDB); err != nil {
		log.Fatal().Msgf("### Failed to migrate DB: %v", err)
	}

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

	ctx := context.Background()
	updateSrv, err := updater.New(ctx, db, cfg.ConfigUpdateInterval)
	if err != nil {
		log.Fatal().Msgf("### Failed to initialize updater: %v", err)
	}
	go updateSrv.Run(shutdown)

	var k8sCfg *rest.Config
	if cfg.KubeConfig != "" {
		k8sCfg, err = clientcmd.BuildConfigFromFlags("", cfg.KubeConfig)
		if err != nil {
			log.Fatal().Msgf("### Failed to initialize Kubernetes config: %v", err)
		}
	} else {
		k8sCfg, err = rest.InClusterConfig()
		if err != nil {
			log.Fatal().Msgf("### Failed to get in-cluster Kubernetes config: %v", err)
		}
	}

	infs := createInformers()

	configService, podService, nodeService := composeServices(
		db, cfg, updateSrv, verifier, infs,
	)

	inv, err := inventory.New(updateSrv, k8sCfg, cfg.K8SSyncInterval, infs.setters()...)
	if err != nil {
		log.Fatal().Msgf("### Failed to initialize inventory: %v", err)
	}

	if err := inv.Run(shutdown); err != nil {
		log.Fatal().Msgf("### Failed to run inventory: %v", err)
	}
	defer inv.Shutdown()

	api.RegisterPodControllerServer(grpcSrv, podService)
	api.RegisterNodeControllerServer(grpcSrv, nodeService)
	api.RegisterConfigControllerServer(grpcSrv, configService)

	// Register reflection kube-manager on gRPC server
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
			if err := httpSrv.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatal().Msgf("### Can't serve HTTP requests: %v", err)
			}
		} else {
			if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
	_ = httpSrv.Shutdown(ctx) // we don't care about errors here

	log.Info().Msg("Instrumentation HTTP server stopping gracefully")
	_ = iSrv.Shutdown(ctx)
}

func composeServices(
	db *gorm.DB,
	cfg *config.Config,
	updater *updater.Service,
	verifier jwt.Verifier,
	infs *Informers,
) (
	configService api.ConfigControllerServer,
	podService api.PodControllerServer,
	nodeService api.NodeControllerServer,
) {
	configService = &service.ConfigGeneric{
		ConfigUpdater: updater,
		ConfigRepository: &database.ConfigDatabase{
			DB: db,
		},
	}
	podService = &service.PodGeneric{
		Pods: infs.Pods,
	}
	nodeService = &service.NodeGeneric{
		Nodes: infs.Nodes,
	}

	if cfg.Auth {
		configService = &service.ConfigAuth{
			ConfigControllerServer: configService,
			Verifier:               verifier,
		}
		podService = &service.PodAuth{
			PodControllerServer: podService,
			Verifier:            verifier,
		}
		nodeService = &service.NodeAuth{
			NodeControllerServer: nodeService,
			Verifier:             verifier,
		}
	}

	configService = &service.ConfigLogging{ConfigControllerServer: configService}
	podService = &service.PodLogging{PodControllerServer: podService}
	nodeService = &service.NodeLogging{NodeControllerServer: nodeService}

	return configService, podService, nodeService
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

type Informers struct {
	Pods  *informers.Informer[*corev1.Pod]
	Nodes *informers.Informer[*corev1.Node]
}

func (i *Informers) setters() []informers.Setter {
	return []informers.Setter{
		i.Pods, i.Nodes,
	}
}

func createInformers() *Informers {
	return &Informers{
		Pods:  informers.New[*corev1.Pod]("pods"),
		Nodes: informers.New[*corev1.Node]("nodes"),
	}
}
