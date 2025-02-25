// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHDEL(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			commands: []string{"HSET key field value", "HDEL key field"},
			expected: []interface{}{ONE, ONE},
		},
		{
			commands: []string{"HSET key field1 value1", "HDEL key field1"},
			expected: []interface{}{ONE, ONE},
		},
		{
			commands: []string{"HSET key field2 value2 field3 value3", "HDEL key field2 field3"},
			expected: []interface{}{TWO, TWO},
		},
		{
			commands: []string{"HSET key_new field value", "HDEL key_new field", "HDEL key_new"},
			expected: []interface{}{ONE, ONE, "ERR wrong number of arguments for 'hdel' command"},
		},
		{
			commands: []string{"SET k v", "HDEL k f"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
	}

	for _, tc := range testCases {
		for i, cmd := range tc.commands {
			result := client.FireString(cmd)
			assert.Equal(t, tc.expected[i], result)
		}
	}
}
