// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONDEBUG(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	client.FireString("DEL k1 k2 k3 k4 k5 k6 k7")

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name: "jsondebug with no path",
			commands: []string{
				`JSON.SET k1 $ {"a":1}`,
				"JSON.DEBUG MEMORY k1",
			},
			expected: []interface{}{"OK", int64(72)},
		},
		{
			name: "jsondebug with a valid path",
			commands: []string{
				`JSON.SET k2 $ {"a":1,"b":2}`,
				"JSON.DEBUG MEMORY k2 $.a",
			},
			expected: []interface{}{"OK", []interface{}{int64(16)}},
		},
		{
			name: "jsondebug with multiple paths",
			commands: []string{
				`JSON.SET k3 $ {"a":1,"b":"dice"}`,
				"JSON.DEBUG MEMORY k3 $.a $.b",
			},
			expected: []interface{}{"OK", []interface{}{int64(16)}},
		},
		{
			name: "jsondebug with multiple paths",
			commands: []string{
				`JSON.SET k4 $ {"a":1,"b":"dice"}`,
				"JSON.DEBUG MEMORY k4 $.a $.b",
			},
			expected: []interface{}{"OK", []interface{}{int64(16)}},
		},
		{
			name: "jsondebug with single path for array json",
			commands: []string{
				`JSON.SET k5 $ ["roll","the","dices"]`,
				"JSON.DEBUG MEMORY k5 $[1]",
			},
			expected: []interface{}{"OK", []interface{}{int64(19)}},
		},
		{
			name: "jsondebug with multiple paths for array json",
			commands: []string{
				`JSON.SET k6 $ ["roll","the","dices"]`,
				"JSON.DEBUG MEMORY k6 $[1,2]",
			},
			expected: []interface{}{"OK", []interface{}{int64(19), int64(21)}},
		},
		{
			name: "jsondebug with all paths for array json",
			commands: []string{
				`JSON.SET k7 $ ["roll","the","dices"]`,
				"JSON.DEBUG MEMORY k7 $[:]",
			},
			expected: []interface{}{"OK", []interface{}{int64(20), int64(19), int64(21)}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
	client.FireString("FLUSHDB")
}
