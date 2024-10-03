package http

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestBitOp(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		delays   []time.Duration
	}{

		//Should it be int64(0)?
		{
			name: "SETBIT for unitTestKeyA",
			commands: []HTTPCommand{
				{Command: "SETBIT", Body: map[string]interface{}{"key": "unitTestKeyA", "offset": 1, "value": 1}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "unitTestKeyA", "offset": 1, "value": 3}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "unitTestKeyA", "offset": 1, "value": 5}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "unitTestKeyA", "offset": 1, "value": 7}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "unitTestKeyA", "offset": 1, "value": 8}},
			},
			expected: []interface{}{float64(0), float64(0), float64(0), float64(0), float64(0)},
			delays:   []time.Duration{0, 0, 0, 0, 0},
		},
		{
			name: "SETBIT for unitTestKeyB",
			commands: []HTTPCommand{
				{Command: "SETBIT", Body: map[string]interface{}{"key": "unitTestKeyB", "offset": 1, "value": 2}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "unitTestKeyB", "offset": 1, "value": 4}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "unitTestKeyB", "offset": 1, "value": 7}},
			},
			expected: []interface{}{float64(0), float64(0), float64(0)},
			delays:   []time.Duration{0, 0, 0},
		},
		{
			name: "SETBIT for foo",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo", "offset": 1, "value": 2}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo", "offset": 1, "value": 4}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo", "offset": 1, "value": 7}},
				{Command: "GET", Body: map[string]interface{}{"key": "foo"}},
			},
			expected: []interface{}{"OK", float64(1), float64(0), float64(0), "kar"},
			delays:   []time.Duration{0, 0, 0, 0, 0},
		},
		{
			name: "SETBIT for mykey12",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "mykey12", "value": "1343"}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "mykey12", "offset": 1, "value": 2}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "mykey12", "offset": 1, "value": 4}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "mykey12", "offset": 1, "value": 7}},
				{Command: "GET", Body: map[string]interface{}{"key": "mykey12"}},
			},
			expected: []interface{}{"OK", float64(1), float64(0), float64(1), float64(9343)},
			delays:   []time.Duration{0, 0, 0, 0, 0},
		},
		{
			name: "SETBIT for foo12",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo12", "value": "bar"}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo12", "offset": 1, "value": 2}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo12", "offset": 1, "value": 4}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "foo12", "offset": 1, "value": 7}},
				{Command: "GET", Body: map[string]interface{}{"key": "foo12"}},
			},
			expected: []interface{}{"OK", float64(1), float64(0), float64(0), "kar"},
			delays:   []time.Duration{0, 0, 0, 0, 0},
		},
		{
			name: "BITOP NOT",
			commands: []HTTPCommand{
				{Command: "BITOP", Body: map[string]interface{}{"operation": "NOT", "destkey": "unitTestKeyNOT", "key": "unitTestKeyA"}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyNOT", "offset": 1}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyNOT", "offset": 2}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyNOT", "offset": 7}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyNOT", "offset": 8}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyNOT", "offset": 9}},
			},
			expected: []interface{}{float64(2), float64(0), float64(1), float64(0), float64(0), float64(1)},
			delays:   []time.Duration{0, 0, 0, 0, 0, 0},
		},
		{
			name: "BITOP OR",
			commands: []HTTPCommand{
				{Command: "BITOP", Body: map[string]interface{}{"operation": "OR", "destkey": "unitTestKeyOR", "keys": []interface{}{"unitTestKeyB", "unitTestKeyA"}}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyOR", "offset": 1}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyOR", "offset": 2}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyOR", "offset": 3}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyOR", "offset": 7}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyOR", "offset": 8}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyOR", "offset": 9}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyOR", "offset": 12}},
			},
			expected: []interface{}{float64(2), float64(1), float64(1), float64(1), float64(1), float64(1), float64(0), float64(0)},
			delays:   []time.Duration{0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name: "BITOP AND",
			commands: []HTTPCommand{
				{Command: "BITOP", Body: map[string]interface{}{"operation": "AND", "destkey": "unitTestKeyAND", "keys": []interface{}{"unitTestKeyB", "unitTestKeyA"}}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyAND", "offset": 1}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyAND", "offset": 2}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyAND", "offset": 7}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyAND", "offset": 8}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyAND", "offset": 9}},
			},
			expected: []interface{}{float64(2), float64(0), float64(0), float64(1), float64(0), float64(0)},
			delays:   []time.Duration{0, 0, 0, 0, 0, 0},
		},
		{
			name: "BITOP XOR",
			commands: []HTTPCommand{
				{Command: "BITOP", Body: map[string]interface{}{"operation": "XOR", "destkey": "unitTestKeyXOR", "keys": []interface{}{"unitTestKeyB", "unitTestKeyA"}}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyXOR", "offset": 1}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyXOR", "offset": 2}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyXOR", "offset": 3}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyXOR", "offset": 7}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "unitTestKeyXOR", "offset": 8}},
			},
			expected: []interface{}{float64(2), float64(1), float64(1), float64(1), float64(0), float64(1)},
			delays:   []time.Duration{0, 0, 0, 0, 0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}

