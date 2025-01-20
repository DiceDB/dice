// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONRESP(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "jsonresp on array with mixed types",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k1", "path": "$", "json": []interface{}{"dice", 10, 10.5, true, nil}}},
				{Command: "JSON.RESP", Body: map[string]interface{}{"key": "k1", "path": "$"}},
			},
			expected: []interface{}{"OK", []interface{}{"[", "dice", float64(10), float64(10.5), true, nil}},
		},
		{
			name: "jsonresp on nested array with mixed types",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k2", "path": "$", "json": map[string]interface{}{"b": []interface{}{"dice", 10, 10.5, true, nil}}}},
				{Command: "JSON.RESP", Body: map[string]interface{}{"key": "k2", "path": "$.b"}},
			},
			expected: []interface{}{"OK", []interface{}{[]interface{}{"[", "dice", float64(10), float64(10.5), true, nil}}},
		},
		{
			name: "jsonresp on object at root path",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k3", "path": "$", "json": map[string]interface{}{"b": []interface{}{"dice", 10, 10.5, true, nil}}}},
				{Command: "JSON.RESP", Body: map[string]interface{}{"key": "k3"}},
			},
			expected: []interface{}{"OK", []interface{}{"{", "b", []interface{}{"[", "dice", float64(10), float64(10.5), true, nil}}},
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
	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": []interface{}{"k1", "k2", "k3"}}})
}
