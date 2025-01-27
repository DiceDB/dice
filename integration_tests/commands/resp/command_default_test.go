// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package resp

import (
	"fmt"
	"net"
	"testing"

	"github.com/dicedb/dice/internal/eval"
	"github.com/stretchr/testify/assert"
)

func TestCommandDefault(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	commands := getCommandDefault(conn)
	t.Run("Command should not be empty", func(t *testing.T) {
		assert.True(t, len(commands) > 0,
			fmt.Sprintf("Unexpected number of CLI commands found. expected greater than 0, %d found", len(commands)))
	})

	t.Run("Command count matches", func(t *testing.T) {
		assert.True(t, len(commands) == len(eval.DiceCmds),
			fmt.Sprintf("Unexpected number of CLI commands found. expected %d, %d found", len(eval.DiceCmds), len(commands)))
	})
}

func getCommandDefault(connection net.Conn) []interface{} {
	responseValue := FireCommand(connection, "COMMAND")
	if responseValue == nil {
		return nil
	}
	var cmds []interface{}
	cmds = append(cmds, responseValue.([]interface{})...)
	return cmds
}

func BenchmarkCommandDefault(b *testing.B) {
	conn := getLocalConnection()
	defer conn.Close()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		commands := getCommandDefault(conn)
		if len(commands) <= 0 {
			b.Fail()
		}
	}
}
