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

func TestZRANK(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	FireCommand(conn, "DEL key")
	defer FireCommand(conn, "DEL key")

	// Initialize the sorted set with members and their scores
	FireCommand(conn, "ZADD key 1 member1 2 member2 3 member3 4 member4 5 member5")

	testCases := []TestCase{
		{
			name:     "ZRANK of existing member",
			commands: []string{"ZRANK key member1"},
			expected: []interface{}{int64(0)},
		},
		{
			name:     "ZRANK of non-existing member",
			commands: []string{"ZRANK key member6"},
			expected: []interface{}{"(nil)"},
		},
		{
			name:     "ZRANK with WITHSCORE option for existing member",
			commands: []string{"ZRANK key member3 WITHSCORE"},
			expected: []interface{}{[]interface{}{int64(2), "3"}},
		},
		{
			name:     "ZRANK with WITHSCORE option for non-existing member",
			commands: []string{"ZRANK key member6 WITHSCORE"},
			expected: []interface{}{"(nil)"},
		},
		{
			name:     "ZRANK on non-existing key",
			commands: []string{"ZRANK nonexisting member1"},
			expected: []interface{}{"(nil)"},
		},
		{
			name:     "ZRANK with wrong number of arguments",
			commands: []string{"ZRANK key"},
			expected: []interface{}{"ERR wrong number of arguments for 'zrank' command"},
		},
		{
			name:     "ZRANK with invalid option",
			commands: []string{"ZRANK key member1 INVALID_OPTION"},
			expected: []interface{}{"ERR syntax error"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}
