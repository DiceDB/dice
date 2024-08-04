package tests

import (
	"fmt"
	"net"
	"testing"

	"gotest.tools/v3/assert"
)

func TestBitOp(t *testing.T) {
	conn := getLocalConnection()
	testcases := []DTestCase{
		{
			InCmds: []string{"SETBIT unitTestKeyA 1 1", "SETBIT unitTestKeyA 3 1", "SETBIT unitTestKeyA 5 1", "SETBIT unitTestKeyA 7 1", "SETBIT unitTestKeyA 8 1"},
			Out:    []interface{}{int64(0), int64(0), int64(0), int64(0), int64(0)},
		},
		{
			InCmds: []string{"SETBIT unitTestKeyB 2 1", "SETBIT unitTestKeyB 4 1", "SETBIT unitTestKeyB 7 1"},
			Out:    []interface{}{int64(0), int64(0), int64(0)},
		},
		{
			InCmds: []string{"BITOP NOT unitTestKeyNOT unitTestKeyA "},
			Out:    []interface{}{int64(2)},
		},
		{
			InCmds: []string{"GETBIT unitTestKeyNOT 1", "GETBIT unitTestKeyNOT 2", "GETBIT unitTestKeyNOT 7", "GETBIT unitTestKeyNOT 8", "GETBIT unitTestKeyNOT 9"},
			Out:    []interface{}{int64(0), int64(1), int64(0), int64(0), int64(1)},
		},
		{
			InCmds: []string{"BITOP OR unitTestKeyOR unitTestKeyB unitTestKeyA"},
			Out:    []interface{}{int64(2)},
		},
		{
			InCmds: []string{"GETBIT unitTestKeyOR 1", "GETBIT unitTestKeyOR 2", "GETBIT unitTestKeyOR 3", "GETBIT unitTestKeyOR 7", "GETBIT unitTestKeyOR 8", "GETBIT unitTestKeyOR 9", "GETBIT unitTestKeyOR 12"},
			Out:    []interface{}{int64(1), int64(1), int64(1), int64(1), int64(1), int64(0), int64(0)},
		},
		{
			InCmds: []string{"BITOP AND unitTestKeyAND unitTestKeyB unitTestKeyA"},
			Out:    []interface{}{int64(2)},
		},
		{
			InCmds: []string{"GETBIT unitTestKeyAND 1", "GETBIT unitTestKeyAND 2", "GETBIT unitTestKeyAND 7", "GETBIT unitTestKeyAND 8", "GETBIT unitTestKeyAND 9"},
			Out:    []interface{}{int64(0), int64(0), int64(1), int64(0), int64(0)},
		},
		{
			InCmds: []string{"BITOP XOR unitTestKeyXOR unitTestKeyB unitTestKeyA"},
			Out:    []interface{}{int64(2)},
		},
		{
			InCmds: []string{"GETBIT unitTestKeyXOR 1", "GETBIT unitTestKeyXOR 2", "GETBIT unitTestKeyXOR 3", "GETBIT unitTestKeyXOR 7", "GETBIT unitTestKeyXOR 8"},
			Out:    []interface{}{int64(1), int64(1), int64(1), int64(0), int64(1)},
		},
	}

	for _, tcase := range testcases {
		for i := 0; i < len(tcase.InCmds); i++ {
			cmd := tcase.InCmds[i]
			out := tcase.Out[i]
			assert.Equal(t, out, fireCommand(conn, cmd), "Value mismatch for cmd %s\n.", cmd)
		}
	}
}

func TestBitCount(t *testing.T) {
	conn := getLocalConnection()
	testcases := []DTestCase{
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
	}

	for _, tcase := range testcases {
		for i := 0; i < len(tcase.InCmds); i++ {
			cmd := tcase.InCmds[i]
			out := tcase.Out[i]
			assert.Equal(t, out, fireCommand(conn, cmd), "Value mismatch for cmd %s\n.", cmd)
		}
	}
}

func generateSetBitCommand(connection net.Conn, bitPosition int) int64 {
	command := fmt.Sprintf("SETBIT unitTestKeyA %d 1", bitPosition)
	responseValue := fireCommand(connection, command)
	if responseValue == nil {
		return -1
	}
	return responseValue.(int64)
}

func BenchmarkSetBitCommand(b *testing.B) {
	connection := getLocalConnection()
	for n := 0; n < 1000; n++ {
		setBitCommand := generateSetBitCommand(connection, n)
		if setBitCommand < 0 {
			b.Fail()
		}
	}
}

func generateGetBitCommand(connection net.Conn, bitPosition int) int64 {
	command := fmt.Sprintf("GETBIT unitTestKeyA %d", bitPosition)
	responseValue := fireCommand(connection, command)
	if responseValue == nil {
		return -1
	}
	return responseValue.(int64)
}

func BenchmarkGetBitCommand(b *testing.B) {
	connection := getLocalConnection()
	for n := 0; n < 1000; n++ {
		getBitCommand := generateGetBitCommand(connection, n)
		if getBitCommand < 0 {
			b.Fail()
		}
	}
}
