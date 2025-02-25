// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

// The following commands are a part of this test class:
// SETBIT, GETBIT, BITCOUNT, BITOP, BITPOS, BITFIELD, BITFIELD_RO

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// func TestBitOp(t *testing.T) {
// 	client := getLocalConnection()
// 	defer client.Close()
// 	testcases := []struct {
// 		InCmds []string
// 		Out    []interface{}
// 	}{
// 		{
// 			InCmds: []string{"SETBIT unitTestKeyA 1 1", "SETBIT unitTestKeyA 3 1", "SETBIT unitTestKeyA 5 1", "SETBIT unitTestKeyA 7 1", "SETBIT unitTestKeyA 8 1"},
// 			Out:    []interface{}{int64(0), int64(0), int64(0), int64(0), int64(0)},
// 		},
// 		{
// 			InCmds: []string{"SETBIT unitTestKeyB 2 1", "SETBIT unitTestKeyB 4 1", "SETBIT unitTestKeyB 7 1"},
// 			Out:    []interface{}{int64(0), int64(0), int64(0)},
// 		},
// 		{
// 			InCmds: []string{"SET foo bar", "SETBIT foo 2 1", "SETBIT foo 4 1", "SETBIT foo 7 1", "GET foo"},
// 			Out:    []interface{}{"OK", int64(1), int64(0), int64(0), "kar"},
// 		},
// 		{
// 			InCmds: []string{"SET mykey12 1343", "SETBIT mykey12 2 1", "SETBIT mykey12 4 1", "SETBIT mykey12 7 1", "GET mykey12"},
// 			Out:    []interface{}{"OK", int64(1), int64(0), int64(1), int64(9343)},
// 		},
// 		{
// 			InCmds: []string{"SET foo12 bar", "SETBIT foo12 2 1", "SETBIT foo12 4 1", "SETBIT foo12 7 1", "GET foo12"},
// 			Out:    []interface{}{"OK", int64(1), int64(0), int64(0), "kar"},
// 		},
// 		{
// 			InCmds: []string{"BITOP NOT unitTestKeyNOT unitTestKeyA "},
// 			Out:    []interface{}{int64(2)},
// 		},
// 		{
// 			InCmds: []string{"GETBIT unitTestKeyNOT 1", "GETBIT unitTestKeyNOT 2", "GETBIT unitTestKeyNOT 7", "GETBIT unitTestKeyNOT 8", "GETBIT unitTestKeyNOT 9"},
// 			Out:    []interface{}{int64(0), int64(1), int64(0), int64(0), int64(1)},
// 		},
// 		{
// 			InCmds: []string{"BITOP OR unitTestKeyOR unitTestKeyB unitTestKeyA"},
// 			Out:    []interface{}{int64(2)},
// 		},
// 		{
// 			InCmds: []string{"GETBIT unitTestKeyOR 1", "GETBIT unitTestKeyOR 2", "GETBIT unitTestKeyOR 3", "GETBIT unitTestKeyOR 7", "GETBIT unitTestKeyOR 8", "GETBIT unitTestKeyOR 9", "GETBIT unitTestKeyOR 12"},
// 			Out:    []interface{}{int64(1), int64(1), int64(1), int64(1), int64(1), int64(0), int64(0)},
// 		},
// 		{
// 			InCmds: []string{"BITOP AND unitTestKeyAND unitTestKeyB unitTestKeyA"},
// 			Out:    []interface{}{int64(2)},
// 		},
// 		{
// 			InCmds: []string{"GETBIT unitTestKeyAND 1", "GETBIT unitTestKeyAND 2", "GETBIT unitTestKeyAND 7", "GETBIT unitTestKeyAND 8", "GETBIT unitTestKeyAND 9"},
// 			Out:    []interface{}{int64(0), int64(0), int64(1), int64(0), int64(0)},
// 		},
// 		{
// 			InCmds: []string{"BITOP XOR unitTestKeyXOR unitTestKeyB unitTestKeyA"},
// 			Out:    []interface{}{int64(2)},
// 		},
// 		{
// 			InCmds: []string{"GETBIT unitTestKeyXOR 1", "GETBIT unitTestKeyXOR 2", "GETBIT unitTestKeyXOR 3", "GETBIT unitTestKeyXOR 7", "GETBIT unitTestKeyXOR 8"},
// 			Out:    []interface{}{int64(1), int64(1), int64(1), int64(0), int64(1)},
// 		},
// 	}

// 	for _, tcase := range testcases {
// 		for i := 0; i < len(tcase.InCmds); i++ {
// 			cmd := tcase.InCmds[i]
// 			out := tcase.Out[i]
// 			assert.Equal(t, out, client.FireString(cmd), "Value mismatch for cmd %s\n.", cmd)
// 		}
// 	}
// }

// func TestBitOpsString(t *testing.T) {

// 	client := getLocalConnection()
// 	defer client.Close()
// 	// foobar in bits is 01100110 01101111 01101111 01100010 01100001 01110010
// 	fooBarBits := "011001100110111101101111011000100110000101110010"
// 	// randomly get 8 bits for testing
// 	testOffsets := make([]int, 8)

