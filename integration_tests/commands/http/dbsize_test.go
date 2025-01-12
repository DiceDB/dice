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

//go:build ignore
// +build ignore

// Ignored as multishard commands not supported by HTTP
package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBSize(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	exec.FireCommand(HTTPCommand{Command: "FLUSHDB"})

	testCases := []TestCase{
		{
			name: "DBSIZE with 3 keys",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k2", "value": "v2"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k3", "value": "v3"}},
				{Command: "DBSIZE", Body: nil},
			},
			expected: []interface{}{"OK", "OK", "OK", float64(3)},
		},
		{
			name: "DBSIZE with repetitive keys",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k2", "value": "v2"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k3", "value": "v3"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v3"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k2", "value": "v22"}},
				{Command: "DBSIZE", Body: nil},
			},
			expected: []interface{}{"OK", "OK", "OK", "OK", "OK", float64(3)},
		},
		{
			name: "DBSIZE with expired keys",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k2", "value": "v2"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k3", "value": "v3"}},
				{Command: "DBSIZE", Body: nil},
				{Command: "EXPIRE", Body: map[string]interface{}{"key": "k3", "seconds": 1}},
				{Command: "SLEEP", Body: map[string]interface{}{"key": 2}},
				{Command: "DBSIZE", Body: nil},
			},
			expected: []interface{}{"OK", "OK", "OK", float64(3), float64(1), "OK", float64(2)},
		},
		{
			name: "DBSIZE after deleting a key",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k2", "value": "v2"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k3", "value": "v3"}},
				{Command: "DBSIZE", Body: nil},
				{Command: "DEL", Body: map[string]interface{}{"key": "k1"}},
				{Command: "DBSIZE", Body: nil},
			},
			expected: []interface{}{"OK", "OK", "OK", float64(3), float64(1), float64(2)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"k1", "k2", "k3"}}})

			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}
