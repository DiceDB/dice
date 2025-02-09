// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPPEND(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()
	client.FireString("FLUSHDB")

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
		cleanup  []string
	}{
		{
			name:     "APPEND and GET a new Val",
			commands: []string{"APPEND k newVal", "GET k"},
			expected: []interface{}{int64(6), "newVal"},
			cleanup:  []string{"DEL k"},
		},
		{
			name:     "APPEND to an existing key and GET",
			commands: []string{"SET k Bhima", "APPEND k Shankar", "GET k"},
			expected: []interface{}{"OK", int64(12), "BhimaShankar"},
			cleanup:  []string{"DEL k"},
		},
		{
			name:     "APPEND without input value",
			commands: []string{"APPEND k"},
			expected: []interface{}{"ERR wrong number of arguments for 'append' command"},
			cleanup:  []string{"DEL k"},
		},
		{
			name:     "APPEND empty string to an existing key with empty string",
			commands: []string{"SET k \"\"", "APPEND k \"\""},
			expected: []interface{}{"OK", int64(0)},
			cleanup:  []string{"DEL k"},
		},
		{
			name:     "APPEND to key created using LPUSH",
			commands: []string{"LPUSH m bhima", "APPEND m shankar"},
			expected: []interface{}{int64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			cleanup:  []string{"DEL m"},
		},
		{
			name:     "APPEND with leading zeros",
			commands: []string{"DEL key", "APPEND key 0043", "GET key", "APPEND key 0034", "GET key"},
			expected: []interface{}{int64(0), int64(4), "0043", int64(8), "00430034"},
			cleanup:  []string{"DEL key"},
		},
		{
			name:     "APPEND to key created using SADD",
			commands: []string{"SADD key apple", "APPEND key banana"},
			expected: []interface{}{int64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			cleanup:  []string{"DEL key"},
		},
		{
			name:     "APPEND to key created using ZADD",
			commands: []string{"ZADD key 1 one", "APPEND key two"},
			expected: []interface{}{int64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			cleanup:  []string{"DEL key"},
		},
		{
			name:     "APPEND After SET and DEL",
			commands: []string{"SET key value", "APPEND key value", "GET key", "APPEND key 100", "GET key", "DEL key", "APPEND key value", "GET key"},
			expected: []interface{}{"OK", int64(10), "valuevalue", int64(13), "valuevalue100", int64(1), int64(5), "value"},
			cleanup:  []string{"DEL key"},
		},
		{
			name:     "APPEND to Integer Values",
			commands: []string{"DEL key", "APPEND key 1", "APPEND key 2", "GET key", "SET key 1", "APPEND key 2", "GET key"},
			expected: []interface{}{int64(0), int64(1), int64(2), "12", "OK", int64(2), "12"},
			cleanup:  []string{"DEL key"},
		},
		{
			name: "APPEND with Various Data Types",
			commands: []string{
				"LPUSH listKey lValue",
				"SETBIT bitKey 0 1",
				"HSET hashKey hKey hValue",
				"SADD setKey sValue",
				"APPEND listKey value",
				"APPEND bitKey value",
				"APPEND hashKey value",
				"APPEND setKey value",
			},
			expected: []interface{}{
				int64(1),
				int64(0),
				int64(1),
				int64(1),
				"WRONGTYPE Operation against a key holding the wrong kind of value",
				int64(6),
				"WRONGTYPE Operation against a key holding the wrong kind of value",
				"WRONGTYPE Operation against a key holding the wrong kind of value",
			},
			cleanup: []string{"DEL listKey", "DEL bitKey", "DEL hashKey", "DEL setKey"},
		},
		{
			name:     "APPEND to key created using SETBIT",
			commands: []string{"SETBIT bitkey 2 1", "SETBIT bitkey 3 1", "SETBIT bitkey 5 1", "SETBIT bitkey 10 1", "SETBIT bitkey 11 1", "SETBIT bitkey 14 1", "APPEND bitkey 1", "GET bitkey"},
			expected: []interface{}{int64(0), int64(0), int64(0), int64(0), int64(0), int64(0), int64(3), "421"},
			cleanup:  []string{"del bitkey"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			for i := 0; i < len(tc.commands); i++ {
				result := client.FireString(tc.commands[i])
				expected := tc.expected[i]
				assert.Equal(t, expected, result)
			}

			for _, cmd := range tc.cleanup {
				client.FireString(cmd)
			}
		})
	}

	setErrorMsg := "WRONGTYPE Operation against a key holding the wrong kind of value"
	ttlTolerance := int64(2) // Set tolerance in seconds for the TTL checks
	ttlTestCases := []struct {
		name     string
		commands []string
		expected []interface{}
		cleanup  []string
	}{
		{
			name: "APPEND with TTL Set",
			commands: []string{
				"SET key Hello EX 10", // Set a key with a 10-second TTL
				"TTL key",             // Check initial TTL
				"APPEND key World",    // Append a value
				"GET key",             // Get the final value
				"SLEEP 2",             // Sleep for 2 seconds
				"TTL key",             // Check TTL after append.
			},
			expected: []interface{}{"OK", int64(10), int64(10), "HelloWorld", "OK", int64(8), setErrorMsg, setErrorMsg, setErrorMsg, setErrorMsg},
			cleanup:  []string{"DEL key"},
		},
		{
			name: "APPEND before near TTL Expiry",
			commands: []string{
				"SET key Hello EX 3", // Set a key with a 10-second TTL
				"TTL key",            // Check initial TTL
				"APPEND key World",   // Append a value
				"SLEEP 3",            // Sleep for 5 seconds
				"GET key",            // Get the final value which should be (nil)
			},
			expected: []interface{}{"OK", int64(3), int64(10), "OK", "(nil)", int64(-2), setErrorMsg, setErrorMsg, setErrorMsg, setErrorMsg},
			cleanup:  []string{"DEL key"},
		},
	}

	for _, tc := range ttlTestCases {
		t.Run(tc.name, func(t *testing.T) {
			client.FireString("DEL key")

			for i, cmd := range tc.commands {
				result := client.FireString(cmd)
				// If checking TTL, apply a tolerance to account for system performance variability
				if cmd == "TTL key" { // Check if TTL command is executed
					expectedTTL := tc.expected[i].(int64)
					actualTTL := result.GetVInt()

					if actualTTL == -2 { // Key does not exist or is expired
						assert.Equal(t, tc.expected[i], result)
					} else {
						assert.Condition(t, func() bool {
							return actualTTL >= expectedTTL-ttlTolerance && actualTTL <= expectedTTL+ttlTolerance
						}, "TTL %d not within expected range [%d, %d]", actualTTL, expectedTTL-ttlTolerance, expectedTTL+ttlTolerance)
					}
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}

			for _, cmd := range tc.cleanup {
				client.FireString(cmd)
			}
		})
	}
}
