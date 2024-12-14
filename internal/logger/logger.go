// This file is part of DiceDB.
// Copyright (C) 2024 DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/dicedb/dice/config"
	"github.com/rs/zerolog"
)

func getSLogLevel() slog.Level {
	switch config.DiceConfig.Logging.LogLevel {
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
		NoColor:    true,
		TimeFormat: time.RFC3339,
	}).Level(toZerologLevel(getSLogLevel())).With().Timestamp().Logger()
	return slog.New(newZerologHandler(&zerologLogger))
}
