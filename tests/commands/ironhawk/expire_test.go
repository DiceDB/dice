// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"strconv"
	"testing"
	"time"
)

func TestEXPIRE(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name: "Set with EXPIRE command",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key 1",
			},
			expected: []interface{}{"OK", 1},
		},
		{
			name: "Check if key is nil after expiration",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key 1",
				"GET test_key",
			},
			expected: []interface{}{"OK", 1, nil},
			delay:    []time.Duration{0, 0, 2 * time.Second},
		},
		{
			name: "EXPIRE non-existent key",
			commands: []string{
				"EXPIRE non_existent_key 1",
			},
			expected: []interface{}{0},
		},
		{
			name: "EXPIRE with past time",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key -1",
				"GET test_key",
			},
			expected: []interface{}{"OK", errors.New("invalid expire time in 'expire' command"), "test_value"},
		},
		{
			name: "EXPIRE with invalid syntax",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key",
			},
			expected: []interface{}{"OK", errors.New("wrong number of arguments for 'EXPIRE' command")},
		},
		{
			name: "Test(NX): Set the expiration only if the key has no expiration time",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key " + strconv.FormatInt(1, 10) + " NX",
				"EXPIRE test_key " + strconv.FormatInt(1, 10) + " NX",
			},
			expected: []interface{}{"OK", 1, 0},
		},
		{
			name: "Test(XX): Set the expiration only if the key already has an expiration time",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key " + strconv.FormatInt(10, 10) + " XX",
				"TTL test_key",
				"EXPIRE test_key " + strconv.FormatInt(10, 10),
				"EXPIRE test_key " + strconv.FormatInt(10, 10) + " XX",
			},
			expected: []interface{}{"OK", 0, -1, 1, 1},
		},
		{
			name: "TEST(GT): Set the expiration only if the new expiration time is greater than the current one",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key " + strconv.FormatInt(10, 10) + " GT",
				"TTL test_key",
				"EXPIRE test_key " + strconv.FormatInt(10, 10),
				"EXPIRE test_key " + strconv.FormatInt(20, 10) + " GT",
			},
			expected: []interface{}{"OK", 0, -1, 1, 1},
		},
		{
			name: "TEST(LT): Set the expiration only if the new expiration time is less than the current one",
			commands: []string{
				"SET test_key test_value EX 15",
				"EXPIRE test_key " + strconv.FormatInt(10, 10) + " LT",
				"EXPIRE test_key " + strconv.FormatInt(20, 10) + " LT",
			},
			expected: []interface{}{"OK", 1, 0},
		},
		{
			name: "TEST(LT): Set the expiration only if the new expiration time is less than the current one",
			commands: []string{
				"SET test_key test_value EX 15",
				"EXPIRE test_key " + strconv.FormatInt(10, 10) + " LT",
				"EXPIRE test_key " + strconv.FormatInt(20, 10) + " LT",
			},
			expected: []interface{}{"OK", 1, 0},
		},
		{
			name: "TEST(NX + LT/GT)",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key " + strconv.FormatInt(20, 10) + " NX",
				"EXPIRE test_key " + strconv.FormatInt(20, 10) + " NX" + " LT",
				"EXPIRE test_key " + strconv.FormatInt(20, 10) + " NX" + " GT",
				"GET test_key",
			},
			expected: []interface{}{"OK", 1,
				errors.New("NX and XX, GT or LT options at the same time are not compatible"),
				errors.New("NX and XX, GT or LT options at the same time are not compatible"),
				"test_value"},
		},
		{
			name: "TEST(XX + LT/GT)",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key " + strconv.FormatInt(20, 10),
				"EXPIRE test_key " + strconv.FormatInt(5, 10) + " XX" + " LT",
				"EXPIRE test_key " + strconv.FormatInt(10, 10) + " XX" + " GT",
				"EXPIRE test_key " + strconv.FormatInt(20, 10) + " XX" + " GT",
				"GET test_key",
			},
			expected: []interface{}{"OK", 1, 1, 1, 1, "test_value"},
		},
		{
			name: "Test if value is nil after expiration",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key " + strconv.FormatInt(20, 10),
				"EXPIRE test_key " + strconv.FormatInt(2, 10) + " XX" + " LT",
				"GET test_key",
			},
			expected: []interface{}{"OK", 1, 1, nil},
			delay:    []time.Duration{0, 0, 0, 2 * time.Second},
		},
		{
			name: "Test if value is nil after expiration",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key " + strconv.FormatInt(2, 10) + " NX",
				"GET test_key",
			},
			expected: []interface{}{"OK", 1, nil},
			delay:    []time.Duration{0, 0, 2 * time.Second},
		},
		{
			name: "Invalid Command Test",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key " + strconv.FormatInt(1, 10) + " XX" + " " + "rr",
				"EXPIRE test_key " + strconv.FormatInt(1, 10) + " XX" + " " + "NX",
				"EXPIRE test_key " + strconv.FormatInt(1, 10) + " GT" + " " + "lt",
				"EXPIRE test_key " + strconv.FormatInt(1, 10) + " GT" + " " + "lt" + " " + "xx",
				"EXPIRE test_key " + strconv.FormatInt(1, 10) + " GT" + " " + "lt" + " " + "nx",
				"EXPIRE test_key " + strconv.FormatInt(1, 10) + " nx" + " " + "xx" + " " + "gt",
				"EXPIRE test_key " + strconv.FormatInt(1, 10) + " nx" + " " + "xx" + " " + "lt",
			},
			expected: []interface{}{"OK", errors.New("unsupported option rr"),
				errors.New("NX and XX, GT or LT options at the same time are not compatible"),
				errors.New("GT and LT options at the same time are not compatible"),
				errors.New("GT and LT options at the same time are not compatible"),
				errors.New("NX and XX, GT or LT options at the same time are not compatible"),
				errors.New("NX and XX, GT or LT options at the same time are not compatible"),
				errors.New("NX and XX, GT or LT options at the same time are not compatible"),
			},
		},
		{
			name:     "EXPIRE with no keys or arguments",
			commands: []string{"EXPIRE"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'EXPIRE' command"),
			},
		},
	}
	runTestcases(t, client, testCases)
}
