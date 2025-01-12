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

func TestJSONRESP(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	FireCommand(conn, "DEL k1 k2 k3")

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name: "jsonresp on array with mixed types",
			commands: []string{
				`JSON.SET k1 $ ["dice",10,10.5,true,null]`,
				"JSON.RESP k1 $",
			},
			expected: []interface{}{"OK", []interface{}([]interface{}{"[", "dice", int64(10), "10.5", "true", "(nil)"})},
		},
		{
			name: "jsonresp on nested array with mixed types",
			commands: []string{
				`JSON.SET k2 $ {"b":["dice",10,10.5,true,null]}`,
				"JSON.RESP k2 $.b",
			},
			expected: []interface{}{"OK", []interface{}([]interface{}{[]interface{}{"[", "dice", int64(10), "10.5", "true", "(nil)"}})},
		},
		{
			name: "jsonresp on object at root path",
			commands: []string{
				`JSON.SET k3 $ {"b":["dice",10,10.5,true,null]}`,
				"JSON.RESP k3",
			},
			expected: []interface{}{"OK", []interface{}([]interface{}{"{", "b", []interface{}{"[", "dice", int64(10), "10.5", "true", "(nil)"}})},
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
