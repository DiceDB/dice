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

	"github.com/dicedb/dice/testutils"
	"github.com/stretchr/testify/assert"
)

func TestKeys(t *testing.T) {
	conn := getLocalConnection()
	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "k matches with k",
			commands: []string{"SET k v", "KEYS k"},
			expected: []interface{}{"OK", []interface{}{"k"}},
		},
		{
			name:     "g* matches good and great",
			commands: []string{"SET good v", "SET great v", "KEYS g*"},
			expected: []interface{}{"OK", "OK", []interface{}{"good", "great"}},
		},
		{
			name:     "g?od matches good",
			commands: []string{"SET good v", "SET great v", "KEYS g?od"},
			expected: []interface{}{"OK", "OK", []interface{}{"good"}},
		},
		{
			name:     "g?eat matches great",
			commands: []string{"SET good v", "SET great v", "KEYS g?eat"},
			expected: []interface{}{"OK", "OK", []interface{}{"great"}},
		},
		{
			name:     "h[^e]llo matches hallo and hbllo",
			commands: []string{"SET hallo v", "SET hbllo v", "SET hello v", "KEYS h[^e]llo"},
			expected: []interface{}{"OK", "OK", "OK", []interface{}{"hallo", "hbllo"}},
		},

		{
			name:     "h[a-b]llo matches hallo and hbllo",
			commands: []string{"SET hallo v", "SET hbllo v", "SET hello v", "KEYS h[a-b]llo"},
			expected: []interface{}{"OK", "OK", "OK", []interface{}{"hallo", "hbllo"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)

				// because the order of keys is not guaranteed, we need to check if the result is an array
				if slice, ok := tc.expected[i].([]interface{}); ok {
					assert.True(t, testutils.UnorderedEqual(slice, result))
				} else {
					assert.Equal(t, tc.expected[i], result)
				}

			}
		})
	}
}
