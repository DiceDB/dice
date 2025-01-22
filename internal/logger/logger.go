// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/dicedb/dice/config"
	"github.com/rs/zerolog"
)

func getSLogLevel() slog.Level {
	switch config.Config.LogLevel {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	default:
		return slog.LevelInfo
	}
}

func New() *slog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerologLogger := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		NoColor:    false,
		TimeFormat: time.RFC3339,
	}).Level(toZerologLevel(getSLogLevel())).With().Timestamp().Logger()
	return slog.New(newZerologHandler(&zerologLogger))
}
