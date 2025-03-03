// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"strconv"
	"testing"
	"time"
)

func TestEXPIRETIME(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	futureUnixTimestamp := time.Now().Unix() + 1

	testCases := []TestCase{
		{
			name: "EXPIRETIME command",
			commands: []string{
				"SET test_key test_value",
				"EXPIREAT test_key " + strconv.FormatInt(futureUnixTimestamp, 10),
				"EXPIRETIME test_key",
			},
			expected: []interface{}{"OK", 1, futureUnixTimestamp},
		},
		{
			name: "EXPIRETIME non-existent key",
			commands: []string{
				"EXPIRETIME non_existent_key",
			},
			expected: []interface{}{int64(-2)},
		},
		{
			name: "EXPIRETIME with past time",
			commands: []string{
				"SET test_key test_value",
				"EXPIREAT test_key 1724167183",
				"EXPIRETIME test_key",
			},
			expected: []interface{}{"OK", 1, int64(-2)},
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
		},
	}
	runTestcases(t, client, testCases)
}
