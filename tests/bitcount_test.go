package tests

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestBitSet(t *testing.T) {
	conn := getLocalConnection()
	for _, tcase := range []DTestCase{
		{
			InCmds: []string{"SETBIT mykey 7 1"},
			Out:    []interface{}{int64(0)},
		},
		{
			InCmds: []string{"SETBIT mykey 7 1"},
			Out:    []interface{}{int64(1)},
		},
		{
			InCmds: []string{"SETBIT mykey 122 1"},
			Out:    []interface{}{int64(0)},
		},
		{
			InCmds: []string{"GETBIT mykey 122"},
			Out:    []interface{}{int64(1)},
		},
		{
			InCmds: []string{"SETBIT mykey 122 0"},
			Out:    []interface{}{int64(1)},
		},
		{
			InCmds: []string{"GETBIT mykey 122"},
			Out:    []interface{}{int64(0)},
		},
		{
			InCmds: []string{"GETBIT mykey 1223232"},
			Out:    []interface{}{int64(0)},
		},
		{
			InCmds: []string{"GETBIT mykey 7"},
			Out:    []interface{}{int64(1)},
		},
		{
			InCmds: []string{"GETBIT mykey 8"},
			Out:    []interface{}{int64(0)},
		},
		{
			InCmds: []string{"BITCOUNT mykey 3 7 BIT"},
			Out:    []interface{}{int64(1)},
		},
		{
			InCmds: []string{"BITCOUNT mykey 3 7"},
			Out:    []interface{}{int64(0)},
		},
		{
			InCmds: []string{"BITCOUNT mykey 0 0"},
			Out:    []interface{}{int64(1)},
		},
	} {
		for i := 0; i < len(tcase.InCmds); i++ {
			cmd := tcase.InCmds[i]
			out := tcase.Out[i]
			assert.Equal(t, out, fireCommand(conn, cmd), "Value mismatch for cmd %s\n.", cmd)
		}
	}
}

func TestBitCount(t *testing.T) {
	conn := getLocalConnection()
	for _, tcase := range []DTestCase{
		{
			InCmds: []string{"SET k foobar", "GET k"},
			Out:    []interface{}{"OK", "foobar"},
		},
		{
			InCmds: []string{"BITCOUNT k"},
			Out:    []interface{}{int64(26)},
		},
		{
			InCmds: []string{"BITCOUNT k 1 3 BYTE"},
			Out:    []interface{}{int64(15)},
		},
		{
			InCmds: []string{"BITCOUNT k 0 3 BYTE"},
			Out:    []interface{}{int64(19)},
		},
		{
			InCmds: []string{"BITCOUNT k 1 5 BYTE"},
			Out:    []interface{}{int64(22)},
		},
		{
			InCmds: []string{"BITCOUNT k 0 15 BIT"},
			Out:    []interface{}{int64(10)},
		},
		{
			InCmds: []string{"BITCOUNT k 0 23 BIT"},
			Out:    []interface{}{int64(16)},
		},
		{
			InCmds: []string{"BITCOUNT k 5 30 BIT"},
			Out:    []interface{}{int64(17)},
		},
	} {
		for i := 0; i < len(tcase.InCmds); i++ {
			cmd := tcase.InCmds[i]
			out := tcase.Out[i]
			assert.Equal(t, out, fireCommand(conn, cmd), "Value mismatch for cmd %s\n.", cmd)
		}
	}
}
