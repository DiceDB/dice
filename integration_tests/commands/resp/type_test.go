// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package resp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestType(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "TYPE with invalid number of arguments",
			commands: []string{"TYPE"},
			expected: []interface{}{"ERR wrong number of arguments for 'type' command"},
		},
		{
			name:     "TYPE for non-existent key",
			commands: []string{"TYPE k1"},
			expected: []interface{}{"none"},
		},
		{
			name:     "TYPE for key with String value",
			commands: []string{"SET k1 v1", "TYPE k1"},
			expected: []interface{}{"OK", "string"},
		},
		{
			name:     "TYPE for key with List value",
			commands: []string{"LPUSH k1 v1", "TYPE k1"},
			expected: []interface{}{int64(1), "list"},
		},
		{
			name:     "TYPE for key with Set value",
			commands: []string{"SADD k1 v1", "TYPE k1"},
			expected: []interface{}{int64(1), "set"},
		},
		{
			name:     "TYPE for key with Hash value",
			commands: []string{"HSET k1 field1 v1", "TYPE k1"},
			expected: []interface{}{int64(1), "hash"},
		},
		{
			name:     "TYPE for key with value created from SETBIT command",
			commands: []string{"SETBIT k1 1 1", "TYPE k1"},
			expected: []interface{}{int64(0), "string"},
		},
		// TODO: uncomment when bitop is added
		// {
		// 	name:     "TYPE for key with value created from SETOP command",
		// 	commands: []string{"SET key1 \"foobar\"", "SET key2 \"abcdef\"", "TYPE dest"},
		// 	expected: []interface{}{"OK", "OK", "string"},
		// },
		{
			name: "TYPE for key with Zset value",
			commands: []string{
				"ZADD myzetset 1 'one' 2 'two'",
				"TYPE myzetset",
			},
			expected: []interface{}{int64(2), "zset"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			FireCommand(conn, "DEL k1")

			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}
