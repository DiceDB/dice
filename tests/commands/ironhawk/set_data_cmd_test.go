// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"sort"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func CustomDeepEqual(t *testing.T, a, b interface{}) {
	if a == nil || b == nil {
		assert.DeepEqual(t, a, b)
	}

	switch a.(type) {
	case []any:
		sort.Slice(a.([]any), func(i, j int) bool {
			return a.([]any)[i].(string) < a.([]any)[j].(string)
		})
		sort.Slice(b.([]any), func(i, j int) bool {
			return b.([]any)[i].(string) < b.([]any)[j].(string)
		})
	}

	assert.DeepEqual(t, a, b)
}

func TestSetDataCommand(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name       string
		cmd        []string
		expected   []interface{}
		assertType []string
		delay      []time.Duration
	}{
		// SADD
		{
			name:       "SADD Simple Value",
			cmd:        []string{"SADD foo bar", "SMEMBERS foo"},
			expected:   []interface{}{int64(1), []any{"bar"}},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
		{
			name:       "SADD Multiple Values",
			cmd:        []string{"SADD foo bar", "SADD foo baz", "SMEMBERS foo"},
			expected:   []interface{}{int64(1), int64(1), []any{"bar", "baz"}},
			assertType: []string{"equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0},
		},
		{
			name:       "SADD Duplicate Values",
			cmd:        []string{"SADD foo bar", "SADD foo bar", "SMEMBERS foo"},
			expected:   []interface{}{int64(1), int64(0), []any{"bar"}},
			assertType: []string{"equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0},
		},
		{
			name:       "SADD Wrong Key Value Type",
			cmd:        []string{"SET foo bar", "SADD foo baz"},
			expected:   []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
		{
			name:       "SADD Multiple add and multiple kind of values",
			cmd:        []string{"SADD foo bar", "SADD foo baz", "SADD foo 1", "SMEMBERS foo"},
			expected:   []interface{}{int64(1), int64(1), int64(1), []any{"bar", "baz", "1"}},
			assertType: []string{"equal", "equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0},
		},
		// SCARD
		{
			name:       "SADD & SCARD",
			cmd:        []string{"SADD foo bar", "SADD foo baz", "SCARD foo"},
			expected:   []interface{}{int64(1), int64(1), int64(2)},
			assertType: []string{"equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0},
		},
		{
			name:       "SADD & CARD with non existing key",
			cmd:        []string{"SADD foo bar", "SADD foo baz", "SCARD bar"},
			expected:   []interface{}{int64(1), int64(1), int64(0)},
			assertType: []string{"equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0},
		},
		{
			name:       "SADD & SCARD with wrong key type",
			cmd:        []string{"SET foo bar", "SCARD foo"},
			expected:   []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
		// SMEMBERS
		{
			name:       "SADD & SMEMBERS",
			cmd:        []string{"SADD foo bar", "SADD foo baz", "SMEMBERS foo"},
			expected:   []interface{}{int64(1), int64(1), []any{"bar", "baz"}},
			assertType: []string{"equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0},
		},
		{
			name:       "SADD & SMEMBERS with non existing key",
			cmd:        []string{"SMEMBERS foo"},
			expected:   []interface{}{[]any{}},
			assertType: []string{"equal"},
			delay:      []time.Duration{0},
		},
		{
			name:       "SADD & SMEMBERS with wrong key type",
			cmd:        []string{"SET foo bar", "SMEMBERS foo"},
			expected:   []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
		// SREM
		{
			name:       "SADD & SREM",
			cmd:        []string{"SADD foo bar", "SADD foo baz", "SREM foo bar", "SMEMBERS foo"},
			expected:   []interface{}{int64(1), int64(1), int64(1), []any{"baz"}},
			assertType: []string{"equal", "equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0},
		},
		{
			name:       "SADD & SREM with non existing key",
			cmd:        []string{"SREM foo bar"},
			expected:   []interface{}{int64(0)},
			assertType: []string{"equal"},
			delay:      []time.Duration{0},
		},
		{
			name:       "SADD & SREM with wrong key type",
			cmd:        []string{"SET foo bar", "SREM foo bar"},
			expected:   []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
		{
			name:       "SADD & SREM with non existing value",
			cmd:        []string{"SADD foo bar baz bax", "SMEMBERS foo", "SREM foo bat", "SMEMBERS foo"},
			expected:   []interface{}{int64(3), []any{"bar", "baz", "bax"}, int64(0), []any{"bar", "baz", "bax"}},
			assertType: []string{"equal", "equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client.FireString("DEL foo")
			client.FireString("DEL foo2")
			for i, cmd := range tc.cmd {
				if tc.delay[i] > 0 {
					time.Sleep(tc.delay[i])
				}
				result := client.FireString(cmd)
				CustomDeepEqual(t, result, tc.expected[i])
			}
		})
	}

}
