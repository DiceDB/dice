package http

// The following commands are a part of this test class:
// SETBIT, GETBIT, BITCOUNT, BITOP, BITPOS, BITFIELD, BITFIELD_RO

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	testifyAssert "github.com/stretchr/testify/assert"
	"golang.org/x/exp/rand"
	"gotest.tools/v3/assert"
)

func TestBitOp(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testcases := []struct {
		InCmds []HTTPCommand
		Out    []interface{}
	}{
		{
			InCmds: []HTTPCommand{
				{Command: "SETBIT", Body: map[string]interface{}{"key": "unitTestKeyA", "values": []interface{}{1, 1}}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "unitTestKeyA", "values": []interface{}{3, 1}}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "unitTestKeyA", "values": []interface{}{5, 1}}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "unitTestKeyA", "values": []interface{}{7, 1}}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "unitTestKeyA", "values": []interface{}{8, 1}}},
			},
			Out: []interface{}{float64(0), float64(0), float64(0), float64(0), float64(0)},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "SETBIT", Body: map[string]interface{}{"key": "unitTestKeyB", "values": []interface{}{2, 1}}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "unitTestKeyB", "values": []interface{}{4, 1}}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "unitTestKeyB", "values": []interface{}{7, 1}}},
			},
			Out: []interface{}{float64(0), float64(0), float64(0)},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo", "values": []interface{}{2, 1}}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo", "values": []interface{}{4, 1}}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo", "values": []interface{}{7, 1}}},
				{Command: "GET", Body: map[string]interface{}{"key": "foo"}},
			},
			Out: []interface{}{"OK", float64(1), float64(0), float64(0), "kar"},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "mykey12", "value": "1343"}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "mykey12", "values": []interface{}{2, 1}}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "mykey12", "values": []interface{}{4, 1}}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "mykey12", "values": []interface{}{7, 1}}},
				{Command: "GET", Body: map[string]interface{}{"key": "mykey12"}},
			},
			Out: []interface{}{"OK", float64(1), float64(0), float64(1), float64(9343)},
		},
		{
			InCmds: []HTTPCommand{{Command: "SET", Body: map[string]interface{}{"key": "foo12", "value": "bar"}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo12", "values": []interface{}{2, 1}}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo12", "values": []interface{}{4, 1}}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo12", "values": []interface{}{7, 1}}},
				{Command: "GET", Body: map[string]interface{}{"key": "foo12"}},
			},
			Out: []interface{}{"OK", float64(1), float64(0), float64(0), "kar"},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "BITOP", Body: map[string]interface{}{"values": []interface{}{"NOT", "unitTestKeyNOT", "unitTestKeyA"}}},
			},
			Out: []interface{}{float64(2)},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyNOT", "values": []interface{}{1}}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyNOT", "values": []interface{}{2}}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyNOT", "values": []interface{}{7}}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyNOT", "values": []interface{}{8}}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyNOT", "values": []interface{}{9}}},
			},
			Out: []interface{}{float64(0), float64(1), float64(0), float64(0), float64(1)},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "BITOP", Body: map[string]interface{}{"values": []interface{}{"OR", "unitTestKeyOR", "unitTestKeyB", "unitTestKeyA"}}},
			},
			Out: []interface{}{float64(2)},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyOR", "values": []interface{}{1}}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyOR", "values": []interface{}{2}}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyOR", "values": []interface{}{3}}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyOR", "values": []interface{}{7}}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyOR", "values": []interface{}{8}}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyOR", "values": []interface{}{9}}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyOR", "values": []interface{}{12}}},
			},
			Out: []interface{}{float64(1), float64(1), float64(1), float64(1), float64(1), float64(0), float64(0)},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "BITOP", Body: map[string]interface{}{"values": []interface{}{"AND", "unitTestKeyAND", "unitTestKeyB", "unitTestKeyA"}}},
			},
			Out: []interface{}{float64(2)},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyAND", "values": []interface{}{1}}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyAND", "values": []interface{}{2}}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyAND", "values": []interface{}{7}}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyAND", "values": []interface{}{8}}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyAND", "values": []interface{}{9}}},
			},
			Out: []interface{}{float64(0), float64(0), float64(1), float64(0), float64(0)},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "BITOP", Body: map[string]interface{}{"values": []interface{}{"XOR", "unitTestKeyXOR", "unitTestKeyB", "unitTestKeyA"}}},
			},
			Out: []interface{}{float64(2)},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyXOR", "values": []interface{}{1}}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyXOR", "values": []interface{}{2}}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyXOR", "values": []interface{}{3}}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyXOR", "values": []interface{}{7}}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyXOR", "values": []interface{}{8}}},
			},
			Out: []interface{}{float64(1), float64(1), float64(1), float64(0), float64(1)},
		},
	}

	for _, tcase := range testcases {
		for i := 0; i < len(tcase.InCmds); i++ {
			cmd := tcase.InCmds[i]
			out := tcase.Out[i]
			res, _ := exec.FireCommand(cmd)
			testifyAssert.Equal(t, out, res, "Value mismatch for cmd %s\n.", cmd)
		}
	}
}

