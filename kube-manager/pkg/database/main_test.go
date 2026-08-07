package database

import (
	"os"
	"runtime/debug"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/config"
	"github.com/runtime-radar/runtime-radar/kube-manager/tests/utils/docker"
	"gorm.io/gorm"
)

var db *gorm.DB

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) (res int) {
	defer func() {
		err := recover()
		if err != nil {
			log.Error().Msgf("%v", err)
			log.Error().Msgf("%v", debug.Stack())
			res = 1
		}
	}()

	cfg := config.New()

	var cleaner func() error
	var err error

	if cfg.TestInCompose {
		db, cleaner, err = New(
			cfg.PostgresAddr,
			cfg.PostgresDB+"_test", // <-- use test DB
			cfg.PostgresUser,
			cfg.PostgresPassword,
			cfg.PostgresSSLMode,
			cfg.PostgresSSLCheckCert,
		)
	} else {
		cfg.ListenGRPCAddr = ":48000"
		db, cleaner, err = docker.SetupDB()
	}
	if err != nil {
		log.Fatal().Err(err).Msg("setup database failed")
	}
	defer func() { _ = cleaner() }()

	err = Migrate(db, cfg.NewDB)
	if err != nil {
		log.Fatal().Err(err).Msg("migrate database failed")
	}

	return m.Run()
}
