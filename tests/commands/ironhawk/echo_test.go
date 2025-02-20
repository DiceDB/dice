// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"

	"github.com/dicedb/dice/testutils"
	"github.com/stretchr/testify/assert"
)

func TestEcho(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "ECHO with invalid number of arguments",
			commands: []string{"ECHO"},
			expected: []interface{}{"wrong number of arguments for 'ECHO' command"},
		},
		{
			name:     "ECHO with one argument",
			commands: []string{"ECHO Hello"},
			expected: []interface{}{"Hello"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expected[i], testutils.ParseTestResponse(result), "Value mismatch for cmd %s", cmd)
			}
		})
	}
}
