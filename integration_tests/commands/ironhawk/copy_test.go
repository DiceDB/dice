// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"
	"time"

	"github.com/dicedb/dice/testutils"
	"github.com/dicedb/dicedb-go/wire"
	"github.com/stretchr/testify/assert"
)

func TestCopy(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	simpleJSON := `{"name":"John","age":30}`

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
		delays   []time.Duration
	}{
		{
			name:     "COPY when source key doesn't exist",
			commands: []string{"COPY k1 k2"},
			expected: []interface{}{int64(0)},
			delays:   []time.Duration{0},
		},
		{
			name:     "COPY with no REPLACE",
			commands: []string{"SET k1 v1", "COPY k1 k2", "GET k1", "GET k2"},
			expected: []interface{}{"OK", int64(1), "v1", "v1"},
			delays:   []time.Duration{0, 0, 0, 0},
		},
		{
			name:     "COPY with REPLACE",
			commands: []string{"SET k1 v1", "SET k2 v2", "GET k2", "COPY k1 k2 REPLACE", "GET k2"},
			expected: []interface{}{"OK", "OK", "v2", int64(1), "v1"},
			delays:   []time.Duration{0, 0, 0, 0, 0},
		},
		{
			name:     "COPY with JSON integer",
			commands: []string{"JSON.SET k1 $ 2", "COPY k1 k2", "JSON.GET k2"},
			expected: []interface{}{"OK", int64(1), "2"},
			delays:   []time.Duration{0, 0, 0},
		},
		{
			name:     "COPY with JSON boolean",
			commands: []string{"JSON.SET k1 $ true", "COPY k1 k2", "JSON.GET k2"},
			expected: []interface{}{"OK", int64(1), "true"},
			delays:   []time.Duration{0, 0, 0},
		},
		{
			name:     "COPY with JSON array",
			commands: []string{`JSON.SET k1 $ [1,2,3]`, "COPY k1 k2", "JSON.GET k2"},
			expected: []interface{}{"OK", int64(1), `[1,2,3]`},
			delays:   []time.Duration{0, 0, 0},
		},
		{
			name:     "COPY with JSON simple JSON",
			commands: []string{`JSON.SET k1 $ ` + simpleJSON, "COPY k1 k2", "JSON.GET k2"},
			expected: []interface{}{"OK", int64(1), simpleJSON},
			delays:   []time.Duration{0, 0, 0},
		},
		{
			name:     "COPY with no expiry",
			commands: []string{"SET k1 v1", "COPY k1 k2", "TTL k1", "TTL k2"},
			expected: []interface{}{"OK", int64(1), int64(-1), int64(-1)},
			delays:   []time.Duration{0, 0, 0, 0},
		},
		{
			name:     "COPY with expiry making sure copy expires",
			commands: []string{"SET k1 v1 EX 3", "COPY k1 k2", "GET k1", "GET k2", "GET k1", "GET k2"},
			expected: []interface{}{"OK", int64(1), "v1", "v1", "(nil)", "(nil)"},
			delays:   []time.Duration{0, 0, 0, 0, 3 * time.Second, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Doesn't seem to work for integration tests where command is not
			// cleaning up keys local store but for store in the server.
			// deleteTestKeys([]string{"k1", "k2"}, store)

			// Using this instead to clean up state before tests
			client.FireString("DEL k1")
			client.FireString("DEL k2")
			for i, cmd := range tc.commands {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result := client.FireString(cmd)
				resStr := result.Value.(*wire.Response_VStr).VStr
				expStr, expOk := tc.expected[i].(string)

				// If both are valid JSON strings, then compare the JSON values.
				// else compare the values as is.
				// This is to handle cases where the expected value is a json string with a different key order.
				if expOk && testutils.IsJSONResponse(resStr) && testutils.IsJSONResponse(expStr) {
					assert.JSONEq(t, expStr, resStr)
				} else {
					assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s", cmd)
				}
			}
		})
	}
}
