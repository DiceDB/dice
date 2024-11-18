package http

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHScan(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		delays   []time.Duration
	}{
		{
			name: "HSCAN with wrong number of arguments",
			commands: []HTTPCommand{
				{Command: "HSCAN", Body: map[string]interface{}{"key": "KEY"}},
			},
			expected: []interface{}{
				"ERR wrong number of arguments for 'hscan' command"},
			delays: []time.Duration{0},
		},
		{
			name: "HSCAN with wrong key",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hScan1", "field": "field", "value": "value"}},
				{Command: "HSCAN", Body: map[string]interface{}{"values": []interface{}{"wrong_key_hScan1", 0}}},
			},
			expected: []interface{}{float64(1), []interface{}{"0", []interface{}{}}},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "HSCAN with non hash",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "string_key", "value": "string_value"}},
				{Command: "HSCAN", Body: map[string]interface{}{"values": []interface{}{"string_key", 0}}},
			},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "HSCAN with valid key and cursor",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hScan2", "key_values": map[string]interface{}{"field1": "value1", "field2": "value2"}}},
				{Command: "HSCAN", Body: map[string]interface{}{"values": []interface{}{"key_hScan2", 0}}},
			},
			expected: []interface{}{float64(2), []interface{}{"0", []interface{}{"field1", "value1", "field2", "value2"}}},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "HSCAN with cursor at the end",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hScan3", "key_values": map[string]interface{}{"field1": "value1", "field2": "value2"}}},
				{Command: "HSCAN", Body: map[string]interface{}{"values": []interface{}{"key_hScan3", 2}}},
			},
			expected: []interface{}{float64(2), []interface{}{"0", []interface{}{}}},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "HSCAN with cursor at the beginning",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hScan4", "key_values": map[string]interface{}{"field1": "value1", "field2": "value2"}}},
				{Command: "HSCAN", Body: map[string]interface{}{"values": []interface{}{"key_hScan4", 0}}},
			},
			expected: []interface{}{float64(2), []interface{}{"0", []interface{}{"field1", "value1", "field2", "value2"}}},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "HSCAN with cursor in the middle",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hScan5", "key_values": map[string]interface{}{"field1": "value1", "field2": "value2"}}},
				{Command: "HSCAN", Body: map[string]interface{}{"values": []interface{}{"key_hScan5", 1}}},
			},
			expected: []interface{}{float64(2), []interface{}{"0", []interface{}{"field2", "value2"}}},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "HSCAN with MATCH argument",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hScan6", "key_values": map[string]interface{}{"field1": "value1", "field2": "value2", "field3": "value3"}}},
				{Command: "HSCAN", Body: map[string]interface{}{"values": []interface{}{"key_hScan6", 0, "MATCH", "field[12]*"}}},
			},
			expected: []interface{}{float64(3), []interface{}{"0", []interface{}{"field1", "value1", "field2", "value2"}}},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "HSCAN with COUNT argument",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hScan7", "key_values": map[string]interface{}{"field1": "value1", "field2": "value2", "field3": "value3"}}},
				{Command: "HSCAN", Body: map[string]interface{}{"values": []interface{}{"key_hScan7", 0, "COUNT", 2}}},
			},
			expected: []interface{}{float64(3), []interface{}{"2", []interface{}{"field1", "value1", "field2", "value2"}}},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "HSCAN with MATCH and COUNT arguments",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hScan8", "key_values": map[string]interface{}{"field1": "value1", "field2": "value2", "field3": "value3", "field4": "value4"}}},
				{Command: "HSCAN", Body: map[string]interface{}{"values": []interface{}{"key_hScan8", 0, "MATCH", "field[13]*", "COUNT", 1}}},
			},
			expected: []interface{}{float64(4), []interface{}{"1", []interface{}{"field1", "value1"}}},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "HSCAN with invalid MATCH pattern",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hScan9", "key_values": map[string]interface{}{"field1": "value1", "field2": "value2"}}},
				{Command: "HSCAN", Body: map[string]interface{}{"values": []interface{}{"key_hScan9", 0, "MATCH", "[invalid"}}},
			},
			expected: []interface{}{float64(2), "ERR Invalid glob pattern: unexpected end of input"},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "HSCAN with invalid COUNT value",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hScan10", "key_values": map[string]interface{}{"field1": "value1", "field2": "value2"}}},
				{Command: "HSCAN", Body: map[string]interface{}{"values": []interface{}{"key_hScan10", 0, "COUNT", "invalid"}}},
			},
			expected: []interface{}{float64(2), "ERR value is not an integer or out of range"},
			delays:   []time.Duration{0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"KEY", "string_key", "key_hScan1", "key_hScan2", "key_hScan3", "key_hScan4", "key_hScan5", "key_hScan6", "key_hScan7", "key_hScan8", "key_hScan9", "key_hScan10"}}})

			for i, cmd := range tc.commands {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result, err := exec.FireCommand(cmd)
				if err != nil {
					// Check if the error message matches the expected result
					log.Println(tc.expected[i])
					assert.Equal(t, tc.expected[i], err.Error(), "Error message mismatch for cmd %s", cmd)
				} else {
					assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s, expected %v, got %v", cmd, tc.expected[i], result)
				}
			}
		})
	}
}
