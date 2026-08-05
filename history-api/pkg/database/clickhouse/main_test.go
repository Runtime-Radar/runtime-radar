package clickhouse

import (
	"os"
	"runtime/debug"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/history-api/pkg/config"
	"gorm.io/gorm"
)

var clickhouseDB *gorm.DB

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

	var closeCH func() error
	var err error

	clickhouseDB, closeCH, err = New(
		cfg.ClickhouseAddr,
		cfg.ClickhouseDB+"_test", // <-- use test DB
		cfg.ClickhouseUser,
		cfg.ClickhousePassword,
		cfg.ClickhouseSSLMode,
		cfg.ClickhouseSSLCheckCert,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("setup clickhouse database failed")
	}
	defer func() { _ = closeCH() }()

	err = Migrate(clickhouseDB, cfg.NewDB, cfg.PopulateNum)
	if err != nil {
		log.Fatal().Err(err).Msg("migrate clickhouse database failed")
	}

	return m.Run()
}
