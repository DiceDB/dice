// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"fmt"
	"math"
	"strconv"
	"testing"

	"github.com/dicedb/dicedb-go/wire"
)

func TestINCR(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name: "Increment multiple keys",
			commands: []string{
				"SET key1 0",
				"INCR key1",
				"INCR key1",
				"INCR key2",
				"GET key1",
				"GET key2",
			},
			expected: []interface{}{"OK", 1, 2, 1, 2, 1},
		},
		{
			name: "Increment to and from max int64",
			commands: []string{
				"SET max_int " + strconv.FormatInt(int64(math.MaxInt64-1), 10),
				"INCR max_int",
				"INCR max_int",
				"SET max_int " + strconv.FormatInt(int64(math.MaxInt64), 10),
				"INCR max_int",
			},
			expected: []interface{}{
				"OK", int64(math.MaxInt64), int64(math.MinInt64), "OK", int64(math.MinInt),
			},
		},
		{
			name: "Increment from min int64",
			commands: []string{
				"SET min_int " + strconv.FormatInt(int64(math.MinInt64), 10),
				"INCR min_int",
				"INCR min_int",
			},
			expected: []interface{}{
				"OK", int64(math.MinInt64 + 1), int64(math.MinInt64 + 2),
			},
		},
		{
			name: "Increment non-integer values",
			commands: []string{
				"SET float_key 3.14",
				"INCR float_key",
				"SET string_key hello",
				"INCR string_key",
				"SET bool_key true",
				"INCR bool_key",
			},
			expected: []interface{}{
				"OK",
				fmt.Errorf("wrong type operation for 'INCR' command"),
				"OK",
				fmt.Errorf("wrong type operation for 'INCR' command"),
				"OK",
				fmt.Errorf("wrong type operation for 'INCR' command"),
			},
		},
		{
			name: "Increment non-existent key",
			commands: []string{
				"INCR non_existent",
				"GET non_existent",
				"INCR non_existent",
			},
			expected: []interface{}{
				1, 1, 2,
			},
		},
		{
			name: "Increment string representing integers",
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
		},
	}
	// Clean up keys before each test case
	keys := []string{
		"key1", "key2", "max_int", "min_int", "float_key", "string_key", "bool_key",
		"non_existent", "str_int1", "str_int2", "str_int3", "expiry_key",
	}
	for _, key := range keys {
		client.Fire(&wire.Command{
			Cmd:  "DEL",
			Args: []string{key},
		})
	}
	runTestcases(t, client, testCases)

}

func TestINCRBY(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	type SetCommand struct {
		key string
		val int64
	}

	type IncrByCommand struct {
		key         string
		decrValue   any
		expectedVal int64
		expectedErr string
	}

	type GetCommand struct {
		key         string
		expectedVal int64
	}

	testCases := []TestCase{
		{
			name: "happy flow",
			commands: []string{
				"SET key 3", "GET key", "INCRBY key 2", "INCRBY key 1", "GET key",
			},
			expected: []interface{}{
				"OK", 3, 5, 6, 6,
			},
		},
		{
			name: "happy flow with negative increment",
			commands: []string{
				"SET key 100",
				"INCRBY key -2",
				"INCRBY key -10",
				"INCRBY key -88",
				"INCRBY key -100",
				"GET key",
			},
			expected: []interface{}{
				"OK", 98, 88, 0, -100, -100,
			},
		},
		{
			name: "happy flow with unset key",
			commands: []string{
				"SET key 3",
				"INCRBY unsetKey 2",
				"GET key",
				"GET unsetKey",
			},
			expected: []interface{}{
				"OK", 2, 3, 2, -100, -100,
			},
		},
		{
			name: "edge case with maxInt64",
			commands: []string{
				"SET key " + strconv.FormatInt(math.MaxInt64-1, 10),
				"INCRBY key 1",
				"INCRBY key 1",
				"GET key",
			},
			expected: []interface{}{
				"OK", math.MaxInt64, math.MinInt64, math.MinInt64,
			},
		},
		{
			name: "edge case with negative increment",
			commands: []string{
				"SET key " + strconv.FormatInt(math.MinInt64+1, 10),
				"INCRBY key -1",
				"INCRBY key -1",
				"GET key",
			},
			expected: []interface{}{
				"OK", math.MinInt64, math.MaxInt64, math.MaxInt64,
			},
		},
		{
			name: "edge case with string values",
			commands: []string{
				"SET key 1",
				"INCRBY key abc",
			},
			expected: []interface{}{
				"OK", fmt.Errorf("ERR value is not an integer or out of range"),
			},
		},
	}
	keys := []string{"key", "unsetKey", "stringkey"}
	for _, key := range keys {
		client.Fire(&wire.Command{
			Cmd:  "DEL",
			Args: []string{key},
		})
	}
	runTestcases(t, client, testCases)
}
