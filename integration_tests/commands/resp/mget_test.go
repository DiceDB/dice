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

	"github.com/dicedb/dice/testutils"
	"github.com/stretchr/testify/assert"
)

func TestMGET(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	defer FireCommand(conn, "DEL k1")
	defer FireCommand(conn, "DEL k2")

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "MGET With non-existing keys",
			commands: []string{"MGET k1 k2"},
			expected: []interface{}{[]interface{}{"(nil)", "(nil)"}},
		},
		{
			name:     "MGET With existing keys",
			commands: []string{"MSET k1 v1 k2 v2", "MGET k1 k2"},
			expected: []interface{}{"OK", []interface{}{"v1", "v2"}},
		},
		{
			name:     "MGET with existing and non existing keys",
			commands: []string{"set k1 v1", "MGET k1 k2"},
			expected: []interface{}{"OK", []interface{}{"v1", "(nil)"}},
		},
		{
			name:     "MGET without any keys",
			commands: []string{"MGET"},
			expected: []interface{}{"ERR wrong number of arguments for 'mget' command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			FireCommand(conn, "DEL k1")
			FireCommand(conn, "DEL k2")

			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				if slice, ok := tc.expected[i].([]interface{}); ok {
					assert.True(t, testutils.UnorderedEqual(slice, result))
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}
}
