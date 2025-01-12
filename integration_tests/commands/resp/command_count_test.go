// This file is part of DiceDB.
// Copyright (C) 2025  DiceDB (dicedb.io).
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

func TestCommandCount(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	t.Run("Command count should be positive", func(t *testing.T) {
		commandCount := getCommandCount(conn)
		assert.True(t, commandCount > 0,
			fmt.Sprintf("Unexpected number of CLI commands found. expected greater than 0, %d found", commandCount))
	})
}

func getCommandCount(connection net.Conn) int64 {
	responseValue := FireCommand(connection, "COMMAND COUNT")
	if responseValue == nil {
		return -1
	}
	return responseValue.(int64)
}

func BenchmarkCountCommand(b *testing.B) {
	conn := getLocalConnection()
	defer conn.Close()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		commandCount := getCommandCount(conn)
		if commandCount <= 0 {
			b.Fail()
		}
	}
}
