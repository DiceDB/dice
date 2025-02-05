// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

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
			exp = append(exp, int64(0))
		}
	}

	cmds = append(cmds, "GETRANGE byteArrayKey 0 4")
	exp = append(exp, "hello")

	return cmds, exp
}

func TestGETRANGE(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	client.FireString("FLUSHDB")
	defer client.FireString("FLUSHDB")

	byteArrayCmds, byteArrayExp := generateByteArrayForGetrangeTestCase()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
		cleanup  []string
	}{
		{
			name:     "Get range on a string",
			commands: []string{"SET test1 shankar", "GETRANGE test1 0 7"},
			expected: []interface{}{"OK", "shankar"},
			cleanup:  []string{"del test1"},
		},
		{
			name:     "Get range on a non existent key",
			commands: []string{"GETRANGE test2 0 7"},
			expected: []interface{}{""},
			cleanup:  []string{"del test2"},
		},
		{
			name:     "Get range on wrong key type",
			commands: []string{"LPUSH test3 shankar", "GETRANGE test3 0 7"},
			expected: []interface{}{int64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			cleanup:  []string{"del test3"},
		},
		{
			name:     "GETRANGE against string value: 0, -1",
			commands: []string{"SET test4 apple", "GETRANGE test4 0 -1"},
			expected: []interface{}{"OK", "apple"},
			cleanup:  []string{"del test4"},
		},
		{
			name:     "GETRANGE against string value: 5, 3",
			commands: []string{"SET test5 apple", "GETRANGE test5 5 3"},
			expected: []interface{}{"OK", ""},
			cleanup:  []string{"del test5"},
		},
		{
			name:     "GETRANGE against integer value: -1, -100",
			commands: []string{"SET test6 apple", "GETRANGE test6 -1 -100"},
			expected: []interface{}{"OK", ""},
			cleanup:  []string{"del test6"},
		},
		{
			name:     "GETRANGE against byte array",
			commands: byteArrayCmds,
			expected: byteArrayExp,
			cleanup:  []string{"del byteArrayKey"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			for i := 0; i < len(tc.commands); i++ {
				result := client.FireString(tc.commands[i])
				expected := tc.expected[i]
				assert.Equal(t, expected, result)
			}

			for _, cmd := range tc.cleanup {
				client.FireString(cmd)
			}
		})
	}
}
