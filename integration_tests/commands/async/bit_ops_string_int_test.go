package async

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBitOpsString(t *testing.T) {
	// test code

	conn := getLocalConnection()
	defer conn.Close()
	// foobar in bits is 01100110 01101111 01101111 01100010 01100001 01110010
	fooBarBits := "011001100110111101101111011000100110000101110010"
	// randomly get 8 bits for testing
	testOffsets := make([]int, 8)

	for i := 0; i < 8; i++ {
		testOffsets[i] = rand.Intn(len(fooBarBits))
	}

	getBitTestCommands := make([]string, 8+1)
	getBitTestExpected := make([]interface{}, 8+1)

	getBitTestCommands[0] = "SET foo foobar"
	getBitTestExpected[0] = "OK"

	for i := 1; i < 8+1; i++ {
		getBitTestCommands[i] = fmt.Sprintf("GETBIT foo %d", testOffsets[i-1])
		getBitTestExpected[i] = int64(fooBarBits[testOffsets[i-1]] - '0')
	}

	testCases := []struct {
		name       string
		cmds       []string
		expected   []interface{}
		assertType []string
	}{
		{
			name:       "Getbit of a key containing a string",
			cmds:       getBitTestCommands,
			expected:   getBitTestExpected,
			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal"},
		},
		{
			name:       "Getbit of a key containing an integer",
			cmds:       []string{"SET foo 10", "GETBIT foo 0", "GETBIT foo 1", "GETBIT foo 2", "GETBIT foo 3", "GETBIT foo 4", "GETBIT foo 5", "GETBIT foo 6", "GETBIT foo 7"},
			expected:   []interface{}{"OK", int64(0), int64(0), int64(1), int64(1), int64(0), int64(0), int64(0), int64(1)},
			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal"},
		}, {
			name:       "Getbit of a key containing an integer 2nd byte",
			cmds:       []string{"SET foo 10", "GETBIT foo 8", "GETBIT foo 9", "GETBIT foo 10", "GETBIT foo 11", "GETBIT foo 12", "GETBIT foo 13", "GETBIT foo 14", "GETBIT foo 15"},
			expected:   []interface{}{"OK", int64(0), int64(0), int64(1), int64(1), int64(0), int64(0), int64(0), int64(0)},
			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal"},
		},
		{
			name:       "Getbit of a key with an offset greater than the length of the string in bits",
			cmds:       []string{"SET foo foobar", "GETBIT foo 100", "GETBIT foo 48", "GETBIT foo 47"},
			expected:   []interface{}{"OK", int64(0), int64(0), int64(0)},
			assertType: []string{"equal", "equal", "equal", "equal"},
		},
		{
			name:       "Bitcount of a key containing a string",
			cmds:       []string{"SET foo foobar", "BITCOUNT foo 0 -1", "BITCOUNT foo", "BITCOUNT foo 0 0", "BITCOUNT foo 1 1", "BITCOUNT foo 1 1 Byte", "BITCOUNT foo 5 30 BIT"},
			expected:   []interface{}{"OK", int64(26), int64(26), int64(4), int64(6), int64(6), int64(17)},
			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal"},
		},
		{
			name:       "Bitcount of a key containing an integer",
			cmds:       []string{"SET foo 10", "BITCOUNT foo 0 -1", "BITCOUNT foo", "BITCOUNT foo 0 0", "BITCOUNT foo 1 1", "BITCOUNT foo 1 1 Byte", "BITCOUNT foo 5 30 BIT"},
			expected:   []interface{}{"OK", int64(5), int64(5), int64(3), int64(2), int64(2), int64(3)},
			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal"},
		},
		{
			name:       "Setbit of a key containing a string",
			cmds:       []string{"SET foo foobar", "setbit foo 7 1", "get foo", "setbit foo 49 1", "setbit foo 50 1", "get foo", "setbit foo 49 0", "get foo"},
			expected:   []interface{}{"OK", int64(0), "goobar", int64(0), int64(0), "goobar`", int64(1), "goobar "},
			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal"},
		},
		{
			name:       "Setbit of a key must not change the expiry of the key if expiry is set",
			cmds:       []string{"SET foo foobar", "EXPIRE foo 100", "TTL foo", "SETBIT foo 7 1", "TTL foo"},
			expected:   []interface{}{"OK", int64(1), int64(100), int64(0), int64(100)},
			assertType: []string{"equal", "equal", "less", "equal", "less"},
		},
		{
			name:       "Setbit of a key must not add expiry to the key if expiry is not set",
			cmds:       []string{"SET foo foobar", "TTL foo", "SETBIT foo 7 1", "TTL foo"},
			expected:   []interface{}{"OK", int64(-1), int64(0), int64(-1)},
			assertType: []string{"equal", "equal", "equal", "equal"},
		},
		{
			name:       "Bitop not of a key containing a string",
			cmds:       []string{"SET foo foobar", "BITOP NOT baz foo", "GET baz", "BITOP NOT bazz baz", "GET bazz"},
			expected:   []interface{}{"OK", int64(6), "\x99\x90\x90\x9d\x9e\x8d", int64(6), "foobar"},
			assertType: []string{"equal", "equal", "equal", "equal", "equal"},
		},
		{
			name:       "Bitop not of a key containing an integer",
			cmds:       []string{"SET foo 10", "BITOP NOT baz foo", "GET baz", "BITOP NOT bazz baz", "GET bazz"},
			expected:   []interface{}{"OK", int64(2), "\xce\xcf", int64(2), int64(10)},
			assertType: []string{"equal", "equal", "equal", "equal", "equal"},
		},
		{
			name:       "Get a string created with setbit",
			cmds:       []string{"SETBIT foo 1 1", "SETBIT foo 3 1", "GET foo"},
			expected:   []interface{}{int64(0), int64(0), "P"},
			assertType: []string{"equal", "equal", "equal"},
		},
		{
			name:       "Bitop and of keys containing a string and get the destkey",
			cmds:       []string{"SET foo foobar", "SET baz abcdef", "BITOP AND bazz foo baz", "GET bazz"},
			expected:   []interface{}{"OK", "OK", int64(6), "`bc`ab"},
			assertType: []string{"equal", "equal", "equal", "equal"},
		},
		{
			name:       "BITOP AND of keys containing integers and get the destkey",
			cmds:       []string{"SET foo 10", "SET baz 5", "BITOP AND bazz foo baz", "GET bazz"},
			expected:   []interface{}{"OK", "OK", int64(2), "1\x00"},
			assertType: []string{"equal", "equal", "equal", "equal"},
		},
		{
			name:       "Bitop or of keys containing a string, a bytearray and get the destkey",
			cmds:       []string{"MSET foo foobar baz abcdef", "SETBIT bazz 8 1", "BITOP and bazzz foo baz bazz", "GET bazzz"},
			expected:   []interface{}{"OK", int64(0), int64(6), "\x00\x00\x00\x00\x00\x00"},
			assertType: []string{"equal", "equal", "equal", "equal"},
		},
		{
			name:       "BITOP OR of keys containing strings and get the destkey",
			cmds:       []string{"MSET foo foobar baz abcdef", "BITOP OR bazz foo baz", "GET bazz"},
			expected:   []interface{}{"OK", int64(6), "goofev"},
			assertType: []string{"equal", "equal", "equal"},
		},
		{
			name:       "BITOP OR of keys containing integers and get the destkey",
			cmds:       []string{"SET foo 10", "SET baz 5", "BITOP OR bazz foo baz", "GET bazz"},
			expected:   []interface{}{"OK", "OK", int64(2), "50"},
			assertType: []string{"equal", "equal", "equal", "equal"},
		},
		{
			name:       "BITOP OR of keys containing strings and a bytearray and get the destkey",
			cmds:       []string{"MSET foo foobar baz abcdef", "SETBIT bazz 8 1", "BITOP OR bazzz foo baz bazz", "GET bazzz", "SETBIT bazz 8 0", "SETBIT bazz 49 1", "BITOP OR bazzz foo baz bazz", "GET bazzz"},
			expected:   []interface{}{"OK", int64(0), int64(6), "g\xefofev", int64(1), int64(0), int64(7), "goofev@"},
			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal"},
		},
		{
			name:       "BITOP XOR of keys containing strings and get the destkey",
			cmds:       []string{"MSET foo foobar baz abcdef", "BITOP XOR bazz foo baz", "GET bazz"},
			expected:   []interface{}{"OK", int64(6), "\x07\x0d\x0c\x06\x04\x14"},
			assertType: []string{"equal", "equal", "equal"},
		},
		{
			name:       "BITOP XOR of keys containing strings and a bytearray and get the destkey",
			cmds:       []string{"MSET foo foobar baz abcdef", "SETBIT bazz 8 1", "BITOP XOR bazzz foo baz bazz", "GET bazzz", "SETBIT bazz 8 0", "SETBIT bazz 49 1", "BITOP XOR bazzz foo baz bazz", "GET bazzz", "Setbit bazz 49 0", "bitop xor bazzz foo baz bazz", "get bazzz"},
			expected:   []interface{}{"OK", int64(0), int64(6), "\x07\x8d\x0c\x06\x04\x14", int64(1), int64(0), int64(7), "\x07\r\x0c\x06\x04\x14@", int64(1), int64(7), "\x07\r\x0c\x06\x04\x14\x00"},
			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal"},
		},
		{
			name:       "BITOP XOR of keys containing integers and get the destkey",
			cmds:       []string{"SET foo 10", "SET baz 5", "BITOP XOR bazz foo baz", "GET bazz"},
			expected:   []interface{}{"OK", "OK", int64(2), "\x040"},
			assertType: []string{"equal", "equal", "equal", "equal"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Delete the key before running the test
			FireCommand(conn, "DEL foo")
			FireCommand(conn, "DEL baz")
			FireCommand(conn, "DEL bazz")
			FireCommand(conn, "DEL bazzz")
			for i := 0; i < len(tc.cmds); i++ {
				res := FireCommand(conn, tc.cmds[i])

				switch tc.assertType[i] {
				case "equal":
					assert.Equal(t, res, tc.expected[i])
				case "less":
					assert.True(t, res.(int64) <= tc.expected[i].(int64), "CMD: %s Expected %d to be less than or equal to %d", tc.cmds[i], res, tc.expected[i])
				}
			}
		})
	}
}
