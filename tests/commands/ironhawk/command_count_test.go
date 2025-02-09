// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"fmt"
	"testing"

	"github.com/dicedb/dicedb-go"
	"github.com/stretchr/testify/assert"
)

func TestCommandCount(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	t.Run("Command count should be positive", func(t *testing.T) {
		commandCount := getCommandCount(client)
		assert.True(t, commandCount > 0,
			fmt.Sprintf("Unexpected number of CLI commands found. expected greater than 0, %d found", commandCount))
	})
}

func getCommandCount(client *dicedb.Client) int64 {
	responseValue := client.FireString("COMMAND COUNT")
	if responseValue == nil {
		return -1
	}
	return responseValue.GetVInt()
}

func BenchmarkCountCommand(b *testing.B) {
	client := getLocalConnection()
	defer client.Close()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		commandCount := getCommandCount(client)
		if commandCount <= 0 {
			b.Fail()
		}
	}
}
