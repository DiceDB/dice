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
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHExists(t *testing.T) {
	cmdExec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		delays   []time.Duration
	}{
		{
			name: "HTTP Check if field exists when k f and v are set",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "k", "field": "f", "value": "v"}},
				{Command: "HEXISTS", Body: map[string]interface{}{"key": "k", "field": "f"}},
			},
			expected: []interface{}{float64(1), float64(1)},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "HTTP Check if field exists when k exists but not f and v",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "k", "field": "f1", "value": "v"}},
				{Command: "HEXISTS", Body: map[string]interface{}{"key": "k", "field": "f"}},
			},
			expected: []interface{}{float64(1), float64(0)},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "HTTP Check if field exists when no k,f and v exist",
			commands: []HTTPCommand{
				{Command: "HEXISTS", Body: map[string]interface{}{"key": "k", "field": "f"}},
			},
			expected: []interface{}{float64(0)},
			delays:   []time.Duration{0},
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
			cmdExec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"key": "k"},
			})

			for i, cmd := range tc.commands {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}

				result, err := cmdExec.FireCommand(cmd)
				if err != nil {
					// Check if the error message matches the expected result
					assert.Equal(t, tc.expected[i], err.Error(), "Error message mismatch for cmd %s", cmd)
				} else {
					assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s, expected %v, got %v", cmd, tc.expected[i], result)
				}
			}
		})
	}

	// Deleting the used keys
	cmdExec.FireCommand(HTTPCommand{
		Command: "HDEL",
		Body:    map[string]interface{}{"key": "k", "field": "f"},
	})
	cmdExec.FireCommand(HTTPCommand{
		Command: "HDEL",
		Body:    map[string]interface{}{"key": "k", "field": "f1"},
	})
	cmdExec.FireCommand(HTTPCommand{
		Command: "DEL",
		Body:    map[string]interface{}{"key": "k"},
	})
}