func TestBitOpsString(t *testing.T) {

	exec := NewHTTPCommandExecutor()

	// foobar in bits is 01100110 01101111 01101111 01100010 01100001 01110010
	fooBarBits := "011001100110111101101111011000100110000101110010"
	// randomly get 8 bits for testing
	testOffsets := make([]int, 8)

	for i := 0; i < 8; i++ {
		testOffsets[i] = rand.Intn(len(fooBarBits))
	}

	getBitTestCommands := make([]HTTPCommand, 8+1)
	getBitTestExpected := make([]interface{}, 8+1)

	getBitTestCommands[0] = HTTPCommand{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "foobar"}}
	getBitTestExpected[0] = "OK"

	for i := 1; i < 8+1; i++ {
		getBitTestCommands[i] = HTTPCommand{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "value": fmt.Sprintf("%d", testOffsets[i-1])}}
		getBitTestExpected[i] = float64(fooBarBits[testOffsets[i-1]] - '0')
	}

	testCases := []struct {
		name       string
		cmds       []HTTPCommand
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
			name: "Getbit of a key containing an integer",
			cmds: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "10"}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "offset": "0"}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "offset": "1"}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "offset": "2"}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "offset": "3"}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "offset": "4"}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "offset": "5"}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "offset": "6"}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "offset": "7"}},
			},
			expected:   []interface{}{"OK", float64(0), float64(0), float64(1), float64(1), float64(0), float64(0), float64(0), float64(1)},
			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal"},
		},
		{
			name: "Getbit of a key containing an integer 2nd byte",
			cmds: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "10"}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "offset": "8"}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "offset": "9"}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "offset": "10"}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "offset": "11"}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "offset": "12"}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "offset": "13"}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "offset": "14"}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "offset": "15"}},
			},
			expected:   []interface{}{"OK", float64(0), float64(0), float64(1), float64(1), float64(0), float64(0), float64(0), float64(0)},
			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal"},
		},
		{
			name: "Getbit of a key with an offset greater than the length of the string in bits",
			cmds: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "foobar"}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "offset": "100"}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "offset": "48"}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "offset": "47"}},
			},
			expected:   []interface{}{"OK", float64(0), float64(0), float64(0)},
			assertType: []string{"equal", "equal", "equal", "equal"},
		},
		{
			name: "Bitcount of a key containing a string",
			cmds: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "foobar"}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo", "values": []interface{}{0, -1}}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo"}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo", "values": []interface{}{0, 0}}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo", "values": []interface{}{1, 1}}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo", "values": []interface{}{1, 1, "BYTE"}}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo", "values": []interface{}{5, 30, "BIT"}}},
			},
			expected:   []interface{}{"OK", float64(26), float64(26), float64(4), float64(6), float64(6), float64(17)},
			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal"},
		},
		{
			name: "Bitcount of a key containing an integer",
			cmds: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "10"}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo", "values": []interface{}{0, -1}}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo"}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo", "values": []interface{}{0, 0}}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo", "values": []interface{}{1, 1}}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo", "values": []interface{}{1, 1, "BYTE"}}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo", "values": []interface{}{5, 30, "BIT"}}},
			},
			expected:   []interface{}{"OK", float64(5), float64(5), float64(3), float64(2), float64(2), float64(3)},
			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal"},
		},
		{
			name: "Setbit of a key containing a string",
			cmds: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "foobar"}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo", "values": []interface{}{7, 1}}},
				{Command: "GET", Body: map[string]interface{}{"key": "foo"}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo", "values": []interface{}{49, 1}}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo", "values": []interface{}{50, 1}}},
				{Command: "GET", Body: map[string]interface{}{"key": "foo"}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo", "values": []interface{}{49, 0}}},
				{Command: "GET", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{"OK", float64(0), "goobar", float64(0), float64(0), "goobar`", float64(1), "goobar "},
			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal"},
		},
		{
			name: "Setbit of a key must not change the expiry of the key if expiry is set",
			cmds: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "foobar"}},
				{Command: "EXPIRE", Body: map[string]interface{}{"key": "foo", "values": []interface{}{100}}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo", "values": []interface{}{7, 1}}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{"OK", float64(1), float64(100), float64(0), float64(100)},
			assertType: []string{"equal", "equal", "less", "equal", "less"},
		},
		{
			name: "Setbit of a key must not add expiry to the key if expiry is not set",
			cmds: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "foobar"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo", "values": []interface{}{7, 1}}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{"OK", float64(-1), float64(0), float64(-1)},
			assertType: []string{"equal", "equal", "equal", "equal"},
		},
		{
			name: "Bitop not of a key containing a string",
			cmds: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "foobar"}},
				{Command: "BITOP", Body: map[string]interface{}{"values": []interface{}{"NOT", "baz", "foo"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "baz"}},
				{Command: "BITOP", Body: map[string]interface{}{"values": []interface{}{"NOT", "bazz", "baz"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "bazz"}},
			},
			expected:   []interface{}{"OK", float64(6), "\x99\x90\x90\x9d\x9e\x8d", float64(6), "foobar"},
			assertType: []string{"equal", "equal", "equal", "equal", "equal"},
		},
		{
			name: "Bitop not of a key containing an integer",
			cmds: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": 10}},
				{Command: "BITOP", Body: map[string]interface{}{"values": []interface{}{"NOT", "baz", "foo"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "baz"}},
				{Command: "BITOP", Body: map[string]interface{}{"values": []interface{}{"NOT", "bazz", "baz"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "bazz"}},
			},
			expected:   []interface{}{"OK", float64(2), "\xce\xcf", float64(2), float64(10)},
			assertType: []string{"equal", "equal", "equal", "equal", "equal"},
		},
		{
			name: "Get a string created with setbit",
			cmds: []HTTPCommand{
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo", "values": []interface{}{1, 1}}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo", "values": []interface{}{3, 1}}},
				{Command: "GET", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{float64(0), float64(0), "P"},
			assertType: []string{"equal", "equal", "equal"},
		},
		{
			name: "Bitop and of keys containing a string and get the destkey",
			cmds: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "foobar"}},
				{Command: "SET", Body: map[string]interface{}{"key": "baz", "value": "abcdef"}},
				{Command: "BITOP", Body: map[string]interface{}{"values": []interface{}{"AND", "bazz", "foo", "baz"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "bazz"}},
			},
			expected:   []interface{}{"OK", "OK", float64(6), "`bc`ab"},
			assertType: []string{"equal", "equal", "equal", "equal"},
		},
		{
			name: "BITOP AND of keys containing integers and get the destkey",
			cmds: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": 10}},
				{Command: "SET", Body: map[string]interface{}{"key": "baz", "value": 5}},
				{Command: "BITOP", Body: map[string]interface{}{"values": []interface{}{"AND", "bazz", "foo", "baz"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "bazz"}},
			},
			expected:   []interface{}{"OK", "OK", float64(2), "1\x00"},
			assertType: []string{"equal", "equal", "equal", "equal"},
		},
		{
			name: "Bitop or of keys containing a string, a bytearray and get the destkey",
			cmds: []HTTPCommand{
				{Command: "MSET", Body: map[string]interface{}{"keys": []interface{}{"foo", "foobar", "baz", "abcdef"}}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "bazz", "values": []interface{}{8, 1}}},
				{Command: "BITOP", Body: map[string]interface{}{"values": []interface{}{"AND", "bazzz", "foo", "baz", "bazz"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "bazzz"}},
			},
			expected:   []interface{}{"OK", float64(0), float64(6), "\x00\x00\x00\x00\x00\x00"},
			assertType: []string{"equal", "equal", "equal", "equal"},
		},
		{
			name: "BITOP OR of keys containing strings and get the destkey",
			cmds: []HTTPCommand{
				{Command: "MSET", Body: map[string]interface{}{"keys": []interface{}{"foo", "foobar", "baz", "abcdef"}}},
				{Command: "BITOP", Body: map[string]interface{}{"values": []interface{}{"OR", "bazz", "foo", "baz"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "bazz"}},
			},
			expected:   []interface{}{"OK", float64(6), "goofev"},
			assertType: []string{"equal", "equal", "equal"},
		},
		{
			name: "BITOP OR of keys containing integers and get the destkey",
			cmds: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": 10}},
				{Command: "SET", Body: map[string]interface{}{"key": "baz", "value": 5}},
				{Command: "BITOP", Body: map[string]interface{}{"values": []interface{}{"OR", "bazz", "foo", "baz"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "bazz"}},
			},
			expected:   []interface{}{"OK", "OK", float64(2), "50"},
			assertType: []string{"equal", "equal", "equal", "equal"},
		},
		{
			name: "BITOP OR of keys containing strings and a bytearray and get the destkey",
			cmds: []HTTPCommand{
				{Command: "MSET", Body: map[string]interface{}{"keys": []interface{}{"foo", "foobar", "baz", "abcdef"}}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "bazz", "values": []interface{}{8, 1}}},
				{Command: "BITOP", Body: map[string]interface{}{"values": []interface{}{"OR", "bazzz", "foo", "baz", "bazz"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "bazzz"}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "bazz", "values": []interface{}{8, 0}}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "bazz", "values": []interface{}{49, 1}}},
				{Command: "BITOP", Body: map[string]interface{}{"values": []interface{}{"OR", "bazzz", "foo", "baz", "bazz"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "bazzz"}},
			},
			expected:   []interface{}{"OK", float64(0), float64(6), "g\xefofev", float64(1), float64(0), float64(7), "goofev@"},
			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal"},
		},
		{
			name: "BITOP XOR of keys containing strings and get the destkey",
			cmds: []HTTPCommand{
				{Command: "MSET", Body: map[string]interface{}{"keys": []interface{}{"foo", "foobar", "baz", "abcdef"}}},
				{Command: "BITOP", Body: map[string]interface{}{"values": []interface{}{"XOR", "bazz", "foo", "baz"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "bazz"}},
			},
			expected:   []interface{}{"OK", float64(6), "\x07\x0d\x0c\x06\x04\x14"},
			assertType: []string{"equal", "equal", "equal"},
		},
		{
			name: "BITOP XOR of keys containing strings and a bytearray and get the destkey",
			cmds: []HTTPCommand{
				{Command: "MSET", Body: map[string]interface{}{"keys": []interface{}{"foo", "foobar", "baz", "abcdef"}}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "bazz", "values": []interface{}{8, 1}}},
				{Command: "BITOP", Body: map[string]interface{}{"values": []interface{}{"XOR", "bazzz", "foo", "baz", "bazz"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "bazzz"}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "bazz", "values": []interface{}{8, 0}}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "bazz", "values": []interface{}{49, 1}}},
				{Command: "BITOP", Body: map[string]interface{}{"values": []interface{}{"XOR", "bazzz", "foo", "baz", "bazz"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "bazzz"}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "bazz", "values": []interface{}{49, 0}}},
				{Command: "BITOP", Body: map[string]interface{}{"values": []interface{}{"XOR", "bazzz", "foo", "baz", "bazz"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "bazzz"}},
			},
			expected:   []interface{}{"OK", float64(0), float64(6), "\x07\x8d\x0c\x06\x04\x14", float64(1), float64(0), float64(7), "\x07\r\x0c\x06\x04\x14@", float64(1), float64(7), "\x07\r\x0c\x06\x04\x14\x00"},
			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal", "equal"},
		},
		{
			name: "BITOP XOR of keys containing integers and get the destkey",
			cmds: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": 10}},
				{Command: "SET", Body: map[string]interface{}{"key": "baz", "value": 5}},
				{Command: "BITOP", Body: map[string]interface{}{"values": []interface{}{"XOR", "bazz", "foo", "baz"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "bazz"}},
			},
			expected:   []interface{}{"OK", "OK", float64(2), "\x040"},
			assertType: []string{"equal", "equal", "equal", "equal"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Delete the key before running the test
			exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "foo"}})
			exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "baz"}})
			exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "bazz"}})
			exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "bazzz"}})
			for i := 0; i < len(tc.cmds); i++ {
				res, _ := exec.FireCommand(tc.cmds[i])

				switch tc.assertType[i] {
				case "equal":
					testifyAssert.Equal(t, tc.expected[i], res)
				case "less":
					assert.Assert(t, res.(float64) <= tc.expected[i].(float64), "CMD: %s Expected %d to be less than or equal to %d", tc.cmds[i], res, tc.expected[i])
				}
			}
		})
	}
}

