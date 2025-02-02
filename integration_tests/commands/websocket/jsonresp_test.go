// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package websocket

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONRESP(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	defer conn.Close()

	DeleteKey(t, conn, exec, "k1")
	DeleteKey(t, conn, exec, "k2")
	DeleteKey(t, conn, exec, "k3")

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
			expected: []interface{}{"OK", []interface{}([]interface{}{"[", "dice", float64(10), float64(10.5), true, nil})},
		},
		{
			name: "jsonresp on nested array with mixed types",
			commands: []string{
				`JSON.SET k2 $ {"b":["dice",10,10.5,true,null]}`,
				"JSON.RESP k2 $.b",
			},
			expected: []interface{}{"OK", []interface{}([]interface{}{[]interface{}{"[", "dice", float64(10), float64(10.5), true, nil}})},
		},
		{
			name: "jsonresp on object at root path",
			commands: []string{
				`JSON.SET k3 $ {"b":["dice",10,10.5,true,null]}`,
				"JSON.RESP k3",
			},
			expected: []interface{}{"OK", []interface{}([]interface{}{"{", "b", []interface{}{"[", "dice", float64(10), float64(10.5), true, nil}})},
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
