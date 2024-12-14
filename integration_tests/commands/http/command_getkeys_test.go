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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandGetKeys(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "Set command",
			commands: []HTTPCommand{
				{Command: "COMMAND/GETKEYS", Body: map[string]interface{}{"key": "SET", "keys": []interface{}{"1", "2"}, "values": []interface{}{"2", "3"}}},
			},
			expected: []interface{}{[]interface{}{"1"}},
		},
		{
			name: "Get command",
			commands: []HTTPCommand{
				{Command: "COMMAND/GETKEYS", Body: map[string]interface{}{"key": "GET", "field": "key"}},
			},
			expected: []interface{}{[]interface{}{"key"}},
		},
		{
			name: "TTL command",
			commands: []HTTPCommand{
				{Command: "COMMAND/GETKEYS", Body: map[string]interface{}{"key": "TTL", "field": "key"}},
			},
			expected: []interface{}{[]interface{}{"key"}},
		},
		{
			name: "Del command",
			commands: []HTTPCommand{
				{Command: "COMMAND/GETKEYS", Body: map[string]interface{}{"key": "DEL", "field": "1 2 3 4 5 6 7"}},
			},
			expected: []interface{}{[]interface{}{"1 2 3 4 5 6 7"}},
		},
		// Skipping these tests until multishards cmds supported by http
		//{
		//	name: "MSET command",
		//	commands: []HTTPCommand{
		//		{Command: "COMMAND/GETKEYS", Body: map[string]interface{}{"key": "MSET", "keys": []interface{}{"key1 key2"}, "values": []interface{}{" val1 val2"}}},
		//	},
		//	expected: []interface{}{"ERR invalid command specified"},
		//},
		{
			name: "Expire command",
			commands: []HTTPCommand{
				{Command: "COMMAND/GETKEYS", Body: map[string]interface{}{"key": "EXPIRE", "field": "key", "values": []interface{}{"time", "extra"}}},
			},
			expected: []interface{}{[]interface{}{"key"}},
		},
		{
			name: "PING command",
			commands: []HTTPCommand{
				{Command: "COMMAND/GETKEYS", Body: map[string]interface{}{"key": "PING"}},
			},
			expected: []interface{}{"ERR the command has no key arguments"},
		},
		{
			name: "Invalid Get command",
			commands: []HTTPCommand{
				{Command: "COMMAND/GETKEYS", Body: map[string]interface{}{"key": "GET"}},
			},
			expected: []interface{}{"ERR invalid number of arguments specified for command"},
		},
		{
			name: "Abort command",
			commands: []HTTPCommand{
				{Command: "COMMAND/GETKEYS", Body: map[string]interface{}{"key": "ABORT"}},
			},
			expected: []interface{}{"ERR the command has no key arguments"},
		},
		{
			name: "Invalid command",
			commands: []HTTPCommand{
				{Command: "COMMAND/GETKEYS", Body: map[string]interface{}{"key": "NotValidCommand"}},
			},
			expected: []interface{}{"ERR invalid command specified"},
		},
		{
			name: "Wrong number of arguments",
			commands: []HTTPCommand{
				{Command: "COMMAND/GETKEYS", Body: map[string]interface{}{"key": ""}},
			},
			expected: []interface{}{"ERR invalid command specified"},
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
