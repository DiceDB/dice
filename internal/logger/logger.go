package logger

import (
	"io"
	"log/slog"
	"os"

	"github.com/dicedb/dice/config"
	"github.com/rs/zerolog"
)

func getLogLevel() slog.Leveler {
	var level slog.Leveler
	switch config.DiceConfig.Server.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	return level
}

type Opts struct {
	WithTimestamp bool
}

func New(opts Opts) *slog.Logger {
	var writer io.Writer = os.Stderr
	if config.DiceConfig.Server.PrettyPrintLogs {
		writer = zerolog.ConsoleWriter{Out: os.Stderr}
	}
	zerologLogger := zerolog.New(writer)
	if opts.WithTimestamp {
		zerologLogger = zerologLogger.With().Timestamp().Logger()
	}
	logger := slog.New(newZerologHandler(zerologLogger))

	return logger
}
