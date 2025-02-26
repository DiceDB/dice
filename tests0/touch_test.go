// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"
	"time"
)

func TestTouch(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "Touch Simple Value",
			commands: []string{"SET foo bar", "OBJECT IDLETIME foo", "TOUCH foo", "OBJECT IDLETIME foo"},
			expected: []interface{}{"OK", int64(2), int64(1), int64(0)},
			delay:    []time.Duration{0, 2 * time.Second, 0, 0},
		},
		{
			name:     "Touch Multiple Existing Keys",
			commands: []string{"SET foo bar", "SET foo1 bar", "TOUCH foo foo1"},
			expected: []interface{}{"OK", "OK", int64(2)},
		},
		{
			name:     "Touch Multiple Existing and Non-Existing Keys",
			commands: []string{"SET foo bar", "TOUCH foo foo1"},
			expected: []interface{}{"OK", int64(1)},
		},
	}

	runTestcases(t, client, testCases)
}
