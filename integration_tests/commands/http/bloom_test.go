package http

import (
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func TestBloomFilter(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		// create tests for bloom filter from resp/bloom_test.go
		{
			name: "BF.RESERVE and BF.ADD",
			commands: []HTTPCommand{
				{
					Command: "BF.RESERVE",
					Body:    map[string]interface{}{"key": "bf", "values": []interface{}{0.01, 1000}},
				},
				{
					Command: "BF.ADD",
					Body:    map[string]interface{}{"key": "bf", "value": "item1"},
				},
				{
					Command: "BF.EXISTS",
					Body:    map[string]interface{}{"key": "bf", "value": "item1"},
				},
			},
			expected: []interface{}{"OK", "1", "1"},
		},
		{
			name: "BF.EXISTS returns false for non-existing item",
			commands: []HTTPCommand{
				{
					Command: "BF.RESERVE",
					Body:    map[string]interface{}{"key": "bf", "values": []interface{}{0.01, 1000}},
				},
				{
					Command: "BF.EXISTS",
					Body:    map[string]interface{}{"key": "bf", "value": "item2"},
				},
			},
			expected: []interface{}{"OK", "0"},
		},
		{
			name: "BF.INFO provides correct information",
			commands: []HTTPCommand{
				{
					Command: "BF.RESERVE",
					Body:    map[string]interface{}{"key": "bf", "values": []interface{}{0.01, 1000}},
				},
				{
					Command: "BF.ADD",
					Body:    map[string]interface{}{"key": "bf", "value": "item1"},
				},
				{
					Command: "BF.INFO",
					Body:    map[string]interface{}{"key": "bf"},
				},
			},
			expected: []interface{}{
				"OK",
				"1",
				[]interface{}{
					"Capacity", float64(1000),
					"Size", float64(10104),
					"Number of filters", float64(7),
					"Number of items inserted", float64(1),
					"Expansion rate", float64(2)},
			},
		},
		{
			name: "BF.RESERVE with duplicate filter name",
			commands: []HTTPCommand{
				{
					Command: "BF.RESERVE",
					Body:    map[string]interface{}{"key": "bf", "values": []interface{}{0.01, 1000}},
				},
				{
					Command: "BF.RESERVE",
					Body:    map[string]interface{}{"key": "bf", "values": []interface{}{0.01, 2000}},
				},
			},
			expected: []interface{}{"OK", "ERR item exists"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// Execute test commands and validate results
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result, "Value mismatch for command %s", cmd.Command)
			}

			exec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"key": "bf"},
			})
		})
	}
}

