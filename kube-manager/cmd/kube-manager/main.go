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
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/autoscaler"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/build"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/client"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/config"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/database"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/informers"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/inventory"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/metrics"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/server"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/service"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/updater"
	"github.com/runtime-radar/runtime-radar/lib/logger"
	"github.com/runtime-radar/runtime-radar/lib/rabbit"
	"github.com/runtime-radar/runtime-radar/lib/security"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	"github.com/runtime-radar/runtime-radar/lib/server/healthcheck"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
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

	if err := cfg.ValidateAutoscaler(); err != nil {
		log.Fatal().Msgf("### Invalid autoscaler config: %v", err)
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

	k8sClient, err := client.NewKubernetes(k8sCfg)
	if err != nil {
		log.Fatal().Msgf("### Failed to initialize Kubernetes client: %v", err)
	}

	infs := createInformers()

	inv, err := inventory.New(updateSrv, k8sCfg, cfg.K8SSyncInterval, infs.setters()...)
	if err != nil {
		log.Fatal().Msgf("### Failed to initialize inventory: %v", err)
	}

	configService, podService, nodeService := composeServices(
		db, cfg, updateSrv, verifier, k8sClient, infs, inv,
	)
	if err != nil {
		log.Fatal().Msgf("### Failed to initialize informers: %v", err)
	}

	if err := inv.Run(shutdown); err != nil {
		log.Fatal().Msgf("### Failed to run inventory: %v", err)
	}
	defer inv.Shutdown()

	autoscalerCtx, cancelAutoscaler := context.WithCancel(context.Background())
	defer cancelAutoscaler()

	autoscalerDone := make(chan struct{})

	if cfg.AutoscalerEnabled {
		identity, err := os.Hostname()
		if err != nil {
			log.Fatal().Msgf("### Failed to get hostname: %v", err)
		}

		clientset := k8sClient.Clientset()

		newProber, closeProbers := rabbitProbers(cfg)
		defer closeProbers()

		broadcaster := record.NewBroadcaster()
		broadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: clientset.CoreV1().Events("")})
		defer broadcaster.Shutdown()

		autoscalerSrv := buildAutoscaler(cfg, clientset, newProber,
			broadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: autoscaler.LeaseName}))

		go func() {
			defer close(autoscalerDone)

			if err := autoscaler.RunLeader(autoscalerCtx, clientset, cfg.CSNamespace, identity, autoscalerSrv.Run); err != nil {
				log.Fatal().Msgf("### Autoscaler leader election failed: %v", err)
			}
		}()
	} else {
		close(autoscalerDone)
		log.Info().Msg("Autoscaler disabled")
	}

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

	cancelAutoscaler()

	// One budget for the whole sequence: the pod is killed when the termination grace period ends,
	// however the steps below divide it up.
	ctx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
	defer cancel()

	log.Info().Msg("gRPC server stopping gracefully")
	grpcSrv.GracefulStop()

	log.Info().Msg("HTTP server stopping gracefully")
	_ = httpSrv.Shutdown(ctx) // we don't care about errors here

	log.Info().Msg("Instrumentation HTTP server stopping gracefully")
	_ = iSrv.Shutdown(ctx)

	// Releasing the lease is an API call that must finish before the process exits; it runs while
	// the servers above drain.
	select {
	case <-autoscalerDone:

	case <-ctx.Done():
		log.Warn().Msg("Autoscaler didn't stop in time")
	}
}

// rabbitProbers returns a factory of per-target queue probers, each backed by its own message
// broker opened on the first probe, and a closer for the brokers that were opened. Nothing dials
// until the replica leads and starts sampling, so an unreachable broker delays the scaling instead
// of holding up the rest of the service.
func rabbitProbers(cfg *config.Config) (newProber func(queue string) autoscaler.Prober, closeProbers func()) {
	var probers []*autoscaler.LazyProber

	newProber = func(queue string) autoscaler.Prober {
		p := &autoscaler.LazyProber{
			Dial: func() (autoscaler.ProberCloser, error) {
				return rabbit.NewMessageBroker(cfg.RabbitAddr, cfg.RabbitUser, cfg.RabbitPassword, queue)
			},
		}
		probers = append(probers, p)

		return p
	}

	closeProbers = func() {
		for _, p := range probers {
			_ = p.Close()
		}
	}

	return newProber, closeProbers
}

// buildAutoscaler wires the enabled targets, taking one prober per target from newProber.
func buildAutoscaler(
	cfg *config.Config,
	k8s kubernetes.Interface,
	newProber func(queue string) autoscaler.Prober,
	recorder record.EventRecorder,
) *autoscaler.Autoscaler {
	var targets []autoscaler.TargetConfig

	for _, target := range cfg.AutoscalerTargets() {
		if !target.Enabled {
			continue
		}

		targets = append(targets, autoscaler.TargetConfig{
			Name:        target.Name,
			Deployment:  target.Deployment,
			Queue:       target.Queue,
			Min:         int32(target.MinReplicas),
			Max:         int32(target.MaxReplicas),
			UpBacklog:   target.ScaleUpBacklog,
			UpWindow:    target.ScaleUpWindow,
			DownBacklog: target.ScaleDownBacklog,
			DownWindow:  target.ScaleDownWindow,
			Prober:      newProber(target.Queue),
		})
	}

	return &autoscaler.Autoscaler{
		Targets: targets,
		Scaler: &autoscaler.Scaler{
			Client:    k8s,
			Recorder:  recorder,
			Namespace: cfg.CSNamespace,
		},
		Interval: cfg.AutoscalerInterval,
		Cooldown: cfg.AutoscalerCooldown,
	}
}

func composeServices(
	db *gorm.DB,
	cfg *config.Config,
	updater *updater.Service,
	verifier jwt.Verifier,
	k8s *client.Kubernetes,
	infs *Informers,
	namespaceInformer informers.Getter[*corev1.Namespace],
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
		Namespaces: namespaceInformer,
		Pods:       infs.Pods,
		Kubernetes: k8s,
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
