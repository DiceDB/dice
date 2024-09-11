package tests

// All commands related to Hyperloglog are part of this test class
// PFADD, PFCOUNT, PFMERGE, PFDEBUG, PFSELFTEST etc.
import (
	"gotest.tools/v3/assert"
	"testing"
)

func TestHyperLogLogCommands(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "PFADD with one key-value pair",
			commands: []string{"PFADD hll0 v1", "PFCOUNT hll0"},
			expected: []interface{}{int64(1), int64(1)},
		},
		{
			name:     "PFADD with multiple key-value pairs",
			commands: []string{"PFADD hll a b c d e f g", "PFCOUNT hll"},
			expected: []interface{}{int64(1), int64(7)},
		},
		{
			name:     "PFADD with duplicate key-value pairs",
			commands: []string{"PFADD hll1 foo bar zap", "PFADD hll1 zap zap zap", "PFADD hll1 foo bar", "PFCOUNT hll1"},
			expected: []interface{}{int64(1), int64(0), int64(0), int64(3)},
		},
		{
			name: "PFADD with multiple keys",
			commands: []string{
				"PFADD hll2 foo bar zap", "PFADD hll2 zap zap zap", "PFCOUNT hll2",
				"PFADD some-other-hll 1 2 3", "PFCOUNT hll2 some-other-hll"},
			expected: []interface{}{int64(1), int64(0), int64(3), int64(1), int64(6)},
		},
		{
			name: "PFADD with non-existing key",
			commands: []string{
				"PFADD hll3 foo bar zap", "PFADD hll3 zap zap zap", "PFCOUNT hll3",
				"PFCOUNT hll3 non-exist-hll", "PFADD some-new-hll abc", "PFCOUNT hll3 non-exist-hll some-new-hll"},
			expected: []interface{}{int64(1), int64(0), int64(3), int64(3), int64(1), int64(4)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}
