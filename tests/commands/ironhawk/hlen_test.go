// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHLEN(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()
	client.FireString("FLUSHDB")

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
		delays []time.Duration
	}{

		{
			name:   "HLEN with wrong number of args",
			cmds:   []string{"HLEN"},
			expect: []interface{}{"ERR wrong number of arguments for 'hlen' command"},
			delays: []time.Duration{0},
		},
		{
			name:   "HLEN with non-existent key",
			cmds:   []string{"HLEN non_existent_key"},
			expect: []interface{}{int64(0)},
			delays: []time.Duration{0},
		},
		{
			name:   "HLEN with non-hash",
			cmds:   []string{"SET string_key string_value", "HLEN string_key"},
			expect: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			delays: []time.Duration{0, 0},
		},
		{
			name:   "HLEN with empty hash",
			cmds:   []string{"HSET key_hLen1 field value", "HDEL key_hLen1 field", "HLEN key_hLen1"},
			expect: []interface{}{int64(1), int64(1), int64(0)},
			delays: []time.Duration{0, 0, 0},
		},
		{
			name:   "HLEN with single field",
			cmds:   []string{"HSET key_hLen2 field1 value1", "HLEN key_hLen2"},
			expect: []interface{}{int64(1), int64(1)},
			delays: []time.Duration{0, 0},
		},
		{
			name:   "HLEN with multiple fields",
			cmds:   []string{"HSET key_hLen3 field1 value1 field2 value2 field3 value3", "HLEN key_hLen3"},
			expect: []interface{}{int64(3), int64(3)},
			delays: []time.Duration{0, 0},
		},
		{
			name:   "HLEN with multiple HSET",
			cmds:   []string{"HSET key_hLen4 field1 value1 field2 value2", "HLEN key_hLen4", "HSET key_hLen4 field3 value3", "HLEN key_hLen4"},
			expect: []interface{}{int64(2), int64(2), int64(1), int64(3)},
			delays: []time.Duration{0, 0, 0, 0},
		},
		{
			name:   "HLEN with HDEL",
			cmds:   []string{"HSET key_hLen5 field1 value1 field2 value2 field3 value3", "HLEN key_hLen5", "HDEL key_hLen5 field3", "HLEN key_hLen5"},
			expect: []interface{}{int64(3), int64(3), int64(1), int64(2)},
			delays: []time.Duration{0, 0, 0, 0},
		},
		{
			name:   "HLEN with DEL",
			cmds:   []string{"HSET key_hLen6 field1 value1 field2 value2 field3 value3", "HLEN key_hLen6", "DEL key_hLen6", "HLEN key_hLen6"},
			expect: []interface{}{int64(3), int64(3), int64(1), int64(0)},
			delays: []time.Duration{0, 0, 0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			for i, cmd := range tc.cmds {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result := client.FireString(cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}
