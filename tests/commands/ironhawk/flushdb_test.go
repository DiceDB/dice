// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"

	"github.com/dicedb/dicedb-go/wire"
)

func extractValueFLUSHDB(result *wire.Result) interface{} {
	return result.GetMessage()
}

func TestFLUSHDB(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name: "FLUSHDB",
			commands: []string{
				"SET k1 v1",
				"SET k2 v2",
				"SET k3 v3",
				"FLUSHDB",
				"GET k1",
				"GET k2",
				"GET k3",
			},
			expected:       []interface{}{"OK", "OK", "OK", "OK", "", "", ""},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueSET, extractValueSET, extractValueFLUSHDB, extractValueGET, extractValueGET, extractValueGET},
		},
	}

	runTestcases(t, client, testCases)
}
