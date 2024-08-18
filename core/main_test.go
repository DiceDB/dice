package core_test

import (
	"os"
	"testing"

	"github.com/dicedb/dice/core"
)

func TestMain(m *testing.M) {
	store := core.NewStore()
	store.ResetStore()

	exitCode := m.Run()

	os.Exit(exitCode)
}
