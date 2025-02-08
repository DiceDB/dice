// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHSCAN(t *testing.T) {
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
			name:   "HSCAN with wrong number of args",
			cmds:   []string{"HSCAN key"},
			expect: []interface{}{"ERR wrong number of arguments for 'hscan' command"},
			delays: []time.Duration{0},
		},
		{
			name:   "HSCAN with non-existent key",
			cmds:   []string{"HSCAN non_existent_key 5"},
			expect: []interface{}{[]interface{}{"0", []interface{}{}}},
			delays: []time.Duration{0},
		},
		{
			name: "HSCAN with non-hash",
			cmds: []string{"SET string_key string_value",
				"HSCAN string_key 0"},
			expect: []interface{}{"OK",
				"WRONGTYPE Operation against a key holding the wrong kind of value"},
			delays: []time.Duration{0, 0},
		},
		{
			name:   "HSCAN with empty hash",
			cmds:   []string{"HSCAN empty_hash 0"},
			expect: []interface{}{[]interface{}{"0", []interface{}{}}},
			delays: []time.Duration{0},
		},
		{
			name: "HSCAN with valid key and cursor",
			cmds: []string{"HSET key_hScan0 field1 value1 field2 value2",
				"HSCAN key_hScan0 0"},
			expect: []interface{}{int64(2),
				[]interface{}{"0", []interface{}{"field1", "value1", "field2", "value2"}}},
			delays: []time.Duration{0, 0},
		},
		{
			name: "HSCAN with cursor at the end",
			cmds: []string{"HSET key_hScan1 field1 value1 field2 value2",
				"HSCAN key_hScan1 2"},
			expect: []interface{}{int64(2),
				[]interface{}{"0", []interface{}{}}},
			delays: []time.Duration{0, 0},
		},
		{
			name: "HSCAN with cursor at the beginning",
			cmds: []string{"HSET key_hScan2 field1 value1 field2 value2",
				"HSCAN key_hScan2 0"},
			expect: []interface{}{int64(2),
				[]interface{}{"0", []interface{}{"field1", "value1", "field2", "value2"}}},
			delays: []time.Duration{0, 0},
		},
		{
			name: "HSCAN with cursor in the middle",
			cmds: []string{"HSET key_hScan3 field1 value1 field2 value2",
				"HSCAN key_hScan3 1"},
			expect: []interface{}{int64(2),
				[]interface{}{"0", []interface{}{"field2", "value2"}}},
			delays: []time.Duration{0, 0},
		},
		{
			name: "HSCAN with MATCH argument",
			cmds: []string{"HSET key_hScan4 field1 value1 field2 value2 field3 value3",
				"HSCAN key_hScan4 0 MATCH field[12]*"},
			expect: []interface{}{int64(3),
				[]interface{}{"0", []interface{}{"field1", "value1", "field2", "value2"}}},
			delays: []time.Duration{0, 0},
		},
		{
			name: "HSCAN with COUNT argument",
			cmds: []string{"HSET key_hScan5 field1 value1 field2 value2 field3 value3 field4 value4",
				"HSCAN key_hScan5 0 COUNT 2",
				"HSCAN key_hScan5 2 COUNT 2"},
			expect: []interface{}{int64(4),
				[]interface{}{"2", []interface{}{"field1", "value1", "field2", "value2"}},
				[]interface{}{"0", []interface{}{"field3", "value3", "field4", "value4"}}},
			delays: []time.Duration{0, 0, 0},
		},
		{
			name: "HSCAN with MATCH and COUNT arguments",
			cmds: []string{"HSET key_hScan6 field1 value1 field2 value2 field3 value3 field4 value4",
				"HSCAN key_hScan6 0 MATCH field[13]* COUNT 1"},
			expect: []interface{}{int64(4),
				[]interface{}{"1", []interface{}{"field1", "value1"}}},
			delays: []time.Duration{0, 0},
		},
		{
			name: "HSCAN with invalid MATCH pattern",
			cmds: []string{"HSET key_hScan7 field1 value1 field2 value2",
				"HSCAN key_hScan7 0 MATCH [invalid"},
			expect: []interface{}{int64(2),
				"ERR Invalid glob pattern: unexpected end of input"},
			delays: []time.Duration{0, 0},
		},
		{
			name: "HSCAN with invalid COUNT value",
			cmds: []string{"HSET key_hScan8 field1 value1 field2 value2",
				"HSCAN key_hScan8 0 COUNT invalid"},
			expect: []interface{}{int64(2),
				"ERR value is not an integer or out of range"},
			delays: []time.Duration{0, 0},
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
