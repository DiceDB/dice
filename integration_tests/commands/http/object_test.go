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

//go:build ignore
// +build ignore

// Ignored as multishard commands not supported by HTTP
package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOBJECT(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name        string
		commands    []HTTPCommand
		expected    []interface{}
		assert_type []string
		delay       []time.Duration
	}{
		{
			name: "Object Idletime",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "OBJECT", Body: map[string]interface{}{"values": []interface{}{"IDLETIME", "foo"}}},
				{Command: "OBJECT", Body: map[string]interface{}{"values": []interface{}{"IDLETIME", "foo"}}},
				{Command: "TOUCH", Body: map[string]interface{}{"keys": []interface{}{"foo"}}},
				{Command: "OBJECT", Body: map[string]interface{}{"values": []interface{}{"IDLETIME", "foo"}}},
			},
			expected:    []interface{}{"OK", float64(2), float64(3), float64(1), float64(0)},
			assert_type: []string{"equal", "assert", "assert", "equal", "assert"},
			delay:       []time.Duration{0, 2 * time.Second, 3 * time.Second, 0, 0},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Ensure key is deleted before the test
			exec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"key": "foo"},
			})

			for i, cmd := range tc.commands {
				if tc.delay[i] > 0 {
					time.Sleep(tc.delay[i])
				}
				result, _ := exec.FireCommand(cmd)
				if tc.assert_type[i] == "equal" {
					assert.Equal(t, tc.expected[i], result)
				} else if tc.assert_type[i] == "assert" {
					assert.True(t, result.(float64) >= tc.expected[i].(float64), "Expected %v to be less than or equal to %v", result, tc.expected[i])
				}
			}
		})
	}
}
