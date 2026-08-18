package config

import (
	"flag"
	"time"

	"github.com/runtime-radar/runtime-radar/lib/config"
)

// Config represents system configuration.
type Config struct {
	NewDB                                 bool          // forces recreation of DB
	PostgresAddr                          string        // Postgres address in host[:port] format
	PostgresDB                            string        // Postgres db name
	PostgresUser                          string        // Postgres user
	PostgresPassword                      string        // Postgres password
	PostgresSSLMode                       bool          // Postgres SSL mode
	PostgresSSLCheckCert                  bool          // Check postgres SSL cert
	LogLevel                              string        // log level can be INFO, WARN, ERROR, FATAL, DEBUG or ALL
	CORSAllowedOrigins                    string        // comma separated origins allowed cross-origin
	LogFile                               string        // path to log file
	ListenGRPCAddr                        string        // address "[host]:port" that server should be listening on
	ListenHTTPAddr                        string        // address "[host]:port" that server should be listening on
	InstrumentationAddr                   string        // address "[host]:port" that instrumentation server should be listening for health checks and metrics
	TLS                                   bool          // is TLS enabled?
	TokenKey                              string        // key for jwt token
	Auth                                  bool          // is auth enabled?
	ConfigUpdateInterval                  time.Duration // interval for config periodic update check
	RabbitAddr                            string        // RabbitMQ address in host[:port] format
	RabbitUser                            string        // RabbitMQ user
	RabbitPassword                        string        // RabbitMQ password
	RabbitRuntimeEventsQueue              string        // RabbitMQ queue name for consuming runtime events
	RabbitRuntimeEventsQueuePrefetchCount int           // RabbitMQ prefetch count when consuming runtime events
	RabbitHistoryEventsQueue              string        // RabbitMQ queue name for publishing processed events
	GopsAddr                              string        // gops listen address
	WorkersPoolSize                       int           // how many workers will process the events in parallel, 0 means auto-adjust to available cpu cores (minimum 2)
	JobsBufferSize                        int           // how many events will be read and kept in memory when all workers are busy
	DeployDir                             string        // directory to load detectors from
	PolicyEnforcerGRPCAddr                string        // Policy Enforcer address in host[:port] format
	NotifierGRPCAddr                      string        // Notifier address in host[:port] format
	KubeManagerGRPCAddr                   string        // Kube Manager address in host[:port] format
	OwnCSURL                              string        // URL of current CS (http(s)://host[:port]).
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
	flag.StringVar(&c.TokenKey, "tokenKey", config.LookupEnvString("TOKEN_KEY", ""), "Hex encoded token key to verify jwt token. Supported key sizes are 16, 24 and 32 bytes.")
	flag.BoolVar(&c.Auth, "auth", config.LookupEnvBool("AUTH", false), "Set to enable JWT auth.")
	flag.DurationVar(&c.ConfigUpdateInterval, "configUpdateInterval", config.LookupEnvDuration("CONFIG_UPDATE_INTERVAL", 30*time.Second), "Set interval for Tetragon config periodic update check.")
	flag.StringVar(&c.RabbitAddr, "rabbitAddr", config.LookupEnvString("RABBIT_ADDR", "rabbitmq.default:5672"), "Set RabbitMQ address in host[:port] format.")
	flag.StringVar(&c.RabbitUser, "rabbitUser", config.LookupEnvString("RABBIT_USER", "guest"), "Set RabbitMQ user.")
	flag.StringVar(&c.RabbitPassword, "rabbitPassword", config.LookupEnvString("RABBIT_PASSWORD", "guest"), "Set RabbitMQ password.")
	flag.StringVar(&c.RabbitRuntimeEventsQueue, "rabbitRuntimeEventsQueue", config.LookupEnvString("RABBIT_RUNTIME_EVENTS_QUEUE", "runtime_events"), "Set RabbitMQ queue name for consuming runtime events.")
	flag.IntVar(&c.RabbitRuntimeEventsQueuePrefetchCount, "rabbitRuntimeEventsQueuePrefetchCount", config.LookupEnvInt("RABBIT_RUNTIME_EVENTS_QUEUE_PREFETCH_COUNT", 100), "Set RabbitMQ prefetch count when consuming runtime events.")
	flag.StringVar(&c.RabbitHistoryEventsQueue, "rabbitHistoryEventsQueue", config.LookupEnvString("RABBIT_HISTORY_EVENTS_QUEUE", "history_events"), "Set RabbitMQ prefetch count when consuming runtime events.")
	flag.StringVar(&c.GopsAddr, "listenGopsAddr", config.LookupEnvString("LISTEN_GOPS_ADDR", "127.0.0.1:7000"), `Address in form of "[host]:port" that gops agent should be listening on. It's not safe to listen to interfaces other than loopback in production.`)
	flag.IntVar(&c.WorkersPoolSize, "workersPoolSize", config.LookupEnvInt("WORKERS_POOL_SIZE", 0), "Set how many workers will process the events in parallel, 0 means auto-adjust to available cpu cores (minimum 2).")
	flag.IntVar(&c.JobsBufferSize, "jobsBufferSize", config.LookupEnvInt("JOBS_BUFFER_SIZE", 100), "Set how many events will be read and kept in memory when all workers are busy.")
	flag.StringVar(&c.DeployDir, "deployDir", config.LookupEnvString("DEPLOY_DIR", "deploy"), "Set directory to load detectors from (doesn't apply if DB is not empty).")
	flag.StringVar(&c.PolicyEnforcerGRPCAddr, "policyEnforcerGRPCAddr", config.LookupEnvString("POLICY_ENFORCER_GRPC_ADDR", "127.0.0.1:10000"), "Policy Enforcer gRPC address in host[:port] format.")
	flag.StringVar(&c.NotifierGRPCAddr, "notifierGRPCAddr", config.LookupEnvString("NOTIFIER_GRPC_ADDR", "127.0.0.1:11000"), "Notifier gRPC address in host[:port] format.")
	flag.StringVar(&c.KubeManagerGRPCAddr, "kubeManagerGRPCAddr", config.LookupEnvString("KUBE_MANAGER_GRPC_ADDR", "127.0.0.1:12000"), "Kube Manager gRPC address in host[:port] format.")
	flag.StringVar(&c.OwnCSURL, "ownCSURL", config.LookupEnvString("OWN_CS_URL", ""), "URL of current CS (http(s)://host[:port]).")
	flag.StringVar(&c.InstrumentationAddr, "listenInstrumentationAddr", config.LookupEnvString("LISTEN_INSTRUMENTATION_ADDR", ":9090"), `Address in form of "[host]:port" that instrumentation HTTP server should be listening on.`)

	flag.StringVar(&c.CORSAllowedOrigins, "corsAllowedOrigins", config.LookupEnvString("CORS_ALLOWED_ORIGINS", ""), "Comma separated list of origins allowed to call the API cross-origin. Empty value allows same-origin requests only.")

	flag.Parse()

	return c
}

// CORSOrigins returns the allowed origins; empty means same-origin only.
func (c *Config) CORSOrigins() []string {
	return config.SplitList(c.CORSAllowedOrigins)
}
