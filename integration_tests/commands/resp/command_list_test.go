// This file is part of DiceDB.
// Copyright (C) 2024 DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package resp

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandList(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	t.Run("Command list should not be empty", func(t *testing.T) {
		commandList := getCommandList(conn)
		assert.True(t, len(commandList) > 0,
			fmt.Sprintf("Unexpected number of CLI commands found. expected greater than 0, %d found", len(commandList)))
	})
}

func getCommandList(connection net.Conn) []string {
	responseValue := FireCommand(connection, "COMMAND LIST")
	if responseValue == nil {
		return nil
	}

	var cmds []string
	for _, v := range responseValue.([]interface{}) {
		cmds = append(cmds, v.(string))
	}
	return cmds
}

func BenchmarkCommandList(b *testing.B) {
	conn := getLocalConnection()
	defer conn.Close()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		commandList := getCommandList(conn)
		if len(commandList) <= 0 {
			b.Fail()
		}
	}
}
