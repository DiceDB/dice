package http

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCMSInitByDim(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
	}{
		{
			name: "wrong number of arguments",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "cms_key"}},
			},
			expected: []interface{}{"ERR wrong number of arguments for 'cms.initbydim' command"},
		},
		{
			name: "wrong type of width",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "cms_key", "values": [...]string{"not_a_number", "5"}}},
			},
			expected: []interface{}{"ERR invalid width for 'cms.initbydim' command"},
		},
		{
			name: "wrong type of depth",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "cms_key", "values": [...]string{"5", "not_a_number"}}},
			},
			expected: []interface{}{"ERR invalid depth for 'cms.initbydim' command"},
		},
		{
			name: "negative width",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "cms_key", "values": [...]string{"-1000", "5"}}},
			},
			expected: []interface{}{"ERR invalid width for 'cms.initbydim' command"},
		},
		{
			name: "negative depth",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "cms_key", "values": [...]string{"5", "-100"}}},
			},
			expected: []interface{}{"ERR invalid depth for 'cms.initbydim' command"},
		},
		{
			name: "correct width and depth",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "cms_key", "values": [...]string{"1000", "5"}}},
			},
			expected: []interface{}{"OK"},
		},
		{
			name: "key already exists",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "new_cms_key", "values": [...]string{"1000", "5"}}},
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "new_cms_key", "values": [...]string{"1000", "5"}}},
			},
			expected: []interface{}{"OK", "ERR key already exists for 'cms.initbydim' command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "cms_key"}})

			for i, cmd := range tc.commands {
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

func TestCMSInitByProb(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
	}{
		{
			name: "wrong number of arguments",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYPROB", Body: map[string]interface{}{"key": "cms_key"}},
			},
			expected: []interface{}{"ERR wrong number of arguments for 'cms.initbyprob' command"},
		},
		{
			name: "wrong type of width",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYPROB", Body: map[string]interface{}{"key": "cms_key", "values": [...]string{"not_a_number", "0.01"}}},
			},
			expected: []interface{}{"ERR invalid overestimation value for 'cms.initbyprob' command"},
		},
		{
			name: "wrong type of depth",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYPROB", Body: map[string]interface{}{"key": "cms_key", "values": [...]string{"0.01", "not_a_number"}}},
			},
			expected: []interface{}{"ERR invalid prob value for 'cms.initbyprob' command"},
		},
		{
			name: "negative width",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYPROB", Body: map[string]interface{}{"key": "cms_key", "values": [...]string{"1", "0.01"}}},
			},
			expected: []interface{}{"ERR invalid overestimation value for 'cms.initbyprob' command"},
		},
		{
			name: "negative depth",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYPROB", Body: map[string]interface{}{"key": "cms_key", "values": [...]string{"0.01", "1"}}},
			},
			expected: []interface{}{"ERR invalid prob value for 'cms.initbyprob' command"},
		},
		{
			name: "correct width and depth",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYPROB", Body: map[string]interface{}{"key": "cms_key", "values": [...]string{"0.01", "0.01"}}},
			},
			expected: []interface{}{"OK"},
		},
		{
			name: "key already exists",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYPROB", Body: map[string]interface{}{"key": "new_cms_key", "values": [...]string{"0.01", "0.01"}}},
				{Command: "CMS.INITBYPROB", Body: map[string]interface{}{"key": "new_cms_key", "values": [...]string{"0.01", "0.01"}}},
			},
			expected: []interface{}{"OK", "ERR key already exists for 'cms.initbyprob' command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"cms_key", "new_cms_key"}}})

			for i, cmd := range tc.commands {
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

func TestCMSInfo(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
	}{
		{
			name: "wrong number of arguments",
			commands: []HTTPCommand{
				{Command: "CMS.INFO", Body: map[string]interface{}{"key": nil}},
			},
			expected: []interface{}{"ERR key does not exist for 'cms.info' command"},
		},
		{
			name: "key doesn't exist",
			commands: []HTTPCommand{
				{Command: "CMS.INFO", Body: map[string]interface{}{"key": "cms_key2"}},
			},
			expected: []interface{}{"ERR key does not exist for 'cms.info' command"},
		},
		{
			name: "one argument",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "cms_key2", "values": [...]string{"1000", "5"}}},
				{Command: "CMS.INFO", Body: map[string]interface{}{"key": "cms_key2"}},
			},
			expected: []interface{}{"OK", []interface{}{"width", float64(1000), "depth", float64(5), "count", float64(0)}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "cms_key2"}})

			for i, cmd := range tc.commands {
				result, err := exec.FireCommand(cmd)
				if err != nil {
					// Check if the error message matches the expected result
					log.Println(tc.expected[i])
					assert.Equal(t, tc.expected[i], err.Error(), "Error message mismatch for cmd %s", cmd)
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}
}

