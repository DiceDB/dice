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

func TestHKeys(t *testing.T) {
	cmdExec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "HTTP One or more keys exist",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "k", "field": "f", "value": "v"}},
				{Command: "HSET", Body: map[string]interface{}{"key": "k", "field": "f1", "value": "v1"}},
				{Command: "HKEYS", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{float64(1), float64(1), []interface{}{"f", "f1"}},
		},
		{
			name: "HTTP No keys exist",
			commands: []HTTPCommand{
				{Command: "HKEYS", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{nil},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdExec.FireCommand(HTTPCommand{
				Command: "HDEL",
				Body:    map[string]interface{}{"key": "k", "field": "f"},
			})
			cmdExec.FireCommand(HTTPCommand{
				Command: "HDEL",
				Body:    map[string]interface{}{"key": "k", "field": "f1"},
			})

			for i, cmd := range tc.commands {
				result, _ := cmdExec.FireCommand(cmd)
				switch e := tc.expected[i].(type) {
				case []interface{}:
					assert.ElementsMatch(t, e, tc.expected[i])
				default:
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}

	cmdExec.FireCommand(HTTPCommand{
		Command: "HDEL",
		Body:    map[string]interface{}{"key": "k", "field": "f"},
	})
	cmdExec.FireCommand(HTTPCommand{
		Command: "HDEL",
		Body:    map[string]interface{}{"key": "k", "field": "f1"},
	})
}