// 	for i := 0; i < 8; i++ {
// 		testOffsets[i] = rand.Intn(len(fooBarBits))
// 	}

// 	getBitTestCommands := make([]string, 8+1)
// 	getBitTestExpected := make([]interface{}, 8+1)

// 	getBitTestCommands[0] = "SET foo foobar"
// 	getBitTestExpected[0] = "OK"

// 	for i := 1; i < 8+1; i++ {
// 		getBitTestCommands[i] = fmt.Sprintf("GETBIT foo %d", testOffsets[i-1])
// 		getBitTestExpected[i] = int64(fooBarBits[testOffsets[i-1]] - '0')
// 	}

// 	testCases := []struct {
// 		name       string
// 		cmds       []string
// 		expected   []interface{}
// 		assertType []string
// 	}{
// 		{
// 			name:       "Getbit of a key containing a string",
// 			cmds:       getBitTestCommands,
// 			expected:   getBitTestExpected,
// 			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal"},
// 		},
// 		{
// 			name:       "Getbit of a key containing an integer",
// 			cmds:       []string{"SET foo 10", "GETBIT foo 0", "GETBIT foo 1", "GETBIT foo 2", "GETBIT foo 3", "GETBIT foo 4", "GETBIT foo 5", "GETBIT foo 6", "GETBIT foo 7"},
// 			expected:   []interface{}{"OK", int64(0), int64(0), int64(1), int64(1), int64(0), int64(0), int64(0), int64(1)},
// 			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal"},
// 		}, {
// 			name:       "Getbit of a key containing an integer 2nd byte",
// 			cmds:       []string{"SET foo 10", "GETBIT foo 8", "GETBIT foo 9", "GETBIT foo 10", "GETBIT foo 11", "GETBIT foo 12", "GETBIT foo 13", "GETBIT foo 14", "GETBIT foo 15"},
// 			expected:   []interface{}{"OK", int64(0), int64(0), int64(1), int64(1), int64(0), int64(0), int64(0), int64(0)},
// 			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal"},
// 		},
// 		{
// 			name:       "Getbit of a key with an offset greater than the length of the string in bits",
// 			cmds:       []string{"SET foo foobar", "GETBIT foo 100", "GETBIT foo 48", "GETBIT foo 47"},
// 			expected:   []interface{}{"OK", int64(0), int64(0), int64(0)},
// 			assertType: []string{"equal", "equal", "equal", "equal"},
// 		},
// 		{
// 			name:       "Bitcount of a key containing a string",
// 			cmds:       []string{"SET foo foobar", "BITCOUNT foo 0 -1", "BITCOUNT foo", "BITCOUNT foo 0 0", "BITCOUNT foo 1 1", "BITCOUNT foo 1 1 Byte", "BITCOUNT foo 5 30 BIT"},
// 			expected:   []interface{}{"OK", int64(26), int64(26), int64(4), int64(6), int64(6), int64(17)},
// 			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal"},
// 		},
// 		{
// 			name:       "Bitcount of a key containing an integer",
// 			cmds:       []string{"SET foo 10", "BITCOUNT foo 0 -1", "BITCOUNT foo", "BITCOUNT foo 0 0", "BITCOUNT foo 1 1", "BITCOUNT foo 1 1 Byte", "BITCOUNT foo 5 30 BIT"},
// 			expected:   []interface{}{"OK", int64(5), int64(5), int64(3), int64(2), int64(2), int64(3)},
// 			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal"},
// 		},
// 		{
// 			name:       "Setbit of a key containing a string",
// 			cmds:       []string{"SET foo foobar", "setbit foo 7 1", "get foo", "setbit foo 49 1", "setbit foo 50 1", "get foo", "setbit foo 49 0", "get foo"},
// 			expected:   []interface{}{"OK", int64(0), "goobar", int64(0), int64(0), "goobar`", int64(1), "goobar "},
// 			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal"},
// 		},
// 		{
// 			name:       "Setbit of a key must not change the expiry of the key if expiry is set",
// 			cmds:       []string{"SET foo foobar", "EXPIRE foo 100", "TTL foo", "SETBIT foo 7 1", "TTL foo"},
// 			expected:   []interface{}{"OK", int64(1), int64(100), int64(0), int64(100)},
// 			assertType: []string{"equal", "equal", "less", "equal", "less"},
// 		},
// 		{
// 			name:       "Setbit of a key must not add expiry to the key if expiry is not set",
// 			cmds:       []string{"SET foo foobar", "TTL foo", "SETBIT foo 7 1", "TTL foo"},
// 			expected:   []interface{}{"OK", int64(-1), int64(0), int64(-1)},
// 			assertType: []string{"equal", "equal", "equal", "equal"},
// 		},
// 		{
// 			name:       "Bitop not of a key containing a string",
// 			cmds:       []string{"SET foo foobar", "BITOP NOT baz foo", "GET baz", "BITOP NOT bazz baz", "GET bazz"},
// 			expected:   []interface{}{"OK", int64(6), "\x99\x90\x90\x9d\x9e\x8d", int64(6), "foobar"},
// 			assertType: []string{"equal", "equal", "equal", "equal", "equal"},
// 		},
// 		{
// 			name:       "Bitop not of a key containing an integer",
// 			cmds:       []string{"SET foo 10", "BITOP NOT baz foo", "GET baz", "BITOP NOT bazz baz", "GET bazz"},
// 			expected:   []interface{}{"OK", int64(2), "\xce\xcf", int64(2), int64(10)},
// 			assertType: []string{"equal", "equal", "equal", "equal", "equal"},
// 		},
// 		{
// 			name:       "Get a string created with setbit",
// 			cmds:       []string{"SETBIT foo 1 1", "SETBIT foo 3 1", "GET foo"},
// 			expected:   []interface{}{int64(0), int64(0), "P"},
// 			assertType: []string{"equal", "equal", "equal"},
// 		},
// 		{
// 			name:       "Bitop and of keys containing a string and get the destkey",
// 			cmds:       []string{"SET foo foobar", "SET baz abcdef", "BITOP AND bazz foo baz", "GET bazz"},
// 			expected:   []interface{}{"OK", "OK", int64(6), "`bc`ab"},
// 			assertType: []string{"equal", "equal", "equal", "equal"},
// 		},
// 		{
// 			name:       "BITOP AND of keys containing integers and get the destkey",
// 			cmds:       []string{"SET foo 10", "SET baz 5", "BITOP AND bazz foo baz", "GET bazz"},
// 			expected:   []interface{}{"OK", "OK", int64(2), "1\x00"},
// 			assertType: []string{"equal", "equal", "equal", "equal"},
// 		},
// 		{
// 			name:       "Bitop or of keys containing a string, a bytearray and get the destkey",
// 			cmds:       []string{"MSET foo foobar baz abcdef", "SETBIT bazz 8 1", "BITOP and bazzz foo baz bazz", "GET bazzz"},
// 			expected:   []interface{}{"OK", int64(0), int64(6), "\x00\x00\x00\x00\x00\x00"},
// 			assertType: []string{"equal", "equal", "equal", "equal"},
// 		},
// 		{
// 			name:       "BITOP OR of keys containing strings and get the destkey",
// 			cmds:       []string{"MSET foo foobar baz abcdef", "BITOP OR bazz foo baz", "GET bazz"},
// 			expected:   []interface{}{"OK", int64(6), "goofev"},
// 			assertType: []string{"equal", "equal", "equal"},
// 		},
// 		{
// 			name:       "BITOP OR of keys containing integers and get the destkey",
// 			cmds:       []string{"SET foo 10", "SET baz 5", "BITOP OR bazz foo baz", "GET bazz"},
// 			expected:   []interface{}{"OK", "OK", int64(2), "50"},
// 			assertType: []string{"equal", "equal", "equal", "equal"},
// 		},
// 		{
// 			name:       "BITOP OR of keys containing strings and a bytearray and get the destkey",
// 			cmds:       []string{"MSET foo foobar baz abcdef", "SETBIT bazz 8 1", "BITOP OR bazzz foo baz bazz", "GET bazzz", "SETBIT bazz 8 0", "SETBIT bazz 49 1", "BITOP OR bazzz foo baz bazz", "GET bazzz"},
// 			expected:   []interface{}{"OK", int64(0), int64(6), "g\xefofev", int64(1), int64(0), int64(7), "goofev@"},
// 			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal"},
// 		},
// 		{
// 			name:       "BITOP XOR of keys containing strings and get the destkey",
// 			cmds:       []string{"MSET foo foobar baz abcdef", "BITOP XOR bazz foo baz", "GET bazz"},
// 			expected:   []interface{}{"OK", int64(6), "\x07\x0d\x0c\x06\x04\x14"},
// 			assertType: []string{"equal", "equal", "equal"},
// 		},
// 		{
// 			name:       "BITOP XOR of keys containing strings and a bytearray and get the destkey",
// 			cmds:       []string{"MSET foo foobar baz abcdef", "SETBIT bazz 8 1", "BITOP XOR bazzz foo baz bazz", "GET bazzz", "SETBIT bazz 8 0", "SETBIT bazz 49 1", "BITOP XOR bazzz foo baz bazz", "GET bazzz", "Setbit bazz 49 0", "BITOP XOR bazzz foo baz bazz", "GET bazzz"},
// 			expected:   []interface{}{"OK", int64(0), int64(6), "\x07\x8d\x0c\x06\x04\x14", int64(1), int64(0), int64(7), "\x07\r\x0c\x06\x04\x14@", int64(1), int64(7), "\x07\r\x0c\x06\x04\x14\x00"},
// 			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal"},
// 		},
// 		{
// 			name:       "BITOP XOR of keys containing integers and get the destkey",
// 			cmds:       []string{"SET foo 10", "SET baz 5", "BITOP XOR bazz foo baz", "GET bazz"},
// 			expected:   []interface{}{"OK", "OK", int64(2), "\x040"},
// 			assertType: []string{"equal", "equal", "equal", "equal"},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			// Delete the key before running the test
// 			client.FireString("DEL foo")
// 			client.FireString("DEL baz")
// 			client.FireString("DEL bazz")
// 			client.FireString("DEL bazzz")
// 			for i := 0; i < len(tc.cmds); i++ {
// 				res := client.FireString(tc.cmds[i])

