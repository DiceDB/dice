// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package http

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHSet(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		delays   []time.Duration
	}{
		{
			name: "HSET with wrong number of arguments",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": nil}},
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hSet1", "field": nil}},
			},
			expected: []interface{}{
				"ERR wrong number of arguments for 'hset' command",
				"ERR wrong number of arguments for 'hset' command"},
			delays: []time.Duration{0, 0},
		},
		{
			name: "HSET with single field",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hSet2", "field": "field1", "value": "value1"}},
				{Command: "HLEN", Body: map[string]interface{}{"key": "key_hSet2"}},
			},
			expected: []interface{}{float64(1), float64(1)},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "HSET with multiple fields",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hSet3", "key_values": map[string]interface{}{"field1": "value1", "field2": "value2", "field3": "value3"}}},
				{Command: "HLEN", Body: map[string]interface{}{"key": "key_hSet3"}},
			},
			expected: []interface{}{float64(3), float64(3)},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "HSET on existing hash",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hSet4", "key_values": map[string]interface{}{"field1": "value1", "field2": "value2"}}},
				{Command: "HGET", Body: map[string]interface{}{"key": "key_hSet4", "field": "field2"}},
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hSet4", "key_values": map[string]interface{}{"field2": "newvalue2"}}},
				{Command: "HGET", Body: map[string]interface{}{"key": "key_hSet4", "field": "field2"}},
			},
			expected: []interface{}{float64(2), "value2", float64(0), "newvalue2"},
			delays:   []time.Duration{0, 0, 0, 0},
		},
		{
			name: "HSET with wrong type",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "string_key", "value": "value"}},
				{Command: "HSET", Body: map[string]interface{}{"key": "string_key", "field": "field", "value": "value"}},
			},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			delays:   []time.Duration{0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"key_hSet1", "key_hSet2", "key_hSet3", "key_hSet4", "string_key"}}})

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
