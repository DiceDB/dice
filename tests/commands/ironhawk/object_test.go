// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestObjectCommand(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()
	defer client.FireString("FLUSHDB")

	testCases := []struct {
		name       string
		commands   []string
		expected   []interface{}
		assertType []string
		delay      []time.Duration
		cleanup    []string
	}{
		{
			name:       "Object Idletime",
			commands:   []string{"SET foo bar", "OBJECT IDLETIME foo", "OBJECT IDLETIME foo", "TOUCH foo", "OBJECT IDLETIME foo"},
			expected:   []interface{}{"OK", int64(2), int64(3), int64(1), int64(0)},
			assertType: []string{"equal", "assert", "assert", "equal", "assert"},
			delay:      []time.Duration{0, 2 * time.Second, 3 * time.Second, 0, 0},
			cleanup:    []string{"DEL foo"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// deleteTestKeys([]string{"foo"}, store)
			client.FireString("DEL foo")

			for i, cmd := range tc.commands {
				if tc.delay[i] != 0 {
					time.Sleep(tc.delay[i])
				}

				result := client.FireString(cmd)

				if tc.assertType[i] == "equal" {
					assert.Equal(t, tc.expected[i], result)
				} else {
					assert.True(t, result.GetVInt() >= tc.expected[i].(int64), "Expected %v to be less than or equal to %v", result, tc.expected[i])
				}
			}
			for _, cmd := range tc.cleanup { // run cleanup
				client.FireString(cmd)
			}
		})
	}
}