// 				switch tc.assertType[i] {
// 				case "equal":
// 					assert.Equal(t, tc.expected[i], res)
// 				case "less":
// 					assert.True(t, res.(int64) <= tc.expected[i].(int64), "CMD: %s Expected %d to be less than or equal to %d", tc.cmds[i], res, tc.expected[i])
// 				}
// 			}
// 		})
// 	}
// }

func TestBitCount(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()
	testcases := []struct {
		InCmds []string
		Out    []interface{}
	}{
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
			InCmds: []string{"SETBIT mykey -1 1"},
			Out:    []interface{}{"ERR bit offset is not an integer or out of range"},
		},
		{
			InCmds: []string{"SETBIT mykey -1 0"},
			Out:    []interface{}{"ERR bit offset is not an integer or out of range"},
		},
		{
			InCmds: []string{"SETBIT mykey -10000 1"},
			Out:    []interface{}{"ERR bit offset is not an integer or out of range"},
		},
		{
			InCmds: []string{"SETBIT mykey -10000 0"},
			Out:    []interface{}{"ERR bit offset is not an integer or out of range"},
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
		{
			InCmds: []string{"BITCOUNT"},
			Out:    []interface{}{"ERR wrong number of arguments for 'bitcount' command"},
		},
		{
			InCmds: []string{"BITCOUNT mykey"},
			Out:    []interface{}{int64(1)},
		},
		{
			InCmds: []string{"BITCOUNT mykey 0"},
			Out:    []interface{}{"ERR syntax error"},
		},
	}

	for _, tcase := range testcases {
		for i := 0; i < len(tcase.InCmds); i++ {
			cmd := tcase.InCmds[i]
			out := tcase.Out[i]
			assert.Equal(t, out, client.FireString(cmd), "Value mismatch for cmd %s\n.", cmd)
		}
	}
}

