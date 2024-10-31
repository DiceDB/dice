package async

import (
	"fmt"
	"math"
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBitOp(t *testing.T) {
	conn := getLocalConnection()
	testcases := []struct {
		InCmds []string
		Out    []interface{}
	}{
		{
			InCmds: []string{"SETBIT unitTestKeyA 1 1", "SETBIT unitTestKeyA 3 1", "SETBIT unitTestKeyA 5 1", "SETBIT unitTestKeyA 7 1", "SETBIT unitTestKeyA 8 1"},
			Out:    []interface{}{int64(0), int64(0), int64(0), int64(0), int64(0)},
		},
		{
			InCmds: []string{"SETBIT unitTestKeyB 2 1", "SETBIT unitTestKeyB 4 1", "SETBIT unitTestKeyB 7 1"},
			Out:    []interface{}{int64(0), int64(0), int64(0)},
		},
		{
			InCmds: []string{"SET foo bar", "SETBIT foo 2 1", "SETBIT foo 4 1", "SETBIT foo 7 1", "GET foo"},
			Out:    []interface{}{"OK", int64(1), int64(0), int64(0), "kar"},
		},
		{
			InCmds: []string{"SET mykey12 1343", "SETBIT mykey12 2 1", "SETBIT mykey12 4 1", "SETBIT mykey12 7 1", "GET mykey12"},
			Out:    []interface{}{"OK", int64(1), int64(0), int64(1), int64(9343)},
		},
		{
			InCmds: []string{"SET foo12 bar", "SETBIT foo12 2 1", "SETBIT foo12 4 1", "SETBIT foo12 7 1", "GET foo12"},
			Out:    []interface{}{"OK", int64(1), int64(0), int64(0), "kar"},
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
			assert.Equal(t, out, FireCommand(conn, cmd), "Value mismatch for cmd %s\n.", cmd)
		}
	}
}

func TestBitCount(t *testing.T) {
	conn := getLocalConnection()
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
			assert.Equal(t, out, FireCommand(conn, cmd), "Value mismatch for cmd %s\n.", cmd)
		}
	}
}

func TestBitPos(t *testing.T) {
	conn := getLocalConnection()
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
				FireCommand(conn, setCmd)
			}

			result := FireCommand(conn, tc.inCmd)
			assert.Equal(t, tc.out, result, "Mismatch for cmd %s\n", tc.inCmd)
		})
	}
}

func generateSetBitCommand(connection net.Conn, bitPosition int) int64 {
	command := fmt.Sprintf("SETBIT unitTestKeyA %d 1", bitPosition)
	responseValue := FireCommand(connection, command)
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
	responseValue := FireCommand(connection, command)
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
