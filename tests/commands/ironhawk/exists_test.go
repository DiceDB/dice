// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/dicedb/dicedb-go/wire"
)

func TestExists(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name     string
		command  []string
		expected []any
		delay    []time.Duration
	}{
		{
			name:     "Test EXISTS command",
			command:  []string{"SET key value", "EXISTS key", "EXISTS key2"},
			expected: []interface{}{
				&wire.Response_VStr{VStr: "OK"},
				&wire.Response_VInt{VInt: 1},
				&wire.Response_VInt{VInt: 0},
			},
			delay:    []time.Duration{0, 0, 0},
		},
		{
		 	// TODO: expected response should be updated for all exists command once multi shard is impl
			name:     "Test EXISTS command with multiple keys",
			command:  []string{"SET key value", "SET key2 value2", "EXISTS key key2 key3", "EXISTS key key2 key3 key4", "DEL key", "EXISTS key key2 key3 key4"},
			expected: []interface{}{
				&wire.Response_VStr{VStr: "OK"},
				&wire.Response_VStr{VStr: "OK"},
				&wire.Response_VInt{VInt: 1},
				&wire.Response_VInt{VInt: 1},
				&wire.Response_VInt{VInt: 1},
				&wire.Response_VInt{VInt: 0},
			},
			delay:    []time.Duration{0, 0, 0, 0, 0, 0},
		},
		{
			name:     "Test EXISTS an expired key",
			command:  []string{"SET key value ex 1", "EXISTS key", "EXISTS key"},
			expected: []interface{}{
				&wire.Response_VStr{VStr: "OK"},
				&wire.Response_VInt{VInt: 1},
				&wire.Response_VInt{VInt: 0},
			},
			delay:    []time.Duration{0, 0, 2 * time.Second},
		},
		{
			// TODO: expected response should be updated for all exists command once multi shard is impl
			name:     "Test EXISTS with multiple keys and expired key",
			command:  []string{"SET key value ex 2", "SET key2 value2", "SET key3 value3", "EXISTS key key2 key3", "EXISTS key key2 key3"},
			expected: []interface{}{
				&wire.Response_VStr{VStr: "OK"},
				&wire.Response_VStr{VStr: "OK"},
				&wire.Response_VStr{VStr: "OK"},
				&wire.Response_VInt{VInt: 1},
				&wire.Response_VInt{VInt: 0},
			},
			delay:    []time.Duration{0, 0, 0, 0, 2 * time.Second},
		},
	}
	for _, tcase := range testCases {
		t.Run(tcase.name, func(t *testing.T) {
			// deleteTestKeys([]string{"key", "key2", "key3", "key4"}, store)
			client.FireString("DEL key")
			client.FireString("DEL key2")
			client.FireString("DEL key3")
			client.FireString("DEL key4")

			for i := 0; i < len(tcase.command); i++ {
				if tcase.delay[i] > 0 {
					time.Sleep(tcase.delay[i])
				}
				cmd := tcase.command[i]
				result := client.FireString(cmd)
				assert.Equal(t, tcase.expected[i], result.GetValue(), "Value mismatch for cmd %s", cmd)
			}
		})
	}
}
