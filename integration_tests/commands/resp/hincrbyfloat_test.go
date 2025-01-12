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

package resp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHINCRBYFLOAT(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	FireCommand(conn, "FLUSHDB")

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
		delays []time.Duration
	}{
		{
			name:   "HINCRBYFLOAT on non-existing key",
			cmds:   []string{"HINCRBYFLOAT key_hincrfloat field1 10.1"},
			expect: []interface{}{"10.1"},
			delays: []time.Duration{0},
		},
		{
			name:   "HINCRBYFLOAT on existing key",
			cmds:   []string{"HINCRBYFLOAT key_hincrfloat field1 10.5"},
			expect: []interface{}{"20.6"},
			delays: []time.Duration{0},
		},
		{
			name:   "HINCRBYFLOAT on non-float or non-integer value",
			cmds:   []string{"HSET keys field value", "HINCRBYFLOAT keys field 1.2"},
			expect: []interface{}{int64(1), "ERR value is not an integer or a float"},
			delays: []time.Duration{0, 0},
		},
		{
			name:   "HINCRBYFLOAT on non-hashmap key",
			cmds:   []string{"SET key value", "HINCRBYFLOAT key value 10"},
			expect: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			delays: []time.Duration{0, 0},
		},
		{
			name:   "HINCRBYFLOAT using a non integer / non-float value",
			cmds:   []string{"HINCRBYFLOAT key value new"},
			expect: []interface{}{"ERR value is not an integer or a float"},
			delays: []time.Duration{0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}
