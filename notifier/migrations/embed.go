package migrations

import "embed"

// Postgres migrations directory.
//
//go:embed postgres/*.sql
var Postgres embed.FS