func TestBitCount(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	testcases := []struct {
		InCmds []HTTPCommand
		Out    []interface{}
	}{
		{
			InCmds: []HTTPCommand{
				{Command: "SETBIT", Body: map[string]interface{}{"key": "mykey", "values": []interface{}{7, 1}}},
			},
			Out: []interface{}{float64(0)},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "SETBIT", Body: map[string]interface{}{"key": "mykey", "values": []interface{}{7, 1}}},
			},
			Out: []interface{}{float64(1)},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "SETBIT", Body: map[string]interface{}{"key": "mykey", "values": []interface{}{122, 1}}},
			},
			Out: []interface{}{float64(0)},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "GETBIT", Body: map[string]interface{}{"key": "mykey", "values": []interface{}{122}}},
			},
			Out: []interface{}{float64(1)},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "SETBIT", Body: map[string]interface{}{"key": "mykey", "values": []interface{}{122, 0}}},
			},
			Out: []interface{}{float64(1)},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "GETBIT", Body: map[string]interface{}{"key": "mykey", "values": []interface{}{122}}},
			},
			Out: []interface{}{float64(0)},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "GETBIT", Body: map[string]interface{}{"key": "mykey", "value": 1223232}},
			},
			Out: []interface{}{float64(0)},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "GETBIT", Body: map[string]interface{}{"key": "mykey", "values": []interface{}{7}}},
			},
			Out: []interface{}{float64(1)},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "GETBIT", Body: map[string]interface{}{"key": "mykey", "values": []interface{}{8}}},
			},
			Out: []interface{}{float64(0)},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "mykey", "values": []interface{}{3, 7, "BIT"}}},
			},
			Out: []interface{}{float64(1)},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "mykey", "values": []interface{}{3, 7}}},
			},
			Out: []interface{}{float64(0)},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "mykey", "values": []interface{}{0, 0}}},
			},
			Out: []interface{}{float64(1)},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "BITCOUNT"},
			},
			Out: []interface{}{"ERR wrong number of arguments for 'bitcount' command"},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "mykey"}},
			},
			Out: []interface{}{float64(1)},
		},
		{
			InCmds: []HTTPCommand{
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "mykey", "values": []interface{}{0}}},
			},
			Out: []interface{}{"ERR syntax error"},
		},
	}

	for _, tcase := range testcases {
		for i := 0; i < len(tcase.InCmds); i++ {
			cmd := tcase.InCmds[i]
			out := tcase.Out[i]
			res, _ := exec.FireCommand(cmd)
			testifyAssert.Equal(t, out, res, "Value mismatch for cmd %s\n.", cmd)
		}
	}
}

