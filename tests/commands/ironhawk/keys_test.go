// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"

	"github.com/dicedb/dicedb-go/wire"
)

func extractValueKEYS(res *wire.Result) interface{} {
	return res.GetKEYSRes().Keys
}

func TestKEYS(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:           "KEYS with more than one matching key",
			commands:       []string{"SET k v", "SET k1 v1", "KEYS k*"},
			expected:       []interface{}{"OK", "OK", []string{"k", "k1"}},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueSET, extractValueKEYS},
		},
		{
			name:           "KEYS with no matching keys",
			commands:       []string{"KEYS a*"},
			expected:       []interface{}{[]string{}},
			valueExtractor: []ValueExtractorFn{extractValueKEYS},
		},
		{
			name:           "KEYS with single character wildcard",
			commands:       []string{"SET k1 v1", "SET k2 v2", "SET ka va", "KEYS k?"},
			expected:       []interface{}{"OK", "OK", "OK", []string{"k1", "k2", "ka"}},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueSET, extractValueSET, extractValueKEYS},
		},
		{
			name:           "KEYS with single matching key",
			commands:       []string{"SET unique_key value", "KEYS unique*"},
			expected:       []interface{}{"OK", []string{"unique_key"}},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueKEYS},
		},
	}

	runTestcases(t, client, testCases)
}
