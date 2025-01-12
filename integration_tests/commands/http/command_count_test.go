// This file is part of DiceDB.
// Copyright (C) 2025DiceDB (dicedb.io).
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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandCount(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name          string
		commands      []HTTPCommand
		expected      []interface{}
		errorExpected bool
		assertType    []string
	}{
		{
			name: "Command count should be greather than zero",
			commands: []HTTPCommand{
				{Command: "COMMAND/COUNT"},
			},
			expected:   []interface{}{float64(0)},
			assertType: []string{"greater"},
		},
		{
			name: "Command count should not support any argument",
			commands: []HTTPCommand{
				{Command: "COMMAND/COUNT", Body: map[string]interface{}{"key": ""}},
			},
			expected:   []interface{}{"ERR wrong number of arguments for 'command|count' command"},
			assertType: []string{"equal"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for c, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				switch tc.assertType[c] {
				case "equal":
					assert.Equal(t, tc.expected[c], result)
				case "greater":
					assert.True(t, result.(float64) >= tc.expected[c].(float64))
				}
			}

		})
	}
}

func getCommandCount(exec *HTTPCommandExecutor) float64 {
	cmd := HTTPCommand{Command: "COMMAND/COUNT", Body: map[string]interface{}{"key": ""}}
	responseValue, _ := exec.FireCommand(cmd)
	if responseValue == nil {
		return -1
	}
	return responseValue.(float64)
}

func BenchmarkCountCommand(b *testing.B) {
	exec := NewHTTPCommandExecutor()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		commandCount := getCommandCount(exec)
		if commandCount <= 0 {
			b.Fail()
		}
	}
}
