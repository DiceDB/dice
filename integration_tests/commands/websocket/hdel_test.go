// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package websocket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHDEL(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	DeleteKey(t, conn, exec, "key_hDel1")
	DeleteKey(t, conn, exec, "key_hDel2")
	DeleteKey(t, conn, exec, "key_hDel3")
	DeleteKey(t, conn, exec, "key_hDel4")
	DeleteKey(t, conn, exec, "key_hDel5")
	DeleteKey(t, conn, exec, "string_key")

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
		delays []time.Duration
	}{
		{
			name: "HDEL with wrong number of arguments",
			cmds: []string{
				"HDEL",
				"HDEL key_hDel1",
			},
			expect: []interface{}{
				"ERR wrong number of arguments for 'hdel' command",
				"ERR wrong number of arguments for 'hdel' command"},
			delays: []time.Duration{0, 0},
		},
		{
			name: "HDEL with single field",
			cmds: []string{
				"HSET key_hDel2 field1 value1",
				"HLEN key_hDel2",
				"HDEL key_hDel2 field1",
				"HLEN key_hDel2",
			},
			expect: []interface{}{float64(1), float64(1), float64(1), float64(0)},
			delays: []time.Duration{0, 0, 0, 0},
		},
		{
			name: "HDEL with multiple fields",
			cmds: []string{
				"HSET key_hDel3 field1 value1 field2 value2 field3 value3 field4 value4",
				"HLEN key_hDel3",
				"HDEL key_hDel3 field1 field2",
				"HLEN key_hDel3",
			},
			expect: []interface{}{float64(4), float64(4), float64(2), float64(2)},
			delays: []time.Duration{0, 0, 0, 0},
		},
		{
			name: "HDEL on non-existent field",
			cmds: []string{
				"HSET key_hDel4 field1 value1 field2 value2",
				"HDEL key_hDel4 field3",
			},
			expect: []interface{}{float64(2), float64(0)},
			delays: []time.Duration{0, 0},
		},
		{
			name: "HDEL on non-existent hash",
			cmds: []string{
				"HSET key_hDel5 field1 value1 field2 value2",
				"HDEL wrong_key_hDel5 field1",
			},
			expect: []interface{}{float64(2), float64(0)},
			delays: []time.Duration{0, 0},
		},
		{
			name: "HDEL with wrong type",
			cmds: []string{
				"SET string_key value",
				"HDEL string_key field",
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
