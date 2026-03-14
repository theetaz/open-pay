package observability

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// NewLogger creates a configured zerolog logger.
func NewLogger(service string, level string) zerolog.Logger {
	var lvl zerolog.Level
	switch level {
	case "debug":
		lvl = zerolog.DebugLevel
	case "info":
		lvl = zerolog.InfoLevel
	case "warn":
		lvl = zerolog.WarnLevel
	case "error":
		lvl = zerolog.ErrorLevel
	default:
		lvl = zerolog.InfoLevel
	}

	return zerolog.New(os.Stdout).
		Level(lvl).
		With().
		Timestamp().
		Str("service", service).
		Logger()
}

// NewTestLogger creates a logger that writes to the provided writer (for testing).
func NewTestLogger(w io.Writer) zerolog.Logger {
	return zerolog.New(w).
		Level(zerolog.DebugLevel).
		With().
		Timestamp().
		Logger()
}

func init() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
}
