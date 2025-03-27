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
		{
			name:     "KEYS with single character wildcard",
			commands: []string{"SET k1 v1", "SET k2 v2", "SET ka va", "KEYS k?"},
			expected: []interface{}{"OK", "OK", "OK", []interface{}{"k1", "k2", "ka"}},
		},
		{
			name:     "KEYS with single matching key",
			commands: []string{"SET unique_key value", "KEYS unique*"},
			expected: []interface{}{"OK", []interface{}{"unique_key"}},
		},
	}

	runTestcases(t, client, testCases)
}
