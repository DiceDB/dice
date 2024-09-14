package eval_test

import (
	"log/slog"
	"os"
	"testing"

	dstore "github.com/dicedb/dice/internal/store"
	"github.com/rs/zerolog"
	slogzerolog "github.com/samber/slog-zerolog/v2"
)

func TestMain(m *testing.M) {
	zerologLogger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	logger := slog.New(slogzerolog.Option{Logger: &zerologLogger}.NewZerologHandler())
	slog.SetDefault(logger)

	store := dstore.NewStore(nil)
	store.ResetStore()

	exitCode := m.Run()

	os.Exit(exitCode)
}
