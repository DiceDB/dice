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

func extractValueEXPIRETIME(res *wire.Result) interface{} {
	return res.GetEXPIRETIMERes().UnixSec
}

func TestEXPIRETIME(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	expireAtUnixSec := time.Now().Unix() + 10

	testCases := []TestCase{
		{
			name: "EXPIRETIME command",
			commands: []string{
				"SET test_key test_value",
				"EXPIREAT test_key " + strconv.FormatInt(expireAtUnixSec, 10),
				"EXPIRETIME test_key",
			},
			expected:       []interface{}{"OK", true, expireAtUnixSec},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueEXPIREAT, extractValueEXPIRETIME},
		},
		{
			name: "EXPIRETIME non-existent key",
			commands: []string{
				"EXPIRETIME non_existent_key",
			},
			expected:       []interface{}{-2},
			valueExtractor: []ValueExtractorFn{extractValueEXPIRETIME},
		},
		{
			name: "EXPIRETIME with past time",
			commands: []string{
				"SET k1 v1",
				"EXPIREAT k1 100",
				"EXPIRETIME k1",
			},
			expected:       []interface{}{"OK", true, -2},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueEXPIREAT, extractValueEXPIRETIME},
		},
		{
			name: "EXPIRETIME with invalid syntax",
			commands: []string{
				"SET test_key test_value",
				"EXPIRETIME",
				"EXPIRETIME key1 key2",
			},
			expected: []interface{}{
				"OK",
				errors.New("wrong number of arguments for 'EXPIRETIME' command"),
				errors.New("wrong number of arguments for 'EXPIRETIME' command"),
			},
			valueExtractor: []ValueExtractorFn{extractValueSET, nil, nil},
		},
	}
	runTestcases(t, client, testCases)
}
