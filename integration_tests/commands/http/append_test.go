// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPPEND(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		cleanup  []HTTPCommand
	}{
		{
			name: "APPEND and GET a new Val",
			commands: []HTTPCommand{
				{Command: "APPEND", Body: map[string]interface{}{"key": "k", "value": "newVal"}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{float64(6), "newVal"},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "k"}},
			},
		},
		{
			name: "APPEND to an exisiting key and GET",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "Bhima"}},
				{Command: "APPEND", Body: map[string]interface{}{"key": "k", "value": "Shankar"}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", float64(12), "BhimaShankar"},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "k"}},
			},
		},
		{
			name: "APPEND without input value",
			commands: []HTTPCommand{
				{Command: "APPEND", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"ERR wrong number of arguments for 'append' command"},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "k"}},
			},
		},
		{
			name: "APPEND empty string to an exsisting key with empty string",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": ""}},
				{Command: "APPEND", Body: map[string]interface{}{"key": "k", "value": ""}},
			},
			expected: []interface{}{"OK", float64(0)},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "k"}},
			},
		},
		{
			name: "APPEND to key created using LPUSH",
			commands: []HTTPCommand{
				{Command: "LPUSH", Body: map[string]interface{}{"key": "m", "value": "bhima"}},
				{Command: "APPEND", Body: map[string]interface{}{"key": "m", "value": "shankar"}},
			},
			expected: []interface{}{float64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "m"}},
			},
		},
		{
			name: "APPEND with leading zeros",
			commands: []HTTPCommand{
				{Command: "DEL", Body: map[string]interface{}{"key": "key"}},
				{Command: "APPEND", Body: map[string]interface{}{"key": "key", "value": "0043"}},
				{Command: "GET", Body: map[string]interface{}{"key": "key"}},
				{Command: "APPEND", Body: map[string]interface{}{"key": "key", "value": "0034"}},
				{Command: "GET", Body: map[string]interface{}{"key": "key"}},
			},
			expected: []interface{}{float64(0), float64(4), "0043", float64(8), "00430034"},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "key"}},
			},
		},
		{
			name: "APPEND to key created using SADD",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "key", "value": "apple"}},
				{Command: "APPEND", Body: map[string]interface{}{"key": "key", "value": "banana"}},
			},
			expected: []interface{}{float64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "key"}},
			},
		},
		{
			name: "APPEND to key created using ZADD",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset", "values": []string{"1", "one"}}},
				{Command: "APPEND", Body: map[string]interface{}{"key": "myzset", "value": "two"}},
			},
			expected: []interface{}{float64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "myzset"}},
			},
		},
		{
			name: "APPEND to key created using SETBIT",
			commands: []HTTPCommand{
				{Command: "SETBIT", Body: map[string]interface{}{"key": "bitkey", "values": []string{"2", "1"}}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "bitkey", "values": []string{"3", "1"}}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "bitkey", "values": []string{"5", "1"}}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "bitkey", "values": []string{"10", "1"}}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "bitkey", "values": []string{"11", "1"}}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "bitkey", "values": []string{"14", "1"}}},
				{Command: "APPEND", Body: map[string]interface{}{"key": "bitkey", "value": "1"}},
				{Command: "GET", Body: map[string]interface{}{"key": "bitkey"}},
			},
			expected: []interface{}{float64(0), float64(0), float64(0), float64(0), float64(0), float64(0), float64(3), "421"},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "bitkey"}},
			},
		},
		{
			name: "SET and SETBIT commands followed by GET",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "key", "value": "10"}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "key", "values": []string{"1", "1"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "key"}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "key", "values": []string{"0", "1"}}},
			},
			expected: []interface{}{"OK", float64(0), "q0", float64(0)},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "key"}},
			},
		},
		{
			name: "APPEND After SET and DEL",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "key", "value": "value"}},
				{Command: "APPEND", Body: map[string]interface{}{"key": "key", "value": "value"}},
				{Command: "GET", Body: map[string]interface{}{"key": "key"}},
				{Command: "APPEND", Body: map[string]interface{}{"key": "key", "value": "100"}},
				{Command: "GET", Body: map[string]interface{}{"key": "key"}},
				{Command: "DEL", Body: map[string]interface{}{"key": "key"}},
				{Command: "APPEND", Body: map[string]interface{}{"key": "key", "value": "value"}},
				{Command: "GET", Body: map[string]interface{}{"key": "key"}},
			},
			expected: []interface{}{"OK", float64(10), "valuevalue", float64(13), "valuevalue100", float64(1), float64(5), "value"},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "key"}},
			},
		},
		{
			name: "APPEND to Integer Values",
			commands: []HTTPCommand{
				{Command: "DEL", Body: map[string]interface{}{"key": "key"}},
				{Command: "APPEND", Body: map[string]interface{}{"key": "key", "value": "1"}},
				{Command: "APPEND", Body: map[string]interface{}{"key": "key", "value": "2"}},
				{Command: "GET", Body: map[string]interface{}{"key": "key"}},
				{Command: "SET", Body: map[string]interface{}{"key": "key", "value": "1"}},
				{Command: "APPEND", Body: map[string]interface{}{"key": "key", "value": "2"}},
				{Command: "GET", Body: map[string]interface{}{"key": "key"}},
			},
			expected: []interface{}{float64(0), float64(1), float64(2), "12", "OK", float64(2), "12"},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "key"}},
			},
		},
		{
			name: "APPEND with Various Data Types",
			commands: []HTTPCommand{
				{Command: "LPUSH", Body: map[string]interface{}{"key": "listKey", "value": "lValue"}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "bitKey", "offset": 0, "value": 1}},
				{Command: "HSET", Body: map[string]interface{}{"key": "hashKey", "field": "hKey", "value": "hValue"}},
				{Command: "SADD", Body: map[string]interface{}{"key": "setKey", "value": "sValue"}},
				{Command: "APPEND", Body: map[string]interface{}{"key": "listKey", "value": "value"}},
				{Command: "APPEND", Body: map[string]interface{}{"key": "bitKey", "value": "value"}},
				{Command: "APPEND", Body: map[string]interface{}{"key": "hashKey", "value": "value"}},
				{Command: "APPEND", Body: map[string]interface{}{"key": "setKey", "value": "value"}},
			},
			expected: []interface{}{
				float64(1),
				float64(0),
				float64(1),
				float64(1),
				"WRONGTYPE Operation against a key holding the wrong kind of value",
				float64(6),
				"WRONGTYPE Operation against a key holding the wrong kind of value",
				"WRONGTYPE Operation against a key holding the wrong kind of value",
			},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "listKey"}},
				{Command: "del", Body: map[string]interface{}{"key": "bitKey"}},
				{Command: "del", Body: map[string]interface{}{"key": "hashKey"}},
				{Command: "del", Body: map[string]interface{}{"key": "setKey"}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
			exec.FireCommand(tc.cleanup[0])
		})
	}

	ttlTestCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		cleanup  []HTTPCommand
	}{
		{
			name: "APPEND with TTL Set",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "key", "value": "Hello", "ex": 10}}, // Set a key with a 10-second TTL
				{Command: "TTL", Body: map[string]interface{}{"key": "key"}},                             // Check initial TTL
				{Command: "APPEND", Body: map[string]interface{}{"key": "key", "value": "World"}},        // Append a value
				{Command: "GET", Body: map[string]interface{}{"key": "key"}},                             // Get the final value
				{Command: "SLEEP", Body: map[string]interface{}{"seconds": 2}},                           // Sleep for 2 seconds
				{Command: "TTL", Body: map[string]interface{}{"key": "key"}},                             // Check TTL after append
			},
			expected: []interface{}{"OK", float64(10), float64(10), "HelloWorld", "OK", float64(8)},
			cleanup: []HTTPCommand{
				{Command: "DEL", Body: map[string]interface{}{"key": "key"}},
			},
		},
		{
			name: "APPEND before near TTL Expiry",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "key", "value": "Hello", "ex": 3}}, // Set a key with a 3-second TTL
				{Command: "TTL", Body: map[string]interface{}{"key": "key"}},                            // Check initial TTL
				{Command: "APPEND", Body: map[string]interface{}{"key": "key", "value": "World"}},       // Append a value
				{Command: "SLEEP", Body: map[string]interface{}{"seconds": 3}},                          // Sleep for 3 seconds
				{Command: "GET", Body: map[string]interface{}{"key": "key"}},                            // Get the final value which should be nil
			},
			expected: []interface{}{"OK", float64(3), float64(10), "OK", (nil)},
			cleanup: []HTTPCommand{
				{Command: "DEL", Body: map[string]interface{}{"key": "key"}},
			},
		},
	}

	for _, tc := range ttlTestCases {
		t.Run(tc.name, func(t *testing.T) {
			exec := NewHTTPCommandExecutor()

			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)

				// Apply TTL tolerance checks if the command is "TTL"
				if cmd.Command == "TTL" {
					expectedTTL := tc.expected[i].(float64)
					actualTTL := result.(float64)

					if actualTTL == -2 { // Key does not exist or has expired
						assert.Equal(t, tc.expected[i], result)
					} else {
						assert.Condition(t, func() bool {
							return actualTTL >= expectedTTL-2 && actualTTL <= expectedTTL+2
						}, "TTL %f not within expected range [%f, %f]", actualTTL, expectedTTL-2, expectedTTL+2)
					}
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}

			for _, cleanupCmd := range tc.cleanup {
				exec.FireCommand(cleanupCmd)
			}
		})
	}
}
