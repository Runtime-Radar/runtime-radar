package server

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/runtime-radar/runtime-radar/lib/server/healthcheck"
	"github.com/runtime-radar/runtime-radar/lib/server/middleware"
	local_middleware "github.com/runtime-radar/runtime-radar/public-api/pkg/server/middleware"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/service"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/service/constructor"
)

const (
	readTimeout  = 5 * time.Second
	writeTimeout = 5 * time.Second
)

// New constructs and configures new *http.Server capable of serving application endpoints.
func New(
	httpAddr string,
	tlsConfig *tls.Config,
	accessTokenSvc service.AccessToken,
	ruleSvc service.Rule,
	runtimeHistorySvc service.RuntimeHistory,
	configSvc service.Config,
	nodeSvc service.Node,
	podSvc service.Pod,
) *http.Server {
	r := mux.NewRouter()

	return &http.Server{
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		Addr:         httpAddr,
		Handler:      setupRouter(r, accessTokenSvc, ruleSvc, runtimeHistorySvc, configSvc, nodeSvc, podSvc),
		TLSConfig:    tlsConfig,
	}
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
		ReadTimeout: readTimeout,
		Addr:        listenAddress,
		Handler:     h,
	}
}

func setupRouter(
	r *mux.Router,
	accessTokenSvc service.AccessToken,
	ruleSvc service.Rule,
	runtimeHistorySvc service.RuntimeHistory,
	configSvc service.Config,
	nodeSvc service.Node,
	podSvc service.Pod,
) http.Handler {
	r.StrictSlash(true)

	corsOpts := cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders: []string{"Origin", "Accept", "Content-Type", "X-Requested-With", "Authorization", "X-Auth-Key"},
	}

	h := alice.New(
		middleware.Log,
		middleware.Recovery,
		cors.New(corsOpts).Handler,
		local_middleware.JWT,
		local_middleware.AccessToken,
		local_middleware.Correlation,
		local_middleware.Metrics,
	).Then(r)

	r.Handle("/api/v1/access-token", constructor.AccessTokenCreate(accessTokenSvc)).Methods(http.MethodPost)
	r.Handle("/api/v1/access-token/page/{page_num:[0-9]+}", constructor.AccessTokenListPage(accessTokenSvc)).Methods(http.MethodGet)
	r.Handle("/api/v1/access-token/{id}", constructor.AccessTokenDelete(accessTokenSvc)).Methods(http.MethodDelete)
	r.Handle("/api/v1/access-token/{id}", constructor.AccessTokenGetByID(accessTokenSvc)).Methods(http.MethodGet)
	r.Handle("/api/v1/access-token/invalidate-access-tokens", constructor.AccessTokenInvalidateAll(accessTokenSvc)).Methods(http.MethodPost)

	// policy-enforcer
	r.Handle("/api/v1/public-api/rule", constructor.RuleCreate(ruleSvc)).Methods(http.MethodPost)
	r.Handle("/api/v1/public-api/rule/page/{page_num:[0-9]+}", constructor.RuleListPage(ruleSvc)).Methods(http.MethodGet)
	r.Handle("/api/v1/public-api/rule/notify-targets-in-use", constructor.RuleNotifyTargetsInUse(ruleSvc)).Methods(http.MethodGet)
	r.Handle("/api/v1/public-api/rule/{id}", constructor.RuleRead(ruleSvc)).Methods(http.MethodGet)
	r.Handle("/api/v1/public-api/rule/{id}", constructor.RuleUpdate(ruleSvc)).Methods(http.MethodPatch)
	r.Handle("/api/v1/public-api/rule/{id}", constructor.RuleDelete(ruleSvc)).Methods(http.MethodDelete)

	// history-api
	r.Handle("/api/v1/public-api/runtime-event/slice/{direction:left|right}", constructor.RuntimeHistoryListEventsSlice(runtimeHistorySvc)).
		Methods(http.MethodGet).
		Queries("cursor", `{cursor:[a-zA-Z0-9\-:.]+}`)

	// kube-manager
	r.Handle("/api/v1/public-api/config/kube-manager", constructor.ConfigAdd(configSvc)).Methods(http.MethodPost)
	r.Handle("/api/v1/public-api/config/kube-manager", constructor.ConfigRead(configSvc)).Methods(http.MethodGet)
	r.Handle("/api/v1/public-api/node", constructor.NodeGet(nodeSvc)).Methods(http.MethodGet)
	r.Handle("/api/v1/public-api/node/list", constructor.NodeListMeta(nodeSvc)).Methods(http.MethodGet)
	r.Handle("/api/v1/public-api/node/page/{page_num:[0-9]+}", constructor.NodeListPage(nodeSvc)).Methods(http.MethodGet)
	r.Handle("/api/v1/public-api/pod", constructor.PodGet(podSvc)).Methods(http.MethodGet)
	r.Handle("/api/v1/public-api/pod/list", constructor.PodListMeta(podSvc)).Methods(http.MethodGet)
	r.Handle("/api/v1/public-api/pod/page/{page_num:[0-9]+}", constructor.PodListPage(podSvc)).Methods(http.MethodGet)

	return h
}
