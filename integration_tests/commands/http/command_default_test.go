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

package http

import (
	"fmt"
	"testing"

	"github.com/dicedb/dice/internal/eval"
	"github.com/stretchr/testify/assert"
)

func TestCommandDefault(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	commands := getCommandDefault(exec)
	t.Run("Command should not be empty", func(t *testing.T) {
		assert.True(t, len(commands) > 0,
			fmt.Sprintf("Unexpected number of CLI commands found. expected greater than 0, %d found", len(commands)))
	})

	t.Run("Command count matches", func(t *testing.T) {
		assert.True(t, len(commands) == len(eval.DiceCmds),
			fmt.Sprintf("Unexpected number of CLI commands found. expected %d, %d found", len(eval.DiceCmds), len(commands)))
	})
}

func getCommandDefault(exec *HTTPCommandExecutor) []interface{} {
	cmd := HTTPCommand{Command: "COMMAND", Body: map[string]interface{}{"values": []interface{}{}}}
	responseValue, _ := exec.FireCommand(cmd)
	if responseValue == nil {
		return nil
	}
	var cmds []interface{}
	cmds = append(cmds, responseValue.([]interface{})...)
	return cmds
}

func BenchmarkCommandDefault(b *testing.B) {
	exec := NewHTTPCommandExecutor()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		commands := getCommandDefault(exec)
		if len(commands) <= 0 {
			b.Fail()
		}
	}
}
