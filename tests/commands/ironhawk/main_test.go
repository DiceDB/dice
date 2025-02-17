// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"os"
	"strings"
	"testing"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dicedb-go/wire"
)

func TestMain(m *testing.M) {
	config.ForceInit(&config.DiceDBConfig{})
	os.Exit(m.Run())
}

type TestCase struct {
	name     string
	commands []string
	expected []interface{}
	keysUsed []string
}

func assertEqual(t *testing.T, expected interface{}, actual *wire.Response) bool {
	var areEqual bool
	switch v := expected.(type) {
	case string:
		areEqual = v == actual.GetVStr()
		if strings.HasPrefix(v, "ERR") {
			areEqual = v == actual.GetErr()
		}
	case int64:
		areEqual = v == actual.GetVInt()
	case int:
		areEqual = int64(v) == actual.GetVInt()
	case nil:
		areEqual = actual.GetVNil()
	}
	if !areEqual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
	return areEqual
}
