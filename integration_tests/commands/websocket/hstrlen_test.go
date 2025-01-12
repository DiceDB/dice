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

package websocket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHSTRLEN(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	DeleteKey(t, conn, exec, "key_hStrLen1")
	DeleteKey(t, conn, exec, "key_hStrLen2")
	DeleteKey(t, conn, exec, "key_hStrLen3")

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
		delays []time.Duration
	}{

		{
			name:   "HSTRLEN with wrong number of args",
			cmds:   []string{"HSTRLEN"},
			expect: []interface{}{"ERR wrong number of arguments for 'hstrlen' command"},
			delays: []time.Duration{0},
		},
		{
			name:   "HSTRLEN with missing field",
			cmds:   []string{"HSET key_hStrLen1 field value", "HSTRLEN key_hStrLen1"},
			expect: []interface{}{float64(1), "ERR wrong number of arguments for 'hstrlen' command"},
			delays: []time.Duration{0, 0},
		},
		{
			name:   "HSTRLEN with non-existent key",
			cmds:   []string{"HSTRLEN non_existent_key field"},
			expect: []interface{}{float64(0)},
			delays: []time.Duration{0},
		},
		{
			name:   "HSTRLEN with non-existent field",
			cmds:   []string{"HSET key_hStrLen2 field value", "HSTRLEN key_hStrLen2 wrong_field"},
			expect: []interface{}{float64(1), float64(0)},
			delays: []time.Duration{0, 0},
		},

		{
			name:   "HSTRLEN with existing key and field",
			cmds:   []string{"HSET key_hStrLen3 field HelloWorld", "HSTRLEN key_hStrLen3 field"},
			expect: []interface{}{float64(1), float64(10)},
			delays: []time.Duration{0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			for i, cmd := range tc.cmds {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}
