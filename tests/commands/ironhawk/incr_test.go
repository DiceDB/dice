// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/dicedb/dicedb-go/wire"
)

func extractValueINCR(result *wire.Result) interface{} {
	return result.GetINCRRes().GetValue()
}

func TestINCR(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name: "Increment multiple keys",
			commands: []string{
				"SET key1 0", "INCR key1", "INCR key1", "INCR key2", "GET key1", "GET key2",
			},
			expected:       []interface{}{"OK", 1, 2, 1, "2", "1"},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueINCR, extractValueINCR, extractValueINCR, extractValueGET, extractValueGET},
		},
		{
			name: "Increment max int64 and expect min int64 (rollover)",
			commands: []string{
				"SET max_int " + fmt.Sprintf("%d", math.MaxInt64-1),
				"INCR max_int",
				"INCR max_int",
				"SET max_int " + fmt.Sprintf("%d", math.MaxInt64),
				"INCR max_int",
			},
			expected: []interface{}{
				"OK", math.MaxInt64, math.MinInt64, "OK", math.MinInt,
			},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueINCR, extractValueINCR, extractValueSET, extractValueINCR},
		},
		{
			name: "Increment from min int64",
			commands: []string{
				"SET min_int " + fmt.Sprintf("%d", math.MinInt64),
				"INCR min_int",
				"INCR min_int",
			},
			expected: []interface{}{
				"OK", math.MinInt64 + 1, math.MinInt64 + 2,
			},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueINCR, extractValueINCR},
		},
		{
			name: "Increment non-integer values and get type error",
			commands: []string{
				"SET float_key 3.14",
				"INCR float_key",
				"SET string_key hello",
				"INCR string_key",
			},
			expected: []interface{}{
				"OK",
				errors.New("wrongtype operation against a key holding the wrong kind of value"),
				"OK",
				errors.New("wrongtype operation against a key holding the wrong kind of value"),
			},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueINCR, extractValueSET, extractValueINCR},
		},
		{
			name: "Increment non-existent key and expect keys to be created",
			commands: []string{
				"INCR non_existent",
				"GET non_existent",
				"INCR non_existent",
			},
			expected: []interface{}{
				1, "1", 2,
			},
			valueExtractor: []ValueExtractorFn{extractValueINCR, extractValueGET, extractValueINCR},
		},
		{
			name: "Increment string representing integers and get type error",
			commands: []string{
				"SET str_int1 42",
				"INCR str_int1",
				"SET str_int2 -10",
				"INCR str_int2",
				"SET str_int2 0",
				"INCR str_int3",
			},
			expected: []interface{}{
				"OK", 43, "OK", -9, "OK", 1,
			},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueINCR,
				extractValueSET, extractValueINCR, extractValueSET, extractValueINCR},
		},
	}
	runTestcases(t, client, testCases)
}
