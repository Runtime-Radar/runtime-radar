package docker

import (
	"fmt"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/lib/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gl "gorm.io/gorm/logger"
)

// containerForceRemoveAfter is the duration after which the container will be force removed.
// This is to prevent the container from being left behind in case of any tests panics.
const containerForceRemoveAfter = 10 * time.Minute

type dockerData struct {
	Addr       string
	db         *gorm.DB
	gormLogger gl.Interface
	pool       *dockertest.Pool
	resource   *dockertest.Resource
}

func (d *dockerData) Run() error {
	var err error

	d.pool, err = dockertest.NewPool("")
	if err != nil {
		return fmt.Errorf("can't create docker pool: %w", err)
	}

	// uses pool to try to connect to Docker
	err = d.pool.Client.Ping()
	if err != nil {
		return fmt.Errorf("can't connect to Docker: %w", err)
	}

	d.resource, err = d.pool.Run("postgres", "17", []string{
		"POSTGRES_DB=cs_test",
		"POSTGRES_USER=cs",
		"POSTGRES_PASSWORD=cs",
	})
	if err != nil {
		return fmt.Errorf("can't start resource: %w", err)
	}

	err = d.resource.Expire(uint(containerForceRemoveAfter.Seconds()))
	if err != nil {
		return fmt.Errorf("can't set container force remove duration: %w", err)
	}

	d.Addr = fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=%s",
		"cs",
		"cs",
		d.resource.GetHostPort("5432/tcp"),
		"cs_test",
		"disable",
	)

	d.gormLogger = gl.New(
		&logger.GORM{&log.Logger},
		gl.Config{
			SlowThreshold: 100 * time.Millisecond,
			Colorful:      false,
			LogLevel:      gl.Info,
		},
	)

	if err := d.pool.Retry(func() error {
		var err error
		d.db, err = gorm.Open(postgres.Open(d.Addr), &gorm.Config{
			Logger: d.gormLogger,
		})
		if err != nil {
			return err
		}

		return d.db.Exec("select 1").Error

	}); err != nil {
		return fmt.Errorf("can't connect to database: %w", err)
	}

	return nil
}

func (d *dockerData) DB() *gorm.DB {
	return d.db
}

func (d *dockerData) Cleanup() error {
	sqlDB, err := d.db.DB()
	if err == nil && sqlDB != nil {
		_ = sqlDB.Close()
	}
	d.db = nil

	// When you're done, kill and remove the container
	if d.pool != nil {
		if err := d.pool.Purge(d.resource); err != nil {
			return fmt.Errorf("can't purge resource: %w", err)
		}

		log.Info().Msg("removing postgres container")
		err := d.pool.RemoveContainerByName(d.resource.Container.Name)
		if err != nil {
			return fmt.Errorf("can't remove postgres container: %w", err)
		}
	}

	return nil
}

func SetupDB() (*gorm.DB, func() error, error) {
	instance := &dockerData{}
	err := instance.Run()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	db := instance.DB()

	return db, instance.Cleanup, nil
}
