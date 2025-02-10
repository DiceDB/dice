// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"fmt"
	"testing"

	"github.com/dicedb/dicedb-go"
	"github.com/stretchr/testify/assert"
)

func TestCommandList(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	t.Run("Command list should not be empty", func(t *testing.T) {
		commandList := getCommandList(client)
		assert.True(t, len(commandList) > 0,
			fmt.Sprintf("Unexpected number of CLI commands found. expected greater than 0, %d found", len(commandList)))
	})
}

func getCommandList(client *dicedb.Client) []string {
	resp := client.FireString("COMMAND LIST")
	if resp == nil {
		return nil
	}

	var cmds []string
	for _, v := range resp.GetVStr() {
		cmds = append(cmds, string(v))
	}
	return cmds
}

func BenchmarkCommandList(b *testing.B) {
	client := getLocalConnection()
	defer client.Close()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		commandList := getCommandList(client)
		if len(commandList) <= 0 {
			b.Fail()
		}
	}
}
