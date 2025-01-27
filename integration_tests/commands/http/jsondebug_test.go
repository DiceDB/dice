// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONDEBUG(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "jsondebug with no path",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k1", "path": "$", "json": map[string]interface{}{"a": 1}}},
				{Command: "JSON.DEBUG", Body: map[string]interface{}{"values": []interface{}{"MEMORY", "k1"}}},
			},
			expected: []interface{}{"OK", float64(72)},
		},
		{
			name: "jsondebug with a valid path",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k2", "path": "$", "json": map[string]interface{}{"a": 1, "b": 2}}},
				{Command: "JSON.DEBUG", Body: map[string]interface{}{"values": []interface{}{"MEMORY", "k2", "$.a"}}},
			},
			expected: []interface{}{"OK", []interface{}{float64(16)}},
		},
		{
			name: "jsondebug with multiple paths",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k3", "path": "$", "json": map[string]interface{}{"a": 1, "b": "dice"}}},
				{Command: "JSON.DEBUG", Body: map[string]interface{}{"values": []interface{}{"MEMORY", "k3", "$.a", "$.b"}}},
			},
			expected: []interface{}{"OK", []interface{}{float64(16)}},
		},
		{
			name: "jsondebug with multiple paths",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k4", "path": "$", "json": map[string]interface{}{"a": 1, "b": "dice"}}},
				{Command: "JSON.DEBUG", Body: map[string]interface{}{"values": []interface{}{"MEMORY", "k4", "$.a", "$.b"}}},
			},
			expected: []interface{}{"OK", []interface{}{float64(16)}},
		},
		{
			name: "jsondebug with single path for array json",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k5", "path": "$", "json": []interface{}{"roll", "the", "dices"}}},
				{Command: "JSON.DEBUG", Body: map[string]interface{}{"values": []interface{}{"MEMORY", "k5", "$[1]"}}},
			},
			expected: []interface{}{"OK", []interface{}{float64(19)}},
		},
		{
			name: "jsondebug with multiple paths for array json",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k6", "path": "$", "json": []interface{}{"roll", "the", "dices"}}},
				{Command: "JSON.DEBUG", Body: map[string]interface{}{"values": []interface{}{"MEMORY", "k6", "$[1,2]"}}},
			},
			expected: []interface{}{"OK", []interface{}{float64(19), float64(21)}},
		},
		{
			name: "jsondebug with all paths for array json",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k7", "path": "$", "json": []interface{}{"roll", "the", "dices"}}},
				{Command: "JSON.DEBUG", Body: map[string]interface{}{"values": []interface{}{"MEMORY", "k7", "$[:]"}}},
			},
			expected: []interface{}{"OK", []interface{}{float64(20), float64(19), float64(21)}},
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

	// Cleanup the used keys
	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": []interface{}{"k1", "k2", "k3", "k4", "k5", "k6", "k7"}}})
}
