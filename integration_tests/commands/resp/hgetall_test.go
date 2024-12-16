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
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHGETALL(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	defer FireCommand(conn, "DEL key_hGetAll key_hGetAll02")

	testCases := []TestCase{
		{
			commands: []string{"HSET key_hGetAll field value", "HSET key_hGetAll field2 value_new", "HGETALL key_hGetAll"},
			expected: []interface{}{ONE, ONE, []string{"field", "value", "field2", "value_new"}},
		},
		{
			commands: []string{"HGETALL key_hGetAll01"},
			expected: []interface{}{[]interface{}{}},
		},
		{
			commands: []string{"SET key_hGetAll02 field", "HGETALL key_hGetAll02"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			commands: []string{"HGETALL key_hGetAll03 x", "HGETALL"},
			expected: []interface{}{"ERR wrong number of arguments for 'hgetall' command",
				"ERR wrong number of arguments for 'hgetall' command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				expectedResults, ok := tc.expected[i].([]string)
				results, ok2 := result.([]interface{})

				if ok && ok2 && len(results) == len(expectedResults) {
					expectedResultsMap := make(map[string]string)
					resultsMap := make(map[string]string)

					for i := 0; i < len(results); i += 2 {
						expectedResultsMap[expectedResults[i]] = expectedResults[i+1]
						resultsMap[results[i].(string)] = results[i+1].(string)
					}
					if !reflect.DeepEqual(resultsMap, expectedResultsMap) {
						t.Fatalf("Assertion failed: expected true, got false")
					}

				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}
}
