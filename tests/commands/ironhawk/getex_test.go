// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"strconv"
	"testing"
	"time"
)

func TestGETEX(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()
	Etime5 := strconv.FormatInt(time.Now().Unix()+5, 10)
	Etime10 := strconv.FormatInt(time.Now().Unix()+10, 10)

	testCases := []TestCase{
		{
			name:     "GETEX Simple Value",
			commands: []string{"SET foo bar", "GETEX foo", "GETEX foo", "TTL foo"},
			expected: []interface{}{"OK", "bar", "bar", -1},
			delay:    []time.Duration{0, 0, 0, 0},
		},
		{
			name:     "GETEX Non-Existent Key",
			commands: []string{"GETEX nonExecFoo", "TTL nonExecFoo"},
			expected: []interface{}{nil, -2},
			delay:    []time.Duration{0, 0},
		},
		{
			name:     "GETEX with EX option",
			commands: []string{"SET foo bar", "GETEX foo", "TTL foo", "GETEX foo ex 2", "TTL foo", "GETEX foo", "TTL foo"},
			expected: []interface{}{"OK", "bar", -1, "bar", 2, nil, -2},
			delay:    []time.Duration{0, 0, 0, 0, 0, 2 * time.Second, 0},
		},
		{
			name:     "GETEX with PX option",
			commands: []string{"SET foo bar", "GETEX foo", "TTL foo", "GETEX foo px 2000", "TTL foo", "GETEX foo", "TTL foo"},
			expected: []interface{}{"OK", "bar", -1, "bar", 2, nil, -2},
			delay:    []time.Duration{0, 0, 0, 0, 0, 2 * time.Second, 0},
		},
		{
			name:     "GETEX with EX option and invalid value",
			commands: []string{"SET foo bar", "GETEX foo", "TTL foo", "GETEX foo ex -1", "TTL foo", "GETEX foo", "TTL foo"},
			expected: []interface{}{"OK", "bar", -1, "ERR invalid expire time in 'getex' command", -1, "bar", -1},
			delay:    []time.Duration{0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:     "GETEX with PX option and invalid value",
			commands: []string{"SET foo bar", "GETEX foo", "TTL foo", "GETEX foo px -1", "TTL foo", "GETEX foo", "TTL foo"},
			expected: []interface{}{"OK", "bar", -1, "ERR invalid expire time in 'getex' command", -1, "bar", -1},
			delay:    []time.Duration{0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:     "GETEX with EXAT option",
			commands: []string{"SET foo bar", "GETEX foo", "TTL foo", "GETEX foo exat " + Etime5, "TTL foo", "GETEX foo", "TTL foo"},
			expected: []interface{}{"OK", "bar", -1, "bar", 5, nil, -2},
			delay:    []time.Duration{0, 0, 0, 0, 0, 5 * time.Second, 0},
		},
		{
			name:     "GETEX with PXAT option",
			commands: []string{"SET foo bar", "GETEX foo", "TTL foo", "GETEX foo pxat " + Etime10 + "000", "TTL foo", "GETEX foo", "TTL foo"},
			expected: []interface{}{"OK", "bar", -1, "bar", 5, nil, -2},
			delay:    []time.Duration{0, 0, 0, 0, 0, 5 * time.Second, 0},
		},
		{
			name:     "GETEX with EXAT option and invalid value",
			commands: []string{"SET foo bar", "GETEX foo", "TTL foo", "GETEX foo exat 123123", "GETEX foo", "TTL foo"},
			expected: []interface{}{"OK", "bar", -1, "bar", nil, -2},
			delay:    []time.Duration{0, 0, 0, 0, 0, 0},
		},
		{
			name:     "GETEX with PXAT option and invalid value",
			commands: []string{"SET foo bar", "GETEX foo", "TTL foo", "GETEX foo pxat 123123", "GETEX foo", "TTL foo"},
			expected: []interface{}{"OK", "bar", -1, "bar", nil, -2},
			delay:    []time.Duration{0, 0, 0, 0, 0, 0},
		},
		{
			name:     "GETEX with Persist option",
			commands: []string{"SET foo bar", "GETEX foo ex 2", "TTL foo", "GETEX foo persist", "TTL foo"},
			expected: []interface{}{"OK", "bar", 2, "bar", -1},
			delay:    []time.Duration{0, 0, 0, 0, 0},
		},
		{
			name:     "GETEX with multiple expiry options",
			commands: []string{"SET foo bar", "GETEX foo ex 2 px 123123", "TTL foo", "GETEX foo"},
			expected: []interface{}{"OK", "ERR syntax error", -1, "bar"},
			delay:    []time.Duration{0, 0, 0, 0},
		},
		{
			name:     "GETEX with persist and ex options",
			commands: []string{"SET foo bar", "GETEX foo ex 2", "TTL foo", "GETEX foo persist ex 2", "TTL foo"},
			expected: []interface{}{"OK", "bar", 2, "ERR syntax error", 2},
			delay:    []time.Duration{0, 0, 0, 0, 0},
		},
		{
			name:     "GETEX with persist and px options",
			commands: []string{"SET foo bar", "GETEX foo px 2000", "TTL foo", "GETEX foo px 2000 persist", "TTL foo"},
			expected: []interface{}{"OK", "bar", 2, "ERR syntax error", 2},
			delay:    []time.Duration{0, 0, 0, 0, 0},
		},
	}

	runTestcases(t, client, testCases)
}
