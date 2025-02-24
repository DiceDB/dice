// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHGETALL(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()
	defer client.FireString("DEL key_hGetAll key_hGetAll02")

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
				result := client.FireString(cmd)
				expectedResults, ok := tc.expected[i].([]string)
				ok2 := result.GetVStr() != ""

				if ok && ok2 && len(result.GetVList()) == len(expectedResults) {
					expectedResultsMap := make(map[string]string)
					resultsMap := make(map[string]string)

					for i := 0; i < len(result.GetVList()); i += 2 {
						expectedResultsMap[expectedResults[i]] = expectedResults[i+1]
						resultsMap[result.GetVList()[i].GetStringValue()] = result.GetVList()[i+1].GetStringValue()
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
