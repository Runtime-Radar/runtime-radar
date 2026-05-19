package config

import (
	"flag"
	"time"

	"github.com/runtime-radar/runtime-radar/lib/config"
)

// Config represents system configuration.
type Config struct {
	NewDB                      bool          // forces recreation of DB
	PopulateNum                int           // populates DB with some test data according to given number
	PostgresAddr               string        // Postgres address in host[:port] format
	PostgresDB                 string        // Postgres db name
	PostgresUser               string        // Postgres user
	PostgresPassword           string        // Postgres password
	PostgresSSLMode            bool          // Postgres SSL mode
	PostgresSSLCheckCert       bool          // Check postgres SSL cert
	ClickhouseAddr             string        // Clickhouse address in host[:port] format
	ClickhouseDB               string        // Clickhouse db name
	ClickhouseUser             string        // Clickhouse user
	ClickhousePassword         string        // Clickhouse password
	ClickhouseSSLMode          bool          // Clickhouse SSL mode
	ClickhouseSSLCheckCert     bool          // Check clickhouse SSL cert
	RabbitAddr                 string        // RabbitMQ address in host[:port] format
	RabbitUser                 string        // RabbitMQ user
	RabbitPassword             string        // RabbitMQ password
	RabbitQueue                string        // RabbitMQ queue name to consume events from
	RabbitQueuePrefetchCount   int           // RabbitMQ prefetch count for queue to consume events from
	RuntimeEventsBatchSize     int           // Size of runtime events buffer
	RuntimeEventsSaveInterval  time.Duration // Interval between savings of runtime buffer
	RuntimeEventsLimit         int           // Max number of runtime events to be stored in database
	RuntimeEventsCleanInterval time.Duration // Interval between cleans of runtime events table
	LogLevel                   string        // log level can be INFO, WARN, ERROR, FATAL, DEBUG or ALL
	LogFile                    string        // path to log file
	ListenGRPCAddr             string        // address "[host]:port" that server should be listening on
	ListenHTTPAddr             string        // address "[host]:port" that server should be listening for health checks
	InstrumentationAddr        string        // address "[host]:port" that instrumentation server should be listening on
	TLS                        bool          // is TLS enabled?
	RetentionInterval          time.Duration // Interval during which events are kept in DB
	TokenKey                   string        // key for jwt token
	Auth                       bool          // is auth enabled?
	OwnCSURL                   string        // URL of current CS (http(s)://host[:port]).
	GopsAddr                   string        // gops listen address
}

