// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"

	"github.com/dicedb/dice/testutils"
	"github.com/stretchr/testify/assert"
)

func TestKeys(t *testing.T) {
	client := getLocalConnection()
	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "k matches with k",
			commands: []string{"SET k v", "KEYS k"},
			expected: []interface{}{"OK", []interface{}{"k"}},
		},
		{
			name:     "g* matches good and great",
			commands: []string{"SET good v", "SET great v", "KEYS g*"},
			expected: []interface{}{"OK", "OK", []interface{}{"good", "great"}},
		},
		{
			name:     "g?od matches good",
			commands: []string{"SET good v", "SET great v", "KEYS g?od"},
			expected: []interface{}{"OK", "OK", []interface{}{"good"}},
		},
		{
			name:     "g?eat matches great",
			commands: []string{"SET good v", "SET great v", "KEYS g?eat"},
			expected: []interface{}{"OK", "OK", []interface{}{"great"}},
		},
		{
			name:     "h[^e]llo matches hallo and hbllo",
			commands: []string{"SET hallo v", "SET hbllo v", "SET hello v", "KEYS h[^e]llo"},
			expected: []interface{}{"OK", "OK", "OK", []interface{}{"hallo", "hbllo"}},
		},

		{
			name:     "h[a-b]llo matches hallo and hbllo",
			commands: []string{"SET hallo v", "SET hbllo v", "SET hello v", "KEYS h[a-b]llo"},
			expected: []interface{}{"OK", "OK", "OK", []interface{}{"hallo", "hbllo"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := client.FireString(cmd)

				// because the order of keys is not guaranteed, we need to check if the result is an array
				if slice, ok := tc.expected[i].([]interface{}); ok {
					assert.True(t, testutils.UnorderedEqual(slice, result))
				} else {
					assert.Equal(t, tc.expected[i], result)
				}

			}
		})
	}
}
