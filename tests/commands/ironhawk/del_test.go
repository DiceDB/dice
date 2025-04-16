// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"

	"github.com/dicedb/dicedb-go/wire"
)

func extractValueDEL(result *wire.Result) interface{} {
	return result.GetDELRes().GetCount()
}

func TestDEL(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:           "DEL with set key",
			commands:       []string{"SET k1 v1", "DEL k1", "GET k1"},
			expected:       []interface{}{"OK", 1, ""},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueDEL, extractValueGET},
		},
		{
			name:           "DEL multiple keys",
			commands:       []string{"SET k1 v1", "SET k2 v2", "DEL k1 k2", "GET k1", "GET k2"},
			expected:       []interface{}{"OK", "OK", 2, "", ""},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueSET, extractValueDEL, extractValueGET, extractValueGET},
		},
		{
			name:           "DEL with key not set",
			commands:       []string{"GET k3", "DEL k3"},
			expected:       []interface{}{"", 0},
			valueExtractor: []ValueExtractorFn{extractValueGET, extractValueDEL},
		},
		{
			name:           "DEL with no keys or arguments",
			commands:       []string{"DEL"},
			expected:       []interface{}{0},
			valueExtractor: []ValueExtractorFn{extractValueDEL},
		},
	}
	runTestcases(t, client, testCases)
}
