package eval_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/dicedb/dice/internal/logger"
	dstore "github.com/dicedb/dice/internal/store"
)

func TestMain(m *testing.M) {
	logger := logger.New(logger.Opts{WithTimestamp: false})
	slog.SetDefault(logger)

	store := dstore.NewStore(nil)
	store.ResetStore()

	exitCode := m.Run()

	os.Exit(exitCode)
}
