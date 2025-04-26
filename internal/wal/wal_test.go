// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package wal_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/dicedb/dice/internal/wal"
)

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
		// Create WAL entry using protobuf message
		wl.LogCommand([]byte("SET K V"))
	}
}
