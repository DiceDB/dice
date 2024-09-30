package http

import (
	"fmt"
	"testing"
	"time"

	"golang.org/x/exp/rand"
	"gotest.tools/v3/assert"
)

func TestBitOpsString(t *testing.T) {
	fooBarBits := "011001100110111101101111011000100110000101110010"
	// randomly get 8 bits for testing
	testOffsets := make([]int, 8)

	for i := 0; i < 8; i++ {
		testOffsets[i] = rand.Intn(len(fooBarBits))
	}
	getBitTestCommands := make([]string, 8+1)
	getBitTestCommandsArgs := make([]map[string]interface{}, 8+1)
	getBitTestExpected := make([]interface{}, 8+1)

	getBitTestCommands[0] = "SET"
	getBitTestCommandsArgs[0] = map[string]interface{}{"key": "foo", "value": "foobar"}
	getBitTestExpected[0] = "OK"

	for i := 1; i < 8+1; i++ {
		getBitTestCommandsArgs[i] = map[string]interface{}{"key": "foo", "value": fmt.Sprintf("%d", testOffsets[i-1])}
		getBitTestCommands[i] = "GETBIT"
		getBitTestExpected[i] = float64(fooBarBits[testOffsets[i-1]] - '0')
	}

	httpCommandsForString := make([]HTTPCommand, 8+1)
	delays := make([]time.Duration, 8+1)
	for i := 0; i < 8+1; i++ {
		httpCommandsForString[i] = HTTPCommand{Command: getBitTestCommands[i], Body: getBitTestCommandsArgs[i]}
	}

	exec := NewHTTPCommandExecutor()
	testCases := []struct {
		name       string
		commands   []HTTPCommand
		expected   []interface{}
		delays     []time.Duration
		assertType []string
	}{
		{
			name:     "Getbit of a key containing a string",
			commands: httpCommandsForString,
			expected: getBitTestExpected,
			delays:   delays,
		},
		{
			name: "Getbit of a key containing an integer",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": 10}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "value": 0}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "value": 1}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "value": 2}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "value": 3}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "value": 4}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "value": 5}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "value": 6}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "value": 7}},
			},
			expected: []interface{}{"OK", float64(0), float64(0), float64(1), float64(1), float64(0), float64(0), float64(0), float64(1)},
			delays:   delays,
		},
		{
			name: "Getbit of a key containing an integer 2nd byte",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": 10}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "value": 8}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "value": 9}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "value": 10}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "value": 11}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "value": 12}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "value": 13}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "value": 14}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "value": 15}},
			},
			expected: []interface{}{"OK", float64(0), float64(0), float64(1), float64(1), float64(0), float64(0), float64(0), float64(0)},
			delays:   delays,
		},
		{
			name: "Getbit of a key with an offset greater than the length of the string in bits",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "foobar"}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "value": 100}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "value": 48}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "value": 47}},
			},
			expected: []interface{}{"OK", float64(0), float64(0), float64(0)},
			delays:   delays,
		},
		{
			name: "Bitcount of a key containing a string",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "foobar"}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo", "start": 0, "end": -1}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo"}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo", "start": 0, "end": 0}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo", "start": 1, "end": 1}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo", "start": 1, "end": 1, "unit": "BYTE"}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo", "start": 5, "end": 30, "unit": "BIT"}},
			},
			expected: []interface{}{"OK", float64(26), float64(26), float64(4), float64(6), float64(6), float64(17)},
			delays:   delays,
		},
		{
			name: "Bitcount of a key containing a integer",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": 10}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo", "start": 0, "end": -1}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo"}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo", "start": 0, "end": 0}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo", "start": 1, "end": 1}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo", "start": 1, "end": 1, "unit": "BYTE"}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo", "start": 5, "end": 30, "unit": "BIT"}},
			},
			expected: []interface{}{"OK", float64(5), float64(5), float64(3), float64(2), float64(2), float64(3)},
			delays:   delays,
		},
		{
			name: "Setbit of a key containing a string",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "foobar"}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo", "offset": 1, "value": 7}},
				{Command: "GET", Body: map[string]interface{}{"key": "foo"}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo", "offset": 1, "value": 49}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo", "offset": 1, "value": 50}},
				{Command: "GET", Body: map[string]interface{}{"key": "foo"}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo", "offset": 0, "value": 49}},
				{Command: "GET", Body: map[string]interface{}{"key": "foo"}},
			},
			expected: []interface{}{"OK", float64(0), "goobar", float64(0), float64(0), "goobar`", float64(1), "goobar "},
			delays:   delays,
		},
		{
			name: "Setbit of a key must not change the expiry of the key if expiry is set", //:TODO is this working as expected?
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "foobar"}},
				{Command: "EXPIRE", Body: map[string]interface{}{"key": "foo", "value": 100}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo", "offset": 0, "value": 7}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{"OK", float64(1), float64(100), float64(0), float64(100)},
			delays:     delays,
			assertType: []string{"equal", "equal", "less", "equal", "less"},
		},
		{
			name: "Setbit of a key must not add expiry to the key if expiry is not set",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "foobar"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo", "offset": 1, "value": 7}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
			},
			expected: []interface{}{"OK", float64(-1), float64(0), float64(-1)},
			delays:   delays,
		},
		{
			name: "Bitop not of a key containing a string",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "foobar"}},
				{Command: "BITOP", Body: map[string]interface{}{"operation": "NOT", "destkey": "baz", "key": "foo"}},
				{Command: "GET", Body: map[string]interface{}{"key": "baz"}},
				// {Command: "BITOP", Body: map[string]interface{}{"operation": "NOT", "destkey": "baz", "key": "baz"}},
				// {Command: "GET", Body: map[string]interface{}{"key": "bazz"}},
			},
			expected: []interface{}{"OK", float64(6), "\x99\x90\x90\x9d\x9e\x8d", float64(6), "foobar"},
			delays:   delays,
		},
		{
			name: "Bitop not of a key containing an integer",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": 10}},
				{Command: "BITOP", Body: map[string]interface{}{"operation": "NOT", "destkey": "baz", "key": "foo"}},
				{Command: "GET", Body: map[string]interface{}{"key": "bazz"}},
				{Command: "BITOP", Body: map[string]interface{}{"operation": "NOT", "destkey": "baz", "key": "baz"}},
				{Command: "GET", Body: map[string]interface{}{"key": "bazz"}},
			},
			expected: []interface{}{"OK", float64(2), "\xce\xcf", float64(2), float64(10)},
			delays:   delays,
		},
		{
			name: "Get a string created with setbit",
			commands: []HTTPCommand{
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo", "offset": 1, "value": 1}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo", "offset": 1, "value": 3}},
				{Command: "GET", Body: map[string]interface{}{"key": "foo"}},
			},
			expected: []interface{}{float64(0), float64(0), "P"},
			delays:   delays,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "foo"}})
			exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "baz"}})
			exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "bazz"}})
			exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "bazzz"}})
			for i, cmd := range tc.commands {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result, _ := exec.FireCommand(cmd)
				if len(tc.assertType) == 0 {
					assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s", cmd)
				} else {
					switch tc.assertType[i] {
					case "equal":
						assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s", cmd)
					case "less":
						assert.Assert(t, tc.expected[i].(float64) >= result.(float64), "CMD: %s Expected %d to be less than or equal to %d", cmd, result, tc.expected[i])
					}

				}
			}
		})
	}
}
