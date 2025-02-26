// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"
)

func TestFLUSHDB(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name: "FLUSHDB",
			commands: []string{
				"SET k1 v1",
				"SET k2 v2",
				"SET k3 v3",
				"FLUSHDB",
				"GET k1",
				"GET k2",
				"GET k3",
			},
			expected: []interface{}{"OK", "OK", "OK", "OK", nil, nil, nil},
		},
	}

	runTestcases(t, client, testCases)
}
