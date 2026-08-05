package database

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
	"github.com/runtime-radar/runtime-radar/event-processor/migrations"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/build"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/model"
	"github.com/runtime-radar/runtime-radar/lib/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"
)

const (
	// Postgres CA cert file name.
	postgresCAFile = "db_ca.pem"
)

func New(address, database, user, password string, sslMode, sslCheckCert bool) (*gorm.DB, func() error, error) {
	mode := "disable"
	if sslMode {
		mode = "require"
		if sslCheckCert {
			mode = "verify-full"
		}
	}

	url := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s", user, url.QueryEscape(password), address, database, mode)

	if sslMode && sslCheckCert {
		url += fmt.Sprintf("&sslrootcert=%s", postgresCAFile)
	}

	ll := gorm_logger.Silent
	if e := log.Debug(); e.Enabled() {
		ll = gorm_logger.Info
	}
	gormLogger := gorm_logger.New(
		&logger.GORM{&log.Logger},
		gorm_logger.Config{
			SlowThreshold: 100 * time.Millisecond, // Slow SQL threshold
			Colorful:      false,                  // Disable color
			LogLevel:      ll,                     // Log level
		},
	)

	db, err := gorm.Open(postgres.Open(url), &gorm.Config{
		Logger: gormLogger,
		// SkipDefaultTransaction: true, // disable DB transactions
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

func Migrate(db *gorm.DB, newDB bool) error {
	ctx := context.TODO()

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("can't migrate postgresql db: %w", err)
	}

	migrationsFS, err := fs.Sub(migrations.Postgres, "postgres")
	if err != nil {
		return err
	}

	// Services share a single database, so each of them keeps its own migration history.
	provider, err := goose.NewProvider("postgres", sqlDB, migrationsFS, goose.WithTableName(fmt.Sprintf("goose_db_version_%s", strings.ReplaceAll(build.AppName, "-", "_"))))
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

	return ensureDefaultConfig(db)
}

func ensureDefaultConfig(db *gorm.DB) error {
	ctx := context.Background()
	config := model.DefaultConfig
	configDB := &ConfigDatabase{db}

	if _, err := configDB.GetLast(ctx, false); errors.Is(err, gorm.ErrRecordNotFound) {
		log.Debug().Interface("config", config).Msg("Creating default config")

		if err := configDB.Add(ctx, config); err != nil {
			return fmt.Errorf("can't create default config '%+v': %w", config, err)
		}
	} else if err != nil {
		return fmt.Errorf("can't get last config: %w", err)
	}

	return nil
}
