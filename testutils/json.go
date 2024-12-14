// This file is part of DiceDB.
// Copyright (C) 2024 DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package testutils

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/dicedb/dice/internal/server/utils"
	"gotest.tools/v3/assert"
)

func IsJSONResponse(s string) bool {
	return s != utils.EmptyStr && (sonic.ValidString(s))
}

func AssertJSONEqualList(t *testing.T, expected []string, actual string) {
	res := false
	for _, exp := range expected {
		if compareJSONs(t, exp, actual) {
			res = true
			break
		}
	}
	if !res {
		t.Errorf("JSON not equal.\nExpected one of: %s\nActual: %s", expected, actual)
	}
}

func compareJSONs(t *testing.T, expected, actual string) bool {
	var expectedJSON, actualJSON interface{}

	err := sonic.UnmarshalString(expected, &expectedJSON)
	assert.NilError(t, err, "Failed to unmarshal expected JSON")

	err = sonic.UnmarshalString(actual, &actualJSON)
	assert.NilError(t, err, "Failed to unmarshal actual JSON")

	return reflect.DeepEqual(NormalizeJSON(expectedJSON), NormalizeJSON(actualJSON))
}

func NormalizeJSON(v interface{}) interface{} {
	switch v := v.(type) {
	case map[string]interface{}:
		normalizedMap := make(map[string]interface{})
		for k, val := range v {
			normalizedMap[k] = NormalizeJSON(val)
		}
		return normalizedMap
	case []interface{}:
		normalizedArray := make([]interface{}, len(v))
		for i, e := range v {
			normalizedArray[i] = NormalizeJSON(e)
		}
		return normalizedArray
	default:
		return v
	}
}

func CompareJSON(t *testing.T, expected, actual string) {
	var expectedMap map[string]interface{}
	var actualMap map[string]interface{}

	err1 := json.Unmarshal([]byte(expected), &expectedMap)
	err2 := json.Unmarshal([]byte(actual), &actualMap)

	assert.NilError(t, err1)
	assert.NilError(t, err2)

	assert.DeepEqual(t, expectedMap, actualMap)
}

func ConvertToArray(input string) []string {
	input = strings.Trim(input, `"[`)
	input = strings.Trim(input, `]"`)
	elements := strings.Split(input, ",")
	for i, element := range elements {
		elements[i] = strings.TrimSpace(element)
	}
	return elements
}

func IsJSONString(s string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
}

func ArraysArePermutations[T comparable](a, b []T) bool {
	// If lengths are different, they cannot be permutations
	if len(a) != len(b) {
		return false
	}

	// Count occurrences of each element in array 'a'
	countA := make(map[T]int)
	for _, elem := range a {
		countA[elem]++
	}

	// Subtract occurrences based on array 'b'
	for _, elem := range b {
		countA[elem]--
		if countA[elem] < 0 {
			return false
		}
	}

	// Check if all counts are zero
	for _, count := range countA {
		if count != 0 {
			return false
		}
	}

	return true
}
