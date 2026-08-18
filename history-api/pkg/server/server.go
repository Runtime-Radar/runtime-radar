package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/justinas/alice"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/runtime-radar/runtime-radar/history-api/api"
	"github.com/runtime-radar/runtime-radar/lib/server/healthcheck"
	"github.com/runtime-radar/runtime-radar/lib/server/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	readTimeout  = 60 * time.Second
	writeTimeout = 60 * time.Second
	// Maximum message size for grpc request
	MaxRecvMsgSize = 10 * 1024 * 1024 // 10MB
)

// New constructs and configures new *http.Server capable of serving application and gRPC gateway endpoints.
func New(httpAddr, grpcAddr string, tlsConfig *tls.Config, corsAllowedOrigins []string) (*http.Server, error) {
	mux := http.NewServeMux()
	gwMux, err := newGWMux(context.Background(), grpcAddr, tlsConfig)
	if err != nil {
		return nil, err
	}

	h := setupRouter(mux, gwMux, corsAllowedOrigins)

	s := &http.Server{
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		Addr:         httpAddr,
		Handler:      h,
		TLSConfig:    tlsConfig,
	}

	return s, nil
}

func NewInstrumentation(listenAddress string, gatherer prometheus.Gatherer) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/ready", healthcheck.ReadyHandler)
	mux.HandleFunc("/live", healthcheck.LiveHandler)

	handler := promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{})
	mux.Handle("/metrics", handler)

	h := alice.New(
		middleware.Log,
		middleware.Recovery,
	).Then(mux)

	return &http.Server{
		Addr:    listenAddress,
		Handler: h,
	}
}

func setupRouter(mux *http.ServeMux, gwMux *runtime.ServeMux, corsAllowedOrigins []string) http.Handler {
	mux.Handle("/", gwMux)

	h := alice.New(
		middleware.Log,
		middleware.Recovery,
		middleware.CORS(corsAllowedOrigins, middleware.DefaultCORSHeaders),
	).Then(mux)

	return h
}

func newGWMux(ctx context.Context, grpcAddr string, tlsConfig *tls.Config) (*runtime.ServeMux, error) {
	m := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		// Configure to always output same snake_case field names as in .*proto
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}))

	var creds credentials.TransportCredentials
	if tlsConfig != nil {
		creds = credentials.NewTLS(tlsConfig)
	} else {
		creds = insecure.NewCredentials()
	}

	opts := []grpc.DialOption{grpc.WithTransportCredentials(creds), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(MaxRecvMsgSize))}

	if err := api.RegisterRuntimeHistoryHandlerFromEndpoint(ctx, m, grpcAddr, opts); err != nil {
		return nil, fmt.Errorf("can't register runtime history handler: %w", err)
	}

	if err := api.RegisterRuntimeStatsHandlerFromEndpoint(ctx, m, grpcAddr, opts); err != nil {
		return nil, fmt.Errorf("can't register runtime stats handler: %w", err)
	}

	return m, nil
}
