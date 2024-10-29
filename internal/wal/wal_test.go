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