// New reads config from environment and returns pointer to a new Config.
func New() *Config {
	c := &Config{}

	flag.BoolVar(&c.NewDB, "newDB", config.LookupEnvBool("NEW_DB", false), "Set this flag to recreate DB from scratch. ALL EXISTING DATA WILL BE LOST.")
	flag.IntVar(&c.PopulateNum, "populateNum", config.LookupEnvInt("POPULATE_NUM", 0), "Set the number of test entries DB should be populated with.")
	flag.StringVar(&c.PostgresAddr, "postgresAddr", config.LookupEnvString("POSTGRES_ADDR", "127.0.0.1:5432"), "Set PostgreSQL address as host:port, where port is optional (without TLS).")
	flag.StringVar(&c.PostgresDB, "postgresDB", config.LookupEnvString("POSTGRES_DB", "cs"), "Set PostgreSQL DB.")
	flag.StringVar(&c.PostgresUser, "postgresUser", config.LookupEnvString("POSTGRES_USER", "cs"), "Set PostgreSQL user.")
	flag.StringVar(&c.PostgresPassword, "postgresPassword", config.LookupEnvString("POSTGRES_PASSWORD", "cs"), "Set PostgreSQL password.")
	flag.BoolVar(&c.PostgresSSLMode, "postgresSSLMode", config.LookupEnvBool("POSTGRES_SSL_MODE", false), "Set to enable PostgreSQL SSL mode.")
	flag.BoolVar(&c.PostgresSSLCheckCert, "postgresSSLCheckCert", config.LookupEnvBool("POSTGRES_SSL_CHECK_CERT", false), "Set to check PostgreSQL SSL cert.")
	flag.StringVar(&c.ClickhouseAddr, "clickhouseAddr", config.LookupEnvString("CLICKHOUSE_ADDR", "127.0.0.1:19000"), "Set Clickhouse address as host:port, where port is optional (without TLS).")
	flag.StringVar(&c.ClickhouseDB, "clickhouseDB", config.LookupEnvString("CLICKHOUSE_DB", "cs"), "Set Clickhouse DB.")
	flag.StringVar(&c.ClickhouseUser, "clickhouseUser", config.LookupEnvString("CLICKHOUSE_USER", "clickhouse"), "Set Clickhouse user.")
	flag.StringVar(&c.ClickhousePassword, "clickhousePassword", config.LookupEnvString("CLICKHOUSE_PASSWORD", "clickhouse"), "Set Clickhouse password.")
	flag.BoolVar(&c.ClickhouseSSLMode, "clickhouseSSLMode", config.LookupEnvBool("CLICKHOUSE_SSL_MODE", false), "Set to enable Clickhouse SSL mode.")
	flag.BoolVar(&c.ClickhouseSSLCheckCert, "clickhouseSSLCheckCert", config.LookupEnvBool("CLICKHOUSE_SSL_CHECK_CERT", false), "Set to check Clickhouse SSL cert.")
	flag.StringVar(&c.RabbitAddr, "rabbitAddr", config.LookupEnvString("RABBIT_ADDR", "127.0.0.1:5672"), "Set RabbitMQ address in host[:port] format.")
	flag.StringVar(&c.RabbitUser, "rabbitUser", config.LookupEnvString("RABBIT_USER", "guest"), "Set RabbitMQ user.")
	flag.StringVar(&c.RabbitPassword, "rabbitPassword", config.LookupEnvString("RABBIT_PASSWORD", "guest"), "Set RabbitMQ password.")
	flag.StringVar(&c.RabbitQueue, "rabbitQueue", config.LookupEnvString("RABBIT_QUEUE", "history_events"), "Set RabbitMQ queue name to consume runtime events from.")
	flag.IntVar(&c.RabbitQueuePrefetchCount, "rabbitQueuePrefetchCount", config.LookupEnvInt("RABBIT_QUEUE_PREFETCH_COUNT", 100), "Set RabbitMQ prefetch count for queue to consume events from.")
	flag.IntVar(&c.RuntimeEventsBatchSize, "runtimeEventsBatchSize", config.LookupEnvInt("RUNTIME_EVENTS_BATCH_SIZE", 10000), "Size of runtime events buffer.")
	flag.DurationVar(&c.RuntimeEventsSaveInterval, "runtimeEventsSaveInterval", config.LookupEnvDuration("RUNTIME_EVENTS_SAVE_INTERVAL", 5*time.Second), `Interval between savings of runtime buffer, as a duration string which is compatibale with time.ParseDuraion, for example "1s"`)
	flag.IntVar(&c.RuntimeEventsLimit, "runtimeEventsLimit", config.LookupEnvInt("RUNTIME_EVENTS_LIMIT", 100000000), "Maximum number of runtime events to be stored in database.")
	flag.DurationVar(&c.RuntimeEventsCleanInterval, "runtimeEventsCleanInterval", config.LookupEnvDuration("RUNTIME_EVENTS_CLEAN_INTERVAL", time.Hour), `Interval between cleans of runtime events table, as a duration string which is compatibale with time.ParseDuraion, for example "1s"`)
	flag.StringVar(&c.LogLevel, "logLevel", config.LookupEnvString("LOG_LEVEL", "TRACE"), "Set log level (DEBUG, INFO, WARN, ERROR, FATAL, any other value means TRACE).")
	flag.StringVar(&c.LogFile, "logFile", config.LookupEnvString("LOG_FILE", ""), "Tail logs to this file. Leave empty to log to stdout without tailing.")
	flag.StringVar(&c.ListenGRPCAddr, "listenGRPCAddr", config.LookupEnvString("LISTEN_GRPC_ADDR", ":8000"), `Address in form of "[host]:port" that gRPC server should be listening on.`)
	flag.StringVar(&c.ListenHTTPAddr, "listenHTTPAddr", config.LookupEnvString("LISTEN_HTTP_ADDR", ":9000"), `Address in form of "[host]:port" that HTTP server should be listening on.`)
	flag.BoolVar(&c.TLS, "tls", config.LookupEnvBool("TLS", false), "Set to enable TLS.")
	flag.DurationVar(&c.RetentionInterval, "retentionInterval", config.LookupEnvDuration("RETENTION_INTERVAL", time.Hour*24*30), `History retention interval, for how long will stored events be kept in DB, as a duration string which is compatibale with time.ParseDuraion, for example "24h"`)
	flag.StringVar(&c.TokenKey, "tokenKey", config.LookupEnvString("TOKEN_KEY", ""), "Hex encoded token key to verify jwt token. Supported key sizes are 16, 24 and 32 bytes.")
	flag.BoolVar(&c.Auth, "auth", config.LookupEnvBool("AUTH", false), "Set to enable JWT auth.")
	flag.StringVar(&c.GopsAddr, "listenGopsAddr", config.LookupEnvString("LISTEN_GOPS_ADDR", "127.0.0.1:7000"), `Address in form of "[host]:port" that gops agent should be listening on. It's not safe to listen to interfaces other than loopback in production.`)
	flag.StringVar(&c.OwnCSURL, "ownCSURL", config.LookupEnvString("OWN_CS_URL", ""), "URL of current CS (http(s)://host[:port]).")
	flag.StringVar(&c.InstrumentationAddr, "listenInstrumentationAddr", config.LookupEnvString("LISTEN_INSTRUMENTATION_ADDR", ":9090"), `Address in form of "[host]:port" that instrumentation (metrics, probes...) HTTP server should be listening on.`)

	flag.Parse()

	return c
}
