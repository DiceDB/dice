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
		wl.LogCommand([]byte("SET key value"))
	}
}
