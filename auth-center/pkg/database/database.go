package database

import (
	"fmt"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/auth-center/pkg/model"
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

func Migrate(
	db *gorm.DB,
	newDB bool,
	adminUsername string,
	adminEmail string,
	adminHashedPassword string,
) error {
	if newDB {
		if err := db.Migrator().DropTable(
			&model.Role{},
			&model.User{},
		); err != nil {
			return err
		}
	}

	if err := db.Migrator().AutoMigrate(&model.Role{}, &model.User{}); err != nil {
		return err
	}

	if err := loadPredeclaredRoles(db); err != nil {
		return err
	}

	if err := createAdmin(db, adminUsername, adminEmail, adminHashedPassword); err != nil {
		return err
	}
	return nil
}

func createAdmin(
	db *gorm.DB,
	username string,
	email string,
	hashedPassword string,
) error {
	return db.Save(&model.User{
		Base:                  model.Base{ID: model.AdminInitID},
		Username:              username,
		Email:                 email,
		AuthType:              model.AuthTypeInternal,
		HashedPassword:        hashedPassword,
		RoleID:                model.AdminRoleID,
		LastPasswordChangedAt: time.Now(),
	}).Error
}

func loadPredeclaredRoles(db *gorm.DB) error {
	tx := db.Begin()

	for _, role := range model.PredeclaredRoles {
		if err := tx.Save(&model.Role{
			ID:              role.ID,
			RoleName:        role.RoleName,
			RolePermissions: role.RolePermissions,
			Description:     role.Description,
		}).Error; err != nil {
			return err
		}
	}

	return tx.Commit().Error
}