func TestBitPos(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()
	testcases := []struct {
		name         string
		val          interface{}
		inCmd        string
		out          interface{}
		setCmdSETBIT bool
	}{
		{
			name:  "String interval BIT 0,-1 ",
			val:   "\\x00\\xff\\x00",
			inCmd: "BITPOS testkey 0 0 -1 bit",
			out:   int64(0),
		},
		{
			name:  "String interval BIT 8,-1",
			val:   "\\x00\\xff\\x00",
			inCmd: "BITPOS testkey 0 8 -1 bit",
			out:   int64(8),
		},
		{
			name:  "String interval BIT 16,-1",
			val:   "\\x00\\xff\\x00",
			inCmd: "BITPOS testkey 0 16 -1 bit",
			out:   int64(16),
		},
		{
			name:  "String interval BIT 16,200",
			val:   "\\x00\\xff\\x00",
			inCmd: "BITPOS testkey 0 16 200 bit",
			out:   int64(16),
		},
		{
			name:  "String interval BIT 8,8",
			val:   "\\x00\\xff\\x00",
			inCmd: "BITPOS testkey 0 8 8 bit",
			out:   int64(8),
		},
		{
			name:  "FindsFirstZeroBit",
			val:   "\xff\xf0\x00",
			inCmd: "BITPOS testkey 0",
			out:   int64(12),
		},
		{
			name:  "FindsFirstOneBit",
			val:   "\x00\x0f\xff",
			inCmd: "BITPOS testkey 1",
			out:   int64(12),
		},
		{
			name:  "NoOneBitFound",
			val:   "\x00\x00\x00",
			inCmd: "BITPOS testkey 1",
			out:   int64(-1),
		},
		{
			name:  "NoZeroBitFound",
			val:   "\xff\xff\xff",
			inCmd: "BITPOS testkey 0",
			out:   int64(24),
		},
		{
			name:  "NoZeroBitFoundWithRangeStartPos",
			val:   "\xff\xff\xff",
			inCmd: "BITPOS testkey 0 2",
			out:   int64(24),
		},
		{
			name:  "NoZeroBitFoundWithOOBRangeStartPos",
			val:   "\xff\xff\xff",
			inCmd: "BITPOS testkey 0 4",
			out:   int64(-1),
		},
		{
			name:  "NoZeroBitFoundWithRange",
			val:   "\xff\xff\xff",
			inCmd: "BITPOS testkey 0 2 2",
			out:   int64(-1),
		},
		{
			name:  "NoZeroBitFoundWithRangeAndRangeType",
			val:   "\xff\xff\xff",
			inCmd: "BITPOS testkey 0 2 2 BIT",
			out:   int64(-1),
		},
		{
			name:  "FindsFirstZeroBitInRange",
			val:   "\xff\xf0\xff",
			inCmd: "BITPOS testkey 0 1 2",
			out:   int64(12),
		},
		{
			name:  "FindsFirstOneBitInRange",
			val:   "\x00\x00\xf0",
			inCmd: "BITPOS testkey 1 2 3",
			out:   int64(16),
		},
		{
			name:  "StartGreaterThanEnd",
			val:   "\xff\xf0\x00",
			inCmd: "BITPOS testkey 0 3 2",
			out:   int64(-1),
		},
		{
			name:  "FindsFirstOneBitWithNegativeStart",
			val:   "\x00\x00\xf0",
			inCmd: "BITPOS testkey 1 -2 -1",
			out:   int64(16),
		},
		{
			name:  "FindsFirstZeroBitWithNegativeEnd",
			val:   "\xff\xf0\xff",
			inCmd: "BITPOS testkey 0 1 -1",
			out:   int64(12),
		},
		{
			name:  "FindsFirstZeroBitInByteRange",
			val:   "\xff\x00\xff",
			inCmd: "BITPOS testkey 0 1 2 BYTE",
			out:   int64(8),
		},
		{
			name:  "FindsFirstOneBitInBitRange",
			val:   "\x00\x01\x00",
			inCmd: "BITPOS testkey 1 0 16 BIT",
			out:   int64(15),
		},
		{
			name:  "NoBitFoundInByteRange",
			val:   "\xff\xff\xff",
			inCmd: "BITPOS testkey 0 0 2 BYTE",
			out:   int64(-1),
		},
		{
			name:  "NoBitFoundInBitRange",
			val:   "\x00\x00\x00",
			inCmd: "BITPOS testkey 1 0 23 BIT",
			out:   int64(-1),
		},
		{
			name:  "EmptyStringReturnsMinusOneForZeroBit",
			val:   "\"\"",
			inCmd: "BITPOS testkey 0",
			out:   int64(-1),
		},
		{
			name:  "EmptyStringReturnsMinusOneForOneBit",
			val:   "\"\"",
			inCmd: "BITPOS testkey 1",
			out:   int64(-1),
		},
		{
			name:  "SingleByteString",
			val:   "\x80",
			inCmd: "BITPOS testkey 1",
			out:   int64(0),
		},
		{
			name:  "RangeExceedsStringLength",
			val:   "\x00\xff",
			inCmd: "BITPOS testkey 1 0 20 BIT",
			out:   int64(8),
		},
		{
			name:  "InvalidBitArgument",
			inCmd: "BITPOS testkey 2",
			out:   "ERR the bit argument must be 1 or 0",
		},
		{
			name:  "NonIntegerStartParameter",
			inCmd: "BITPOS testkey 0 start",
			out:   "ERR value is not an integer or out of range",
		},
		{
			name:  "NonIntegerEndParameter",
			inCmd: "BITPOS testkey 0 1 end",
			out:   "ERR value is not an integer or out of range",
		},
		{
			name:  "InvalidRangeType",
			inCmd: "BITPOS testkey 0 1 2 BYTEs",
			out:   "ERR syntax error",
		},
		{
			name:  "InsufficientArguments",
			inCmd: "BITPOS testkey",
			out:   "ERR wrong number of arguments for 'bitpos' command",
		},
		{
			name:  "NonExistentKeyForZeroBit",
			inCmd: "BITPOS nonexistentkey 0",
			out:   int64(0),
		},
		{
			name:  "NonExistentKeyForOneBit",
			inCmd: "BITPOS nonexistentkey 1",
			out:   int64(-1),
		},
		{
			name:  "IntegerValue",
			val:   65280, // 0xFF00 in decimal
			inCmd: "BITPOS testkey 0",
			out:   int64(0),
		},
		{
			name:  "LargeIntegerValue",
			val:   16777215, // 0xFFFFFF in decimal
			inCmd: "BITPOS testkey 1",
			out:   int64(2),
		},
		{
			name:  "SmallIntegerValue",
			val:   1, // 0x01 in decimal
			inCmd: "BITPOS testkey 0",
			out:   int64(0),
		},
		{
			name:  "ZeroIntegerValue",
			val:   0,
			inCmd: "BITPOS testkey 1",
			out:   int64(2),
		},
		{
			name:  "BitRangeStartGreaterThanBitLength",
			val:   "\xff\xff\xff",
			inCmd: "BITPOS testkey 0 25 30 BIT",
			out:   int64(-1),
		},
		{
			name:  "BitRangeEndExceedsBitLength",
			val:   "\xff\xff\xff",
			inCmd: "BITPOS testkey 0 0 30 BIT",
			out:   int64(-1),
		},
		{
			name:  "NegativeStartInBitRange",
			val:   "\x00\xff\xff",
			inCmd: "BITPOS testkey 1 -16 -1 BIT",
			out:   int64(8),
		},
		{
			name:  "LargeNegativeStart",
			val:   "\x00\xff\xff",
			inCmd: "BITPOS testkey 1 -100 -1",
			out:   int64(8),
		},
		{
			name:  "LargePositiveEnd",
			val:   "\x00\xff\xff",
			inCmd: "BITPOS testkey 1 0 100",
			out:   int64(8),
		},
		{
			name:  "StartAndEndEqualInByteRange",
			val:   "\x0f\xff\xff",
			inCmd: "BITPOS testkey 0 1 1 BYTE",
			out:   int64(-1),
		},
		{
			name:  "StartAndEndEqualInBitRange",
			val:   "\x0f\xff\xff",
			inCmd: "BITPOS testkey 1 1 1 BIT",
			out:   int64(-1),
		},
		{
			name:  "FindFirstZeroBitInNegativeRange",
			val:   "\xff\x00\xff",
			inCmd: "BITPOS testkey 0 -2 -1",
			out:   int64(8),
		},
		{
			name:  "FindFirstOneBitInNegativeRangeBIT",
			val:   "\x00\x00\x80",
			inCmd: "BITPOS testkey 1 -8 -1 BIT",
			out:   int64(16),
		},
		{
			name:  "MaxIntegerValue",
			val:   math.MaxInt64,
			inCmd: "BITPOS testkey 0",
			out:   int64(0),
		},
		{
			name:  "MinIntegerValue",
			val:   math.MinInt64,
			inCmd: "BITPOS testkey 1",
			out:   int64(2),
		},
		{
			name:  "SingleBitStringZero",
			val:   "\x00",
			inCmd: "BITPOS testkey 1",
			out:   int64(-1),
		},
		{
			name:  "SingleBitStringOne",
			val:   "\x01",
			inCmd: "BITPOS testkey 0",
			out:   int64(0),
		},
		{
			name:  "AllBitsSetExceptLast",
			val:   "\xff\xff\xfe",
			inCmd: "BITPOS testkey 0",
			out:   int64(23),
		},
		{
			name:  "OnlyLastBitSet",
			val:   "\x00\x00\x01",
			inCmd: "BITPOS testkey 1",
			out:   int64(23),
		},
		{
			name:  "AlternatingBitsLongString",
			val:   "\xaa\xaa\xaa\xaa\xaa",
			inCmd: "BITPOS testkey 0",
			out:   int64(1),
		},
		{
			name:  "VeryLargeByteString",
			val:   strings.Repeat("\xff", 1000) + "\x00",
			inCmd: "BITPOS testkey 0",
			out:   int64(8000),
		},
		{
			name:         "FindZeroBitOnSetBitKey",
			val:          "8 1",
			inCmd:        "BITPOS testkeysb 1",
			out:          int64(8),
			setCmdSETBIT: true,
		},
		{
			name:         "FindOneBitOnSetBitKey",
			val:          "1 1",
			inCmd:        "BITPOS testkeysb 1",
			out:          int64(1),
			setCmdSETBIT: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var setCmd string
			if tc.setCmdSETBIT {
				setCmd = fmt.Sprintf("SETBIT testkeysb %s", tc.val.(string))
			} else {
				switch v := tc.val.(type) {
				case string:
					setCmd = fmt.Sprintf("SET testkey %s", v)
				case int:
					setCmd = fmt.Sprintf("SET testkey %d", v)
				default:
					// For test cases where we don't set a value (e.g., error cases)
					setCmd = ""
				}
			}

			if setCmd != "" {
				client.FireString(setCmd)
			}

			result := client.FireString(tc.inCmd)
			assert.Equal(t, tc.out, result, "Mismatch for cmd %s\n", tc.inCmd)
		})
	}
}

