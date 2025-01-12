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

func TestMSET(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	testCases := []TestCase{
		{
			name: "MSET with one key-value pair",
			commands: []HTTPCommand{
				{Command: "MSET", Body: map[string]interface{}{"key_values": map[string]interface{}{"k1": "v1"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "k1"}},
			},
			expected: []interface{}{"OK", "v1"},
		},
		{
			name: "MSET with multiple key-value pairs",
			commands: []HTTPCommand{
				{Command: "MSET", Body: map[string]interface{}{"key_values": map[string]interface{}{"k1": "v1", "k2": "v2"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "k1"}},
				{Command: "GET", Body: map[string]interface{}{"key": "k2"}},
			},
			expected: []interface{}{"OK", "v1", "v2"},
		},
		{
			name: "MSET with integers arguments",
			commands: []HTTPCommand{
				{Command: "MSET", Body: map[string]interface{}{"key_values": map[string]interface{}{"k1": 12345, "k2": 12345}}},
				{Command: "GET", Body: map[string]interface{}{"key": "k1"}},
				{Command: "GET", Body: map[string]interface{}{"key": "k2"}},
			},
			expected: []interface{}{"OK", float64(12345), float64(12345)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"keys": []interface{}{"k1"}},
			})
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)

				assert.Equal(t, tc.expected[i], result)
			}
		})
	}

	// Deleting the used keys
	exec.FireCommand(HTTPCommand{
		Command: "DEL",
		Body: map[string]interface{}{
			"keys": []interface{}{"k1", "k2"},
		},
	})
}
