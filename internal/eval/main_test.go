package eval_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/dicedb/dice/internal/logger"
	dstore "github.com/dicedb/dice/internal/store"
)

func TestMain(m *testing.M) {
	l := logger.New(logger.Opts{WithTimestamp: false})
	slog.SetDefault(l)

	store := dstore.NewStore(nil, nil)
	store.ResetStore()

	exitCode := m.Run()

	os.Exit(exitCode)
}
