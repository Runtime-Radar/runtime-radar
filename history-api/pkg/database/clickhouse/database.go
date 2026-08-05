package clickhouse

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"strings"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/history-api/migrations"
	"github.com/runtime-radar/runtime-radar/history-api/pkg/build"
	"github.com/runtime-radar/runtime-radar/lib/logger"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"
)

const (
	dialTimeout = "10s"
	readTimeout = "20s"
)

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
	ctx := context.TODO()

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("can't migrate clickhouse db: %w", err)
	}

	migrationsFS, err := fs.Sub(migrations.Clickhouse, "clickhouse")
	if err != nil {
		return err
	}

	provider, err := goose.NewProvider("clickhouse", sqlDB, migrationsFS, goose.WithTableName(fmt.Sprintf("goose_db_version_%s", strings.ReplaceAll(build.AppName, "-", "_"))))
	if err != nil {
		return err
	}

	if newDB {
		if _, err := provider.DownTo(ctx, 0); err != nil && !errors.Is(err, goose.ErrNoNextVersion) {
			return err
		}
	}

	if _, err := provider.Up(ctx); err != nil {
		return err
	}

	return populate(db, populateNum)
}
