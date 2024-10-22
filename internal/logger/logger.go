package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/dicedb/dice/config"
	"github.com/rs/zerolog"
)

func getLogLevel() slog.Leveler {
	var level slog.Leveler
	switch config.DiceConfig.Logging.LogLevel {
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

func New() *slog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerologLogger := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		NoColor:    true,
		TimeFormat: time.RFC3339,
	}).Level(mapLevel(getLogLevel().Level())).With().Timestamp().Logger()
	return slog.New(newZerologHandler(&zerologLogger))
}
