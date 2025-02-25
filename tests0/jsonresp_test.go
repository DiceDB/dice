// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONRESP(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	client.FireString("DEL k1 k2 k3")

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name: "jsonresp on array with mixed types",
			commands: []string{
				`JSON.SET k1 $ ["dice",10,10.5,true,null]`,
				"JSON.RESP k1 $",
			},
			expected: []interface{}{"OK", []interface{}([]interface{}{"[", "dice", int64(10), "10.5", "true", "(nil)"})},
		},
		{
			name: "jsonresp on nested array with mixed types",
			commands: []string{
				`JSON.SET k2 $ {"b":["dice",10,10.5,true,null]}`,
				"JSON.RESP k2 $.b",
			},
			expected: []interface{}{"OK", []interface{}([]interface{}{[]interface{}{"[", "dice", int64(10), "10.5", "true", "(nil)"}})},
		},
		{
			name: "jsonresp on object at root path",
			commands: []string{
				`JSON.SET k3 $ {"b":["dice",10,10.5,true,null]}`,
				"JSON.RESP k3",
			},
			expected: []interface{}{"OK", []interface{}([]interface{}{"{", "b", []interface{}{"[", "dice", int64(10), "10.5", "true", "(nil)"}})},
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
