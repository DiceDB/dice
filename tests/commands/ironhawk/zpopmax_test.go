// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"fmt"
	"sort"
	"testing"

	"github.com/dicedb/dicedb-go/wire"
)

func extractValueZPOPMAX(res *wire.Result) interface{} {
	elements := res.GetZPOPMAXRes().Elements
	sort.Slice(elements, func(i, j int) bool {
		return elements[i].Score > elements[j].Score
	})

	str := ""
	for _, element := range elements {
		str += fmt.Sprintf("%d, %s\n", element.Score, element.Member)
	}
	return str
}

func TestZPOPMAX(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:           "ZPOPMAX on non-existing key without count argument",
			commands:       []string{"ZPOPMAX NON_EXISTENT_KEY"},
			expected:       []interface{}{""},
			valueExtractor: []ValueExtractorFn{extractValueZPOPMAX},
		},
		{
			name:           "ZPOPMAX with wrong type of key without count argument",
			commands:       []string{"SET stringkey string_value", "ZPOPMAX stringkey"},
			expected:       []interface{}{"OK", errors.New("wrongtype operation against a key holding the wrong kind of value")},
			valueExtractor: []ValueExtractorFn{extractValueSET, nil},
		},
		{
			name:           "ZPOPMAX on existing key without count argument",
			commands:       []string{"ZADD ss 1 m1 2 m2 3 m3", "ZPOPMAX ss", "ZCOUNT ss 1 10"},
			expected:       []interface{}{3, "3, m3\n", 2},
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZPOPMAX, extractValueZCOUNT},
		},
		{
			name:           "ZPOPMAX with normal count argument",
			commands:       []string{"ZADD ss1 1 m1 2 m2 3 m3", "ZPOPMAX ss1 2", "ZCOUNT ss1 1 2"},
			expected:       []interface{}{3, "3, m3\n2, m2\n", 1},
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZPOPMAX, extractValueZCOUNT},
		},
		{
			name:           "ZPOPMAX with count argument but multiple members have the same score",
			commands:       []string{"ZADD ss2 1 m1 1 m2 1 m3", "ZPOPMAX ss2 2", "ZCOUNT ss2 1 1"},
			expected:       []interface{}{3, "1, m3\n1, m2\n", 1},
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZPOPMAX, extractValueZCOUNT},
		},
		{
			name:           "ZPOPMAX with negative count argument",
			commands:       []string{"ZADD ss3 1 m1 2 m2 3 m3", "ZPOPMAX ss3 -1"},
			expected:       []interface{}{3, errors.New("value is not an integer or out of range"), 3},
			valueExtractor: []ValueExtractorFn{extractValueZADD, nil},
		},
		{
			name:           "ZPOPMAX with invalid count argument",
			commands:       []string{"ZADD ss4 1 m1 2 m2 3 m3", "ZPOPMAX ss4 INCORRECT_COUNT_ARGUMENT"},
			expected:       []interface{}{3, errors.New("value is not an integer or out of range")},
			valueExtractor: []ValueExtractorFn{extractValueZADD, nil},
		},
		{
			name:           "ZPOPMAX with count argument greater than length of sorted set",
			commands:       []string{"ZADD ss5 1 m1 2 m2 3 m3", "ZPOPMAX ss5 10", "ZCOUNT ss5 1 3"},
			expected:       []interface{}{3, "3, m3\n2, m2\n1, m1\n", 0},
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZPOPMAX, extractValueZCOUNT},
		},
		{
			name:           "ZPOPMAX on empty sorted set",
			commands:       []string{"ZADD ss6 1 m1", "ZPOPMAX ss6 1", "ZPOPMAX ss6", "ZCOUNT ss6 0 3"},
			expected:       []interface{}{1, "1, m1\n", "", 0},
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZPOPMAX, extractValueZPOPMAX, extractValueZCOUNT},
		},
	}

	runTestcases(t, client, testCases)
}
