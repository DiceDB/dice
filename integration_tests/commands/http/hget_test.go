// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package http

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHGet(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		delays   []time.Duration
	}{
		{
			name: "HGET with wrong number of arguments",
			commands: []HTTPCommand{
				{Command: "HGET", Body: map[string]interface{}{"key": nil}},
				{Command: "HGET", Body: map[string]interface{}{"key": "key_hGet1"}},
			},
			expected: []interface{}{
				"ERR wrong number of arguments for 'hget' command",
				"ERR wrong number of arguments for 'hget' command"},
			delays: []time.Duration{0, 0},
		},
		{
			name: "HGET on existent hash",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hGet2", "key_values": map[string]interface{}{"field1": "value1", "field2": "value2", "field3": "value3"}}},
				{Command: "HGET", Body: map[string]interface{}{"key": "key_hGet2", "field": "field2"}},
			},
			expected: []interface{}{float64(3), "value2"},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "HGET on non-existent field",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hGet3", "key_values": map[string]interface{}{"field1": "value1", "field2": "value2"}}},
				{Command: "HGET", Body: map[string]interface{}{"key": "key_hGet3", "field": "field2"}},
				{Command: "HDEL", Body: map[string]interface{}{"key": "key_hGet3", "field": "field2"}},
				{Command: "HGET", Body: map[string]interface{}{"key": "key_hGet3", "field": "field2"}},
				{Command: "HGET", Body: map[string]interface{}{"key": "key_hGet3", "field": "field3"}},
			},
			expected: []interface{}{float64(2), "value2", float64(1), nil, nil},
			delays:   []time.Duration{0, 0, 0, 0, 0},
		},
		{
			name: "HGET on non-existent hash",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hGet4", "key_values": map[string]interface{}{"field1": "value1", "field2": "value2"}}},
				{Command: "HGET", Body: map[string]interface{}{"key": "wrong_key_hGet4", "field": "field2"}},
			},
			expected: []interface{}{float64(2), nil},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "HGET with wrong type",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "string_key", "value": "value"}},
				{Command: "HGET", Body: map[string]interface{}{"key": "string_key", "field": "field"}},
			},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			delays:   []time.Duration{0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"key_hGet1", "key_hGet2", "key_hGet3", "key_hGet4", "key_hGet5", "string_key"}}})

			for i, cmd := range tc.commands {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result, err := exec.FireCommand(cmd)
				if err != nil {
					log.Println(tc.expected[i])
					assert.Equal(t, tc.expected[i], err.Error(), "Error message mismatch for cmd %s", cmd)
				} else {
					assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s, expected %v, got %v", cmd, tc.expected[i], result)
				}
			}
		})
	}
}
