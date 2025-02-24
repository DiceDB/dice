// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestINCRBYFLOAT(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()
	invalidArgMessage := "ERR wrong number of arguments for 'incrbyfloat' command"
	invalidIncrTypeMessage := "ERR value is not a valid float"
	valueOutOfRangeMessage := "ERR value is out of range"

	testCases := []struct {
		name      string
		setupData string
		commands  []string
		expected  []interface{}
	}{
		{
			name:      "Invalid number of arguments",
			setupData: "",
			commands:  []string{"INCRBYFLOAT", "INCRBYFLOAT foo"},
			expected:  []interface{}{invalidArgMessage, invalidArgMessage},
		},
		{
			name:      "Increment a non existing key",
			setupData: "",
			commands:  []string{"INCRBYFLOAT foo 0.1", "GET foo"},
			expected:  []interface{}{"0.1", "0.1"},
		},
		{
			name:      "Increment a key with an integer value",
			setupData: "SET foo 1",
			commands:  []string{"INCRBYFLOAT foo 0.1", "GET foo"},
			expected:  []interface{}{"1.1", "1.1"},
		},
		{
			name:      "Increment and then decrement a key with the same value",
			setupData: "SET foo 1",
			commands:  []string{"INCRBYFLOAT foo 0.1", "GET foo", "INCRBYFLOAT foo -0.1", "GET foo"},
			expected:  []interface{}{"1.1", "1.1", "1", "1"},
		},
		{
			name:      "Increment a non numeric value",
			setupData: "SET foo bar",
			commands:  []string{"INCRBYFLOAT foo 0.1"},
			expected:  []interface{}{invalidIncrTypeMessage},
		},
		{
			name:      "Increment by a non numeric value",
			setupData: "SET foo 1",
			commands:  []string{"INCRBYFLOAT foo bar"},
			expected:  []interface{}{invalidIncrTypeMessage},
		},
		{
			name:      "Increment by both integer and float",
			setupData: "SET foo 1",
			commands:  []string{"INCRBYFLOAT foo 1", "INCRBYFLOAT foo 0.1"},
			expected:  []interface{}{"2", "2.1"},
		},
		{
			name:      "Increment that would make the value Inf",
			setupData: "SET foo 1e308",
			commands:  []string{"INCRBYFLOAT foo 1e308", "INCRBYFLOAT foo -1e308"},
			expected:  []interface{}{valueOutOfRangeMessage, "0"},
		},
		{
			name:      "Increment that would make the value -Inf",
			setupData: "SET foo -1e308",
			commands:  []string{"INCRBYFLOAT foo -1e308", "INCRBYFLOAT foo 1e308"},
			expected:  []interface{}{valueOutOfRangeMessage, "0"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer client.FireString("DEL foo")
			if tc.setupData != "" {
				assert.Equal(t, client.FireString(tc.setupData), "OK")
			}
			for i := 0; i < len(tc.commands); i++ {
				cmd := tc.commands[i]
				out := tc.expected[i]
				result := client.FireString(cmd)
				assert.Equal(t, out, result)
			}
		})
	}
}
