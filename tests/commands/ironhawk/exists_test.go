// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"
	"time"
)

func TestExists(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "Test EXISTS command",
			commands: []string{"SET key value", "EXISTS key", "EXISTS key2"},
			expected: []interface{}{"OK", 1, 0},
		},
		{
			name:     "Test EXISTS command with multiple keys",
			commands: []string{"SET key value", "SET key2 value2", "EXISTS key key2 key3", "EXISTS key key2 key3 key4", "DEL key", "EXISTS key key2 key3 key4"},
			expected: []interface{}{"OK", "OK", 2, 2, 1, 1},
		},
		{
			name:     "Test EXISTS an expired key",
			commands: []string{"SET key value ex 1", "EXISTS key", "EXISTS key"},
			expected: []interface{}{"OK", 1, 0},
			delay:    []time.Duration{0, 0, 2 * time.Second},
		},
		{
			name:     "Test EXISTS with multiple keys and expired key",
			commands: []string{"SET key value ex 2", "SET key2 value2", "SET key3 value3", "EXISTS key key2 key3", "EXISTS key key2 key3"},
			expected: []interface{}{"OK", "OK", "OK", 3, 2},
			delay:    []time.Duration{0, 0, 0, 0, 2 * time.Second},
		},
		{
			name:     "EXISTS with no keys or arguments",
			commands: []string{"EXISTS"},
			expected: []interface{}{0},
		},
		{
			name:     "EXISTS with duplicate keys",
			commands: []string{"SET key value", "EXISTS key key"},
			expected: []interface{}{"OK", 2},
		},
		{
			name:     "EXISTS with duplicate non existent keys",
			commands: []string{"SET key value", "EXISTS key neq neq2"},
			expected: []interface{}{"OK", 1},
		},
		{
			name:     "EXISTS with duplicate keys multiple keys",
			commands: []string{"SET key value", "SET key1 value", "EXISTS key key key1"},
			expected: []interface{}{"OK", "OK", 3},
		},
	}

	runTestcases(t, client, testCases)
}
