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

func extractValueSET(result *wire.Result) interface{} {
	// TODO: Changed this to return the IsChanged value
	// and make the test cases consume and reflect the same
	return result.GetMessage()
}

func TestSET(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	expiryTimeInt := time.Now().Add(1 * time.Minute).UnixMilli()
	expiryTime := strconv.FormatInt(expiryTimeInt, 10)

	testCases := []TestCase{
		{
			name:           "Set and Get Simple Value",
			commands:       []string{"SET k v", "GET k"},
			expected:       []interface{}{"OK", "v"},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueGET},
		},
		{
			name:           "Set and Get Integer Value",
			commands:       []string{"SET k 123456789", "GET k"},
			expected:       []interface{}{"OK", "123456789"},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueGET},
		},
		{
			name:           "Overwrite Existing Key",
			commands:       []string{"SET k v1", "SET k 5", "GET k"},
			expected:       []interface{}{"OK", "OK", "5"},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueSET, extractValueGET},
		},
		{
			name:           "Set with EX option",
			commands:       []string{"SET k100 v1 EX 2", "GET k100", "GET k100"},
			expected:       []interface{}{"OK", "v1", ""},
			delay:          []time.Duration{0, 0, 3 * time.Second},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueGET, extractValueGET},
		},
		{
			name:           "Set with PX option",
			commands:       []string{"SET k200 v2 PX 2000", "GET k200", "GET k200"},
			expected:       []interface{}{"OK", "v2", ""},
			delay:          []time.Duration{0, 0, 3 * time.Second},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueGET, extractValueGET},
		},
		{
			name:     "Set with EX and PX option",
			commands: []string{"SET k v EX 2 PX 2000"},
			expected: []interface{}{
				errors.New("invalid syntax for 'SET' command"),
			},
			valueExtractor: []ValueExtractorFn{nil},
		},
		{
			name:           "XX on non-existing key",
			commands:       []string{"SET k99 v XX", "GET k99"},
			expected:       []interface{}{"OK", ""},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueGET},
		},
		{
			name:           "NX on non-existing key",
			commands:       []string{"SET k1729 v NX", "GET k1729"},
			expected:       []interface{}{"OK", "v"},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueGET},
		},
		{
			name:           "NX on existing key",
			commands:       []string{"SET k1730 v NX", "GET k1730", "SET k1730 v2 NX", "GET k1730"},
			expected:       []interface{}{"OK", "v", "OK", "v"},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueGET, extractValueSET, extractValueGET},
		},
		{
			name:           "PXAT option",
			commands:       []string{"SET k v PXAT " + expiryTime, "GET k"},
			expected:       []interface{}{"OK", "v"},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueGET},
		},
		{
			name:     "PXAT option with invalid unix time ms",
			commands: []string{"SET k2 v2 PXAT 123123"},
			expected: []interface{}{
				errors.New("invalid value for a parameter in 'SET' command for PXAT parameter"),
			},
			valueExtractor: []ValueExtractorFn{nil},
		},
		{
			name:           "XX on existing key",
			commands:       []string{"SET k v1", "SET k v2 XX", "GET k"},
			expected:       []interface{}{"OK", "OK", "v2"},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueSET, extractValueGET},
		},
		{
			name:           "Multiple XX operations",
			commands:       []string{"SET k v1", "SET k v2 XX", "SET k v3 XX", "GET k"},
			expected:       []interface{}{"OK", "OK", "OK", "v3"},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueSET, extractValueSET, extractValueGET},
		},
		{
			name:           "EX option",
			commands:       []string{"SET k v EX 1", "GET k", "GET k"},
			expected:       []interface{}{"OK", "v", ""},
			delay:          []time.Duration{0, 0, 3 * time.Second},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueGET, extractValueGET},
		},
		{
			name:           "XX option",
			commands:       []string{"SET k9 v9 XX", "GET k9", "SET k9 v9", "GET k9", "SET k9 v10 XX", "GET k9"},
			expected:       []interface{}{"OK", "", "OK", "v9", "OK", "v10"},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueGET, extractValueSET, extractValueGET, extractValueSET, extractValueGET},
		},
		{
			name: "SET with KEEPTTL option",
			commands: []string{"SET k v EX 2", "SET k vv KEEPTTL", "GET k", "SET kk vv", "SET kk vvv KEEPTTL", "GET kk",
				"SET K V EX 2 KEEPTTL",
				"SET K1 vv PX 2000 KEEPTTL",
				"SET K2 vv EXAT " + expiryTime + " KEEPTTL"},
			expected: []interface{}{"OK", "OK", "vv", "OK", "OK", "vvv",
				errors.New("invalid syntax for 'SET' command"),
				errors.New("invalid syntax for 'SET' command"),
				errors.New("invalid syntax for 'SET' command"),
			},
			valueExtractor: []ValueExtractorFn{
				extractValueSET, extractValueSET, extractValueGET,
				extractValueSET, extractValueSET, extractValueGET,
				nil, nil, nil},
		},
		{
			name:     "SET with no keys or arguments",
			commands: []string{"SET"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'SET' command"),
			},
			valueExtractor: []ValueExtractorFn{extractValueSET},
		},
	}
	runTestcases(t, client, testCases)
}
