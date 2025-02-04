// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHKeys(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "RESP HKEYS with key containing hash with multiple fields",
			commands: []string{"HSET key_hkeys field1 value1", "HSET key_hkeys field2 value2", "HKEYS key_hkeys"},
			expected: []interface{}{int64(1), int64(1), []interface{}{"field1", "field2"}},
		},
		{
			name:     "RESP HKEYS with non-existent key",
			commands: []string{"HKEYS key_hkeys01"},
			expected: []interface{}{[]interface{}{}},
		},
		{
			name:     "RESP HKEYS with key containing a non-hash value",
			commands: []string{"SET key_hkeys02 field1", "HKEYS key_hkeys02"},
			expected: []interface{}{"OK", "ERR -WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name:     "RESP HKEYS with wrong number of arguments",
			commands: []string{"HKEYS key_hkeys03 x", "HKEYS"},
			expected: []interface{}{"ERR wrong number of arguments for 'hkeys' command",
				"ERR wrong number of arguments for 'hkeys' command"},
		},
		{
			name:     "RESP One or more keys exist",
			commands: []string{"HSET key field value", "HSET key field1 value1", "HKEYS key"},
			expected: []interface{}{int64(1), int64(1), []interface{}{"field", "field1"}},
		},
		{
			name:     "RESP No keys exist",
			commands: []string{"HKEYS key"},
			expected: []interface{}{[]interface{}{}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client.FireString("HDEL key field")
			client.FireString("HDEL key field1")
			client.FireString("DEL key")

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
	client.FireString("DEL key")
}