func TestBitPos(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	exec.FireCommand(HTTPCommand{Command: "FLUSHDB"})
	defer exec.FireCommand(HTTPCommand{Command: "FLUSHDB"}) // clean up after all test cases
	testcases := []struct {
		name         string
		val          interface{}
		inCmd        HTTPCommand
		out          interface{}
		setCmdSETBIT bool
	}{
		{
			name:  "String interval BIT 0,-1 ",
			val:   "\\x00\\xff\\x00",
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, 0, -1, "bit"}}},
			out:   float64(0),
		},
		{
			name:  "String interval BIT 8,-1",
			val:   "\\x00\\xff\\x00",
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, 8, -1, "bit"}}},
			out:   float64(8),
		},
		{
			name:  "String interval BIT 16,-1",
			val:   "\\x00\\xff\\x00",
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, 16, -1, "bit"}}},
			out:   float64(16),
		},
		{
			name:  "String interval BIT 16,200",
			val:   "\\x00\\xff\\x00",
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, 16, 200, "bit"}}},
			out:   float64(16),
		},
		{
			name:  "String interval BIT 8,8",
			val:   "\\x00\\xff\\x00",
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, 8, 8, "bit"}}},
			out:   float64(8),
		},
		{
			name:  "FindsFirstZeroBit",
			val:   []byte("\xff\xf0\x00"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "value": 0}},
			out:   float64(12),
		},
		{
			name:  "FindsFirstOneBit",
			val:   []byte("\x00\x0f\xff"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "value": 1}},
			out:   float64(12),
		},
		{
			name:  "NoOneBitFound",
			val:   "\x00\x00\x00",
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "value": 1}},
			out:   float64(-1),
		},
		{
			name:  "NoZeroBitFound",
			val:   []byte("\xff\xff\xff"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "value": 0}},
			out:   float64(24),
		},
		{
			name:  "NoZeroBitFoundWithRangeStartPos",
			val:   []byte("\xff\xff\xff"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, 2}}},
			out:   float64(24),
		},
		{
			name:  "NoZeroBitFoundWithOOBRangeStartPos",
			val:   []byte("\xff\xff\xff"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, 4}}},
			out:   float64(-1),
		},
		{
			name:  "NoZeroBitFoundWithRange",
			val:   []byte("\xff\xff\xff"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, 2, 2}}},
			out:   float64(-1),
		},
		{
			name:  "NoZeroBitFoundWithRangeAndRangeType",
			val:   []byte("\xff\xff\xff"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, 2, 2, "BIT"}}},
			out:   float64(-1),
		},
		{
			name:  "FindsFirstZeroBitInRange",
			val:   []byte("\xff\xf0\xff"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, 1, 2}}},
			out:   float64(12),
		},
		{
			name:  "FindsFirstOneBitInRange",
			val:   "\x00\x00\xf0",
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{1, 2, 3}}},
			out:   float64(16),
		},
		{
			name:  "StartGreaterThanEnd",
			val:   "\xff\xf0\x00",
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, 3, 2}}},
			out:   float64(-1),
		},
		{
			name:  "FindsFirstOneBitWithNegativeStart",
			val:   []byte("\x00\x00\xf0"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{1, -2, -1}}},
			out:   float64(16),
		},
		{
			name:  "FindsFirstZeroBitWithNegativeEnd",
			val:   []byte("\xff\xf0\xff"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, 1, -1}}},
			out:   float64(12),
		},
		{
			name:  "FindsFirstZeroBitInByteRange",
			val:   []byte("\xff\x00\xff"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, 1, 2, "BYTE"}}},
			out:   float64(8),
		},
		{
			name:  "FindsFirstOneBitInBitRange",
			val:   "\x00\x01\x00",
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{1, 0, 16, "BIT"}}},
			out:   float64(15),
		},
		{
			name:  "NoBitFoundInByteRange",
			val:   []byte("\xff\xff\xff"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, 0, 2, "BYTE"}}},
			out:   float64(-1),
		},
		{
			name:  "NoBitFoundInBitRange",
			val:   "\x00\x00\x00",
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{1, 0, 23, "BIT"}}},
			out:   float64(-1),
		},
		{
			name:  "EmptyStringReturnsMinusOneForZeroBit",
			val:   []byte(""),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "value": 0}},
			out:   float64(-1),
		},
		{
			name:  "EmptyStringReturnsMinusOneForOneBit",
			val:   []byte(""),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "value": 1}},
			out:   float64(-1),
		},
		{
			name:  "SingleByteString",
			val:   "\x80",
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "value": 1}},
			out:   float64(0),
		},
		{
			name:  "RangeExceedsStringLength",
			val:   "\x00\xff",
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{1, 0, 20, "BIT"}}},
			out:   float64(8),
		},
		{
			name:  "InvalidBitArgument",
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "value": 2}},
			out:   "ERR the bit argument must be 1 or 0",
		},
		{
			name:  "NonIntegerStartParameter",
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, "start"}}},
			out:   "ERR value is not an integer or out of range",
		},
		{
			name:  "NonIntegerEndParameter",
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, 1, "end"}}},
			out:   "ERR value is not an integer or out of range",
		},
		{
			name:  "InvalidRangeType",
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, 1, 2, "BYTEs"}}},
			out:   "ERR syntax error",
		},
		{
			name:  "InsufficientArguments",
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey"}},
			out:   "ERR wrong number of arguments for 'bitpos' command",
		},
		{
			name:  "NonExistentKeyForZeroBit",
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "nonexistentkey", "value": 0}},
			out:   float64(0),
		},
		{
			name:  "NonExistentKeyForOneBit",
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "nonexistentkey", "value": 1}},
			out:   float64(-1),
		},
		{
			name:  "IntegerValue",
			val:   65280, // 0xFF00 in decimal
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "value": 0}},
			out:   float64(0),
		},
		{
			name:  "LargeIntegerValue",
			val:   16777215, // 0xFFFFFF in decimal
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "value": 1}},
			out:   float64(2),
		},
		{
			name:  "SmallIntegerValue",
			val:   1, // 0x01 in decimal
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "value": 0}},
			out:   float64(0),
		},
		{
			name:  "ZeroIntegerValue",
			val:   0,
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "value": 1}},
			out:   float64(2),
		},
		{
			name:  "BitRangeStartGreaterThanBitLength",
			val:   []byte("\xff\xff\xff"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, 25, 30, "BIT"}}},
			out:   float64(-1),
		},
		{
			name:  "BitRangeEndExceedsBitLength",
			val:   []byte("\xff\xff\xff"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, 0, 30, "BIT"}}},
			out:   float64(-1),
		},
		{
			name:  "NegativeStartInBitRange",
			val:   []byte("\x00\xff\xff"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{1, -16, -1, "BIT"}}},
			out:   float64(8),
		},
		{
			name:  "LargeNegativeStart",
			val:   []byte("\x00\xff\xff"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{1, -100, -1}}},
			out:   float64(8),
		},
		{
			name:  "LargePositiveEnd",
			val:   "\x00\xff\xff",
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{1, 0, 100}}},
			out:   float64(8),
		},
		{
			name:  "StartAndEndEqualInByteRange",
			val:   []byte("\x0f\xff\xff"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, 1, 1, "BYTE"}}},
			out:   float64(-1),
		},
		{
			name:  "StartAndEndEqualInBitRange",
			val:   "\x0f\xff\xff",
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{1, 1, 1, "BIT"}}},
			out:   float64(-1),
		},
		{
			name:  "FindFirstZeroBitInNegativeRange",
			val:   []byte("\xff\x00\xff"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, -2, -1}}},
			out:   float64(8),
		},
		{
			name:  "FindFirstOneBitInNegativeRangeBIT",
			val:   []byte("\x00\x00\x80"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{1, -8, -1, "BIT"}}},
			out:   float64(16),
		},
		{
			name:  "MaxIntegerValue",
			val:   math.MaxInt64,
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "value": 0}},
			out:   float64(0),
		},
		{
			name:  "MinIntegerValue",
			val:   math.MinInt64,
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "value": 1}},
			out:   float64(2),
		},
		{
			name:  "SingleBitStringZero",
			val:   "\x00",
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "value": 1}},
			out:   float64(-1),
		},
		{
			name:  "SingleBitStringOne",
			val:   "\x01",
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "value": 0}},
			out:   float64(0),
		},
		{
			name:  "AllBitsSetExceptLast",
			val:   []byte("\xff\xff\xfe"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "value": 0}},
			out:   float64(23),
		},
		{
			name:  "OnlyLastBitSet",
			val:   "\x00\x00\x01",
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "value": 1}},
			out:   float64(23),
		},
		{
			name:  "AlternatingBitsLongString",
			val:   []byte("\xaa\xaa\xaa\xaa\xaa"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "value": 0}},
			out:   float64(1),
		},
		{
			name:  "VeryLargeByteString",
			val:   []byte(strings.Repeat("\xff", 1000) + "\x00"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "value": 0}},
			out:   float64(8000),
		},
		{
			name:         "FindZeroBitOnSetBitKey",
			val:          []interface{}{8, 1},
			inCmd:        HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkeysb", "value": 1}},
			out:          float64(8),
			setCmdSETBIT: true,
		},
		{
			name:         "FindOneBitOnSetBitKey",
			val:          []interface{}{1, 1},
			inCmd:        HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkeysb", "value": 1}},
			out:          float64(1),
			setCmdSETBIT: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var setCmd HTTPCommand
			if tc.setCmdSETBIT {
				setCmd = HTTPCommand{
					Command: "SETBIT",
					Body:    map[string]interface{}{"key": "testkeysb", "values": tc.val},
				}
			} else {
				switch v := tc.val.(type) {
				case []byte:
					setCmd = HTTPCommand{
						Command: "SET",
						Body:    map[string]interface{}{"key": "testkey", "value": v, "isByteEncodedVal": true},
					}
				case string:
					setCmd = HTTPCommand{
						Command: "SET",
						Body:    map[string]interface{}{"key": "testkey", "value": v},
					}
				case int:
					setCmd = HTTPCommand{
						Command: "SET",
						Body:    map[string]interface{}{"key": "testkey", "value": fmt.Sprintf("%d", v)},
					}
				default:
					// For test cases where we don't set a value (e.g., error cases)
					setCmd = HTTPCommand{Command: ""}
				}
			}

			if setCmd.Command != "" {
				_, _ = exec.FireCommand(setCmd)
			}

			result, _ := exec.FireCommand(tc.inCmd)
			testifyAssert.Equal(t, tc.out, result, "Mismatch for cmd %s\n", tc.inCmd)
		})
	}
}

