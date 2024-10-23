package http

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestHyperLogLogCommands(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		delays   []time.Duration
	}{
		{
			name: "PFADD with one key-value pair",
			commands: []HTTPCommand{
				{Command: "PFADD", Body: map[string]interface{}{"key": "hll0", "value": "v1"}},
				{Command: "PFCOUNT", Body: map[string]interface{}{"key": "hll0"}},
			},
			expected: []interface{}{float64(1), float64(1)},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "PFADD with multiple key-value pair",
			commands: []HTTPCommand{
				{Command: "PFADD", Body: map[string]interface{}{"key": "hll", "values": [...]string{"a", "b", "c", "d", "e", "f"}}},
				{Command: "PFCOUNT", Body: map[string]interface{}{"key": "hll"}},
			},
			expected: []interface{}{float64(1), float64(6)},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "PFADD with duplicate key-value pairs",
			commands: []HTTPCommand{
				{Command: "PFADD", Body: map[string]interface{}{"key": "hll1", "values": [...]string{"foo", "bar", "zap"}}},
				{Command: "PFADD", Body: map[string]interface{}{"key": "hll1", "values": [...]string{"zap", "zap", "zap"}}},
				{Command: "PFADD", Body: map[string]interface{}{"key": "hll1", "values": [...]string{"foo", "bar"}}},
				{Command: "PFCOUNT", Body: map[string]interface{}{"key": "hll1"}},
			},
			expected: []interface{}{float64(1), float64(0), float64(0), float64(3)},
			delays:   []time.Duration{0, 0, 0, 0},
		},
		{
			name: "PFADD with multiple keys",
			commands: []HTTPCommand{
				{Command: "PFADD", Body: map[string]interface{}{"key": "hll2", "values": [...]string{"foo", "bar", "zap"}}},
				{Command: "PFADD", Body: map[string]interface{}{"key": "hll2", "values": [...]string{"zap", "zap", "zap"}}},
				{Command: "PFCOUNT", Body: map[string]interface{}{"key": "hll2"}},
				{Command: "PFADD", Body: map[string]interface{}{"key": "some-other-hll", "values": [...]string{"1", "2", "3"}}},
				{Command: "PFCOUNT", Body: map[string]interface{}{"keys": [...]string{"hll2", "some-other-hll"}}},
			},
			expected: []interface{}{float64(1), float64(0), float64(3), float64(1), float64(6)},
			delays:   []time.Duration{0, 0, 0, 0, 0},
		},
		{
			name: "PFADD with non-existing key",
			commands: []HTTPCommand{
				{Command: "PFADD", Body: map[string]interface{}{"key": "hll3", "values": [...]string{"foo", "bar", "zap"}}},
				{Command: "PFADD", Body: map[string]interface{}{"key": "hll3", "values": [...]string{"zap", "zap", "zap"}}},
				{Command: "PFCOUNT", Body: map[string]interface{}{"key": "hll3"}},
				{Command: "PFCOUNT", Body: map[string]interface{}{"keys": [...]string{"hll3", "non-exist-hll"}}},
				{Command: "PFADD", Body: map[string]interface{}{"key": "some-new-hll", "value": "abc"}},
				{Command: "PFCOUNT", Body: map[string]interface{}{"keys": [...]string{"hll3", "non-exist-hll", "some-new-hll"}}},
			},
			expected: []interface{}{float64(1), float64(0), float64(3), float64(3), float64(1), float64(4)},
			delays:   []time.Duration{0, 0, 0, 0, 0, 0},
		},
		{
			name: "PFMERGE with srcKey non-existing",
			commands: []HTTPCommand{
				{Command: "PFMERGE", Body: map[string]interface{}{"key": "NON_EXISTING_SRC_KEY"}},
				{Command: "PFCOUNT", Body: map[string]interface{}{"key": "NON_EXISTING_SRC_KEY"}},
			},
			expected: []interface{}{"OK", float64(0)},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "PFMERGE with destKey non-existing",
			commands: []HTTPCommand{
				{Command: "PFMERGE", Body: map[string]interface{}{"keys": []string{"EXISTING_SRC_KEY", "NON_EXISTING_SRC_KEY"}}},
				{Command: "PFCOUNT", Body: map[string]interface{}{"key": "EXISTING_SRC_KEY"}},
			},
			expected: []interface{}{"OK", float64(0)},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "PFMERGE with destKey existing",
			commands: []HTTPCommand{
				{Command: "PFADD", Body: map[string]interface{}{"key": "DEST_KEY_1", "values": [...]string{"foo", "bar", "zap", "a"}}},
				{Command: "PFADD", Body: map[string]interface{}{"key": "DEST_KEY_2", "values": [...]string{"a", "b", "c", "foo"}}},
				{Command: "PFMERGE", Body: map[string]interface{}{"keys": [...]string{"SRC_KEY_1", "DEST_KEY_1", "DEST_KEY_2"}}},
				{Command: "PFCOUNT", Body: map[string]interface{}{"key": "SRC_KEY_1"}},
			},
			expected: []interface{}{float64(1), float64(1), "OK", float64(6)},
			delays:   []time.Duration{0, 0, 0, 0},
		},
		{
			name: "PFMERGE with only one destKey existing",
			commands: []HTTPCommand{
				{Command: "PFADD", Body: map[string]interface{}{"key": "DEST_KEY_3", "values": [...]string{"foo", "bar", "zap", "a"}}},
				{Command: "PFMERGE", Body: map[string]interface{}{"keys": [...]string{"SRC_KEY_2", "DEST_KEY_3", "NON_EXISTING_DEST_KEY"}}},
				{Command: "PFCOUNT", Body: map[string]interface{}{"key": "SRC_KEY_2"}},
			},
			expected: []interface{}{float64(1), "OK", float64(4)},
			delays:   []time.Duration{0, 0, 0},
		},
		{
			name: "PFMERGE with invalid object",
			commands: []HTTPCommand{
				{Command: "PFADD", Body: map[string]interface{}{"key": "INVALID_HLL", "values": [...]string{"a", "b", "c"}}},
				{Command: "SET", Body: map[string]interface{}{"key": "INVALID_HLL", "value": "1"}},
				{Command: "PFMERGE", Body: map[string]interface{}{"key": "INVALID_HLL"}},
			},
			expected: []interface{}{float64(1), "OK", "WRONGTYPE Key is not a valid HyperLogLog string value"},
			delays:   []time.Duration{0, 0, 0},
		},
		{
			name: "PFMERGE with invalid src object",
			commands: []HTTPCommand{
				{Command: "PFADD", Body: map[string]interface{}{"key": "INVALID_SRC_HLL", "values": [...]string{"a", "b", "c"}}},
				{Command: "SET", Body: map[string]interface{}{"key": "INVALID_SRC_HLL", "value": "1"}},
				{Command: "PFMERGE", Body: map[string]interface{}{"keys": [...]string{"HLL", "INVALID_SRC_HLL"}}},
			},
			expected: []interface{}{float64(1), "OK", "WRONGTYPE Key is not a valid HyperLogLog string value"},
			delays:   []time.Duration{0, 0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result, err := exec.FireCommand(cmd)
				if err != nil {
					// Check if the error message matches the expected result
					assert.Equal(t, tc.expected[i], err.Error(), "Error message mismatch for cmd %s", cmd)
				} else {
					assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s, expected %v, got %v", cmd, tc.expected[i], result)
				}
			}
		})
	}
}
