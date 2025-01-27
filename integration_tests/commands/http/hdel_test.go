// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package http

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHDel(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		delays   []time.Duration
	}{
		{
			name: "HDEL with wrong number of arguments",
			commands: []HTTPCommand{
				{Command: "HDEL", Body: map[string]interface{}{"key": nil}},
				{Command: "HDEL", Body: map[string]interface{}{"key": "key_hDel1"}},
			},
			expected: []interface{}{
				"ERR wrong number of arguments for 'hdel' command",
				"ERR wrong number of arguments for 'hdel' command"},
			delays: []time.Duration{0, 0},
		},
		{
			name: "HDEL with single field",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hDel2", "field": "field1", "value": "value1"}},
				{Command: "HLEN", Body: map[string]interface{}{"key": "key_hDel2"}},
				{Command: "HDEL", Body: map[string]interface{}{"key": "key_hDel2", "field": "field1"}},
				{Command: "HLEN", Body: map[string]interface{}{"key": "key_hDel2"}},
			},
			expected: []interface{}{float64(1), float64(1), float64(1), float64(0)},
			delays:   []time.Duration{0, 0, 0, 0},
		},
		{
			name: "HDEL with multiple fields",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hDel3", "key_values": map[string]interface{}{"field1": "value1", "field2": "value2", "field3": "value3", "field4": "value4"}}},
				{Command: "HLEN", Body: map[string]interface{}{"key": "key_hDel3"}},
				{Command: "HDEL", Body: map[string]interface{}{"key": "key_hDel3", "values": []string{"field1", "field2"}}},
				{Command: "HLEN", Body: map[string]interface{}{"key": "key_hDel3"}},
			},
			expected: []interface{}{float64(4), float64(4), float64(2), float64(2)},
			delays:   []time.Duration{0, 0, 0, 0},
		},
		{
			name: "HDEL on non-existent field",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hDel4", "key_values": map[string]interface{}{"field1": "value1", "field2": "value2"}}},
				{Command: "HDEL", Body: map[string]interface{}{"key": "key_hDel4", "field": "field3"}},
			},
			expected: []interface{}{float64(2), float64(0)},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "HDEL on non-existent hash",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hDel5", "key_values": map[string]interface{}{"field1": "value1", "field2": "value2"}}},
				{Command: "HDEL", Body: map[string]interface{}{"key": "wrong_key_hDel5", "field": "field1"}},
			},
			expected: []interface{}{float64(2), float64(0)},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "HDEL with wrong type",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "string_key", "value": "value"}},
				{Command: "HDEL", Body: map[string]interface{}{"key": "string_key", "field": "field"}},
			},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			delays:   []time.Duration{0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"key_hDel1", "key_hDel2", "key_hDel3", "key_hDel4", "key_hDel5", "string_key"}}})

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
