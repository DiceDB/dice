// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"
)

func TestGet(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "Get with expiration",
			commands: []string{"SET k v", "GET k"},
			expected: []interface{}{"OK", "v"},
		},
		{
			name:     "Get with non existent key",
			commands: []string{"GET nek"},
			expected: []interface{}{nil},
		},
	}

	runTestcases(t, client, testCases)
}
