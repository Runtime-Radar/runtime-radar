package main

import (
	"context"
	"crypto/tls"
	"encoding/hex"
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
	"github.com/runtime-radar/runtime-radar/auth-center/api"
	"github.com/runtime-radar/runtime-radar/auth-center/pkg/build"
	"github.com/runtime-radar/runtime-radar/auth-center/pkg/config"
	"github.com/runtime-radar/runtime-radar/auth-center/pkg/database"
	"github.com/runtime-radar/runtime-radar/auth-center/pkg/metrics"
	"github.com/runtime-radar/runtime-radar/auth-center/pkg/server"
	"github.com/runtime-radar/runtime-radar/auth-center/pkg/service"
	"github.com/runtime-radar/runtime-radar/auth-center/pkg/utils"
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
	// TLS CA cert file name.
	caFile = "ca.pem"

	// Timeout on graceful shutdown.
	gracefulTimeout = 15 * time.Second
)

// Channel for stopping the program.
var shutdown = make(chan struct{})

func main() {
	cfg := config.New()
	logger.Init("", cfg.LogLevel)

	log.Info().
		Str("build_release", build.Release).
		Str("build_branch", build.Branch).
		Str("build_commit", build.Commit).
		Str("build_date", build.Date).
		Msgf("-> %s started", build.AppName)
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
	db, closeDB, err := database.New(
		cfg.PostgresAddr,
		cfg.PostgresDB,
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresSSLMode,
		cfg.PostgresSSLCheckCert,
	)
	if err != nil {
		log.Fatal().Msgf("### Failed to open DB: %v", err)
	}
	defer closeDB()

	// Recreate DB from scratch, or migrate automatically when needed
	hashedAdminPassword, err := service.HashPassword(cfg.AdminPassword)
	if err != nil {
		log.Fatal().Msgf("### Failed to hash admin password: %v", err)
	}
	if err := database.Migrate(
		db, cfg.NewDB, cfg.AdminUsername,
		cfg.AdminEmail, hashedAdminPassword,
	); err != nil {
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

	tokenKey, err := hex.DecodeString(cfg.TokenKey)
	if err != nil {
		log.Fatal().Msgf("### Failed to decode token key: %v", err)
	}

	passwordCheckMap, err := utils.ReadPasswordDictionary(cfg.PathToPasswordSecList)
	if err != nil {
		log.Fatal().Msgf("### Can't open file with password dictionary: %v", err)
	}

	roleService, userService, authService := composeServices(
		db,
		verifier,
		cfg.Auth,
		tokenKey,
		passwordCheckMap,
		cfg.AccessTokenExpiration,
		cfg.RefreshTokenExpiration,
	)

	api.RegisterRoleControllerServer(grpcSrv, roleService)
	api.RegisterUserControllerServer(grpcSrv, userService)
	api.RegisterAuthControllerServer(grpcSrv, authService)

	// Register reflection auth-center on gRPC server
	reflection.Register(grpcSrv)

	// Initialize metrics.
	m, err := metrics.PrepareRegistry(build.AppName, cfg.OwnCSURL, grpcMetrics)
	if err != nil {
		log.Fatal().Msgf("### Failed to initialize metrics: %v", err)
	}

	// Create and Run the instrumentation HTTP server for probes, metrics, etc.
	iSrv := server.NewInstrumentation(cfg.InstrumentationAddr, m)
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
			if err := httpSrv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
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
	_ = httpSrv.Shutdown(ctx) // we don't care about errors here

	log.Info().Msg("Instrumentation HTTP server stopping gracefully")
	_ = iSrv.Shutdown(ctx)
}

func composeServices(
	db *gorm.DB,
	verifier jwt.Verifier,
	isAuth bool,
	tokenKey []byte,
	passMap map[string]struct{},
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,

) (roleService api.RoleControllerServer, userService api.UserControllerServer, authService api.AuthControllerServer) {
	roleService = &service.RoleGeneric{
		RoleRepository: &database.RoleDatabase{db}}

	userService = &service.UserGeneric{
		UserRepository:   &database.UserDatabase{db},
		TokenKey:         tokenKey,
		PasswordCheckMap: passMap,
		AccessTokenTTL:   accessTokenTTL,
		RefreshTokenTTL:  refreshTokenTTL,
	}

	authService = &service.AuthGeneric{
		UserRepository:   &database.UserDatabase{db},
		TokenKey:         tokenKey,
		PasswordCheckMap: passMap,
		AccessTokenTTL:   accessTokenTTL,
		RefreshTokenTTL:  refreshTokenTTL,
	}

	if isAuth {
		roleService = &service.RoleAuth{
			RoleControllerServer: roleService,
			Verifier:             verifier,
		}
		userService = &service.UserAuth{
			UserControllerServer: userService,
			Verifier:             verifier,
		}
		authService = &service.AuthAuth{
			AuthControllerServer: authService,
		}
	}

	roleService = &service.RoleLogging{RoleControllerServer: roleService}
	userService = &service.UserLogging{UserControllerServer: userService}
	authService = &service.AuthLogging{AuthControllerServer: authService}

	return roleService, userService, authService
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