func TestBFEdgeCasesAndErrors(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "BF.RESERVE with incorrect number of arguments",
			commands: []HTTPCommand{
				{
					Command: "BF.RESERVE",
					Body:    map[string]interface{}{"key": "bf"},
				},
				{
					Command: "BF.RESERVE",
					Body:    map[string]interface{}{"key": "bf", "values": []interface{}{"a"}},
				},
			},
			expected: []interface{}{"ERR wrong number of arguments for 'bf.reserve' command", "ERR wrong number of arguments for 'bf.reserve' command"},
		},
		{
			name: "BF.RESERVE with zero capacity",
			commands: []HTTPCommand{
				{
					Command: "BF.RESERVE",
					Body:    map[string]interface{}{"key": "bf", "values": []interface{}{0.01, 0}},
				},
			},
			expected: []interface{}{"ERR (capacity should be larger than 0)"},
		},
		{
			name: "BF.RESERVE with negative capacity",
			commands: []HTTPCommand{
				{
					Command: "BF.RESERVE",
					Body:    map[string]interface{}{"key": "bf", "values": []interface{}{0.01, -1}},
				},
			},
			expected: []interface{}{"ERR (capacity should be larger than 0)"},
		},
		{
			name: "BF.RESERVE with invalid capacity",
			commands: []HTTPCommand{
				{
					Command: "BF.RESERVE",
					Body:    map[string]interface{}{"key": "bf", "values": []interface{}{0.01, "a"}},
				},
			},
			expected: []interface{}{"ERR bad capacity"},
		},
		{
			name: "BF.RESERVE with invalid error rate",
			commands: []HTTPCommand{
				{
					Command: "BF.RESERVE",
					Body:    map[string]interface{}{"key": "bf", "values": []interface{}{-0.01, 1000}},
				},
			},
			expected: []interface{}{"ERR (0 < error rate range < 1) "},
		},
		{
			name: "BF.RESERVE with invalid error rate",
			commands: []HTTPCommand{
				{
					Command: "BF.RESERVE",
					Body:    map[string]interface{}{"key": "bf", "values": []interface{}{"a", 1000}},
				},
			},
			expected: []interface{}{"ERR bad error rate"},
		},
		{
			name: "BF.ADD to a Bloom filter without reserving",
			commands: []HTTPCommand{
				{
					Command: "BF.ADD",
					Body:    map[string]interface{}{"key": "bf", "value": "item1"},
				},
			},
			expected: []interface{}{"1"},
		},
		{
			name: "BF.EXISTS on an unreserved filter",
			commands: []HTTPCommand{
				{
					Command: "BF.EXISTS",
					Body:    map[string]interface{}{"key": "bf", "value": "item1"},
				},
			},
			expected: []interface{}{"0"},
		},
		{
			name: "BF.INFO on a non-existent filter",
			commands: []HTTPCommand{
				{
					Command: "BF.INFO",
					Body:    map[string]interface{}{"key": "bf"},
				},
			},
			expected: []interface{}{"ERR not found"},
		},
		{
			name: "BF.RESERVE with a very high error rate",
			commands: []HTTPCommand{
				{
					Command: "BF.RESERVE",
					Body:    map[string]interface{}{"key": "bf", "values": []interface{}{0.99, 1000}},
				},
			},
			expected: []interface{}{"OK"},
		},
		{
			name: "BF.RESERVE with a very low error rate",
			commands: []HTTPCommand{
				{
					Command: "BF.RESERVE",
					Body:    map[string]interface{}{"key": "bf", "values": []interface{}{0.000001, 1000}},
				},
			},
			expected: []interface{}{"OK"},
		},
		{
			name: "BF.ADD multiple items and check existence",
			commands: []HTTPCommand{
				{
					Command: "BF.RESERVE",
					Body:    map[string]interface{}{"key": "bf", "values": []interface{}{0.01, 1000}},
				},
				{
					Command: "BF.ADD",
					Body:    map[string]interface{}{"key": "bf", "value": "item1"},
				},
				{
					Command: "BF.ADD",
					Body:    map[string]interface{}{"key": "bf", "value": "item2"},
				},
				{
					Command: "BF.EXISTS",
					Body:    map[string]interface{}{"key": "bf", "value": "item1"},
				},
				{
					Command: "BF.EXISTS",
					Body:    map[string]interface{}{"key": "bf", "value": "item2"},
				},
				{
					Command: "BF.EXISTS",
					Body:    map[string]interface{}{"key": "bf", "value": "item3"},
				},
			},
			expected: []interface{}{"OK", "1", "1", "1", "1", "0"},
		},
		{
			name: "BF.EXISTS after BF.ADD returns false on non-existing item",
			commands: []HTTPCommand{
				{
					Command: "BF.RESERVE",
					Body:    map[string]interface{}{"key": "bf", "values": []interface{}{0.01, 1000}},
				},
				{
					Command: "BF.ADD",
					Body:    map[string]interface{}{"key": "bf", "value": "item1"},
				},
				{
					Command: "BF.EXISTS",
					Body:    map[string]interface{}{"key": "bf", "value": "item2"},
				},
			},
			expected: []interface{}{"OK", "1", "0"},
		},
		{
			name: "BF.RESERVE with duplicate filter name",
			commands: []HTTPCommand{
				{
					Command: "BF.RESERVE",
					Body:    map[string]interface{}{"key": "bf", "values": []interface{}{0.01, 1000}},
				},
				{
					Command: "BF.RESERVE",
					Body:    map[string]interface{}{"key": "bf", "values": []interface{}{0.01, 2000}},
				},
			},
			expected: []interface{}{"OK", "ERR item exists"},
		},
		{
			name: "BF.INFO after multiple additions",
			commands: []HTTPCommand{
				{
					Command: "BF.RESERVE",
					Body:    map[string]interface{}{"key": "bf", "values": []interface{}{0.01, 1000}},
				},
				{
					Command: "BF.ADD",
					Body:    map[string]interface{}{"key": "bf", "value": "item1"},
				},
				{
					Command: "BF.ADD",
					Body:    map[string]interface{}{"key": "bf", "value": "item2"},
				},
				{
					Command: "BF.INFO",
					Body:    map[string]interface{}{"key": "bf"},
				},
			},
			expected: []interface{}{
				"OK",
				"1",
				"1",
				[]interface{}{
					"Capacity", float64(1000),
					"Size", float64(10104),
					"Number of filters", float64(7),
					"Number of items inserted", float64(2),
					"Expansion rate", float64(2)},
			},
		},
		{
			name: "BF.RESERVE on a key holding a string value",
			commands: []HTTPCommand{
				{
					Command: "SET",
					Body:    map[string]interface{}{"key": "bf", "value": "string"},
				},
				{
					Command: "BF.RESERVE",
					Body:    map[string]interface{}{"key": "bf", "values": []interface{}{0.01, 1000}},
				},
			},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name: "BF.ADD on a key holding a list",
			commands: []HTTPCommand{
				{
					Command: "LPUSH",
					Body:    map[string]interface{}{"key": "bf", "value": "item1"},
				},
				{
					Command: "BF.ADD",
					Body:    map[string]interface{}{"key": "bf", "value": "item2"},
				},
			},
			expected: []interface{}{float64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name: "BF.INFO on a key holding a hash",
			commands: []HTTPCommand{
				{
					Command: "HSET",
					Body:    map[string]interface{}{"key": "bf", "field": "field1", "value": "value1"},
				},
				{
					Command: "BF.INFO",
					Body:    map[string]interface{}{"key": "bf"},
				},
			},
			expected: []interface{}{float64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// Execute test commands and validate results
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result, "Value mismatch for command %s", cmd.Command)
			}
			//delete the key bf and foo
			exec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"key": "bf"},
			})
			exec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"key": "foo"},
			})
		})
	}
}
