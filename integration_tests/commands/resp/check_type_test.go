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

package resp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// this file may contain test cases for checking error messages across all commands
func TestErrorsForSetData(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	setErrorMsg := "WRONGTYPE Operation against a key holding the wrong kind of value"
	testCases := []struct {
		name       string
		cmd        []string
		expected   []interface{}
		assertType []string
		delay      []time.Duration
	}{
		{
			name:       "GET a key holding a set",
			cmd:        []string{"SADD foo bar", "GET foo"},
			expected:   []interface{}{int64(1), setErrorMsg},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
		{
			name:       "GETDEL a key holding a set",
			cmd:        []string{"SADD foo bar", "GETDEL foo"},
			expected:   []interface{}{int64(1), setErrorMsg},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
		{
			name:       "INCR a key holding a set",
			cmd:        []string{"SADD foo bar", "INCR foo"},
			expected:   []interface{}{int64(1), setErrorMsg},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
		{
			name:       "DECR a key holding a set",
			cmd:        []string{"SADD foo bar", "DECR foo"},
			expected:   []interface{}{int64(1), setErrorMsg},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
		{
			name:       "BIT operations on a key holding a set",
			cmd:        []string{"SADD foo bar", "GETBIT foo 1", "BITCOUNT foo"},
			expected:   []interface{}{int64(1), setErrorMsg, setErrorMsg},
			assertType: []string{"equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0},
		},
		{
			name:       "GETEX a key holding a set",
			cmd:        []string{"SADD foo bar", "GETEX foo"},
			expected:   []interface{}{int64(1), setErrorMsg},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
		{
			name:       "GETSET a key holding a set",
			cmd:        []string{"SADD foo bar", "GETSET foo bar"},
			expected:   []interface{}{int64(1), setErrorMsg},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
		{
			name:       "LPUSH, LPOP, RPUSH, RPOP a key holding a set",
			cmd:        []string{"SADD foo bar", "LPUSH foo bar", "LPOP foo", "RPUSH foo bar", "RPOP foo"},
			expected:   []interface{}{int64(1), setErrorMsg, setErrorMsg, setErrorMsg, setErrorMsg},
			assertType: []string{"equal", "equal", "equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Delete the key before running the test
			FireCommand(conn, "DEL foo")
			for i, cmd := range tc.cmd {
				if tc.delay[i] > 0 {
					time.Sleep(tc.delay[i])
				}
				res := FireCommand(conn, cmd)
				if tc.assertType[i] == "equal" {
					assert.Equal(t, res, tc.expected[i])
				}
			}
		})
	}
}
