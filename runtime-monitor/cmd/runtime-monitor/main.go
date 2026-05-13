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
	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/lib/logger"
	"github.com/runtime-radar/runtime-radar/lib/rabbit"
	"github.com/runtime-radar/runtime-radar/lib/security"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	"github.com/runtime-radar/runtime-radar/lib/server/healthcheck"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	"github.com/runtime-radar/runtime-radar/runtime-monitor/api"
	"github.com/runtime-radar/runtime-radar/runtime-monitor/pkg/build"
	"github.com/runtime-radar/runtime-radar/runtime-monitor/pkg/config"
	"github.com/runtime-radar/runtime-radar/runtime-monitor/pkg/database"
	"github.com/runtime-radar/runtime-radar/runtime-monitor/pkg/monitor"
	"github.com/runtime-radar/runtime-radar/runtime-monitor/pkg/monitor/publisher"
	"github.com/runtime-radar/runtime-radar/runtime-monitor/pkg/monitor/updater"
	"github.com/runtime-radar/runtime-radar/runtime-monitor/pkg/server"
	"github.com/runtime-radar/runtime-radar/runtime-monitor/pkg/service"
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
	// var tokenKey []byte
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

	tetra, closeTetra, err := getTetra(db, cfg.TetragonAddr, cfg.TetragonEventsBuffer)
	if err != nil {
		log.Fatal().Msgf("### Can't initialize Tetragon: %v", err)
	}
	defer closeTetra()
	log.Info().Msgf("Connected to tetragon version %s at %s", tetra.Version, cfg.TetragonAddr)

	mb, err := rabbit.NewMessageBroker(cfg.RabbitAddr, cfg.RabbitUser, cfg.RabbitPassword, cfg.RabbitQueue)
	if err != nil {
		log.Fatal().Msgf("### Failed to initialize Message Broker: %v", err)
	}
	defer mb.Close()

	grpcSrv := grpc.NewServer(opts...)
	configSvc := composeServices(db, tetra, verifier, cfg.Auth)

	api.RegisterConfigControllerServer(grpcSrv, configSvc)

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

	// Run tetra monitor
	go func() {
		if err := tetra.Run(shutdown); err != nil {
			log.Error().Msgf("Failed to run monitor: %v", err)
			// Signal the whole system to stop if it's not yet the case, such as when we discover that Tetragon crashed.
			// In some rare cases, when k8s is stopping the pod, SIGTERM can arrive after Tetragon container stopped
			// before or after error is processed, thus creating a race for shutdown, so use special helper instead if just close.
			closeIfNotClosed(shutdown)
		}
	}()

	// Run events publisher
	go eventsPublisher(tetra, mb)

	// Check tetra config for periodic updates
	go tetraUpdater(cfg.ConfigUpdateInterval, tetra, db)

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

func getTetra(db *gorm.DB, addr string, bufferSize int) (*monitor.Tetra, func() error, error) {
	ctx := context.Background()

	tetraRepo := &database.ConfigDatabase{db}
	tetraConfig, err := tetraRepo.GetLast(ctx, true) // preload is on
	if err != nil {
		return nil, nil, err
	}

	tetra, closeTetra, err := monitor.NewTetra(addr, bufferSize)
	if err != nil {
		return nil, nil, err
	}

	if err := tetra.Init(ctx, tetraConfig); err != nil {
		return nil, nil, err
	}

	return tetra, closeTetra, nil
}

func tetraUpdater(interval time.Duration, tetra *monitor.Tetra, db *gorm.DB) {
	u := updater.Updater{
		Interval:         interval,
		ConfigRepository: &database.ConfigDatabase{db},
		Monitor:          tetra,
	}

	u.Run(shutdown)
}

func eventsPublisher(tetra *monitor.Tetra, mb *rabbit.MessageBroker) {
	p := &publisher.Publisher{
		Monitor:         tetra,
		PublishConsumer: mb,
	}

	p.Run(shutdown)
}

func composeServices(db *gorm.DB, monitor monitor.Monitor, verifier jwt.Verifier, isAuth bool) (configSvc api.ConfigControllerServer) {
	configSvc = &service.ConfigGeneric{
		ConfigRepository: &database.ConfigDatabase{db},
		Monitor:          monitor,
	}

	if isAuth {
		configSvc = &service.ConfigAuth{
			ConfigControllerServer: configSvc,
			Verifier:               verifier,
		}
	}

	configSvc = &service.ConfigLogging{configSvc}

	return
}

func signalListener() {
	defer closeIfNotClosed(shutdown)

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

// closeIfNotClosed closes ch in case if it wasn't closed already. For doing this it tries to read from channel and analyses
// result indicator. It panics if ch is nil.
func closeIfNotClosed[T any](ch chan T) {
	if ch == nil {
		panic("can't close nil channel")
	}

	ok := false

	defer func() {
		if ok {
			close(ch)
		}
	}()

	select {
	case _, ok = <-ch:
	default:
		ok = true
	}
}
