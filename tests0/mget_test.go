// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"

	"github.com/dicedb/dice/testutils"
	"github.com/stretchr/testify/assert"
)

func TestMGET(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	defer client.FireString("DEL k1")
	defer client.FireString("DEL k2")

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "MGET With non-existing keys",
			commands: []string{"MGET k1 k2"},
			expected: []interface{}{[]interface{}{"(nil)", "(nil)"}},
		},
		{
			name:     "MGET With existing keys",
			commands: []string{"MSET k1 v1 k2 v2", "MGET k1 k2"},
			expected: []interface{}{"OK", []interface{}{"v1", "v2"}},
		},
		{
			name:     "MGET with existing and non existing keys",
			commands: []string{"set k1 v1", "MGET k1 k2"},
			expected: []interface{}{"OK", []interface{}{"v1", "(nil)"}},
		},
		{
			name:     "MGET without any keys",
			commands: []string{"MGET"},
			expected: []interface{}{"ERR wrong number of arguments for 'mget' command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client.FireString("DEL k1")
			client.FireString("DEL k2")

			for i, cmd := range tc.commands {
				result := client.FireString(cmd)
				if slice, ok := tc.expected[i].([]interface{}); ok {
					assert.True(t, testutils.UnorderedEqual(slice, result))
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}
}
