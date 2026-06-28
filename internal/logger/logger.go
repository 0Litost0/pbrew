package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// L is the package-level logger instance used by all internal packages.
// It defaults to io.Discard so that tests and library usage are silent
// until Init is explicitly called.
var L zerolog.Logger

func init() {
	L = zerolog.New(io.Discard).With().Timestamp().Logger()
}

// Init initializes the global logger.
//
//   - verbose=true  → Debug level + ConsoleWriter (human-readable, colored) to stderr
//   - verbose=false → Warn level + JSON writer to stderr (AI-parseable, quiet by default)
func Init(verbose bool) {
	var w io.Writer
	level := zerolog.WarnLevel
	if verbose {
		level = zerolog.DebugLevel
		w = zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.TimeOnly,
		}
	} else {
		w = os.Stderr
	}
	L = zerolog.New(w).Level(level).With().Timestamp().Logger()
}
