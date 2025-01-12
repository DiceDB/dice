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

package websocket

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONDEBUG(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	defer conn.Close()

	DeleteKey(t, conn, exec, "k1")
	DeleteKey(t, conn, exec, "k2")
	DeleteKey(t, conn, exec, "k3")
	DeleteKey(t, conn, exec, "k4")
	DeleteKey(t, conn, exec, "k5")
	DeleteKey(t, conn, exec, "k6")
	DeleteKey(t, conn, exec, "k7")

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
			expected: []interface{}{"OK", float64(72)},
		},
		{
			name: "jsondebug with a valid path",
			commands: []string{
				`JSON.SET k2 $ {"a":1,"b":2}`,
				"JSON.DEBUG MEMORY k2 $.a",
			},
			expected: []interface{}{"OK", []interface{}{float64(16)}},
		},
		{
			name: "jsondebug with multiple paths",
			commands: []string{
				`JSON.SET k3 $ {"a":1,"b":"dice"}`,
				"JSON.DEBUG MEMORY k3 $.a $.b",
			},
			expected: []interface{}{"OK", []interface{}{float64(16)}},
		},
		{
			name: "jsondebug with multiple paths",
			commands: []string{
				`JSON.SET k4 $ {"a":1,"b":"dice"}`,
				"JSON.DEBUG MEMORY k4 $.a $.b",
			},
			expected: []interface{}{"OK", []interface{}{float64(16)}},
		},
		{
			name: "jsondebug with single path for array json",
			commands: []string{
				`JSON.SET k5 $ ["roll","the","dices"]`,
				"JSON.DEBUG MEMORY k5 $[1]",
			},
			expected: []interface{}{"OK", []interface{}{float64(19)}},
		},
		{
			name: "jsondebug with multiple paths for array json",
			commands: []string{
				`JSON.SET k6 $ ["roll","the","dices"]`,
				"JSON.DEBUG MEMORY k6 $[1,2]",
			},
			expected: []interface{}{"OK", []interface{}{float64(19), float64(21)}},
		},
		{
			name: "jsondebug with all paths for array json",
			commands: []string{
				`JSON.SET k7 $ ["roll","the","dices"]`,
				"JSON.DEBUG MEMORY k7 $[:]",
			},
			expected: []interface{}{"OK", []interface{}{float64(20), float64(19), float64(21)}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
	exec.FireCommandAndReadResponse(conn, "FLUSHDB")
}
