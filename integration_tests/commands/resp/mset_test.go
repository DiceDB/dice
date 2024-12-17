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

package resp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMset(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "MSET with one key-value pair",
			commands: []string{"MSET k1 v1", "GET k1"},
			expected: []interface{}{"OK", "v1"},
		},
		{
			name:     "MSET with multiple key-value pairs",
			commands: []string{"MSET k1 v1 k2 v2", "GET k1", "GET k2"},
			expected: []interface{}{"OK", "v1", "v2"},
		},
		{
			name:     "MSET with odd number of arguments",
			commands: []string{"MSET k1 v1 k2"},
			expected: []interface{}{"ERR wrong number of arguments for 'mset' command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// deleteTestKeys([]string{"k1", "k2"}, store)
			FireCommand(conn, "DEL k1")
			FireCommand(conn, "DEL k1")

			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}
func TestMSETInconsistency(t *testing.T) {

	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "MSET with one key-value pair",
			commands: []string{"MSET k1 v1", "GET k1"},
			expected: []interface{}{"OK", "v1"},
		},
		{
			name:     "MSET with multiple key-value pairs",
			commands: []string{"MSET k1 v1 k2 v2", "GET k1", "GET k2"},
			expected: []interface{}{"OK", "v1", "v2"},
		},
		{
			name:     "MSET with odd number of arguments",
			commands: []string{"MSET k1 v1 k2"},
			expected: []interface{}{"ERR wrong number of arguments for 'mset' command"},
		},
		{
			name:     "MSET with multiple key-value pairs",
			commands: []string{"MSET k1 v1 k2 v2", "GET k1", "GET k2"},
			expected: []interface{}{"OK", "v1", "v2"},
		},
		{
			name:     "MSET with integers arguments",
			commands: []string{"MSET key1 12345 key2 67890", "GET key1", "GET key2"},
			expected: []interface{}{"OK", int64(12345), int64(67890)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// deleteTestKeys([]string{"k1", "k2"}, store)
			FireCommand(conn, "DEL k1")
			FireCommand(conn, "DEL k1")

			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}

}
