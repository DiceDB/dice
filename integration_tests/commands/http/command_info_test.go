// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandInfo(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "Set command",
			commands: []HTTPCommand{
				{Command: "COMMAND/INFO", Body: map[string]interface{}{"key": "SET"}},
			},
			expected: []interface{}{[]interface{}{[]interface{}{"set", float64(-3), float64(1), float64(0), float64(0), interface{}(nil)}}},
		},
		{
			name: "Get command",
			commands: []HTTPCommand{
				{Command: "COMMAND/INFO", Body: map[string]interface{}{"key": "GET"}},
			},
			expected: []interface{}{[]interface{}{[]interface{}{"get", float64(2), float64(1), float64(0), float64(0), interface{}(nil)}}},
		},
		{
			name: "PING command",
			commands: []HTTPCommand{
				{Command: "COMMAND/INFO", Body: map[string]interface{}{"key": "PING"}},
			},
			expected: []interface{}{[]interface{}{[]interface{}{"ping", float64(-1), float64(0), float64(0), float64(0), interface{}(nil)}}},
		},
		{
			name: "Combination of multiple valid commands",
			commands: []HTTPCommand{
				{Command: "COMMAND/INFO", Body: map[string]interface{}{"keys": []interface{}{"SET", "GET"}}},
			},
			expected: []interface{}{[]interface{}{
				[]interface{}{"set", float64(-3), float64(1), float64(0), float64(0), interface{}(nil)},
				[]interface{}{"get", float64(2), float64(1), float64(0), float64(0), interface{}(nil)},
			}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}
