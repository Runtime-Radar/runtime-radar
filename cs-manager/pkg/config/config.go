package config

import (
	"flag"
	"time"

	"github.com/runtime-radar/runtime-radar/lib/config"
)

// Config represents system configuration.
type Config struct {
	NewDB                 bool          // forces recreation of DB
	PostgresAddr          string        // Postgres address in host[:port] format
	PostgresDB            string        // Postgres db name
	PostgresUser          string        // Postgres user
	PostgresPassword      string        // Postgres password
	PostgresSSLMode       bool          // Postgres SSL mode
	PostgresSSLCheckCert  bool          // Check postgres SSL cert
	LogLevel              string        // log level can be INFO, WARN, ERROR, FATAL, DEBUG or ALL
	LogFile               string        // path to log file
	ListenGRPCAddr        string        // address "[host]:port" that server should be listening on
	ListenHTTPAddr        string        // address "[host]:port" that server should be listening for health checks
	InstrumentationAddr   string        // address "[host]:port" that instrumentation server should be listening for health checks and metrics
	TLS                   bool          // is TLS enabled?
	CSVersion             string        // CS version
	TokenKey              string        // key for jwt token
	Auth                  bool          // is auth enabled?
	GopsAddr              string        // gops listen address
	IsChildCluster        bool          // is CS in child cluster
	OwnCSURL              string        // URL of current CS
	CentralCSURL          string        // URL of central CS
	GrafanaURL            string        // URL to Grafana with CS metrics
	CentralCSTLSCheckCert bool          // Check central CS TLS certificate
	RegistrationToken     string        // token of current cluster to register in central CS
	RegistrationInterval  time.Duration // interval for registration in central CS
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
	flag.StringVar(&c.ListenGRPCAddr, "listenGRPCAddr", config.LookupEnvString("LISTEN_GRPC_ADDR", ":8000"), `Address in form of "[host]:port" that gRPC server should be listening on.`)
	flag.StringVar(&c.ListenHTTPAddr, "listenHTTPAddr", config.LookupEnvString("LISTEN_HTTP_ADDR", ":9000"), `Address in form of "[host]:port" that HTTP server should be listening on.`)
	flag.BoolVar(&c.TLS, "tls", config.LookupEnvBool("TLS", false), "Set to enable TLS.")
	flag.StringVar(&c.CSVersion, "csVersion", config.LookupEnvString("CS_VERSION", "v0.0.0"), "CS version.")
	flag.StringVar(&c.TokenKey, "tokenKey", config.LookupEnvString("TOKEN_KEY", ""), "Hex encoded token key to verify jwt token. Supported key sizes are 16, 24 and 32 bytes.")
	flag.BoolVar(&c.Auth, "auth", config.LookupEnvBool("AUTH", false), "Set to enable JWT auth.")
	flag.StringVar(&c.GopsAddr, "listenGopsAddr", config.LookupEnvString("LISTEN_GOPS_ADDR", "127.0.0.1:7000"), `Address in form of "[host]:port" that gops agent should be listening on. It's not safe to listen to interfaces other than loopback in production.`)
	flag.BoolVar(&c.IsChildCluster, "isChildCluster", config.LookupEnvBool("IS_CHILD_CLUSTER", false), "Is CS in child cluster.")
	flag.StringVar(&c.OwnCSURL, "ownCSURL", config.LookupEnvString("OWN_CS_URL", ""), "URL of current CS (http(s)://host[:port]).")
	flag.StringVar(&c.CentralCSURL, "centralCSURL", config.LookupEnvString("CENTRAL_CS_URL", ""), "URL of central CS (http(s)://host[:port]).")
	flag.StringVar(&c.GrafanaURL, "grafanaURL", config.LookupEnvString("GRAFANA_URL", ""), "URL to Grafana with CS metrics (http(s)://host[:port]).")
	flag.BoolVar(&c.CentralCSTLSCheckCert, "centralCSTLSCheckCert", config.LookupEnvBool("CENTRAL_CS_TLS_CHECK_CERT", true), "Set to check central CS TLS certificate.")
	flag.StringVar(&c.RegistrationToken, "registrationToken", config.LookupEnvString("REGISTRATION_TOKEN", ""), `Token of current cluster for use in management procedures with central cluster. Token consists of 16 hex-encoded bytes. Different formats supported, such as "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", "urn:uuid:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" or "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx". For full list and details please refer to uuid.Parse documentation.`)
	flag.DurationVar(&c.RegistrationInterval, "registrationInterval", config.LookupEnvDuration("REGISTRATION_INTERVAL", time.Minute*5), "Interval for registration in central CS.")
	flag.StringVar(&c.InstrumentationAddr, "listenInstrumentationAddr", config.LookupEnvString("LISTEN_INSTRUMENTATION_ADDR", ":9090"), `Address in form of "[host]:port" that instrumentation HTTP server should be listening on.`)

	flag.Parse()

	return c
}
