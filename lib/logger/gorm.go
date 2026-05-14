package logger

import (
	"strings"

	"github.com/rs/zerolog"
)

type GORM struct {
	*zerolog.Logger
}

// Printf wraps Printf calls, parses incoming data and changes it for better logs readability. Unfortunately there is no easy way of customizing
// GORM logger because another option of re-implementing `gorm.io/gorm/logger.Interface` is not very convenient due to its awkward architecture.
func (g *GORM) Printf(format string, toPrint ...any) {
	// First item is always "*.go" source file with corresponding format prefix
	toPrint = toPrint[1:]

	// Prefixes are taken from `gorm.io/gorm/logger`
	r := strings.NewReplacer(
		"%s\n[info] ", "<gorm info> %s",
		"%s\n[warn] ", "<gorm warn> %s",
		"%s\n[error] ", "<gorm error> %s",
		"%s\n[%.3fms] [rows:%v] %s", "SQL [%.3fms] [rows:%v] %s",
		"%s %s\n[%.3fms] [rows:%v] %s", "<%s> SQL [%.3fms] [rows:%v] %s",
	)
	format = r.Replace(format)

	g.Logger.Printf(format, toPrint...)
}
