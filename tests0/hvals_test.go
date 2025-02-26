// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHVals(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "RESP HVALS with multiple fields",
			commands: []string{"HSET hvalsKey field value", "HSET hvalsKey field2 value1", "HVALS hvalsKey"},
			expected: []interface{}{int64(1), int64(1), []interface{}{"value", "value1"}},
		},
		{
			name:     "RESP HVALS with non-existing key",
			commands: []string{"HVALS hvalsKey01"},
			expected: []interface{}{[]any{}},
		},
		{
			name:     "HVALS on wrong key type",
			commands: []string{"SET hvalsKey02 field", "HVALS hvalsKey02"},
			expected: []interface{}{"OK", "ERR -WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name:     "HVALS with wrong number of arguments",
			commands: []string{"HVALS hvalsKey03 x", "HVALS"},
			expected: []interface{}{"ERR wrong number of arguments for 'hvals' command", "ERR wrong number of arguments for 'hvals' command"},
		},
		{
			name:     "RESP One or more vals exist",
			commands: []string{"HSET key field value", "HSET key field1 value1", "HVALS key"},
			expected: []interface{}{int64(1), int64(1), []interface{}{"value", "value1"}},
		},
		{
			name:     "RESP No values exist",
			commands: []string{"HVALS key"},
			expected: []interface{}{[]interface{}{}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client.FireString("HDEL key field")
			client.FireString("HDEL key field1")

			for i, cmd := range tc.commands {
				result := client.FireString(cmd)
				switch e := tc.expected[i].(type) {
				case []interface{}:
					assert.ElementsMatch(t, e, tc.expected[i])
				default:
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}

	client.FireString("HDEL key field")
	client.FireString("HDEL key field1")
}