func TestCMSIncrBy(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
	}{
		{
			name: "wrong number of arguments",
			commands: []HTTPCommand{
				{Command: "CMS.INCRBY", Body: map[string]interface{}{"key": "cms_key3"}},
			},
			expected: []interface{}{"ERR wrong number of arguments for 'cms.incrby' command"},
		},
		{
			name: "key doesn't exist",
			commands: []HTTPCommand{
				{Command: "CMS.INCRBY", Body: map[string]interface{}{"key": "cms_key3", "values": [...]string{"test", "10"}}},
			},
			expected: []interface{}{"ERR key does not exist for 'cms.incrby' command"},
		},
		{
			name: "inserting keys",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "cms_key3", "values": [...]string{"1000", "5"}}},
				{Command: "CMS.INCRBY", Body: map[string]interface{}{"key": "cms_key3", "values": [...]string{"test", "10", "test1", "10"}}},
				{Command: "CMS.QUERY", Body: map[string]interface{}{"key": "cms_key3", "value": "test"}},
				{Command: "CMS.QUERY", Body: map[string]interface{}{"key": "cms_key3", "values": [...]string{"test1", "test2"}}},
				{Command: "CMS.INFO", Body: map[string]interface{}{"key": "cms_key3"}},
			},
			expected: []interface{}{
				"OK",
				[]interface{}{float64(10), float64(10)},
				[]interface{}{float64(10)},
				[]interface{}{float64(10), float64(0)},
				[]interface{}{"width", float64(1000), "depth", float64(5), "count", float64(20)},
			},
		},
		{
			name: "missing values",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "cms_key3", "values": [...]string{"1000", "5"}}},
				{Command: "CMS.INCRBY", Body: map[string]interface{}{"key": "cms_key3", "values": [...]string{"test"}}},
			},
			expected: []interface{}{"OK", "ERR wrong number of arguments for 'cms.incrby' command"},
		},
		{
			name: "negative values",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "cms_key3", "values": [...]string{"1000", "5"}}},
				{Command: "CMS.INCRBY", Body: map[string]interface{}{"key": "cms_key3", "values": [...]string{"test", "-1"}}},
			},
			expected: []interface{}{"OK", "ERR cannot parse number for 'cms.incrby' command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "cms_key3"}})

			for i, cmd := range tc.commands {
				result, err := exec.FireCommand(cmd)
				if err != nil {
					// Check if the error message matches the expected result
					log.Println(tc.expected[i])
					assert.Equal(t, tc.expected[i], err.Error(), "Error message mismatch for cmd %s", cmd)
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}
}

func TestCMSQuery(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
	}{
		{
			name: "wrong number of arguments",
			commands: []HTTPCommand{
				{Command: "CMS.QUERY", Body: map[string]interface{}{"key": "cms_key4"}},
			},
			expected: []interface{}{"ERR wrong number of arguments for 'cms.query' command"},
		},
		{
			name: "key doesn't exist",
			commands: []HTTPCommand{
				{Command: "CMS.QUERY", Body: map[string]interface{}{"key": "cms_key4", "values": [...]string{"test"}}},
			},
			expected: []interface{}{"ERR key does not exist for 'cms.query' command"},
		},
		{
			name: "query keys",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "cms_key4", "values": [...]string{"1000", "5"}}},
				{Command: "CMS.INCRBY", Body: map[string]interface{}{"key": "cms_key4", "values": [...]string{"test", "10000"}}},
				{Command: "CMS.INCRBY", Body: map[string]interface{}{"key": "cms_key4", "values": [...]string{"test1", "100"}}},
				{Command: "CMS.QUERY", Body: map[string]interface{}{"key": "cms_key4", "values": [...]string{"test", "test1"}}},
			},
			expected: []interface{}{"OK", []interface{}{float64(10000)}, []interface{}{float64(100)}, []interface{}{float64(10000), float64(100)}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "cms_key4"}})

			for i, cmd := range tc.commands {
				result, err := exec.FireCommand(cmd)
				if err != nil {
					// Check if the error message matches the expected result
					log.Println(tc.expected[i])
					assert.Equal(t, tc.expected[i], err.Error(), "Error message mismatch for cmd %s", cmd)
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}
}

