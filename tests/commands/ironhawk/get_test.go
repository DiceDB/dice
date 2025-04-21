// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"
	"time"

	"github.com/dicedb/dicedb-go/wire"
)

func extractValueGET(result *wire.Result) interface{} {
	return result.GetGETRes().Value
}

func TestGET(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:           "Get with expiration",
			commands:       []string{"SET k978 v EX 2", "GET k978", "GET k978"},
			expected:       []interface{}{"OK", "v", ""},
			delay:          []time.Duration{0, 0, 5 * time.Second},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueGET, extractValueGET},
		},
		{
			name:           "Get without expiration",
			commands:       []string{"SET k v", "GET k"},
			expected:       []interface{}{"OK", "v"},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueGET},
		},
		{
			name:           "Set Floating Point Value",
			commands:       []string{"SET fp 123.123", "GET fp"},
			expected:       []interface{}{"OK", "123.123000"},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueGET},
		},
		{
			name:           "Get on other ObjType should give error",
			commands:       []string{"HSET map k1 v1", "GET map"},
			expected:       []interface{}{"OK", errors.New("unknown object type")},
			valueExtractor: []ValueExtractorFn{extractValueSET, nil},
		},
		{
			name:           "Get with non existent key",
			commands:       []string{"GET nek"},
			expected:       []interface{}{""},
			valueExtractor: []ValueExtractorFn{extractValueGET},
		},
		{
			name:     "GET with no keys or arguments",
			commands: []string{"GET"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'GET' command"),
			},
			valueExtractor: []ValueExtractorFn{nil},
		},
	}

	runTestcases(t, client, testCases)
}
