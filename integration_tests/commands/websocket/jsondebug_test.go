// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package websocket

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONDEBUG(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	defer conn.Close()

	DeleteKey(t, conn, exec, "k1")
	DeleteKey(t, conn, exec, "k2")
	DeleteKey(t, conn, exec, "k3")
	DeleteKey(t, conn, exec, "k4")
	DeleteKey(t, conn, exec, "k5")
	DeleteKey(t, conn, exec, "k6")
	DeleteKey(t, conn, exec, "k7")

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
			expected: []interface{}{"OK", float64(72)},
		},
		{
			name: "jsondebug with a valid path",
			commands: []string{
				`JSON.SET k2 $ {"a":1,"b":2}`,
				"JSON.DEBUG MEMORY k2 $.a",
			},
			expected: []interface{}{"OK", []interface{}{float64(16)}},
		},
		{
			name: "jsondebug with multiple paths",
			commands: []string{
				`JSON.SET k3 $ {"a":1,"b":"dice"}`,
				"JSON.DEBUG MEMORY k3 $.a $.b",
			},
			expected: []interface{}{"OK", []interface{}{float64(16)}},
		},
		{
			name: "jsondebug with multiple paths",
			commands: []string{
				`JSON.SET k4 $ {"a":1,"b":"dice"}`,
				"JSON.DEBUG MEMORY k4 $.a $.b",
			},
			expected: []interface{}{"OK", []interface{}{float64(16)}},
		},
		{
			name: "jsondebug with single path for array json",
			commands: []string{
				`JSON.SET k5 $ ["roll","the","dices"]`,
				"JSON.DEBUG MEMORY k5 $[1]",
			},
			expected: []interface{}{"OK", []interface{}{float64(19)}},
		},
		{
			name: "jsondebug with multiple paths for array json",
			commands: []string{
				`JSON.SET k6 $ ["roll","the","dices"]`,
				"JSON.DEBUG MEMORY k6 $[1,2]",
			},
			expected: []interface{}{"OK", []interface{}{float64(19), float64(21)}},
		},
		{
			name: "jsondebug with all paths for array json",
			commands: []string{
				`JSON.SET k7 $ ["roll","the","dices"]`,
				"JSON.DEBUG MEMORY k7 $[:]",
			},
			expected: []interface{}{"OK", []interface{}{float64(20), float64(19), float64(21)}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
	exec.FireCommandAndReadResponse(conn, "FLUSHDB")
}