func TestBitfield(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	exec.FireCommand(HTTPCommand{Command: "FLUSHDB"})
	defer exec.FireCommand(HTTPCommand{Command: "FLUSHDB"}) // clean up after all test cases
	syntaxErrMsg := "ERR syntax error"
	bitFieldTypeErrMsg := "ERR Invalid bitfield type. Use something like i16 u8. Note that u64 is not supported but i64 is"
	integerErrMsg := "ERR value is not an integer or out of range"
	overflowErrMsg := "ERR Invalid OVERFLOW type specified"

	testCases := []struct {
		Name     string
		Commands []HTTPCommand
		Expected []interface{}
		Delay    []time.Duration
		CleanUp  []HTTPCommand
	}{
		{
			Name:     "BITFIELD Arity Check",
			Commands: []HTTPCommand{{Command: "BITFIELD"}},
			Expected: []interface{}{"ERR wrong number of arguments for 'bitfield' command"},
			Delay:    []time.Duration{0},
			CleanUp:  []HTTPCommand{},
		},
		{
			Name:     "BITFIELD on unsupported type of SET",
			Commands: []HTTPCommand{{Command: "SADD", Body: map[string]interface{}{"key": "bits", "values": []string{"a", "b", "c"}}}, {Command: "BITFIELD", Body: map[string]interface{}{"key": "bits"}}},
			Expected: []interface{}{float64(3), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "bits"}}},
		},
		{
			Name:     "BITFIELD on unsupported type of JSON",
			Commands: []HTTPCommand{{Command: "json.set", Body: map[string]interface{}{"key": "bits", "path": "$", "value": "1"}}, {Command: "BITFIELD", Body: map[string]interface{}{"key": "bits"}}},
			Expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "bits"}}},
		},
		{
			Name:     "BITFIELD on unsupported type of HSET",
			Commands: []HTTPCommand{{Command: "HSET", Body: map[string]interface{}{"key": "bits", "field": "a", "value": "1"}}, {Command: "BITFIELD", Body: map[string]interface{}{"key": "bits"}}},
			Expected: []interface{}{float64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "bits"}}},
		},
		{
			Name: "BITFIELD with syntax errors",
			Commands: []HTTPCommand{
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"set", "u8", 0, 255, "incrby", "u8", 0, 100, "get", "u8"}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"set", "a8", 0, 255, "incrby", "u8", 0, 100, "get", "u8"}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"set", "u8", "a", 255, "incrby", "u8", 0, 100}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"set", "u8", 0, 255, "incrby", "u8", 0, 100, "overflow", "wraap"}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"set", "u8", 0, "incrby", "u8", 0, 100, "get", "u8", 288}}},
			},
			Expected: []interface{}{
				syntaxErrMsg,
				bitFieldTypeErrMsg,
				"ERR bit offset is not an integer or out of range",
				overflowErrMsg,
				integerErrMsg,
			},
			Delay:   []time.Duration{0, 0, 0, 0, 0},
			CleanUp: []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "bits"}}},
		},
		{
			Name: "BITFIELD signed SET and GET basics",
			Commands: []HTTPCommand{
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"set", "i8", 0, -100}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"set", "i8", 0, 101}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"get", "i8", 0}}},
			},
			Expected: []interface{}{[]interface{}{float64(0)}, []interface{}{float64(-100)}, []interface{}{float64(101)}},
			Delay:    []time.Duration{0, 0, 0},
			CleanUp:  []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "bits"}}},
		},
		{
			Name: "BITFIELD unsigned SET and GET basics",
			Commands: []HTTPCommand{
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"set", "u8", 0, 255}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"set", "u8", 0, 100}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"get", "u8", 0}}},
			},
			Expected: []interface{}{[]interface{}{float64(0)}, []interface{}{float64(255)}, []interface{}{float64(100)}},
			Delay:    []time.Duration{0, 0, 0},
			CleanUp:  []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "bits"}}},
		},
		{
			Name:     "BITFIELD signed SET and GET together",
			Commands: []HTTPCommand{{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"set", "i8", 0, 255, "set", "i8", 0, 100, "get", "i8", 0}}}},
			Expected: []interface{}{[]interface{}{float64(0), float64(-1), float64(100)}},
			Delay:    []time.Duration{0},
			CleanUp:  []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "bits"}}},
		},
		{
			Name:     "BITFIELD unsigned with SET, GET and INCRBY arguments",
			Commands: []HTTPCommand{{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"set", "u8", 0, 255, "incrby", "u8", 0, 100, "get", "u8", 0}}}},
			Expected: []interface{}{[]interface{}{float64(0), float64(99), float64(99)}},
			Delay:    []time.Duration{0},
			CleanUp:  []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "bits"}}},
		},
		{
			Name:     "BITFIELD with only key as argument",
			Commands: []HTTPCommand{{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits"}}},
			Expected: []interface{}{[]interface{}{}},
			Delay:    []time.Duration{0},
			CleanUp:  []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "bits"}}},
		},
		{
			Name: "BITFIELD #<idx> form",
			Commands: []HTTPCommand{
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"set", "u8", "#0", 65}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"set", "u8", "#1", 66}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"set", "u8", "#2", 67}}},
				{Command: "GET", Body: map[string]interface{}{"key": "bits"}},
			},
			Expected: []interface{}{[]interface{}{float64(0)}, []interface{}{float64(0)}, []interface{}{float64(0)}, "ABC"},
			Delay:    []time.Duration{0, 0, 0, 0},
			CleanUp:  []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "bits"}}},
		},
		{
			Name: "BITFIELD basic INCRBY form",
			Commands: []HTTPCommand{
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"set", "u8", "#0", 10}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"incrby", "u8", "#0", 100}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"incrby", "u8", "#0", 100}}},
			},
			Expected: []interface{}{[]interface{}{float64(0)}, []interface{}{float64(110)}, []interface{}{float64(210)}},
			Delay:    []time.Duration{0, 0, 0},
			CleanUp:  []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "bits"}}},
		},
		{
			Name: "BITFIELD chaining of multiple commands",
			Commands: []HTTPCommand{
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"set", "u8", "#0", 10}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"incrby", "u8", "#0", 100, "incrby", "u8", "#0", 100}}},
			},
			Expected: []interface{}{[]interface{}{float64(0)}, []interface{}{float64(110), float64(210)}},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "bits"}}},
		},
		{
			Name: "BITFIELD unsigned overflow wrap",
			Commands: []HTTPCommand{
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"set", "u8", "#0", 100}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"overflow", "wrap", "incrby", "u8", "#0", 257}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"get", "u8", "#0"}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"overflow", "wrap", "incrby", "u8", "#0", 255}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"get", "u8", "#0"}}},
			},
			Expected: []interface{}{
				[]interface{}{float64(0)},
				[]interface{}{float64(101)},
				[]interface{}{float64(101)},
				[]interface{}{float64(100)},
				[]interface{}{float64(100)},
			},
			Delay:   []time.Duration{0, 0, 0, 0, 0},
			CleanUp: []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "bits"}}},
		},
		{
			Name: "BITFIELD unsigned overflow sat",
			Commands: []HTTPCommand{
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"set", "u8", "#0", 100}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"overflow", "sat", "incrby", "u8", "#0", 257}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"get", "u8", "#0"}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"overflow", "sat", "incrby", "u8", "#0", -255}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"get", "u8", "#0"}}},
			},
			Expected: []interface{}{
				[]interface{}{float64(0)},
				[]interface{}{float64(255)},
				[]interface{}{float64(255)},
				[]interface{}{float64(0)},
				[]interface{}{float64(0)},
			},
			Delay:   []time.Duration{0, 0, 0, 0, 0},
			CleanUp: []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "bits"}}},
		},
		{
			Name: "BITFIELD signed overflow wrap",
			Commands: []HTTPCommand{
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"set", "i8", "#0", 100}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"overflow", "wrap", "incrby", "i8", "#0", 257}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"get", "i8", "#0"}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"overflow", "wrap", "incrby", "i8", "#0", 255}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"get", "i8", "#0"}}},
			},
			Expected: []interface{}{
				[]interface{}{float64(0)},
				[]interface{}{float64(101)},
				[]interface{}{float64(101)},
				[]interface{}{float64(100)},
				[]interface{}{float64(100)},
			},
			Delay:   []time.Duration{0, 0, 0, 0, 0},
			CleanUp: []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "bits"}}},
		},
		{
			Name: "BITFIELD signed overflow sat",
			Commands: []HTTPCommand{
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"set", "u8", "#0", 100}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"overflow", "sat", "incrby", "i8", "#0", 257}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"get", "i8", "#0"}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"overflow", "sat", "incrby", "i8", "#0", -255}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"get", "i8", "#0"}}},
			},
			Expected: []interface{}{
				[]interface{}{float64(0)},
				[]interface{}{float64(127)},
				[]interface{}{float64(127)},
				[]interface{}{float64(-128)},
				[]interface{}{float64(-128)},
			},
			Delay:   []time.Duration{0, 0, 0, 0, 0},
			CleanUp: []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "bits"}}},
		},
		{
			Name:     "BITFIELD regression 1",
			Commands: []HTTPCommand{{Command: "SET", Body: map[string]interface{}{"key": "bits", "value": "1"}}, {Command: "BITFIELD", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"get", "u1", 0}}}},
			Expected: []interface{}{"OK", []interface{}{float64(0)}},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "bits"}}},
		},
		{
			Name: "BITFIELD regression 2",
			Commands: []HTTPCommand{
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "mystring", "values": []interface{}{"set", "i8", 0, 10}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "mystring", "values": []interface{}{"set", "i8", 64, 10}}},
				{Command: "BITFIELD", Body: map[string]interface{}{"key": "mystring", "values": []interface{}{"incrby", "i8", 10, 99900}}},
			},
			Expected: []interface{}{[]interface{}{float64(0)}, []interface{}{float64(0)}, []interface{}{float64(60)}},
			Delay:    []time.Duration{0, 0, 0},
			CleanUp:  []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "mystring"}}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {

			for i := 0; i < len(tc.Commands); i++ {
				if tc.Delay[i] > 0 {
					time.Sleep(tc.Delay[i])
				}
				result, _ := exec.FireCommand(tc.Commands[i])
				expected := tc.Expected[i]
				testifyAssert.Equal(t, expected, result)
			}

			for _, cmd := range tc.CleanUp {
				exec.FireCommand(cmd)
			}
		})
	}
}

