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

func TestHDEL(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []TestCase{
		{
			commands: []string{"HSET key field value", "HDEL key field"},
			expected: []interface{}{ONE, ONE},
		},
		{
			commands: []string{"HSET key field1 value1", "HDEL key field1"},
			expected: []interface{}{ONE, ONE},
		},
		{
			commands: []string{"HSET key field2 value2 field3 value3", "HDEL key field2 field3"},
			expected: []interface{}{TWO, TWO},
		},
		{
			commands: []string{"HSET key_new field value", "HDEL key_new field", "HDEL key_new"},
			expected: []interface{}{ONE, ONE, "ERR wrong number of arguments for 'hdel' command"},
		},
		{
			commands: []string{"SET k v", "HDEL k f"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
	}

	for _, tc := range testCases {
		for i, cmd := range tc.commands {
			result := FireCommand(conn, cmd)
			assert.Equal(t, tc.expected[i], result)
		}
	}
}