func TestBitfield(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	client.FireString("FLUSHDB")
	defer client.FireString("FLUSHDB") // clean up after all test cases
	syntaxErrMsg := "ERR syntax error"
	bitFieldTypeErrMsg := "ERR Invalid bitfield type. Use something like i16 u8. Note that u64 is not supported but i64 is"
	integerErrMsg := "ERR value is not an integer or out of range"
	overflowErrMsg := "ERR Invalid OVERFLOW type specified"

	testCases := []struct {
		Name     string
		Commands []string
		Expected []interface{}
		Delay    []time.Duration
		CleanUp  []string
	}{
		{
			Name:     "BITFIELD Arity Check",
			Commands: []string{"bitfield"},
			Expected: []interface{}{"ERR wrong number of arguments for 'bitfield' command"},
			Delay:    []time.Duration{0},
			CleanUp:  []string{},
		},
		{
			Name:     "BITFIELD on unsupported type of SET",
			Commands: []string{"SADD bits a b c", "bitfield bits"},
			Expected: []interface{}{int64(3), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:     "BITFIELD on unsupported type of JSON",
			Commands: []string{"json.set bits $ 1", "bitfield bits"},
			Expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:     "BITFIELD on unsupported type of HSET",
			Commands: []string{"HSET bits a 1", "bitfield bits"},
			Expected: []interface{}{int64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name: "BITFIELD with syntax errors",
			Commands: []string{
				"bitfield bits set u8 0 255 incrby u8 0 100 get u8",
				"bitfield bits set a8 0 255 incrby u8 0 100 get u8",
				"bitfield bits set u8 a 255 incrby u8 0 100 get u8",
				"bitfield bits set u8 0 255 incrby u8 0 100 overflow wraap",
				"bitfield bits set u8 0 incrby u8 0 100 get u8 288",
			},
			Expected: []interface{}{
				syntaxErrMsg,
				bitFieldTypeErrMsg,
				"ERR bit offset is not an integer or out of range",
				overflowErrMsg,
				integerErrMsg,
			},
			Delay:   []time.Duration{0, 0, 0, 0, 0},
			CleanUp: []string{"Del bits"},
		},
		{
			Name:     "BITFIELD signed SET and GET basics",
			Commands: []string{"bitfield bits set i8 0 -100", "bitfield bits set i8 0 101", "bitfield bits get i8 0"},
			Expected: []interface{}{[]interface{}{int64(0)}, []interface{}{int64(-100)}, []interface{}{int64(101)}},
			Delay:    []time.Duration{0, 0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:     "BITFIELD unsigned SET and GET basics",
			Commands: []string{"bitfield bits set u8 0 255", "bitfield bits set u8 0 100", "bitfield bits get u8 0"},
			Expected: []interface{}{[]interface{}{int64(0)}, []interface{}{int64(255)}, []interface{}{int64(100)}},
			Delay:    []time.Duration{0, 0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:     "BITFIELD signed SET and GET together",
			Commands: []string{"bitfield bits set i8 0 255 set i8 0 100 get i8 0"},
			Expected: []interface{}{[]interface{}{int64(0), int64(-1), int64(100)}},
			Delay:    []time.Duration{0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:     "BITFIELD unsigned with SET, GET and INCRBY arguments",
			Commands: []string{"bitfield bits set u8 0 255 incrby u8 0 100 get u8 0"},
			Expected: []interface{}{[]interface{}{int64(0), int64(99), int64(99)}},
			Delay:    []time.Duration{0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:     "BITFIELD with only key as argument",
			Commands: []string{"bitfield bits"},
			Expected: []interface{}{[]interface{}{}},
			Delay:    []time.Duration{0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name: "BITFIELD #<idx> form",
			Commands: []string{
				"bitfield bits set u8 #0 65",
				"bitfield bits set u8 #1 66",
				"bitfield bits set u8 #2 67",
				"get bits",
			},
			Expected: []interface{}{[]interface{}{int64(0)}, []interface{}{int64(0)}, []interface{}{int64(0)}, "ABC"},
			Delay:    []time.Duration{0, 0, 0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name: "BITFIELD basic INCRBY form",
			Commands: []string{
				"bitfield bits set u8 #0 10",
				"bitfield bits incrby u8 #0 100",
				"bitfield bits incrby u8 #0 100",
			},
			Expected: []interface{}{[]interface{}{int64(0)}, []interface{}{int64(110)}, []interface{}{int64(210)}},
			Delay:    []time.Duration{0, 0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name: "BITFIELD chaining of multiple commands",
			Commands: []string{
				"bitfield bits set u8 #0 10",
				"bitfield bits incrby u8 #0 100 incrby u8 #0 100",
			},
			Expected: []interface{}{[]interface{}{int64(0)}, []interface{}{int64(110), int64(210)}},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name: "BITFIELD unsigned overflow wrap",
			Commands: []string{
				"bitfield bits set u8 #0 100",
				"bitfield bits overflow wrap incrby u8 #0 257",
				"bitfield bits get u8 #0",
				"bitfield bits overflow wrap incrby u8 #0 255",
				"bitfield bits get u8 #0",
			},
			Expected: []interface{}{
				[]interface{}{int64(0)},
				[]interface{}{int64(101)},
				[]interface{}{int64(101)},
				[]interface{}{int64(100)},
				[]interface{}{int64(100)},
			},
			Delay:   []time.Duration{0, 0, 0, 0, 0},
			CleanUp: []string{"DEL bits"},
		},
		{
			Name: "BITFIELD unsigned overflow sat",
			Commands: []string{
				"bitfield bits set u8 #0 100",
				"bitfield bits overflow sat incrby u8 #0 257",
				"bitfield bits get u8 #0",
				"bitfield bits overflow sat incrby u8 #0 -255",
				"bitfield bits get u8 #0",
			},
			Expected: []interface{}{
				[]interface{}{int64(0)},
				[]interface{}{int64(255)},
				[]interface{}{int64(255)},
				[]interface{}{int64(0)},
				[]interface{}{int64(0)},
			},
			Delay:   []time.Duration{0, 0, 0, 0, 0},
			CleanUp: []string{"DEL bits"},
		},
		{
			Name: "BITFIELD signed overflow wrap",
			Commands: []string{
				"bitfield bits set i8 #0 100",
				"bitfield bits overflow wrap incrby i8 #0 257",
				"bitfield bits get i8 #0",
				"bitfield bits overflow wrap incrby i8 #0 255",
				"bitfield bits get i8 #0",
			},
			Expected: []interface{}{
				[]interface{}{int64(0)},
				[]interface{}{int64(101)},
				[]interface{}{int64(101)},
				[]interface{}{int64(100)},
				[]interface{}{int64(100)},
			},
			Delay:   []time.Duration{0, 0, 0, 0, 0},
			CleanUp: []string{"DEL bits"},
		},
		{
			Name: "BITFIELD signed overflow sat",
			Commands: []string{
				"bitfield bits set u8 #0 100",
				"bitfield bits overflow sat incrby i8 #0 257",
				"bitfield bits get i8 #0",
				"bitfield bits overflow sat incrby i8 #0 -255",
				"bitfield bits get i8 #0",
			},
			Expected: []interface{}{
				[]interface{}{int64(0)},
				[]interface{}{int64(127)},
				[]interface{}{int64(127)},
				[]interface{}{int64(-128)},
				[]interface{}{int64(-128)},
			},
			Delay:   []time.Duration{0, 0, 0, 0, 0},
			CleanUp: []string{"DEL bits"},
		},
		{
			Name:     "BITFIELD regression 1",
			Commands: []string{"set bits 1", "bitfield bits get u1 0"},
			Expected: []interface{}{"OK", []interface{}{int64(0)}},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name: "BITFIELD regression 2",
			Commands: []string{
				"bitfield mystring set i8 0 10",
				"bitfield mystring set i8 64 10",
				"bitfield mystring incrby i8 10 99900",
			},
			Expected: []interface{}{[]interface{}{int64(0)}, []interface{}{int64(0)}, []interface{}{int64(60)}},
			Delay:    []time.Duration{0, 0, 0},
			CleanUp:  []string{"DEL mystring"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {

			for i := 0; i < len(tc.Commands); i++ {
				if tc.Delay[i] > 0 {
					time.Sleep(tc.Delay[i])
				}
				result := client.FireString(tc.Commands[i])
				expected := tc.Expected[i]
				assert.Equal(t, expected, result)
			}

			for _, cmd := range tc.CleanUp {
				client.FireString(cmd)
			}
		})
	}
}

func TestBitfieldRO(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	client.FireString("FLUSHDB")
	defer client.FireString("FLUSHDB")

	syntaxErrMsg := "ERR syntax error"
	bitFieldTypeErrMsg := "ERR Invalid bitfield type. Use something like i16 u8. Note that u64 is not supported but i64 is"
	unsupportedCmdErrMsg := "ERR BITFIELD_RO only supports the GET subcommand"

	testCases := []struct {
		Name     string
		Commands []string
		Expected []interface{}
		Delay    []time.Duration
		CleanUp  []string
	}{
		{
			Name:     "BITFIELD_RO Arity Check",
			Commands: []string{"bitfield_ro"},
			Expected: []interface{}{"ERR wrong number of arguments for 'bitfield_ro' command"},
			Delay:    []time.Duration{0},
			CleanUp:  []string{},
		},
		{
			Name:     "BITFIELD_RO on unsupported type of SET",
			Commands: []string{"SADD bits a b c", "bitfield_ro bits"},
			Expected: []interface{}{int64(3), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:     "BITFIELD_RO on unsupported type of JSON",
			Commands: []string{"json.set bits $ 1", "bitfield_ro bits"},
			Expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:     "BITFIELD_RO on unsupported type of HSET",
			Commands: []string{"HSET bits a 1", "bitfield_ro bits"},
			Expected: []interface{}{int64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name: "BITFIELD_RO with unsupported commands",
			Commands: []string{
				"bitfield_ro bits set u8 0 255",
				"bitfield_ro bits incrby u8 0 100",
			},
			Expected: []interface{}{
				unsupportedCmdErrMsg,
				unsupportedCmdErrMsg,
			},
			Delay:   []time.Duration{0, 0},
			CleanUp: []string{"Del bits"},
		},
		{
			Name: "BITFIELD_RO with syntax error",
			Commands: []string{
				"set bits 1",
				"bitfield_ro bits get u8",
				"bitfield_ro bits get",
				"bitfield_ro bits get somethingrandom",
			},
			Expected: []interface{}{
				"OK",
				syntaxErrMsg,
				syntaxErrMsg,
				syntaxErrMsg,
			},
			Delay:   []time.Duration{0, 0, 0, 0},
			CleanUp: []string{"Del bits"},
		},
		{
			Name: "BITFIELD_RO with invalid bitfield type",
			Commands: []string{
				"set bits 1",
				"bitfield_ro bits get a8 0",
				"bitfield_ro bits get s8 0",
				"bitfield_ro bits get somethingrandom 0",
			},
			Expected: []interface{}{
				"OK",
				bitFieldTypeErrMsg,
				bitFieldTypeErrMsg,
				bitFieldTypeErrMsg,
			},
			Delay:   []time.Duration{0, 0, 0, 0},
			CleanUp: []string{"Del bits"},
		},
		{
			Name:     "BITFIELD_RO with only key as argument",
			Commands: []string{"bitfield_ro bits"},
			Expected: []interface{}{[]interface{}{}},
			Delay:    []time.Duration{0},
			CleanUp:  []string{"DEL bits"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {

			for i := 0; i < len(tc.Commands); i++ {
				if tc.Delay[i] > 0 {
					time.Sleep(tc.Delay[i])
				}
				result := client.FireString(tc.Commands[i])
				expected := tc.Expected[i]
				assert.Equal(t, expected, result)
			}

			for _, cmd := range tc.CleanUp {
				client.FireString(cmd)
			}
		})
	}
}
