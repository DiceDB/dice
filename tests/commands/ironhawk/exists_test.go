// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"
	"time"

	"github.com/dicedb/dicedb-go/wire"
)

func extractValueEXISTS(result *wire.Result) interface{} {
	return result.GetEXISTSRes().GetCount()
}

func TestEXISTS(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:           "Test EXISTS command",
			commands:       []string{"SET key value", "EXISTS key", "EXISTS key2"},
			expected:       []interface{}{"OK", 1, 0},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueEXISTS, extractValueEXISTS},
		},
		{
			name:           "Test EXISTS command with multiple keys",
			commands:       []string{"SET key value", "SET key2 value2", "EXISTS key key2 key3", "EXISTS key key2 key3 key4", "DEL key", "EXISTS key key2 key3 key4"},
			expected:       []interface{}{"OK", "OK", 2, 2, 1, 1},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueSET, extractValueEXISTS, extractValueEXISTS, extractValueDEL, extractValueEXISTS},
		},
		{
			name:           "Test EXISTS an expired key",
			commands:       []string{"SET key value ex 1", "EXISTS key", "EXISTS key"},
			expected:       []interface{}{"OK", 1, 0},
			delay:          []time.Duration{0, 0, 3 * time.Second},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueEXISTS, extractValueEXISTS},
		},
		{
			name:           "Test EXISTS with multiple keys and expired key",
			commands:       []string{"SET key value ex 2", "SET key2 value2", "SET key3 value3", "EXISTS key key2 key3", "EXISTS key key2 key3"},
			expected:       []interface{}{"OK", "OK", "OK", 3, 2},
			delay:          []time.Duration{0, 0, 0, 0, 4 * time.Second},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueSET, extractValueSET, extractValueEXISTS, extractValueEXISTS},
		},
		{
			name:           "EXISTS with no keys or arguments",
			commands:       []string{"EXISTS"},
			expected:       []interface{}{0},
			valueExtractor: []ValueExtractorFn{extractValueEXISTS},
		},
		{
			name:           "EXISTS with duplicate keys",
			commands:       []string{"SET key value", "EXISTS key key"},
			expected:       []interface{}{"OK", 2},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueEXISTS},
		},
		{
			name:           "EXISTS with duplicate non existent keys",
			commands:       []string{"SET key value", "EXISTS key neq neq2"},
			expected:       []interface{}{"OK", 1},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueEXISTS},
		},
		{
			name:           "EXISTS with duplicate keys multiple keys",
			commands:       []string{"SET key value", "SET key1 value", "EXISTS key key key1"},
			expected:       []interface{}{"OK", "OK", 3},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueSET, extractValueEXISTS},
		},
	}

	runTestcases(t, client, testCases)
}
