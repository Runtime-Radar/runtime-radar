package clickhouse

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/history-api/pkg/model"
	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"github.com/runtime-radar/runtime-radar/lib/logger"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"
)

const (
	dialTimeout = "10s"
	readTimeout = "20s"
)

var errNoIndexFields = errors.New("no fields to index given")

// DateTimeFormat defines the layout of timestamp string to be passed to clickhouse's toDateTime64 function.
// Accodring to clickhouse's documenation, DateTime64 cannot be automatically converted from string so that toDateTime64 has to be called.
// See https://clickhouse.com/docs/en/sql-reference/data-types/datetime64 for details.
const DateTimeFormat = "2006-01-02 15:04:05.000000000"

func New(address, database, user, password string, sslMode, sslCheckCert bool) (*gorm.DB, func() error, error) {
	sslModeValue := 0
	if sslMode {
		sslModeValue = 1
	}

	url := fmt.Sprintf("tcp://%s/%s?username=%s&password=%s&secure=%d&dial_timeout=%s&read_timeout=%s", address, database, user, url.QueryEscape(password), sslModeValue, readTimeout, dialTimeout)

	if sslMode && !sslCheckCert {
		url += fmt.Sprintf("&skip_verify=true")
	}

	var gormLogger gorm_logger.Interface

	if e := log.Debug(); e.Enabled() {
		gormLogger = gorm_logger.New(
			&logger.GORM{&log.Logger},
			gorm_logger.Config{
				SlowThreshold: 100 * time.Millisecond, // Slow SQL threshold
				Colorful:      false,                  // Disable color
				LogLevel:      gorm_logger.Info,       // Log level
			},
		)
	} else {
		gormLogger = gorm_logger.New(
			&logger.GORM{&log.Logger},
			gorm_logger.Config{
				SlowThreshold: 100 * time.Millisecond, // Slow SQL threshold
				Colorful:      false,                  // Disable color
				LogLevel:      gorm_logger.Silent,     // Log level
			},
		)
	}

	db, err := gorm.Open(clickhouse.Open(url), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	return db, sqlDB.Close, nil
}

func Migrate(db *gorm.DB, newDB bool, populateNum int) error {
	if newDB {
		if err := db.Migrator().DropTable(
			&model.RuntimeEvent{},
		); err != nil {
			return err
		}
	}

	if err := db.
		Set("gorm:table_options", "ENGINE = MergeTree() PARTITION BY (toYYYYMM(registered_at)) ORDER BY (registered_at) SETTINGS index_granularity = 8192;").
		AutoMigrate(&model.RuntimeEvent{}); err != nil {
		return err
	}

	const tableName = "runtime_events"

	errs := []error{}

	toIndexBloomFilter := []struct {
		table  string
		fields []string
	}{
		{tableName, []string{"threats_detectors"}},
		{tableName, []string{"block_by", "notify_by"}}, // the exact order doesn't matter
	}
	for _, v := range toIndexBloomFilter {
		if err := ensureBloomFilterIndex(db, v.table, v.fields...); err != nil {
			errs = append(errs, err)
		}
	}

	toIndexTokenbf := []struct{ table, field string }{
		{tableName, "process_pod_name"},
		{tableName, "process_pod_namespace"},
	}
	for _, v := range toIndexTokenbf {
		if err := ensureTokenbfv1Index(db, v.table, v.field); err != nil {
			errs = append(errs, err)
		}
	}

	toIndexSet := []struct{ table, field string }{
		{tableName, "is_incident"},
	}
	for _, v := range toIndexSet {
		if err := ensureSetIndex(db, v.table, v.field); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errcommon.CollectErrors("database.Migrate", errs)
	}

	return populate(db, populateNum)
}

// ensureBloomFilterIndex creates index with type bloom_filter if it doesn't exist yet.
// According to the research on data from Standoffs, this type of index can reduce latency of queries performing searches in array columns (for example, using has() function).
// Cardinality and distribution of arrays' elements by granulas were taken into account.
// The effectiveness of index can be different depending on data distribution in production.
//
// Granularity was chosen by analyzing the number of unique values within different numbers of granules.
//
// Currently, clickhouse doesn't use data-skipping indexes when executing queries with WHERE ... OR ... conditions on multiple indexed columns.
// See https://github.com/ClickHouse/ClickHouse/issues/8168 for details.
// There's workaround which implies creating composite index on fields which take part in OR query. This is the reason this function allows passing multiple fields.
// The function's signature will probably be broken when the issue is resolved.
//
// NOTE: this function does not build the index for records that already exist in a table. They will be indexed during next merge automatically.
func ensureBloomFilterIndex(db *gorm.DB, table string, fields ...string) error {
	if len(fields) == 0 {
		return errNoIndexFields
	}

	index := indexName(table, fields...)
	expr := indexExpr(fields...)

	return db.Exec(fmt.Sprintf("ALTER TABLE %s ADD INDEX IF NOT EXISTS %s %s TYPE bloom_filter GRANULARITY 3", table, index, expr)).Error
}

// ensureTokenbfv1Index creates index with type tokenbf_v1 if it doesn't exist yet.
//
// Created Bloom filter is 256 bytes in size and applies 2 hash functions to value being indexed (token).
// The https://hur.st/bloomfilter/ calculator was used to pick up parameters with number of tokens (n), probability of false positives (p) and number of hash functions (k) as input values.
//
// Number of tokens (n) was chosen by exploring data distribution within tables filled during Standoffs.
// According to the results, process_pod_name can contain up to 55 unique tokens (n) per granule. Actually, 100 was used as n in order to have some reserve.
// As most of other columns have lower cardinality, calculated parameters can be generalized and used to index those columns as well.
// This value may be changed later after exploring tables with production data.
//
// Number of hash functions (k) is kept low because higher values can affect inserts' bandwidth.
//
// Index has granularity = 1 as it isn't very expensive to index every granule.
//
// Currently, clickhouse doesn't use data-skipping indexes when executing queries with WHERE ... OR ... conditions on multiple indexed columns.
// See https://github.com/ClickHouse/ClickHouse/issues/8168 for details.
// There's workaround which implies creating composite index on fields which take part in OR query. This is the reason this function allows passing multiple fields.
// The function's signature will probably be broken when the issue is resolved.
//
// NOTE: this function does not build the index for records that already exist in a table. They will be indexed during next merge automatically.
func ensureTokenbfv1Index(db *gorm.DB, table string, fields ...string) error {
	if len(fields) == 0 {
		return errNoIndexFields
	}

	index := indexName(table, fields...)
	expr := indexExpr(fields...)

	return db.Exec(fmt.Sprintf("ALTER TABLE %s ADD INDEX IF NOT EXISTS %s %s TYPE tokenbf_v1(256, 2, 0) GRANULARITY 1", table, index, expr)).Error
}

// ensureSetIndex creates index with type set.
// Set has size of 0 which means that its size is not limited.
// Index has granularity of 1.
// Size of set and index's granularity makes it potentially useful for the fields with low cardinality.
// Index can lead to performance decrease when using with fields with high cardinality.
//
// Currently, clickhouse doesn't use data-skipping indexes when executing queries with WHERE ... OR ... conditions on multiple indexed columns.
// See https://github.com/ClickHouse/ClickHouse/issues/8168 for details.
// There's workaround which implies creating composite index on fields which take part in OR query. This is the reason this function allows passing multiple fields.
// The function's signature will probably be broken when the issue is resolved.
//
// NOTE: this function does not build the index for records that already exist in a table. They will be indexed during next merge automatically.
func ensureSetIndex(db *gorm.DB, table string, fields ...string) error {
	if len(fields) == 0 {
		return errNoIndexFields
	}

	index := indexName(table, fields...)
	expr := indexExpr(fields...)

	return db.Exec(fmt.Sprintf("ALTER TABLE %s ADD INDEX IF NOT EXISTS %s %s TYPE set(0) GRANULARITY 1", table, index, expr)).Error
}

func indexName(table string, fields ...string) string {
	return fmt.Sprintf("idx_%s_%s", table, strings.Join(fields, "_"))
}

func indexExpr(fields ...string) string {
	return fmt.Sprintf("(%s)", strings.Join(fields, ", "))
}
