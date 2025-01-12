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

func TestINCRBYFLOAT(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	invalidArgMessage := "ERR wrong number of arguments for 'incrbyfloat' command"
	invalidIncrTypeMessage := "ERR value is not a valid float"
	valueOutOfRangeMessage := "ERR value is out of range"

	testCases := []struct {
		name      string
		setupData string
		commands  []string
		expected  []interface{}
	}{
		{
			name:      "Invalid number of arguments",
			setupData: "",
			commands:  []string{"INCRBYFLOAT", "INCRBYFLOAT foo"},
			expected:  []interface{}{invalidArgMessage, invalidArgMessage},
		},
		{
			name:      "Increment a non existing key",
			setupData: "",
			commands:  []string{"INCRBYFLOAT foo 0.1", "GET foo"},
			expected:  []interface{}{"0.1", "0.1"},
		},
		{
			name:      "Increment a key with an integer value",
			setupData: "SET foo 1",
			commands:  []string{"INCRBYFLOAT foo 0.1", "GET foo"},
			expected:  []interface{}{"1.1", "1.1"},
		},
		{
			name:      "Increment and then decrement a key with the same value",
			setupData: "SET foo 1",
			commands:  []string{"INCRBYFLOAT foo 0.1", "GET foo", "INCRBYFLOAT foo -0.1", "GET foo"},
			expected:  []interface{}{"1.1", "1.1", "1", "1"},
		},
		{
			name:      "Increment a non numeric value",
			setupData: "SET foo bar",
			commands:  []string{"INCRBYFLOAT foo 0.1"},
			expected:  []interface{}{invalidIncrTypeMessage},
		},
		{
			name:      "Increment by a non numeric value",
			setupData: "SET foo 1",
			commands:  []string{"INCRBYFLOAT foo bar"},
			expected:  []interface{}{invalidIncrTypeMessage},
		},
		{
			name:      "Increment by both integer and float",
			setupData: "SET foo 1",
			commands:  []string{"INCRBYFLOAT foo 1", "INCRBYFLOAT foo 0.1"},
			expected:  []interface{}{"2", "2.1"},
		},
		{
			name:      "Increment that would make the value Inf",
			setupData: "SET foo 1e308",
			commands:  []string{"INCRBYFLOAT foo 1e308", "INCRBYFLOAT foo -1e308"},
			expected:  []interface{}{valueOutOfRangeMessage, "0"},
		},
		{
			name:      "Increment that would make the value -Inf",
			setupData: "SET foo -1e308",
			commands:  []string{"INCRBYFLOAT foo -1e308", "INCRBYFLOAT foo 1e308"},
			expected:  []interface{}{valueOutOfRangeMessage, "0"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer FireCommand(conn, "DEL foo")
			if tc.setupData != "" {
				assert.Equal(t, FireCommand(conn, tc.setupData), "OK")
			}
			for i := 0; i < len(tc.commands); i++ {
				cmd := tc.commands[i]
				out := tc.expected[i]
				result := FireCommand(conn, cmd)
				assert.Equal(t, out, result)
			}
		})
	}
}
