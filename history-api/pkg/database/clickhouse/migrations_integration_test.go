package clickhouse

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClickhouseMigrationsFromScratch(t *testing.T) {
	t.Parallel()

	require.NotNil(t, clickhouseDB, "clickhouse test database must be initialized in TestMain")

	require.NoError(t, Migrate(clickhouseDB, true, 0), "first clickhouse migration (from scratch) failed")

	err := clickhouseDB.Table("runtime_events").Limit(1).Error
	require.NoError(t, err, "table runtime_events should exist after migrations")

	require.NoError(t, Migrate(clickhouseDB, false, 0), "second clickhouse migration (idempotent up) failed")

	require.NoError(t, Migrate(clickhouseDB, true, 0), "third clickhouse migration (recreate from scratch) failed")
}
