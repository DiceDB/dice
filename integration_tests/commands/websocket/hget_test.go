// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package websocket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHGET(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	DeleteKey(t, conn, exec, "key_hGet1")
	DeleteKey(t, conn, exec, "key_hGet2")
	DeleteKey(t, conn, exec, "key_hGet3")
	DeleteKey(t, conn, exec, "key_hGet4")
	DeleteKey(t, conn, exec, "string_key")

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
		delays []time.Duration
	}{
		{
			name: "HGET with wrong number of arguments",
			cmds: []string{
				"HGET",
				"HGET key_hGet1",
			},
			expect: []interface{}{
				"ERR wrong number of arguments for 'hget' command",
				"ERR wrong number of arguments for 'hget' command"},
			delays: []time.Duration{0, 0},
		},
		{
			name: "HGET on existent hash",
			cmds: []string{
				"HSET key_hGet2 field1 value1 field2 value2 field3 value3",
				"HGET key_hGet2 field2",
			},
			expect: []interface{}{float64(3), "value2"},
			delays: []time.Duration{0, 0},
		},
		{
			name: "HGET on non-existent field",
			cmds: []string{
				"HSET key_hGet3 field1 value1 field2 value2",
				"HGET key_hGet3 field2",
				"HDEL key_hGet3 field2",
				"HGET key_hGet3 field2",
				"HGET key_hGet3 field3",
			},
			expect: []interface{}{float64(2), "value2", float64(1), nil, nil},
			delays: []time.Duration{0, 0, 0, 0, 0},
		},
		{
			name: "HGET on non-existent hash",
			cmds: []string{
				"HSET key_hGet4 field1 value1 field2 value2",
				"HGET wrong_key_hGet4 field2",
			},
			expect: []interface{}{float64(2), nil},
			delays: []time.Duration{0, 0},
		},
		{
			name: "HGET with wrong type",
			cmds: []string{
				"SET string_key value",
				"HGET string_key field",
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
