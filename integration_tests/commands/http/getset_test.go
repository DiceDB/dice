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
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetSet(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		delays   []time.Duration
	}{
		{
			name: "GETSET with INCR",
			commands: []HTTPCommand{
				{Command: "INCR", Body: map[string]interface{}{"key": "mycounter"}},
				{Command: "GETSET", Body: map[string]interface{}{"key": "mycounter", "value": "0"}},
				{Command: "GET", Body: map[string]interface{}{"key": "mycounter"}},
			},
			expected: []interface{}{float64(1), float64(1), float64(0)},
			delays:   []time.Duration{0, 0, 0},
		},
		{
			name: "GETSET with SET",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "mykey", "value": "Hello"}},
				{Command: "GETSET", Body: map[string]interface{}{"key": "mykey", "value": "world"}},
				{Command: "GET", Body: map[string]interface{}{"key": "mykey"}},
			},
			expected: []interface{}{"OK", "Hello", "world"},
			delays:   []time.Duration{0, 0, 0},
		},
		{
			name: "GETSET with TTL",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v", "ex": 60}},
				{Command: "GETSET", Body: map[string]interface{}{"key": "k", "value": "v1"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "v", float64(-1)},
			delays:   []time.Duration{0, 0, 0},
		},
		{
			name: "GETSET error when key exists but does not hold a string value",
			commands: []HTTPCommand{
				{Command: "LPUSH", Body: map[string]interface{}{"key": "k1", "value": "val"}},
				{Command: "GETSET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
			},
			expected: []interface{}{float64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			delays:   []time.Duration{0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}
