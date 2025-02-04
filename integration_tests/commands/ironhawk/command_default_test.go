// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"fmt"
	"testing"

	"github.com/dicedb/dice/internal/eval"
	"github.com/dicedb/dicedb-go"
	"github.com/stretchr/testify/assert"
)

func TestCommandDefault(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	commands := getCommandDefault(client)
	t.Run("Command should not be empty", func(t *testing.T) {
		assert.True(t, len(commands) > 0,
			fmt.Sprintf("Unexpected number of CLI commands found. expected greater than 0, %d found", len(commands)))
	})

	t.Run("Command count matches", func(t *testing.T) {
		assert.True(t, len(commands) == len(eval.DiceCmds),
			fmt.Sprintf("Unexpected number of CLI commands found. expected %d, %d found", len(eval.DiceCmds), len(commands)))
	})
}

func getCommandDefault(client *dicedb.Client) []interface{} {
	resp := client.FireString("COMMAND")
	if resp == nil {
		return nil
	}
	var cmds []interface{}
	cmds = append(cmds, resp.GetVStr())
	return cmds
}

func BenchmarkCommandDefault(b *testing.B) {
	client := getLocalConnection()
	defer client.Close()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		commands := getCommandDefault(client)
		if len(commands) <= 0 {
			b.Fail()
		}
	}
}
