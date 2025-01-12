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

func TestJSONDEBUG(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	FireCommand(conn, "DEL k1 k2 k3 k4 k5 k6 k7")

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name: "jsondebug with no path",
			commands: []string{
				`JSON.SET k1 $ {"a":1}`,
				"JSON.DEBUG MEMORY k1",
			},
			expected: []interface{}{"OK", int64(72)},
		},
		{
			name: "jsondebug with a valid path",
			commands: []string{
				`JSON.SET k2 $ {"a":1,"b":2}`,
				"JSON.DEBUG MEMORY k2 $.a",
			},
			expected: []interface{}{"OK", []interface{}{int64(16)}},
		},
		{
			name: "jsondebug with multiple paths",
			commands: []string{
				`JSON.SET k3 $ {"a":1,"b":"dice"}`,
				"JSON.DEBUG MEMORY k3 $.a $.b",
			},
			expected: []interface{}{"OK", []interface{}{int64(16)}},
		},
		{
			name: "jsondebug with multiple paths",
			commands: []string{
				`JSON.SET k4 $ {"a":1,"b":"dice"}`,
				"JSON.DEBUG MEMORY k4 $.a $.b",
			},
			expected: []interface{}{"OK", []interface{}{int64(16)}},
		},
		{
			name: "jsondebug with single path for array json",
			commands: []string{
				`JSON.SET k5 $ ["roll","the","dices"]`,
				"JSON.DEBUG MEMORY k5 $[1]",
			},
			expected: []interface{}{"OK", []interface{}{int64(19)}},
		},
		{
			name: "jsondebug with multiple paths for array json",
			commands: []string{
				`JSON.SET k6 $ ["roll","the","dices"]`,
				"JSON.DEBUG MEMORY k6 $[1,2]",
			},
			expected: []interface{}{"OK", []interface{}{int64(19), int64(21)}},
		},
		{
			name: "jsondebug with all paths for array json",
			commands: []string{
				`JSON.SET k7 $ ["roll","the","dices"]`,
				"JSON.DEBUG MEMORY k7 $[:]",
			},
			expected: []interface{}{"OK", []interface{}{int64(20), int64(19), int64(21)}},
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
	FireCommand(conn, "FLUSHDB")
}
