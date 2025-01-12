// This file is part of DiceDB.
// Copyright (C) 2025DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package http

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func generateByteArrayForGetrangeTestCase() ([]HTTPCommand, []interface{}) {
	var cmds []HTTPCommand
	var exp []interface{}

	str := "helloworld"
	var binaryStr string

	for _, c := range str {
		binaryStr += fmt.Sprintf("%08b", c)
	}

	for idx, bit := range binaryStr {
		if bit == '1' {
			cmds = append(cmds, HTTPCommand{Command: "SETBIT", Body: map[string]interface{}{"key": "byteArrayKey", "values": []interface{}{idx, 1}}})
			exp = append(exp, float64(0))
		}
	}

	cmds = append(cmds, HTTPCommand{Command: "GETRANGE", Body: map[string]interface{}{"key": "byteArrayKey", "values": []interface{}{0, 4}}})
	exp = append(exp, "hello")

	return cmds, exp
}

func TestGETRANGE(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	byteArrayCmds, byteArrayExp := generateByteArrayForGetrangeTestCase()
	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		cleanup  []HTTPCommand
	}{
		{
			name: "Get range on a string",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "test1", "value": "shankar"}},
				{Command: "GETRANGE", Body: map[string]interface{}{"key": "test1", "values": []interface{}{0, 7}}},
			},
			expected: []interface{}{"OK", "shankar"},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "test1"}},
			},
		},
		{
			name: "Get range on a non existent key",
			commands: []HTTPCommand{
				{Command: "GETRANGE", Body: map[string]interface{}{"key": "test2", "values": []interface{}{0, 7}}},
			},
			expected: []interface{}{""},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "test2"}},
			},
		},
		{
			name: "Get range on wrong key type",
			commands: []HTTPCommand{
				{Command: "LPUSH", Body: map[string]interface{}{"key": "test3", "value": "shankar"}},
				{Command: "GETRANGE", Body: map[string]interface{}{"key": "test3", "values": []interface{}{0, 7}}},
			},
			expected: []interface{}{float64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "test3"}},
			},
		},
		{
			name: "GETRANGE against string value: 0, -1",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "test4", "value": "shankar"}},
				{Command: "GETRANGE", Body: map[string]interface{}{"key": "test4", "values": []interface{}{0, -1}}},
			},
			expected: []interface{}{"OK", "shankar"},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "test4"}},
			},
		},
		{
			name: "GETRANGE against string value: 5, 3",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "test5", "value": "shankar"}},
				{Command: "GETRANGE", Body: map[string]interface{}{"key": "test5", "values": []interface{}{5, 3}}},
			},
			expected: []interface{}{"OK", ""},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "test5"}},
			},
		},
		{
			name:     "GETRANGE against byte array",
			commands: byteArrayCmds,
			expected: byteArrayExp,
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "byteArrayKey"}},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
			exec.FireCommand(tc.cleanup[0])
		})
	}
}
