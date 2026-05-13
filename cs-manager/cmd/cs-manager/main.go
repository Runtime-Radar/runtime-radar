package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/gops/agent"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	cluster_api "github.com/runtime-radar/runtime-radar/cluster-manager/api"
	"github.com/runtime-radar/runtime-radar/cs-manager/api"
	"github.com/runtime-radar/runtime-radar/cs-manager/pkg/build"
	"github.com/runtime-radar/runtime-radar/cs-manager/pkg/client"
	"github.com/runtime-radar/runtime-radar/cs-manager/pkg/config"
	"github.com/runtime-radar/runtime-radar/cs-manager/pkg/database"
	"github.com/runtime-radar/runtime-radar/cs-manager/pkg/registrar"
	"github.com/runtime-radar/runtime-radar/cs-manager/pkg/server"
	"github.com/runtime-radar/runtime-radar/cs-manager/pkg/service"
	"github.com/runtime-radar/runtime-radar/cs-manager/pkg/state"
	"github.com/runtime-radar/runtime-radar/lib/logger"
	"github.com/runtime-radar/runtime-radar/lib/security"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	"github.com/runtime-radar/runtime-radar/lib/server/healthcheck"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
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
		grpc.ChainUnaryInterceptor(
			interceptor.Recovery,
			interceptor.Correlation,
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

	centralCSURL, err := parseCSURL(cfg.CentralCSURL)
	if err != nil {
		log.Fatal().Msgf("### Failed to parse central CS URL: %v", err)
	}

	s, err := getClusterState(cfg.IsChildCluster, db)
	if err != nil {
		log.Fatal().Msgf("### Failed to get cluster state: %v", err)
	}
	state.Set(s)

	if s != state.Central {
		childTLSConfig := tlsConfig.Clone()
		childTLSConfig.InsecureSkipVerify = !cfg.CentralCSTLSCheckCert
		if s == state.ChildUnregistered {
			token, err := uuid.Parse(cfg.RegistrationToken)
			if err != nil {
				log.Fatal().Msgf("### Failed to parse registration token of current CS: %v", err)
			}

			clusterController, closeCC, err := client.NewClusterController(centralCSURL.Host, childTLSConfig, tokenKey)
			if err != nil {
				log.Fatal().Msgf("### Failed to connect to central Cluster Manager: %v", err)
			}
			defer closeCC()

			// Make internal services without authentication to be used by registrar
			go csRegistrar(db, cfg.RegistrationInterval, token, clusterController)
		}
	}

	grpcSrv := grpc.NewServer(opts...)
	infoSvc := composeServices(cfg.CSVersion, centralCSURL, verifier, cfg.Auth)

	api.RegisterInfoControllerServer(grpcSrv, infoSvc)
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
	csVersion string,
	centralCSURL *url.URL,
	verifier jwt.Verifier,
	isAuth bool,
) (infoSvc api.InfoControllerServer) {
	infoSvc = &service.InfoGeneric{
		Version:      csVersion,
		CentralCSURL: centralCSURL,
	}

	if isAuth {
		infoSvc = &service.InfoAuth{
			InfoControllerServer: infoSvc,
			Verifier:             verifier,
		}
	}

	infoSvc = &service.InfoLogging{infoSvc}

	return
}

func csRegistrar(db *gorm.DB, interval time.Duration, token uuid.UUID, clusterController cluster_api.ClusterControllerClient) {
	r := &registrar.Registrar{
		interval,
		token,
		clusterController,
		&database.RegistrationDatabase{db},
	}

	r.Run(shutdown)
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

func getClusterState(isChildCluster bool, db *gorm.DB) (state.State, error) {
	if !isChildCluster {
		return state.Central, nil
	}

	repo := database.RegistrationDatabase{db}
	_, err := repo.GetLastSuccessful(context.Background(), false)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return state.ChildUnregistered, nil
	} else if err != nil {
		return 0, fmt.Errorf("can't get registration: %w", err)
	}

	return state.ChildRegistered, nil
}

func parseCSURL(rawURL string) (*url.URL, error) {
	if rawURL == "" {
		return &url.URL{}, nil
	}

	if !strings.Contains(rawURL, "://") {
		return nil, fmt.Errorf("wrong format, url should contain scheme: %s", rawURL)
	}

	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return nil, fmt.Errorf("can't parse url: %w", err)
	}

	// Explicitly set default port to avoid gRPC connection error when dialing with
	// a host ip address that doesn't include a port number
	// `transport: Error while dialing: dial tcp: address X.X.X.X: missing port in address`
	if parsedURL.Port() == "" {
		switch parsedURL.Scheme {
		case "http":
			parsedURL.Host = net.JoinHostPort(parsedURL.Hostname(), "80")
		case "https":
			parsedURL.Host = net.JoinHostPort(parsedURL.Hostname(), "443")
		default:
			return nil, fmt.Errorf("unsupported scheme: %s", rawURL)
		}
	}

	return parsedURL, nil
}
