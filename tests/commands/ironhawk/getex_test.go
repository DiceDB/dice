// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"strconv"
	"testing"
	"time"
)

func TestGETEX(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "GETEX Simple Value",
			commands: []string{"SET foo bar", "GETEX foo", "GETEX foo", "TTL foo"},
			expected: []interface{}{"OK", "bar", "bar", -1},
		},
		{
			name:     "GETEX Non-Existent Key",
			commands: []string{"GETEX nonExecFoo", "TTL nonExecFoo"},
			expected: []interface{}{nil, -2},
		},
		{
			name:     "GETEX with EX option",
			commands: []string{"SET foo bar", "GETEX foo EX 2", "TTL foo", "GET foo", "GET foo"},
			expected: []interface{}{"OK", "bar", 2, "bar", nil},
			delay:    []time.Duration{0, 0, 0, 0, 3 * time.Second},
		},
		{
			name:     "GETEX with PX option",
			commands: []string{"SET foo bar", "GETEX foo PX 2000", "TTL foo", "GET foo", "GET foo"},
			expected: []interface{}{"OK", "bar", 2, "bar", nil},
			delay:    []time.Duration{0, 0, 0, 0, 3 * time.Second},
		},
		{
			name:     "GETEX with PERSIST option",
			commands: []string{"SET foo bar EX 100", "TTL foo", "GETEX foo PERSIST", "TTL foo"},
			expected: []interface{}{"OK", 100, "bar", -1},
		},

		{
			name:     "GETEX with invalid option",
			commands: []string{"SET foo bar", "GETEX foo INVALID"},
			expected: []interface{}{"OK", "bar"},
		},
		{
			name:     "GETEX with negative expiry",
			commands: []string{"SET foo bar", "GETEX foo EX -1"},
			expected: []interface{}{
				"OK",
				errors.New("invalid value for a parameter in 'GETEX' command for EX parameter"),
			},
		},
		{
			name:     "GETEX with zero expiry",
			commands: []string{"SET foo bar", "GETEX foo PX 0"},
			expected: []interface{}{
				"OK",
				errors.New("invalid value for a parameter in 'GETEX' command for PX parameter"),
			},
		},
		{
			name:     "GETEX with non-numeric expiry",
			commands: []string{"SET foo bar", "GETEX foo EX abc"},
			expected: []interface{}{
				"OK",
				errors.New("invalid value for a parameter in 'GETEX' command for EX parameter"),
			},
		},
		{
			name:     "GETEX with PXAT option",
			commands: []string{
				"SET foo bar",
				"GETEX foo PXAT " + strconv.FormatInt(time.Now().Add(24*time.Hour).UnixMilli(), 10),
			},
			expected: []interface{}{"OK", "bar", 2, "bar", nil},
			delay:    []time.Duration{0, 0, 0, 0, 3 * time.Second},
		},
		{
			name:     "GETEX with past EXAT timestamp",
			commands: []string{"SET foo bar", "GETEX foo EXAT " + strconv.FormatInt(time.Now().Add(-1*time.Hour).Unix(), 10)},
			expected: []interface{}{
				"OK",
				errors.New("invalid value for a parameter in 'GETEX' command for EXAT parameter"),
			},
		},
	}

	runTestcases(t, client, testCases)
}
