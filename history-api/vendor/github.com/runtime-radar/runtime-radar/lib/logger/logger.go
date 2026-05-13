package logger

import (
	"io"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logging configuration
const (
	logMaxNum  = 3
	logMaxSize = 10 * 1024 * 1024 // 10MB
)

func Init(file, level string) {
	var out io.Writer = os.Stdout
	if file != "" {
		rl, err := NewRotateFileWriter(file, logMaxNum, logMaxSize)
		if err != nil {
			log.Fatal().Msgf("### Failed to create log file: %v", err)
		}
		out = io.MultiWriter(os.Stdout, rl)
	}

	l := zerolog.New(out).
		With().
		Timestamp().
		Caller().
		Logger()

	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.ErrorFieldName = "err"
	zerolog.CallerMarshalFunc = func(_ uintptr, file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		return file + ":" + strconv.Itoa(line)
	}

	log.Logger = l

	switch level {
	case "DEBUG":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "INFO":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "WARN":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "ERROR":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "FATAL":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}
}