func TestBitCount(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		delays   []time.Duration
	}{
		{
			name: "Set and GetBit",
			commands: []HTTPCommand{
				{Command: "SETBIT", Body: map[string]interface{}{"key": "mykey", "offset": 1, "value": 7}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "mykey", "offset": 1, "value": 7}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "mykey", "offset": 1, "value": 122}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "mykey", "offset": 122}},
				{Command: "SETBIT", Body: map[string]interface{}{"key": "mykey", "offset": 0, "value": 122}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "mykey", "offset": 122}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "mykey", "offset": "1223232"}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "mykey", "offset": 7}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "mykey", "offset": 8}},
			},
			expected: []interface{}{float64(0), float64(1), float64(0), float64(1), float64(1), float64(0), float64(0), float64(1), float64(0)},
			delays:   []time.Duration{0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name: "BitCount with unit",
			commands: []HTTPCommand{
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "mykey", "start": 1, "end": 7, "unit": "BIT"}},
			},
			expected: []interface{}{float64(1)},
			delays:   []time.Duration{0},
		},
		{
			name: "BitCount",
			commands: []HTTPCommand{
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "mykey", "start": 3, "end": 7}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "mykey", "start": 0, "end": 0}},
				//{Command: "BITCOUNT", Body: nil},//:TODO add tests for ERR wrong number of arguments for 'bitcount' command
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "mykey"}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "mykey", "start": 0}},
			},
			expected: []interface{}{float64(0), float64(1), float64(1), "ERR syntax error"},
			delays:   []time.Duration{0, 0, 0, 0},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

}

func TestBitPos(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	testCases := []struct {
		name         string
		commands     []HTTPCommand
		expected     []interface{}
		delays       []time.Duration
		val          interface{}
		setCmdSETBIT bool
	}{
		{
			name: "BitPos",
			commands: []HTTPCommand{
				{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "bit": 0, "start": 0, "end": -1, "unit": "BIT"}},
				{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "bit": 0, "start": 8, "end": -1, "unit": "BIT"}},
				{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "bit": 0, "start": 16, "end": -1, "unit": "BIT"}},
				{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "bit": 0, "start": 16, "end": 200, "unit": "BIT"}},
				{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "bit": 0, "start": 8, "end": 8, "unit": "BIT"}},
			},
			expected: []interface{}{float64(0), float64(8), float64(16), float64(16), float64(8)},
			delays:   []time.Duration{0, 0, 0, 0, 0},
			val:      "\\x00\\xff\\x00",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setCmdSETBIT {
				exec.FireCommand(HTTPCommand{
					Command: "SETBIT", Body: map[string]interface{}{"key": "testkeysb", "offset": tc.val.(string)},
				})
			} else {
				switch v := tc.val.(type) {
				case string:
					exec.FireCommand(HTTPCommand{
						Command: "SET", Body: map[string]interface{}{"key": "testkey", "value": v},
					})
				case int:
					exec.FireCommand(HTTPCommand{
						Command: "SET", Body: map[string]interface{}{"key": "testkey", "value": v},
					})
				default:
					// For test cases where we don't set a value (e.g., error cases)
				}
			}
			for i, cmd := range tc.commands {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

}
