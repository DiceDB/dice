// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHSETNX(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			commands: []string{"HSETNX key_nx_t1 field value", "HSET key_nx_t1 field value_new"},
			expected: []interface{}{ONE, ZERO},
		},
		{
			commands: []string{"HSETNX key_nx_t2 field1 value1"},
			expected: []interface{}{ONE},
		},
		{
			commands: []string{"HSETNX key_nx_t3 field value", "HSETNX key_nx_t3 field new_value", "HSETNX key_nx_t3"},
			expected: []interface{}{ONE, ZERO, "ERR wrong number of arguments for 'hsetnx' command"},
		},
		{
			commands: []string{"SET key_nx_t4 v", "HSETNX key_nx_t4 f v"},
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
