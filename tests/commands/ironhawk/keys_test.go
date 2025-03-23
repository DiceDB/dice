// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"
)

func TestKEYS(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "KEYS with more than one matching key",
			commands: []string{"SET k v", "SET k1 v1", "KEYS k*"},
			expected: []interface{}{"OK", "OK", []interface{}{"k", "k1"}},
		},
		{
			name:     "KEYS with no matching keys",
			commands: []string{"KEYS a*"},
			expected: []interface{}{[]interface{}{}},
		},
	}

	runTestcases(t, client, testCases)
}
