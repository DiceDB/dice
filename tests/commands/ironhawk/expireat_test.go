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

func extractValueEXPIREAT(res *wire.Result) interface{} {
	return res.GetEXPIREATRes().IsChanged
}

func TestEXPIREAT(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name: "Set with EXPIREAT command",
			commands: []string{
				"SET test_key test_value",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+1, 10),
			},
			expected:       []interface{}{"OK", true},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueEXPIREAT},
		},
		{
			name: "Check if key is nil after expiration",
			commands: []string{
				"SET k1 v1",
				"EXPIREAT k1 " + strconv.FormatInt(time.Now().Unix()+1, 10),
				"GET k1",
			},
			expected:       []interface{}{"OK", true, ""},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueEXPIREAT, extractValueGET},
			delay:          []time.Duration{0, 0, 3 * time.Second},
		},
		{
			name: "EXPIREAT non-existent key",
			commands: []string{
				"EXPIREAT non_existent_key " + strconv.FormatInt(time.Now().Unix()+1, 10),
			},
			expected:       []interface{}{false},
			valueExtractor: []ValueExtractorFn{extractValueEXPIREAT},
		},
		{
			name: "EXPIREAT with past time",
			commands: []string{
				"SET k3 v3",
				"EXPIREAT k3 20",
				"GET k3",
			},
			expected:       []interface{}{"OK", true, ""},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueEXPIREAT, extractValueGET},
			delay:          []time.Duration{0, 0, 1 * time.Second},
		},
		{
			name: "EXPIREAT with invalid syntax",
			commands: []string{
				"EXPIREAT test_key",
			},
			expected:       []interface{}{errors.New("wrong number of arguments for 'EXPIREAT' command")},
			valueExtractor: []ValueExtractorFn{nil},
		},
		{
			name: "Test(NX): Set the expiration only if the key has no expiration time",
			commands: []string{
				"SET test_key test_value",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+10, 10) + " NX",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+1, 10) + " NX",
			},
			expected:       []interface{}{"OK", true, false},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueEXPIREAT, extractValueEXPIREAT},
		},

		{
			name: "Test(XX): Set the expiration only if the key already has an expiration time",
			commands: []string{
				"SET test_key test_value",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+10, 10) + " XX",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+10, 10) + " XX",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+10, 10) + " XX",
			},
			expected:       []interface{}{"OK", false, true, true},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueEXPIREAT, extractValueEXPIREAT, extractValueEXPIREAT},
		},
		{
			name: "Test if value is nil after expiration",
			commands: []string{
				"SET k2 v2",
				"EXPIREAT k2 " + strconv.FormatInt(time.Now().Unix()+2, 10) + " NX",
				"GET k2",
			},
			expected:       []interface{}{"OK", true, ""},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueEXPIREAT, extractValueGET},
			delay:          []time.Duration{0, 0, 4 * time.Second},
		},
		{
			name: "Invalid Command Test",
			commands: []string{
				"SET test_key test_value",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+1, 10) + " XX" + " " + "rr",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+1, 10) + " XX" + " " + "NX",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+1, 10) + " GT" + " " + "lt",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+1, 10) + " GT" + " " + "lt" + " " + "xx",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+1, 10) + " GT" + " " + "lt" + " " + "nx",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+1, 10) + " nx" + " " + "xx" + " " + "gt",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+1, 10) + " nx" + " " + "xx" + " " + "lt",
			},
			expected: []interface{}{"OK", errors.New("unsupported option rr"),
				errors.New("NX and XX, GT or LT options at the same time are not compatible"),
				errors.New("GT and LT options at the same time are not compatible"),
				errors.New("GT and LT options at the same time are not compatible"),
				errors.New("NX and XX, GT or LT options at the same time are not compatible"),
				errors.New("NX and XX, GT or LT options at the same time are not compatible"),
				errors.New("NX and XX, GT or LT options at the same time are not compatible")},
			valueExtractor: []ValueExtractorFn{extractValueSET, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil},
		},
		{
			name: "Test upper bound check for EXPIREAT",
			commands: []string{
				"SET test_key test_value",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().AddDate(11, 0, 0).Unix(), 10),
			},
			expected: []interface{}{
				"OK",
				errors.New("invalid expire time in 'EXPIREAT' command"),
			},
			valueExtractor: []ValueExtractorFn{extractValueSET, nil},
		},
	}
	runTestcases(t, client, testCases)
}
