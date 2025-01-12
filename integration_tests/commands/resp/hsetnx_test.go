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

package resp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHSETNX(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []TestCase{
		{
			commands: []string{"HSETNX key_nx_t1 field value", "HSET key_nx_t1 field value_new"},
			expected: []interface{}{ONE, ZERO},
		},
		{
			commands: []string{"HSETNX key_nx_t2 field1 value1"},
			expected: []interface{}{ONE},
		},
		{
			commands: []string{"HSETNX key_nx_t3 field value", "HSETNX key_nx_t3 field new_value", "HSETNX key_nx_t3"},
			expected: []interface{}{ONE, ZERO, "ERR wrong number of arguments for 'hsetnx' command"},
		},
		{
			commands: []string{"SET key_nx_t4 v", "HSETNX key_nx_t4 f v"},
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
