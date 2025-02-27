// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"strconv"
	"testing"
	"time"
)

func TestEXPIREAT(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name: "Set with EXPIREAT command",
			commands: []string{
				"SET test_key test_value",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+1, 10),
			},
			expected: []interface{}{"OK", 1},
		},
		{
			name: "Check if key is nil after expiration",
			commands: []string{
				"SET test_key test_value",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+1, 10),
				"GET test_key",
			},
			expected: []interface{}{"OK", 1, nil},
			delay:    []time.Duration{0, 2 * time.Second},
		},
		{
			name: "EXPIREAT non-existent key",
			commands: []string{
				"EXPIREAT non_existent_key " + strconv.FormatInt(time.Now().Unix()+1, 10),
			},
			expected: []interface{}{int64(0)},
		},
		{
			name: "EXPIREAT with past time",
			commands: []string{
				"SET test_key test_value",
				"EXPIREAT test_key " + strconv.FormatInt(-1, 10),
				"GET test_key",
			},
			expected: []interface{}{"OK", 1, errors.New("invalid expire time in 'EXPIREAT' command"), "test_value"},
		},
		{
			name: "EXPIREAT with invalid syntax",
			commands: []string{
				"EXPIREAT test_key",
			},
			expected: []interface{}{"ERR wrong number of arguments for 'expireat' command"},
		},
		{
			name: "Test(NX): Set the expiration only if the key has no expiration time",
			commands: []string{
				"SET test_key test_value",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+10, 10) + " NX",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+1, 10) + " NX",
			},
			expected: []interface{}{"OK", 1, int64(0)},
		},

		{
			name: "Test(XX): Set the expiration only if the key already has an expiration time",
			commands: []string{
				"SET test_key test_value",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+10, 10) + " XX",
				"TTL test_key",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+10, 10),
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+10, 10) + " XX",
			},
			expected: []interface{}{"OK", int64(0), int64(-1), 1, 1},
		},

		{
			name: "TEST(GT): Set the expiration only if the new expiration time is greater than the current one",
			commands: []string{
				"SET test_key test_value",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+10, 10) + " GT",
				"TTL test_key",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+10, 10),
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+20, 10) + " GT",
			},
			expected: []interface{}{"OK", int64(0), int64(-1), 1, 1},
		},

		{
			name: "TEST(LT): Set the expiration only if the new expiration time is less than the current one",
			commands: []string{
				"SET test_key test_value",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+10, 10) + " LT",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+20, 10) + " LT",
			},
			expected: []interface{}{"OK", 1, int64(0)},
		},

		{
			name: "TEST(LT): Set the expiration only if the new expiration time is less than the current one",
			commands: []string{
				"SET test_key test_value",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+10, 10) + " LT",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+20, 10) + " LT",
			},
			expected: []interface{}{"OK", 1, int64(0)},
		},

		{
			name: "TEST(NX + LT/GT)",
			commands: []string{
				"SET test_key test_value",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+20, 10) + " NX",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+20, 10) + " NX" + " LT",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+20, 10) + " NX" + " GT",
				"GET test_key",
			},
			expected: []interface{}{"OK", 1,
				"ERR NX and XX, GT or LT options at the same time are not compatible",
				"ERR NX and XX, GT or LT options at the same time are not compatible",
				"test_value"},
		},
		{
			name: "TEST(XX + LT/GT)",
			commands: []string{
				"SET test_key test_value",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+20, 10),
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+5, 10) + " XX" + " LT",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+10, 10) + " XX" + " GT",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+20, 10) + " XX" + " GT",
				"GET test_key",
			},
			expected: []interface{}{"OK", 1, 1, 1, 1, "test_value"},
		},
		{
			name: "Test if value is nil after expiration",
			commands: []string{
				"SET test_key test_value",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+20, 10),
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+2, 10) + " XX" + " LT",
				"GET test_key",
			},
			expected: []interface{}{"OK", 1, 1, nil},
			delay:    []time.Duration{0, 0, 0, 2 * time.Second},
		},
		{
			name: "Test if value is nil after expiration",
			commands: []string{
				"SET test_key test_value",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+2, 10) + " NX",
				"GET test_key",
			},
			expected: []interface{}{"OK", 1, nil},
			delay:    []time.Duration{0, 0, 2 * time.Second},
		},
		{
			name: "Invalid Command Test",
			commands: []string{
				"SET test_key test_value",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+1, 10) + " XX" + " " + "rr",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+1, 10) + " XX" + " " + "NX",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+1, 10) + " GT" + " " + "lt",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+1, 10) + " GT" + " " + "lt" + " " + "xx",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+1, 10) + " GT" + " " + "lt" + " " + "nx",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+1, 10) + " nx" + " " + "xx" + " " + "gt",
				"EXPIREAT test_key " + strconv.FormatInt(time.Now().Unix()+1, 10) + " nx" + " " + "xx" + " " + "lt",
			},
			expected: []interface{}{"OK", "ERR Unsupported option rr",
				"ERR NX and XX, GT or LT options at the same time are not compatible",
				"ERR GT and LT options at the same time are not compatible",
				"ERR GT and LT options at the same time are not compatible",
				"ERR NX and XX, GT or LT options at the same time are not compatible",
				"ERR NX and XX, GT or LT options at the same time are not compatible",
				"ERR NX and XX, GT or LT options at the same time are not compatible"},
		},
	}
	runTestcases(t, client, testCases)
}
