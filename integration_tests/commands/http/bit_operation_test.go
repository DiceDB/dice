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
		// {
		// 	InCmds: []string{"BITOP NOT unitTestKeyNOT unitTestKeyA "},
		// 	Out:    []interface{}{int64(2)},
		// },
		// {
		// 	InCmds: []string{"GETBIT unitTestKeyNOT 1", "GETBIT unitTestKeyNOT 2", "GETBIT unitTestKeyNOT 7", "GETBIT unitTestKeyNOT 8", "GETBIT unitTestKeyNOT 9"},
		// 	Out:    []interface{}{int64(0), int64(1), int64(0), int64(0), int64(1)},
		// },
		// {
		// 	InCmds: []string{"BITOP OR unitTestKeyOR unitTestKeyB unitTestKeyA"},
		// 	Out:    []interface{}{int64(2)},
		// },
		// {
		// 	InCmds: []string{"GETBIT unitTestKeyOR 1", "GETBIT unitTestKeyOR 2", "GETBIT unitTestKeyOR 3", "GETBIT unitTestKeyOR 7", "GETBIT unitTestKeyOR 8", "GETBIT unitTestKeyOR 9", "GETBIT unitTestKeyOR 12"},
		// 	Out:    []interface{}{int64(1), int64(1), int64(1), int64(1), int64(1), int64(0), int64(0)},
		// },
		// {
		// 	InCmds: []string{"BITOP AND unitTestKeyAND unitTestKeyB unitTestKeyA"},
		// 	Out:    []interface{}{int64(2)},
		// },
		// {
		// 	InCmds: []string{"GETBIT unitTestKeyAND 1", "GETBIT unitTestKeyAND 2", "GETBIT unitTestKeyAND 7", "GETBIT unitTestKeyAND 8", "GETBIT unitTestKeyAND 9"},
		// 	Out:    []interface{}{int64(0), int64(0), int64(1), int64(0), int64(0)},
		// },
		// {
		// 	InCmds: []string{"BITOP XOR unitTestKeyXOR unitTestKeyB unitTestKeyA"},
		// 	Out:    []interface{}{int64(2)},
		// },
		// {
		// 	InCmds: []string{"GETBIT unitTestKeyXOR 1", "GETBIT unitTestKeyXOR 2", "GETBIT unitTestKeyXOR 3", "GETBIT unitTestKeyXOR 7", "GETBIT unitTestKeyXOR 8"},
		// 	Out:    []interface{}{int64(1), int64(1), int64(1), int64(0), int64(1)},
		// },
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
		name     string
		commands []HTTPCommand
		expected []interface{}
		delays   []time.Duration
		value    interface{}
	}{
		{
			name: "BitPos",
			commands: []HTTPCommand{
				{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "bit": 0, "start": 0, "end": -1, "unit": "BIT"}},
			},
			expected: []interface{}{float64(0), float64(8)},
			delays:   []time.Duration{0, 0},
			value:    "\\x00\\xff\\x00",
		},
		{
			name: "BitPos",
			commands: []HTTPCommand{
				{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "bit": 0, "start": 0, "end": -1, "unit": "BIT"}},
			},
			expected: []interface{}{float64(0), float64(8)},
			delays:   []time.Duration{0, 0},
			value:    "\\x00\\xff\\x00",
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
