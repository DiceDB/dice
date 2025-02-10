// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var getKeysTestCases = []struct {
	name     string
	inCmd    string
	expected interface{}
}{
	{"Set command", "set 1 2 3 4", []interface{}{"1"}},
	{"Get command", "get key", []interface{}{"key"}},
	{"TTL command", "ttl key", []interface{}{"key"}},
	{"Del command", "del 1 2 3 4 5 6", []interface{}{"1", "2", "3", "4", "5", "6"}},
	// TODO: Fix this for multi shard support
	//{"MSET command", "MSET key1 val1 key2 val2", []interface{}{"key1", "key2"}},
	{"Expire command", "expire key time extra", []interface{}{"key"}},
	{"Ping command", "ping", "ERR the command has no key arguments"},
	{"Invalid Get command", "get", "ERR invalid number of arguments specified for command"},
	{"Abort command", "abort", "ERR the command has no key arguments"},
	{"Invalid command", "NotValidCommand", "ERR invalid command specified"},
	{"Wrong number of arguments", "", "ERR wrong number of arguments for 'command|getkeys' command"},
}

func TestCommandGetKeys(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	for _, tc := range getKeysTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result := client.FireString("COMMAND GETKEYS " + tc.inCmd)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func BenchmarkGetKeysMatch(b *testing.B) {
	client := getLocalConnection()
	defer client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range getKeysTestCases {
			client.FireString("COMMAND GETKEYS " + tc.inCmd)
		}
	}
}
