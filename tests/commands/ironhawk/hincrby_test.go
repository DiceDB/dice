// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHINCRBY(t *testing.T) {
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
			name:   "HINCRBY on non-existing key",
			cmds:   []string{"HINCRBY key field1 10"},
			expect: []interface{}{int64(10)},
			delays: []time.Duration{0},
		},
		{
			name:   "HINCRBY on existing key",
			cmds:   []string{"HINCRBY key field1 5"},
			expect: []interface{}{int64(15)},
			delays: []time.Duration{0},
		},
		{
			name:   "HINCRBY on non-integer value",
			cmds:   []string{"HSET keys field value", "HINCRBY keys field 1"},
			expect: []interface{}{int64(1), "ERR hash value is not an integer"},
			delays: []time.Duration{0, 0, 0},
		},
		{
			name:   "HINCRBY on non-hashmap key",
			cmds:   []string{"SET key value", "HINCRBY key value 10"},
			expect: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			delays: []time.Duration{0, 0, 0},
		},
		{
			name:   "HINCRBY overflow",
			cmds:   []string{"HSET new-key value 9000000000000000000", "HINCRBY new-key value 1000000000000000000"},
			expect: []interface{}{int64(1), "ERR increment or decrement would overflow"},
			delays: []time.Duration{0, 0, 0},
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
