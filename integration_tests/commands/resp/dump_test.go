// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package resp

import (
	"testing"

	"github.com/dicedb/dice/testutils"
	"github.com/stretchr/testify/assert"
)

func TestDumpRestore(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	simpleJSON := `{"name":"John","age":30}`
	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name: "DUMP and RESTORE string value",
			commands: []string{
				"SET foo bar",
				"DUMP foo",
				"DEL foo",
				"RESTORE foo 2 CQAAAAADYmFy/72GUVF+ClKv",
				"GET foo",
			},
			expected: []interface{}{
				"OK",
				"CQAAAAADYmFy/72GUVF+ClKv",
				int64(1),
				"OK",
				"bar",
			},
		},
		{
			name: "DUMP and RESTORE integer value",
			commands: []string{
				"set foo 12345",
				"DUMP foo",
				"DEL foo",
				"RESTORE foo 2 CQUAAAAAAAAwOf8OqbusYAl2pQ==",
			},
			expected: []interface{}{
				"OK",
				"CQUAAAAAAAAwOf8OqbusYAl2pQ==",
				int64(1),
				"OK",
			},
		},
		{
			name: "DUMP non-existent key",
			commands: []string{
				"DUMP nonexistentkey",
			},
			expected: []interface{}{
				"(nil)",
			},
		},
		{
			name: "DUMP JSON",
			commands: []string{
				`JSON.SET foo $ ` + simpleJSON,
				"DUMP foo",
				"del foo",
				"restore foo 2 CQMAAAAYeyJhZ2UiOjMwLCJuYW1lIjoiSm9obiJ9/6PVaIgw0n+C",
				"JSON.GET foo $..name",
			},
			expected: []interface{}{
				"OK",
				"skip",
				int64(1),
				"OK",
				`"John"`,
			},
		},
		{
			name: "DUMP Set",
			commands: []string{
				"sadd foo bar baz bazz",
				"dump foo",
				"del foo",
				"restore foo 2 CQYAAAAAAAAAAwAAAANiYXIAAAADYmF6AAAABGJhenr/DSf4vHxjdYo=",
				"smembers foo",
			},
			expected: []interface{}{
				int64(3),
				"skip",
				int64(1),
				"OK",
				[]interface{}{"bar", "baz", "bazz"},
			},
		},
		{
			name: "DUMP bytearray",
			commands: []string{
				"setbit foo 1 1",
				"dump foo",
				"del foo",
				"restore foo 2 CQQAAAAAAAAAAUD/g00L5pRbaJI=",
				"get foo",
			},
			expected: []interface{}{
				int64(0),
				"CQQAAAAAAAAAAUD/g00L5pRbaJI=",
				int64(1),
				"OK",
				"@",
			},
		},
		{
			name: "DUMP sorted set",
			commands: []string{
				"zadd foo 1 bar 2 bazz",
				"dump foo",
				"del foo",
				"restore foo 2 CQgAAAAAAAAAAgAAAAAAAAADYmFyP/AAAAAAAAAAAAAAAAAABGJhenpAAAAAAAAAAP/POOElibTuYQ==",
				"zrange foo 0 2",
			},
			expected: []interface{}{
				int64(2),
				"skip",
				int64(1),
				"OK",
				[]interface{}{"bar", "bazz"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			FireCommand(conn, "del foo")
			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				expected := tc.expected[i]
				if expected == "skip" {
					// when order of elements define the dump value, we test the restore function and skip dump
					continue
				}
				switch exp := expected.(type) {
				case string:
					assert.Equal(t, exp, result)
				case []interface{}:
					assert.True(t, testutils.UnorderedEqual(exp, result))
				default:
					assert.Equal(t, expected, result)
				}
			}
		})
	}
	FireCommand(conn, "FLUSHDB")
}

func TestDumpRestoreBF(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	res := FireCommand(conn, "bf.add foo bar")
	assert.Equal(t, int64(1), res)

	dumpValue := FireCommand(conn, "dump foo")
	FireCommand(conn, "del foo")

	res = FireCommand(conn, "restore foo 0 "+dumpValue.(string))
	assert.Equal(t, "OK", res)
	res = FireCommand(conn, "bf.exists foo bazz")
	assert.Equal(t, int64(0), res)

	FireCommand(conn, "FLUSHDB")
}

func TestDumpRestoreCMS(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	// Add a value to the CMS
	FireCommand(conn, "CMS.INITBYPROB foo 0.1 0.01")
	res := FireCommand(conn, "cms.incrby foo bar 42")
	assert.Equal(t, []interface{}([]interface{}{int64(42)}), res)

	// Dump the serialized value
	dumpValue := FireCommand(conn, "dump foo")
	FireCommand(conn, "del foo") // Delete the CMS

	// Restore the CMS from the dumped value
	res = FireCommand(conn, "restore foo 0 "+dumpValue.(string))
	assert.Equal(t, "OK", res)

	// Check the value for a key in the restored CMS
	res = FireCommand(conn, "cms.query foo bar")
	assert.Equal(t, []interface{}([]interface{}{int64(42)}), res)

	// Check that another key not in the CMS returns 0
	res = FireCommand(conn, "cms.query foo bar")
	assert.Equal(t, []interface{}([]interface{}{int64(42)}), res)

	FireCommand(conn, "FLUSHDB")
}

func TestDumpRestoreDeque(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	res := FireCommand(conn, "lpush foo bar")
	assert.Equal(t, int64(1), res)
	dumpValue := FireCommand(conn, "dump foo")
	res = FireCommand(conn, "del foo")
	assert.Equal(t, int64(1), res)
	res = FireCommand(conn, "restore foo 0 "+dumpValue.(string))
	assert.Equal(t, "OK", res)
	res = FireCommand(conn, "lpop foo")
	assert.Equal(t, "bar", res)

	FireCommand(conn, "FLUSHDB")
}
