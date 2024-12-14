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

package wal_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/wal"
)

func BenchmarkLogCommandSQLite(b *testing.B) {
	wl, err := wal.NewSQLiteWAL("/tmp/dicedb-lt")
	if err != nil {
		panic(err)
	}

	if err := wl.Init(time.Now()); err != nil {
		slog.Error("could not initialize WAL", slog.Any("error", err))
	} else {
		go wal.InitBG(wl)
	}

	for i := 0; i < b.N; i++ {
		wl.LogCommand(&cmd.DiceDBCmd{
			Cmd:  "SET",
			Args: []string{"key", "value"},
		})
	}
}

func BenchmarkLogCommandAOF(b *testing.B) {
	wl, err := wal.NewAOFWAL("/tmp/dicedb-lt")
	if err != nil {
		panic(err)
	}

	if err := wl.Init(time.Now()); err != nil {
		slog.Error("could not initialize WAL", slog.Any("error", err))
	} else {
		go wal.InitBG(wl)
	}

	for i := 0; i < b.N; i++ {
		wl.LogCommand(&cmd.DiceDBCmd{
			Cmd:  "SET",
			Args: []string{"key", "value"},
		})
	}
}
