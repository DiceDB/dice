// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHExists(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "RESP wrong number of arguments for HEXISTS",
			commands: []string{"HEXISTS", "HEXISTS KEY", "HEXISTS KEY FIELD ANOTHER_FIELD"},
			expected: []interface{}{"ERR wrong number of arguments for 'hexists' command",
				"ERR wrong number of arguments for 'hexists' command",
				"ERR wrong number of arguments for 'hexists' command"},
		},
		{
			name:     "RESP HEXISTS non existent key",
			commands: []string{"HSET key_hExists1 field value", "HEXISTS wrong_key_hExists field"},
			expected: []interface{}{int64(1), int64(0)},
		},
		{
			name:     "RESP HEXISTS non existent field",
			commands: []string{"HSET key_hExists2 field value", "HEXISTS key_hExists2 wrong_field"},
			expected: []interface{}{int64(1), int64(0)},
		},
		{
			name:     "RESP HEXISTS existent key and field",
			commands: []string{"HSET key_hExists3 field HelloWorld", "HEXISTS key_hExists3 field"},
			expected: []interface{}{int64(1), int64(1)},
		},
		{
			name:     "RESP Check if field exists when k f and v are set",
			commands: []string{"HSET key field value", "HEXISTS key field"},
			expected: []interface{}{int64(1), int64(1)},
		},
		{
			name:     "RESP Check if field exists when k exists but not f and v",
			commands: []string{"HSET key field1 value", "HEXISTS key field"},
			expected: []interface{}{int64(1), int64(0)},
		},
		{
			name:     "RESP Check if field exists when no k,f and v exist",
			commands: []string{"HEXISTS key field"},
			expected: []interface{}{int64(0)},
		},
		{
			name:     "RESP HEXISTS operation against a key holding the wrong kind of value",
			commands: []string{"SET key value", "HEXISTS key field"},
			expected: []interface{}{"OK", "ERR -WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client.FireString("HDEL key field")
			client.FireString("DEL key value")
			client.FireString("HDEL key field1")

			for i, cmd := range tc.commands {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}

	client.FireString("HDEL key field")
	client.FireString("DEL key value")
	client.FireString("HDEL key field1")
}
