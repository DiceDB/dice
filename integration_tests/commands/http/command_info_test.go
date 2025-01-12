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

func TestCommandInfo(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "Set command",
			commands: []HTTPCommand{
				{Command: "COMMAND/INFO", Body: map[string]interface{}{"key": "SET"}},
			},
			expected: []interface{}{[]interface{}{[]interface{}{"set", float64(-3), float64(1), float64(0), float64(0), interface{}(nil)}}},
		},
		{
			name: "Get command",
			commands: []HTTPCommand{
				{Command: "COMMAND/INFO", Body: map[string]interface{}{"key": "GET"}},
			},
			expected: []interface{}{[]interface{}{[]interface{}{"get", float64(2), float64(1), float64(0), float64(0), interface{}(nil)}}},
		},
		{
			name: "PING command",
			commands: []HTTPCommand{
				{Command: "COMMAND/INFO", Body: map[string]interface{}{"key": "PING"}},
			},
			expected: []interface{}{[]interface{}{[]interface{}{"ping", float64(-1), float64(0), float64(0), float64(0), interface{}(nil)}}},
		},
		{
			name: "Combination of multiple valid commands",
			commands: []HTTPCommand{
				{Command: "COMMAND/INFO", Body: map[string]interface{}{"keys": []interface{}{"SET", "GET"}}},
			},
			expected: []interface{}{[]interface{}{
				[]interface{}{"set", float64(-3), float64(1), float64(0), float64(0), interface{}(nil)},
				[]interface{}{"get", float64(2), float64(1), float64(0), float64(0), interface{}(nil)},
			}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}
