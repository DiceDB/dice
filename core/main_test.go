package core_test

import (
	"os"
	"testing"

	"github.com/dicedb/dice/core"
)

func TestMain(m *testing.M) {
	core.ResetStore()

	exitCode := m.Run()

	os.Exit(exitCode)
}
