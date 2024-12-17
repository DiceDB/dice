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

func TestType(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "TYPE with invalid number of arguments",
			commands: []HTTPCommand{
				{Command: "TYPE", Body: map[string]interface{}{"keys": []interface{}{}}},
			},
			expected:      []interface{}{"ERR wrong number of arguments for 'type' command"},
			errorExpected: true,
		},
		{
			name: "TYPE for non-existent key",
			commands: []HTTPCommand{
				{Command: "TYPE", Body: map[string]interface{}{"key": "k1"}},
			},
			expected:      []interface{}{"none"},
			errorExpected: false,
		},
		{
			name: "TYPE for key with String value",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "TYPE", Body: map[string]interface{}{"key": "k1"}},
			},
			expected:      []interface{}{"OK", "string"},
			errorExpected: false,
		},
		{
			name: "TYPE for key with List value",
			commands: []HTTPCommand{
				{Command: "LPUSH", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "TYPE", Body: map[string]interface{}{"key": "k1"}},
			},
			expected:      []interface{}{float64(1), "list"},
			errorExpected: false,
		},
		{
			name: "TYPE for key with Set value",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "TYPE", Body: map[string]interface{}{"key": "k1"}},
			},
			expected:      []interface{}{float64(1), "set"},
			errorExpected: false,
		},
		{
			name: "TYPE for key with Hash value",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "k1", "field": "field1", "value": "v1"}},
				{Command: "TYPE", Body: map[string]interface{}{"key": "k1"}},
			},
			expected:      []interface{}{float64(1), "hash"},
			errorExpected: false,
		},
		{
			name: "TYPE for key with value created from SETBIT command",
			commands: []HTTPCommand{
				{Command: "SETBIT", Body: map[string]interface{}{"values": []interface{}{"k1", 1, 1}}},
				{Command: "TYPE", Body: map[string]interface{}{"key": "k1"}},
			},
			expected:      []interface{}{float64(0), "string"},
			errorExpected: false,
		},
		// TODO: uncomment when bitop is added
		// {
		// 	name: "TYPE for key with value created from BITOP command",
		// 	commands: []HTTPCommand{
		// 		{Command: "SET", Body: map[string]interface{}{"key": "key1", "value": "foobar"}},
		// 		{Command: "SET", Body: map[string]interface{}{"key": "key2", "value": "abcdef"}},
		// 		{Command: "TYPE", Body: map[string]interface{}{"key": "dest"}},
		// 	},
		// 	expected:      []interface{}{"OK", "OK", "string"},
		// 	errorExpected: false,
		// },
		{
			name: "TYPE for key with value created from ZADD command",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "k11", "values": [...]string{"1", "member11"}}},
				{Command: "TYPE", Body: map[string]interface{}{"key": "k11"}},
			},
			expected:      []interface{}{float64(1), "zset"},
			errorExpected: false,
		},
		{
			name: "TYPE for key with value created from GEOADD command",
			commands: []HTTPCommand{
				{Command: "GEOADD", Body: map[string]interface{}{"key": "k12", "values": [...]string{"13.361389", "38.115556", "Palermo"}}},
				{Command: "TYPE", Body: map[string]interface{}{"key": "k12"}},
			},
			expected:      []interface{}{float64(1), "zset"},
			errorExpected: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keys := []string{"foo", "k1", "key1", "key2"}

			for _, key := range keys {
				exec.FireCommand(HTTPCommand{
					Command: "DEL",
					Body:    map[string]interface{}{"key": key},
				})
			}

			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}

}
