// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package websocket

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func generateByteArrayForGetrangeTestCase() ([]string, []interface{}) {
	var cmds []string
	var exp []interface{}

	str := "helloworld"
	var binaryStr string

	for _, c := range str {
		binaryStr += fmt.Sprintf("%08b", c)
	}

	for idx, bit := range binaryStr {
		if bit == '1' {
			cmds = append(cmds, string("SETBIT byteArrayKey "+strconv.Itoa(idx)+" 1"))
			exp = append(exp, float64(0))
		}
	}

	cmds = append(cmds, "GETRANGE byteArrayKey 0 4")
	exp = append(exp, "hello")

	return cmds, exp
}

func TestGETRANGE(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	byteArrayCmds, byteArrayExp := generateByteArrayForGetrangeTestCase()
	testCases := []struct {
		name       string
		commands   []string
		expected   []interface{}
		cleanupKey string
	}{
		{
			name:       "Get range on a string",
			commands:   []string{"SET test1 shankar", "GETRANGE test1 0 7"},
			expected:   []interface{}{"OK", "shankar"},
			cleanupKey: "test1",
		},
		{
			name:       "Get range on a non existent key",
			commands:   []string{"GETRANGE test2 0 7"},
			expected:   []interface{}{""},
			cleanupKey: "test2",
		},
		{
			name:       "Get range on wrong key type",
			commands:   []string{"LPUSH test3 shankar", "GETRANGE test3 0 7"},
			expected:   []interface{}{float64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			cleanupKey: "test3",
		},
		{
			name:       "GETRANGE against string value: 0, -1",
			commands:   []string{"SET test4 apple", "GETRANGE test4 0 -1"},
			expected:   []interface{}{"OK", "apple"},
			cleanupKey: "test4",
		},
		{
			name:       "GETRANGE against string value: 5, 3",
			commands:   []string{"SET test5 apple", "GETRANGE test5 5 3"},
			expected:   []interface{}{"OK", ""},
			cleanupKey: "test5",
		},
		{
			name:       "GETRANGE against integer value: -1, -100",
			commands:   []string{"SET test6 apple", "GETRANGE test6 -1 -100"},
			expected:   []interface{}{"OK", ""},
			cleanupKey: "test6",
		},
		{
			name:       "GETRANGE against byte array",
			commands:   byteArrayCmds,
			expected:   byteArrayExp,
			cleanupKey: "byteArrayKey",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conn := exec.ConnectToServer()

			for i, cmd := range tc.commands {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				assert.Equal(t, tc.expected[i], result)
			}
			DeleteKey(t, conn, exec, tc.cleanupKey)
		})
	}
}
