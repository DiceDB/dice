package eval_test

import (
	"os"
	"testing"

	dstore "github.com/dicedb/dice/internal/store"
)

func TestMain(m *testing.M) {
	store := dstore.NewStore(nil, nil)
	store.ResetStore()

	exitCode := m.Run()

	os.Exit(exitCode)
}
