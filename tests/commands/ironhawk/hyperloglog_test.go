// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

// All commands related to Hyperloglog are part of this test class
// PFADD, PFCOUNT, PFMERGE, PFDEBUG, PFSELFTEST etc.
import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHyperLogLogCommands(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

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
		{
			name: "PFMERGE with srcKey non-existing",
			commands: []string{
				"PFMERGE NON_EXISTING_SRC_KEY", "PFCOUNT NON_EXISTING_SRC_KEY"},
			expected: []interface{}{"OK", int64(0)},
		},
		{
			name: "PFMERGE with destKey non-existing",
			commands: []string{
				"PFMERGE EXISTING_SRC_KEY NON_EXISTING_DEST_KEY", "PFCOUNT EXISTING_SRC_KEY"},
			expected: []interface{}{"OK", int64(0)},
		},
		{
			name: "PFMERGE with destKey existing",
			commands: []string{
				"PFADD DEST_KEY_1 foo bar zap a", "PFADD DEST_KEY_2 a b c foo", "PFMERGE SRC_KEY_1 DEST_KEY_1 DEST_KEY_2",
				"PFCOUNT SRC_KEY_1"},
			expected: []interface{}{int64(1), int64(1), "OK", int64(6)},
		},
		{
			name: "PFMERGE with only one destKey existing",
			commands: []string{
				"PFADD DEST_KEY_3 foo bar zap a", "PFMERGE SRC_KEY_2 DEST_KEY_3 NON_EXISTING_DEST_KEY",
				"PFCOUNT SRC_KEY_2"},
			expected: []interface{}{int64(1), "OK", int64(4)},
		},
		{
			name: "PFMERGE with invalid object",
			commands: []string{
				"PFADD INVALID_HLL a b c", "SET INVALID_HLL \"1\"", "PFMERGE INVALID_HLL"},
			expected: []interface{}{int64(1), "OK", "WRONGTYPE Key is not a valid HyperLogLog string value"},
		},
		{
			name: "PFMERGE with invalid src object",
			commands: []string{
				"PFADD INVALID_SRC_HLL a b c", "SET INVALID_SRC_HLL \"1\"", "PFMERGE HLL INVALID_SRC_HLL"},
			expected: []interface{}{int64(1), "OK", "WRONGTYPE Key is not a valid HyperLogLog string value"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}
