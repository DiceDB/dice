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

package http

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestDel(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
	}{
		{
			name: "DEL with set key",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "DEL", Body: map[string]interface{}{"key": "k1"}},
				{Command: "GET", Body: map[string]interface{}{"key": "k1"}},
			},
			expected: []interface{}{"OK", float64(1), nil},
		},
		{
			name: "DEL with multiple keys",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k2", "value": "v2"}},
				{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"k1", "k2"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "k1"}},
				{Command: "GET", Body: map[string]interface{}{"key": "k2"}},
			},
			expected: []interface{}{"OK", "OK", float64(2), nil, nil},
		},
		{
			name: "DEL with key not set",
			commands: []HTTPCommand{
				{Command: "GET", Body: map[string]interface{}{"key": "k3"}},
				{Command: "DEL", Body: map[string]interface{}{"key": "k3"}},
			},
			expected: []interface{}{nil, float64(0)},
		},
		{
			name: "DEL with no keys or arguments",
			commands: []HTTPCommand{
				{Command: "DEL", Body: map[string]interface{}{"key": ""}},
			},
			expected: []interface{}{float64(0)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, err := exec.FireCommand(cmd)
				if err != nil {
					// Check if the error message matches the expected result
					assert.Equal(t, tc.expected[i], err.Error(), "Error message mismatch for cmd %s", cmd)
				} else {
					assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s, expected %v, got %v", cmd, tc.expected[i], result)
				}
			}
		})
	}
}
