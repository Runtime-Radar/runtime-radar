package database

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"github.com/runtime-radar/runtime-radar/lib/logger"
	"github.com/runtime-radar/runtime-radar/notifier/pkg/model"
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
	if newDB {
		if err := db.Migrator().DropTable(
			&model.Notification{},
			&model.Syslog{},
			&model.Email{},
			&model.Webhook{},
		); err != nil {
			return err
		}
	}

	if err := db.Migrator().AutoMigrate(
		&model.Webhook{},
		&model.Email{},
		&model.Syslog{},
		&model.Notification{},
	); err != nil {
		return err
	}

	var errs []error

	return errcommon.CollectErrors("database.Migrate", errs)
}

func uniqueConstraintViolation(err error, table, field string) bool {
	// "idx_table_field" is default index name created by GORM (can be changed via struct tags)
	index := fmt.Sprintf("idx_%s_%s", table, field)

	var pgErr *pgconn.PgError
	// From https://www.postgresql.org/docs/current/errcodes-appendix.html:
	// 23505 => unique_violation
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == index {
		return true
	}

	return false
}
