// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCMSInitByDim(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "wrong number of arguments",
			commands: []string{"CMS.INITBYDIM cms_key"},
			expected: []interface{}{"ERR wrong number of arguments for 'cms.initbydim' command"},
		},
		{
			name:     "wrong type of width",
			commands: []string{"CMS.INITBYDIM cms_key not_a_number 5"},
			expected: []interface{}{"ERR invalid width for 'cms.initbydim' command"},
		},
		{
			name:     "wrong type of depth",
			commands: []string{"CMS.INITBYDIM cms_key 5 not_a_number"},
			expected: []interface{}{"ERR invalid depth for 'cms.initbydim' command"},
		},
		{
			name:     "negative width",
			commands: []string{"CMS.INITBYDIM cms_key -100 5"},
			expected: []interface{}{"ERR invalid width for 'cms.initbydim' command"},
		},
		{
			name:     "negative depth",
			commands: []string{"CMS.INITBYDIM cms_key 5 -100"},
			expected: []interface{}{"ERR invalid depth for 'cms.initbydim' command"},
		},
		{
			name:     "correct width and depth",
			commands: []string{"CMS.INITBYDIM cms_key 1000 5"},
			expected: []interface{}{"OK"},
		},
		{
			name:     "key already exists",
			commands: []string{"CMS.INITBYDIM new_cms_key 1000 5", "CMS.INITBYDIM new_cms_key 1000 5"},
			expected: []interface{}{"OK", "ERR key already exists for 'cms.initbydim' command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client.FireString("DEL cms_key")
			for i, cmd := range tc.commands {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}

func TestCMSInitByProb(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "wrong number of arguments",
			commands: []string{"CMS.INITBYPROB cms_key1"},
			expected: []interface{}{"ERR wrong number of arguments for 'cms.initbyprob' command"},
		},
		{
			name:     "wrong type of error rate",
			commands: []string{"CMS.INITBYPROB cms_key1 not_a_number 0.01"},
			expected: []interface{}{"ERR invalid overestimation value for 'cms.initbyprob' command"},
		},
		{
			name:     "wrong type of probability",
			commands: []string{"CMS.INITBYPROB cms_key1 0.01 not_a_number"},
			expected: []interface{}{"ERR invalid prob value for 'cms.initbyprob' command"},
		},
		{
			name:     "error rate out of range",
			commands: []string{"CMS.INITBYPROB cms_key1 1 0.01"},
			expected: []interface{}{"ERR invalid overestimation value for 'cms.initbyprob' command"},
		},
		{
			name:     "probability out of range",
			commands: []string{"CMS.INITBYPROB cms_key1 0.01 1"},
			expected: []interface{}{"ERR invalid prob value for 'cms.initbyprob' command"},
		},
		{
			name:     "correct error rate and probability",
			commands: []string{"CMS.INITBYPROB cms_key1 0.01 0.01"},
			expected: []interface{}{"OK"},
		},
		{
			name:     "key already exists",
			commands: []string{"CMS.INITBYPROB new_cms_key1 0.01 0.01", "CMS.INITBYPROB new_cms_key1 0.01 0.01"},
			expected: []interface{}{"OK", "ERR key already exists for 'cms.initbyprob' command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client.FireString("DEL cms_key1")
			for i, cmd := range tc.commands {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}

func TestCMSInfo(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "wrong number of arguments",
			commands: []string{"CMS.INFO"},
			expected: []interface{}{"ERR wrong number of arguments for 'cms.info' command"},
		},
		{
			name:     "key doesn't exist",
			commands: []string{"CMS.INFO cms_key2"},
			expected: []interface{}{"ERR key does not exist for 'cms.info' command"},
		},
		{
			name: "one argument",
			commands: []string{
				"CMS.INITBYDIM cms_key2 1000 5",
				"CMS.INFO cms_key2",
			},
			expected: []interface{}{"OK", []interface{}{"width", int64(1000), "depth", int64(5), "count", int64(0)}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client.FireString("DEL cms_key2")
			for i, cmd := range tc.commands {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

func TestCMSIncrBy(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "wrong number of arguments",
			commands: []string{"CMS.INCRBY cms_key3"},
			expected: []interface{}{"ERR wrong number of arguments for 'cms.incrby' command"},
		},
		{
			name:     "key doesn't exist",
			commands: []string{"CMS.INCRBY cms_key3 test 10"},
			expected: []interface{}{"ERR key does not exist for 'cms.incrby' command"},
		},
		{
			name: "inserting keys",
			commands: []string{
				"CMS.INITBYDIM cms_key3 1000 5",
				"CMS.INCRBY cms_key3 test 10 test1 10",
				"CMS.QUERY cms_key3 test",
				"CMS.QUERY cms_key3 test1 test2",
				"CMS.INFO cms_key3",
			},
			expected: []interface{}{
				"OK",
				[]interface{}{int64(10), int64(10)},
				[]interface{}{int64(10)},
				[]interface{}{int64(10), int64(0)},
				[]interface{}{"width", int64(1000), "depth", int64(5), "count", int64(20)},
			},
		},
		{
			name: "missing values",
			commands: []string{
				"CMS.INITBYDIM cms_key3 1000 5",
				"CMS.INCRBY cms_key3 test 10 test1",
			},
			expected: []interface{}{"OK", "ERR wrong number of arguments for 'cms.incrby' command"},
		},
		{
			name: "negative values",
			commands: []string{
				"CMS.INITBYDIM cms_key3 1000 5",
				"CMS.INCRBY cms_key3 test -1",
			},
			expected: []interface{}{"OK", "ERR cannot parse number for 'cms.incrby' command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client.FireString("DEL cms_key3")
			for i, cmd := range tc.commands {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

func TestCMSQuery(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "wrong number of arguments",
			commands: []string{"CMS.QUERY cms_key4"},
			expected: []interface{}{"ERR wrong number of arguments for 'cms.query' command"},
		},
		{
			name:     "key doesn't exist",
			commands: []string{"CMS.QUERY cms_key4 test"},
			expected: []interface{}{"ERR key does not exist for 'cms.query' command"},
		},
		{
			name: "query keys",
			commands: []string{
				"CMS.INITBYDIM cms_key4 1000 5",
				"CMS.INCRBY cms_key4 test 10000",
				"CMS.INCRBY cms_key4 test1 100",
				"CMS.QUERY cms_key4 test test1",
			},
			expected: []interface{}{"OK", []interface{}{int64(10000)}, []interface{}{int64(100)}, []interface{}{int64(10000), int64(100)}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client.FireString("DEL cms_key4")
			for i, cmd := range tc.commands {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

func TestCMSMerge(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "wrong number of arguments",
			commands: []string{"CMS.MERGE cms_key5"},
			expected: []interface{}{
				"ERR wrong number of arguments for 'cms.merge' command",
			},
		},
		{
			name:     "key doesn't exist",
			commands: []string{"CMS.MERGE cms_key5 1 test"},
			expected: []interface{}{
				"ERR key does not exist for 'cms.merge' command",
			},
		},
		{
			name: "wrong type of number of sources",
			commands: []string{
				"CMS.INITBYDIM cms_key5 1000 5",
				"CMS.INITBYDIM test 1000 5",
				"CMS.MERGE cms_key5 not_a_number test",
			},
			expected: []interface{}{
				"OK",
				"OK",
				"ERR cannot parse number for 'cms.merge' command",
			},
		},
		{
			name: "more sources than specified",
			commands: []string{
				"CMS.INITBYDIM cms_key5 1000 5",
				"CMS.INITBYDIM test 1000 5",
				"CMS.INITBYDIM test1 1000 5",
				"CMS.MERGE cms_key5 3 test test1",
			},
			expected: []interface{}{
				"OK",
				"OK",
				"OK",
				"ERR invalid number of arguments to merge for 'cms.merge' command",
			},
		},
		{
			name: "fewer sources than specified",
			commands: []string{
				"CMS.INITBYDIM cms_key5 1000 5",
				"CMS.INITBYDIM test 1000 5",
				"CMS.INITBYDIM test1 1000 5",
				"CMS.MERGE cms_key5 1 test test1",
			},
			expected: []interface{}{
				"OK",
				"OK",
				"OK",
				"ERR invalid number of arguments to merge for 'cms.merge' command",
			},
		},
		{
			name: "missing weights",
			commands: []string{
				"CMS.INITBYDIM cms_key5 1000 5",
				"CMS.INITBYDIM test 1000 5",
				"CMS.MERGE cms_key5 1 test WEIGHTS",
			},
			expected: []interface{}{
				"OK",
				"OK",
				"ERR invalid number of arguments to merge for 'cms.merge' command",
			},
		},
		{
			name: "more weights than needed",
			commands: []string{
				"CMS.INITBYDIM cms_key5 1000 5",
				"CMS.INITBYDIM test 1000 5",
				"CMS.MERGE cms_key5 1 test WEIGHTS 1 2",
			},
			expected: []interface{}{
				"OK",
				"OK",
				"ERR invalid number of arguments to merge for 'cms.merge' command",
			},
		},
		{
			name: "correct case",
			commands: []string{
				"CMS.INITBYDIM cms_key5 1000 5",
				"CMS.INITBYDIM test 1000 5",
				"CMS.INCRBY cms_key5 test 10 test1 10",
				"CMS.INCRBY test test 10 test2 10",
				"CMS.MERGE cms_key5 2 cms_key5 test",
				"CMS.QUERY cms_key5 test test1 test2",
				"CMS.INFO cms_key5",
			},
			expected: []interface{}{
				"OK",
				"OK",
				[]interface{}{int64(10), int64(10)},
				[]interface{}{int64(10), int64(10)},
				"OK",
				[]interface{}{int64(20), int64(10), int64(10)},
				[]interface{}{"width", int64(1000), "depth", int64(5), "count", int64(40)},
			},
		},
		{
			name: "correct case with given weights",
			commands: []string{
				"CMS.INITBYDIM cms_key5 1000 5",
				"CMS.INITBYDIM test 1000 5",
				"CMS.INITBYDIM test1 1000 5",
				"CMS.INCRBY test a 10 b 20",
				"CMS.INCRBY test1 a 20 b 20",
				"CMS.MERGE cms_key5 2 test test1 WEIGHTS 1 2",
				"CMS.QUERY cms_key5 a b",
			},
			expected: []interface{}{
				"OK",
				"OK",
				"OK",
				[]interface{}{int64(10), int64(20)},
				[]interface{}{int64(20), int64(20)},
				"OK",
				[]interface{}{int64(50), int64(60)},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client.FireString("DEL cms_key5 test test1")
			for i, cmd := range tc.commands {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}