func TestCMSMerge(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
	}{
		{
			name: "wrong number of arguments",
			commands: []HTTPCommand{
				{Command: "CMS.MERGE", Body: map[string]interface{}{"key": "cms_key5"}},
			},
			expected: []interface{}{"ERR wrong number of arguments for 'cms.merge' command"},
		},
		{
			name: "key doesn't exist",
			commands: []HTTPCommand{
				{Command: "CMS.MERGE", Body: map[string]interface{}{"key": "cms_key5", "values": [...]string{"1", "test"}}},
			},
			expected: []interface{}{"ERR key does not exist for 'cms.merge' command"},
		},
		{
			name: "wrong type of number of sources",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "cms_key5", "values": [...]string{"1000", "5"}}},
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "test", "values": [...]string{"1000", "5"}}},
				{Command: "CMS.MERGE", Body: map[string]interface{}{"key": "cms_key5", "values": [...]string{"not_a_number", "test"}}},
			},
			expected: []interface{}{"OK", "OK", "ERR cannot parse number for 'cms.merge' command"},
		},
		{
			name: "more sources than specified",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "cms_key5", "values": [...]string{"1000", "5"}}},
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "test", "values": [...]string{"1000", "5"}}},
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "test1", "values": [...]string{"1000", "5"}}},
				{Command: "CMS.MERGE", Body: map[string]interface{}{"key": "cms_key5", "values": [...]string{"3", "test", "test1"}}},
			},
			expected: []interface{}{"OK", "OK", "OK", "ERR invalid number of arguments to merge for 'cms.merge' command"},
		},
		{
			name: "fewer sources than specified",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "cms_key5", "values": [...]string{"1000", "5"}}},
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "test", "values": [...]string{"1000", "5"}}},
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "test1", "values": [...]string{"1000", "5"}}},
				{Command: "CMS.MERGE", Body: map[string]interface{}{"key": "cms_key5", "values": [...]string{"1", "test", "test1"}}},
			},
			expected: []interface{}{"OK", "OK", "OK", "ERR invalid number of arguments to merge for 'cms.merge' command"},
		},
		{
			name: "missing weights",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "cms_key5", "values": [...]string{"1000", "5"}}},
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "test", "values": [...]string{"1000", "5"}}},
				{Command: "CMS.MERGE", Body: map[string]interface{}{"key": "cms_key5", "values": [...]string{"1", "test", "WEIGHTS"}}},
			},
			expected: []interface{}{"OK", "OK", "ERR invalid number of arguments to merge for 'cms.merge' command"},
		},
		{
			name: "more weights than needed",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "cms_key5", "values": [...]string{"1000", "5"}}},
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "test", "values": [...]string{"1000", "5"}}},
				{Command: "CMS.MERGE", Body: map[string]interface{}{"key": "cms_key5", "values": [...]string{"1", "test", "WEIGHTS", "1", "2"}}},
			},
			expected: []interface{}{"OK", "OK", "ERR invalid number of arguments to merge for 'cms.merge' command"},
		},
		{
			name: "correct case",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "cms_key5", "values": [...]string{"1000", "5"}}},
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "test", "values": [...]string{"1000", "5"}}},
				{Command: "CMS.INCRBY", Body: map[string]interface{}{"key": "cms_key5", "values": [...]string{"test", "10", "test1", "10"}}},
				{Command: "CMS.INCRBY", Body: map[string]interface{}{"key": "test", "values": [...]string{"test", "10", "test2", "10"}}},
				{Command: "CMS.MERGE", Body: map[string]interface{}{"key": "cms_key5", "values": [...]string{"2", "cms_key5", "test"}}},
				{Command: "CMS.QUERY", Body: map[string]interface{}{"key": "cms_key5", "values": [...]string{"test", "test1", "test2"}}},
				{Command: "CMS.INFO", Body: map[string]interface{}{"key": "cms_key5"}},
			},
			expected: []interface{}{
				"OK",
				"OK",
				[]interface{}{float64(10), float64(10)},
				[]interface{}{float64(10), float64(10)},
				"OK",
				[]interface{}{float64(20), float64(10), float64(10)},
				[]interface{}{"width", float64(1000), "depth", float64(5), "count", float64(40)},
			},
		},
		{
			name: "correct case with given weights",
			commands: []HTTPCommand{
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "cms_key5", "values": [...]string{"1000", "5"}}},
				{Command: "CMS.INITBYDIM", Body: map[string]interface{}{"key": "test", "values": [...]string{"1000", "5"}}},
				{Command: "CMS.INCRBY", Body: map[string]interface{}{"key": "cms_key5", "values": [...]string{"test", "10", "test1", "10"}}},
				{Command: "CMS.INCRBY", Body: map[string]interface{}{"key": "test", "values": [...]string{"test", "10", "test2", "10"}}},
				{Command: "CMS.MERGE", Body: map[string]interface{}{"key": "cms_key5", "values": [...]string{"2", "cms_key5", "test", "WEIGHTS", "1", "2"}}},
				{Command: "CMS.QUERY", Body: map[string]interface{}{"key": "cms_key5", "values": [...]string{"test", "test1", "test2"}}},
				{Command: "CMS.INFO", Body: map[string]interface{}{"key": "cms_key5"}},
			},
			expected: []interface{}{
				"OK",
				"OK",
				[]interface{}{float64(10), float64(10)},
				[]interface{}{float64(10), float64(10)},
				"OK",
				[]interface{}{float64(30), float64(10), float64(20)},
				[]interface{}{"width", float64(1000), "depth", float64(5), "count", float64(60)},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"cms_key5", "test", "test1"}}})

			for i, cmd := range tc.commands {
				result, err := exec.FireCommand(cmd)
				if err != nil {
					// Check if the error message matches the expected result
					log.Println(tc.expected[i])
					assert.Equal(t, tc.expected[i], err.Error(), "Error message mismatch for cmd %s", cmd)
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}
}
