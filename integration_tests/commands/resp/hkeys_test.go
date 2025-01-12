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

package resp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHKeys(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []TestCase{
		{
			name:     "RESP HKEYS with key containing hash with multiple fields",
			commands: []string{"HSET key_hkeys field1 value1", "HSET key_hkeys field2 value2", "HKEYS key_hkeys"},
			expected: []interface{}{int64(1), int64(1), []interface{}{"field1", "field2"}},
		},
		{
			name:     "RESP HKEYS with non-existent key",
			commands: []string{"HKEYS key_hkeys01"},
			expected: []interface{}{[]interface{}{}},
		},
		{
			name:     "RESP HKEYS with key containing a non-hash value",
			commands: []string{"SET key_hkeys02 field1", "HKEYS key_hkeys02"},
			expected: []interface{}{"OK", "ERR -WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name:     "RESP HKEYS with wrong number of arguments",
			commands: []string{"HKEYS key_hkeys03 x", "HKEYS"},
			expected: []interface{}{"ERR wrong number of arguments for 'hkeys' command",
				"ERR wrong number of arguments for 'hkeys' command"},
		},
		{
			name:     "RESP One or more keys exist",
			commands: []string{"HSET key field value", "HSET key field1 value1", "HKEYS key"},
			expected: []interface{}{int64(1), int64(1), []interface{}{"field", "field1"}},
		},
		{
			name:     "RESP No keys exist",
			commands: []string{"HKEYS key"},
			expected: []interface{}{[]interface{}{}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			FireCommand(conn, "HDEL key field")
			FireCommand(conn, "HDEL key field1")
			FireCommand(conn, "DEL key")

			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				switch e := tc.expected[i].(type) {
				case []interface{}:
					assert.ElementsMatch(t, e, tc.expected[i])
				default:
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}

	FireCommand(conn, "HDEL key field")
	FireCommand(conn, "HDEL key field1")
	FireCommand(conn, "DEL key")
}
