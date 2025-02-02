// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package websocket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHSET(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	DeleteKey(t, conn, exec, "key_hSet1")
	DeleteKey(t, conn, exec, "key_hSet2")
	DeleteKey(t, conn, exec, "key_hSet3")
	DeleteKey(t, conn, exec, "key_hSet4")
	DeleteKey(t, conn, exec, "string_key")

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
		delays []time.Duration
	}{
		{
			name: "HSET with wrong number of arguments",
			cmds: []string{
				"HSET",
				"HSET key_hSet1",
			},
			expect: []interface{}{
				"ERR wrong number of arguments for 'hset' command",
				"ERR wrong number of arguments for 'hset' command"},
			delays: []time.Duration{0, 0},
		},
		{
			name: "HSET with single field",
			cmds: []string{
				"HSET key_hSet2 field1 value1",
				"HLEN key_hSet2",
			},
			expect: []interface{}{float64(1), float64(1)},
			delays: []time.Duration{0, 0},
		},
		{
			name: "HSET with multiple fields",
			cmds: []string{
				"HSET key_hSet3 field1 value1 field2 value2 field3 value3",
				"HLEN key_hSet3",
			},
			expect: []interface{}{float64(3), float64(3)},
			delays: []time.Duration{0, 0},
		},
		{
			name: "HSET on existing hash",
			cmds: []string{
				"HSET key_hSet4 field1 value1 field2 value2",
				"HGET key_hSet4 field2",
				"HSET key_hSet4 field2 newvalue2",
				"HGET key_hSet4 field2",
			},
			expect: []interface{}{float64(2), "value2", float64(0), "newvalue2"},
			delays: []time.Duration{0, 0, 0, 0},
		},
		{
			name: "HSET with wrong type",
			cmds: []string{
				"SET string_key value",
				"HSET string_key field value",
			},
			expect: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			delays: []time.Duration{0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			for i, cmd := range tc.cmds {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}
