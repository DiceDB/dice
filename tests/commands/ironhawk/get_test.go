// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"
	"time"
	"errors"
)

func TestGet(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "Get with expiration",
			commands: []string{"SET k v EX 2", "GET k", "GET k"},
			expected: []interface{}{"OK", "v", nil},
			delay:    []time.Duration{0, 0, 2 * time.Second},
		},
		{
			name:     "Get without expiration",
			commands: []string{"SET k v", "GET k"},
			expected: []interface{}{"OK", "v"},
		},
		{
			name:     "Get with non existent key",
			commands: []string{"GET nek"},
			expected: []interface{}{nil},
		},
		{
			name:     "GET with no keys or arguments",
			commands: []string{"GET"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'GET' command"),
			},
		},
	}

	runTestcases(t, client, testCases)
}
