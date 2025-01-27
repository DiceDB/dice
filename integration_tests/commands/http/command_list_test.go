// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandList(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "Command list should not be empty",
			commands: []HTTPCommand{
				{Command: "COMMAND/LIST", Body: map[string]interface{}{"key": ""}},
			},
			expected: []interface{}{"ERR wrong number of arguments for 'command|list' command"},
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}
