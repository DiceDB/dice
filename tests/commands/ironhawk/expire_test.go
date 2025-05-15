// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/dicedb/dicedb-go/wire"
)

func extractValueEXPIRE(res *wire.Result) interface{} {
	return res.GetEXPIRERes().IsChanged
}

func TestEXPIRE(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name: "Set with EXPIRE command",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key 1",
			},
			expected:       []interface{}{"OK", true},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueEXPIRE},
		},
		{
			name: "Check if key is nil after expiration",
			commands: []string{
				"SET k1 v1",
				"EXPIRE k1 1",
				"GET k1",
			},
			expected:       []interface{}{"OK", true, ""},
			delay:          []time.Duration{0, 0, 3 * time.Second},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueEXPIRE, extractValueGET},
		},
		{
			name: "EXPIRE non-existent key",
			commands: []string{
				"EXPIRE non_existent_key 1",
			},
			expected:       []interface{}{false},
			valueExtractor: []ValueExtractorFn{extractValueEXPIRE},
		},
		{
			name: "EXPIRE with past time",
			commands: []string{
				"EXPIRE test_key -1",
			},
			expected:       []interface{}{errors.New("invalid expire time in 'EXPIRE' command")},
			valueExtractor: []ValueExtractorFn{nil},
		},
		{
			name: "EXPIRE with invalid syntax",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key",
			},
			expected:       []interface{}{"OK", errors.New("wrong number of arguments for 'EXPIRE' command")},
			valueExtractor: []ValueExtractorFn{extractValueSET, nil},
		},
		{
			name: "Test(NX): Set the expiration only if the key has no expiration time",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key " + strconv.FormatInt(1, 10) + " NX",
				"EXPIRE test_key " + strconv.FormatInt(1, 10) + " NX",
			},
			expected:       []interface{}{"OK", true, false},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueEXPIRE, extractValueEXPIRE},
		},
		{
			name: "Test(XX): Set the expiration only if the key already has an expiration time",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key " + strconv.FormatInt(10, 10) + " XX",
				"EXPIRE test_key " + strconv.FormatInt(10, 10),
				"EXPIRE test_key " + strconv.FormatInt(10, 10) + " XX",
			},
			expected:       []interface{}{"OK", false, true, true},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueEXPIRE, extractValueEXPIRE, extractValueEXPIRE},
		},
		{
			name: "Test if value is nil after expiration",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key " + strconv.FormatInt(2, 10),
				"GET test_key",
			},
			expected:       []interface{}{"OK", true, ""},
			delay:          []time.Duration{0, 0, 4 * time.Second},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueEXPIRE, extractValueGET},
		},
		{
			name: "Test if value is nil after expiration with NX",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key " + strconv.FormatInt(2, 10) + " NX",
				"GET test_key",
			},
			expected:       []interface{}{"OK", true, ""},
			delay:          []time.Duration{0, 0, 4 * time.Second},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueEXPIRE, extractValueGET},
		},
		{
			name:     "EXPIRE with no keys or arguments",
			commands: []string{"EXPIRE"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'EXPIRE' command"),
			},
			valueExtractor: []ValueExtractorFn{nil},
		},
	}
	runTestcases(t, client, testCases)
}
