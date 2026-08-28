package config

import (
	"flag"

	"github.com/runtime-radar/runtime-radar/lib/config"
)

// Config represents system configuration.
type Config struct {
	NewDB                  bool              // forces recreation of DB
	PostgresAddr           string            // Postgres address in host[:port] format
	PostgresDB             string            // Postgres db name
	PostgresUser           string            // Postgres user
	PostgresPassword       string            // Postgres password
	PostgresSSLMode        bool              // Postgres SSL mode
	PostgresSSLCheckCert   bool              // Check postgres SSL cert
	LogLevel               string            // log level can be INFO, WARN, ERROR, FATAL, DEBUG or ALL
	CORSAllowedOrigins     config.StringList // origins allowed to call the API cross-origin
	LogFile                string            // path to log file
	ListenHTTPAddr         string            // address "[host]:port" that server should be listening for health checks
	InstrumentationAddr    string            // address "[host]:port" that instrumentation server should be listening for health checks and metrics
	TLS                    bool              // is TLS enabled?
	AuthAPIURL             string            // Auth API URL in schema://host[:port] format
	PolicyEnforcerGRPCAddr string            // Policy Enforcer gRPC address in host[:port] format
	HistoryAPIGRPCAddr     string            // History API gRPC address in host[:port] format
	KubeManagerGRPCAddr    string            // Kube Manager gRPC address in host[:port] format
	TokenKey               string            // key for jwt token
	AccessTokenSalt        string            // salt for access token
	Auth                   bool              // is auth enabled?
	GopsAddr               string            // gops listen address
	OwnCSURL               string            // URL of current CS (http(s)://host[:port]).
}

// New reads config from environment and returns pointer to a new Config.
func New() *Config {
	c := &Config{}

	flag.BoolVar(&c.NewDB, "newDB", config.LookupEnvBool("NEW_DB", false), "Set this flag to recreate DB from scratch. ALL EXISTING DATA WILL BE LOST.")
	flag.StringVar(&c.PostgresAddr, "postgresAddr", config.LookupEnvString("POSTGRES_ADDR", "127.0.0.1:5432"), "Set PostgreSQL address as host:port, where port is optional (without TLS).")
	flag.StringVar(&c.PostgresDB, "postgresDB", config.LookupEnvString("POSTGRES_DB", "cs"), "Set PostgreSQL DB.")
	flag.StringVar(&c.PostgresUser, "postgresUser", config.LookupEnvString("POSTGRES_USER", "cs"), "Set PostgreSQL user.")
	flag.StringVar(&c.PostgresPassword, "postgresPassword", config.LookupEnvString("POSTGRES_PASSWORD", "cs"), "Set PostgreSQL password.")
	flag.BoolVar(&c.PostgresSSLMode, "postgresSSLMode", config.LookupEnvBool("POSTGRES_SSL_MODE", false), "Set to enable PostgreSQL SSL mode.")
	flag.BoolVar(&c.PostgresSSLCheckCert, "postgresSSLCheckCert", config.LookupEnvBool("POSTGRES_SSL_CHECK_CERT", false), "Set to check PostgreSQL SSL cert.")
	flag.StringVar(&c.LogLevel, "logLevel", config.LookupEnvString("LOG_LEVEL", "TRACE"), "Set log level (DEBUG, INFO, WARN, ERROR, FATAL, any other value means TRACE).")
	flag.StringVar(&c.LogFile, "logFile", config.LookupEnvString("LOG_FILE", ""), "Tail logs to this file. Leave empty to log to stdout without tailing.")
	flag.StringVar(&c.ListenHTTPAddr, "listenHTTPAddr", config.LookupEnvString("LISTEN_HTTP_ADDR", ":9000"), `Address in form of "[host]:port" that HTTP server should be listening on.`)
	flag.BoolVar(&c.TLS, "tls", config.LookupEnvBool("TLS", false), "Set to enable TLS.")
	flag.StringVar(&c.AuthAPIURL, "authAPIURL", config.LookupEnvString("AUTH_API_URL", "http://127.0.0.1:8080"), "Auth API URL in schema://host[:port] format.")
	flag.StringVar(&c.PolicyEnforcerGRPCAddr, "policyEnforcerGRPCAddr", config.LookupEnvString("POLICY_ENFORCER_GRPC_ADDR", "127.0.0.1:10000"), "Policy Enforcer gRPC address in host[:port] format.")
	flag.StringVar(&c.HistoryAPIGRPCAddr, "historyAPIGRPCAddr", config.LookupEnvString("HISTORY_API_GRPC_ADDR", "127.0.0.1:10000"), "History API gRPC address in host[:port] format.")
	flag.StringVar(&c.KubeManagerGRPCAddr, "kubeManagerGRPCAddr", config.LookupEnvString("KUBE_MANAGER_GRPC_ADDR", "127.0.0.1:10000"), "Kube Manager gRPC address in host[:port] format.")
	flag.StringVar(&c.TokenKey, "tokenKey", config.LookupEnvString("TOKEN_KEY", ""), "Hex encoded token key to verify jwt token. Supported key sizes are greater than 32 bytes.")
	flag.StringVar(&c.AccessTokenSalt, "accessTokenSalt", config.LookupEnvString("ACCESS_TOKEN_SALT", ""), "Hex encoded token salt to verify access token. Salt must have 64 bytes length.")
	flag.BoolVar(&c.Auth, "auth", config.LookupEnvBool("AUTH", false), "Set to enable JWT auth.")
	flag.StringVar(&c.GopsAddr, "listenGopsAddr", config.LookupEnvString("LISTEN_GOPS_ADDR", "127.0.0.1:7000"), `Address in form of "[host]:port" that gops agent should be listening on. It's not safe to listen to interfaces other than loopback in production.`)
	flag.StringVar(&c.OwnCSURL, "ownCSURL", config.LookupEnvString("OWN_CS_URL", ""), "URL of current CS (http(s)://host[:port]).")
	flag.StringVar(&c.InstrumentationAddr, "listenInstrumentationAddr", config.LookupEnvString("LISTEN_INSTRUMENTATION_ADDR", ":9090"), `Address in form of "[host]:port" that instrumentation HTTP server should be listening on.`)

	config.StringListVar(&c.CORSAllowedOrigins, "corsAllowedOrigins", "CORS_ALLOWED_ORIGINS", "", "Comma separated list of origins allowed to call the API cross-origin. Empty value allows same-origin requests only.")

	flag.Parse()

	return c
}
