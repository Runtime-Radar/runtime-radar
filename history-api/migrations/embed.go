package migrations

import "embed"

// Clickhouse migrations directory.
//
//go:embed clickhouse/*.sql
var Clickhouse embed.FS
