// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package websocket

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppend(t *testing.T) {
	exec := NewWebsocketCommandExecutor()

	testCases := []struct {
		name       string
		commands   []string
		expected   []interface{}
		cleanupKey string
	}{
		{
			name:       "APPEND and GET a new Val",
			commands:   []string{"APPEND k newVal", "GET k"},
			expected:   []interface{}{float64(6), "newVal"},
			cleanupKey: "k",
		},
		{
			name:       "APPEND to an existing key and GET",
			commands:   []string{"SET k Bhima", "APPEND k Shankar", "GET k"},
			expected:   []interface{}{"OK", float64(12), "BhimaShankar"},
			cleanupKey: "k",
		},
		{
			name:       "APPEND without input value",
			commands:   []string{"APPEND k"},
			expected:   []interface{}{"ERR wrong number of arguments for 'append' command"},
			cleanupKey: "k",
		},
		{
			name:       "APPEND to key created using LPUSH",
			commands:   []string{"LPUSH z bhima", "APPEND z shankar"},
			expected:   []interface{}{float64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			cleanupKey: "z",
		},
		{
			name:       "APPEND with leading zeros",
			commands:   []string{"DEL key", "APPEND key 0043", "GET key", "APPEND key 0034", "GET key"},
			expected:   []interface{}{float64(0), float64(4), "0043", float64(8), "00430034"},
			cleanupKey: "key",
		},
		{
			name:       "APPEND to key created using ZADD",
			commands:   []string{"ZADD key 1 one", "APPEND key two"},
			expected:   []interface{}{float64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			cleanupKey: "key",
		},
		{
			name:       "APPEND to key created using SETBIT",
			commands:   []string{"SETBIT bitkey 2 1", "SETBIT bitkey 3 1", "SETBIT bitkey 5 1", "SETBIT bitkey 10 1", "SETBIT bitkey 11 1", "SETBIT bitkey 14 1", "APPEND bitkey 1", "GET bitkey"},
			expected:   []interface{}{float64(0), float64(0), float64(0), float64(0), float64(0), float64(0), float64(3), "421"},
			cleanupKey: "bitkey",
		},
		{
			name:       "APPEND After SET and DEL",
			commands:   []string{"SET key value", "APPEND key value", "GET key", "APPEND key 100", "GET key", "DEL key", "APPEND key value", "GET key"},
			expected:   []interface{}{"OK", float64(10), "valuevalue", float64(13), "valuevalue100", float64(1), float64(5), "value"},
			cleanupKey: "key",
		},
		{
			name:       "APPEND to Integer Values",
			commands:   []string{"DEL key", "APPEND key 1", "APPEND key 2", "GET key", "SET key 1", "APPEND key 2", "GET key"},
			expected:   []interface{}{float64(0), float64(1), float64(2), "12", "OK", float64(2), "12"},
			cleanupKey: "key",
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
				float64(1),
				float64(0),
				float64(1),
				float64(1),
				"WRONGTYPE Operation against a key holding the wrong kind of value",
				float64(6),
				"WRONGTYPE Operation against a key holding the wrong kind of value",
				"WRONGTYPE Operation against a key holding the wrong kind of value",
			},
			cleanupKey: "listKey",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conn := exec.ConnectToServer()

			for i, cmd := range tc.commands {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				assert.Equal(t, tc.expected[i], result)
			}
			DeleteKey(t, conn, exec, tc.cleanupKey)
		})
	}

	ttlTolerance := float64(2) // Set tolerance in seconds for the TTL checks
	ttlTestCases := []struct {
		name       string
		commands   []string
		expected   []interface{}
		cleanupKey string
	}{
		{
			name:       "APPEND with TTL Set",
			commands:   []string{"SET key Hello EX 10", "TTL key", "APPEND key World", "GET key", "SLEEP 2", "TTL key"},
			expected:   []interface{}{"OK", float64(10), float64(10), "HelloWorld", "OK", float64(8)},
			cleanupKey: "key",
		},
		{
			name:       "APPEND before near TTL Expiry",
			commands:   []string{"SET key Hello EX 3", "TTL key", "APPEND key World", "SLEEP 3", "GET key"},
			expected:   []interface{}{"OK", float64(3), float64(10), "OK", (nil)},
			cleanupKey: "key",
		},
	}

	for _, tc := range ttlTestCases {
		t.Run(tc.name, func(t *testing.T) {
			conn := exec.ConnectToServer()

			for i, cmd := range tc.commands {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)

				// TTL tolerance handling
				if cmd == "TTL key" {
					expectedTTL := tc.expected[i].(float64)
					actualTTL := result.(float64)
					if actualTTL == -2 {
						assert.Equal(t, tc.expected[i], result)
					} else {
						assert.Condition(t, func() bool {
							return actualTTL >= expectedTTL-ttlTolerance && actualTTL <= expectedTTL+ttlTolerance
						}, "TTL %f not within range [%f, %f]", actualTTL, expectedTTL-ttlTolerance, expectedTTL+ttlTolerance)
					}
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}

			// Cleanup
			DeleteKey(t, conn, exec, tc.cleanupKey)
		})
	}
}
