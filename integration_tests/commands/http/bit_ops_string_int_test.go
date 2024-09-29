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
			name:     "Bitcount of a key containing a string",
			commands: []HTTPCommand{},
			expected: []interface{}{"OK", float64(0), float64(0), float64(0)},
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
					}
				}
			}
		})
	}
}