func TestBitfieldRO(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	exec.FireCommand(HTTPCommand{Command: "FLUSHDB"})
	defer exec.FireCommand(HTTPCommand{Command: "FLUSHDB"})

	syntaxErrMsg := "ERR syntax error"
	bitFieldTypeErrMsg := "ERR Invalid bitfield type. Use something like i16 u8. Note that u64 is not supported but i64 is"
	unsupportedCmdErrMsg := "ERR BITFIELD_RO only supports the GET subcommand"

	testCases := []struct {
		Name     string
		Commands []HTTPCommand
		Expected []interface{}
		Delay    []time.Duration
		CleanUp  []HTTPCommand
	}{
		{
			Name:     "BITFIELD_RO Arity Check",
			Commands: []HTTPCommand{{Command: "BITFIELD_RO"}},
			Expected: []interface{}{"ERR wrong number of arguments for 'bitfield_ro' command"},
			Delay:    []time.Duration{0},
			CleanUp:  []HTTPCommand{},
		},
		{
			Name:     "BITFIELD_RO on unsupported type of SET",
			Commands: []HTTPCommand{{Command: "SADD", Body: map[string]interface{}{"key": "bits", "values": []string{"a", "b", "c"}}}, {Command: "BITFIELD_RO", Body: map[string]interface{}{"key": "bits"}}},
			Expected: []interface{}{float64(3), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "bits"}}},
		},
		{
			Name:     "BITFIELD_RO on unsupported type of JSON",
			Commands: []HTTPCommand{{Command: "JSON.SET", Body: map[string]interface{}{"key": "bits", "path": "$", "value": "1"}}, {Command: "BITFIELD_RO", Body: map[string]interface{}{"key": "bits"}}},
			Expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "bits"}}},
		},
		{
			Name:     "BITFIELD_RO on unsupported type of HSET",
			Commands: []HTTPCommand{{Command: "HSET", Body: map[string]interface{}{"key": "bits", "field": "a", "value": "1"}}, {Command: "BITFIELD_RO", Body: map[string]interface{}{"key": "bits"}}},
			Expected: []interface{}{float64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "bits"}}},
		},
		{
			Name: "BITFIELD_RO with unsupported commands",
			Commands: []HTTPCommand{
				{Command: "BITFIELD_RO", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"set", "u8", 0, 255}}},
				{Command: "BITFIELD_RO", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"incrby", "u8", 0, 100}}},
			},
			Expected: []interface{}{
				unsupportedCmdErrMsg,
				unsupportedCmdErrMsg,
			},
			Delay:   []time.Duration{0, 0},
			CleanUp: []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "bits"}}},
		},
		{
			Name: "BITFIELD_RO with syntax error",
			Commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "bits", "value": "1"}},
				{Command: "BITFIELD_RO", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"get", "u8"}}},
				{Command: "BITFIELD_RO", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"get"}}},
				{Command: "BITFIELD_RO", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"get", "somethingrandom"}}},
			},
			Expected: []interface{}{
				"OK",
				syntaxErrMsg,
				syntaxErrMsg,
				syntaxErrMsg,
			},
			Delay:   []time.Duration{0, 0, 0, 0},
			CleanUp: []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "bits"}}},
		},
		{
			Name: "BITFIELD_RO with invalid bitfield type",
			Commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "bits", "value": "1"}},
				{Command: "BITFIELD_RO", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"get", "a8", 0}}},
				{Command: "BITFIELD_RO", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"get", "s8", 0}}},
				{Command: "BITFIELD_RO", Body: map[string]interface{}{"key": "bits", "values": []interface{}{"get", "somethingrandom", 0}}},
			},
			Expected: []interface{}{
				"OK",
				bitFieldTypeErrMsg,
				bitFieldTypeErrMsg,
				bitFieldTypeErrMsg,
			},
			Delay:   []time.Duration{0, 0, 0, 0},
			CleanUp: []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "bits"}}},
		},
		{
			Name:     "BITFIELD_RO with only key as argument",
			Commands: []HTTPCommand{{Command: "BITFIELD_RO", Body: map[string]interface{}{"key": "bits"}}},
			Expected: []interface{}{[]interface{}{}},
			Delay:    []time.Duration{0},
			CleanUp:  []HTTPCommand{{Command: "DEL", Body: map[string]interface{}{"key": "bits"}}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {

			for i := 0; i < len(tc.Commands); i++ {
				if tc.Delay[i] > 0 {
					time.Sleep(tc.Delay[i])
				}
				result, _ := exec.FireCommand(tc.Commands[i])
				expected := tc.Expected[i]
				testifyAssert.Equal(t, expected, result)
			}

			for _, cmd := range tc.CleanUp {
				_, _ = exec.FireCommand(cmd)
			}
		})
	}
}
