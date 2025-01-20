// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

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
