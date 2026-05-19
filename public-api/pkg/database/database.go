package database

import (
	"fmt"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"github.com/runtime-radar/runtime-radar/lib/logger"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/model"
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

	db, err := gorm.Open(postgres.Open(url), &gorm.Config{
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

func Migrate(db *gorm.DB, newDB bool) error {
	if newDB {
		if err := db.Migrator().DropTable(
			&model.AccessToken{},
		); err != nil {
			return err
		}
	}

	if err := db.Migrator().AutoMigrate(
		&model.AccessToken{},
	); err != nil {
		return err
	}

	var errs []error

	return errcommon.CollectErrors("database.Migrate", errs)
}
