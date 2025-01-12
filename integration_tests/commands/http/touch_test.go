// This file is part of DiceDB.
// Copyright (C) 2025  DiceDB (dicedb.io).
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

func TestTouch(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		delay    []time.Duration
	}{
		{
			name: "Touch Simple Value",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "OBJECT", Body: map[string]interface{}{"values": []interface{}{"IDLETIME", "foo"}}},
				{Command: "TOUCH", Body: map[string]interface{}{"key": "foo"}},
				{Command: "OBJECT", Body: map[string]interface{}{"values": []interface{}{"IDLETIME", "foo"}}},
			},
			expected: []interface{}{"OK", float64(2), float64(1), float64(0)},
			delay:    []time.Duration{0, 2 * time.Second, 0, 0},
		},
		// Touch Multiple Existing Keys
		{
			name: "Touch Multiple Existing Keys",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SET", Body: map[string]interface{}{"key": "foo1", "value": "bar"}},
				{Command: "TOUCH", Body: map[string]interface{}{"keys": []interface{}{"foo", "foo1"}}},
			},
			expected: []interface{}{"OK", "OK", float64(2)},
			delay:    []time.Duration{0, 0, 0},
		},
		// Touch Multiple Existing and Non-Existing Keys
		{
			name: "Touch Multiple Existing and Non-Existing Keys",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "TOUCH", Body: map[string]interface{}{"keys": []interface{}{"foo", "foo1"}}},
			},
			expected: []interface{}{"OK", float64(1)},
			delay:    []time.Duration{0, 0},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"key": "foo"},
			})
			exec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"key": "foo1"},
			})

			for i, cmd := range tc.commands {
				if tc.delay[i] > 0 {
					time.Sleep(tc.delay[i])
				}
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}

}
