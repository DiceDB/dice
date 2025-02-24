// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var getDocsTestCases = []struct {
	name              string
	inCmd             string
	expected          interface{}
	skipExpectedMatch bool
}{
	{"Without any commands", "", []any{}, true},
	{"Set command", "SET", []interface{}{[]interface{}{
		"set",
		[]interface{}{
			"summary",
			string("SET puts a new <key, value> pair in db as in the args\n\t\targs must contain key and value.\n\t\targs can also contain multiple options -\n\t\tEX or ex which will set the expiry time(in secs) for the key\n\t\tReturns encoded error response if at least a <key, value> pair is not part of args\n\t\tReturns encoded error response if expiry tme value in not integer\n\t\tReturns encoded OK RESP once new entry is added\n\t\tIf the key already exists then the value will be overwritten and expiry will be discarded"),
			"arity", int64(-3),
			"beginIndex", int64(1),
			"lastIndex", int64(0),
			"step", int64(0),
		},
	}}, false},
	{"Get command", "GET", []interface{}{[]interface{}{
		"get",
		[]interface{}{
			"summary",
			string("GET returns the value for the queried key in args\n\t\tThe key should be the only param in args\n\t\tThe RESP value of the key is encoded and then returned\n\t\tGET returns RespNIL if key is expired or it does not exist"),
			"arity", int64(2),
			"beginIndex", int64(1),
			"lastIndex", int64(0),
			"step", int64(0),
		},
	}}, false},
	{"Ping command", "PING", []interface{}{[]interface{}{
		"ping",
		[]interface{}{
			"summary",
			string(`PING returns with an encoded "PONG" If any message is added with the ping command,the message will be returned.`),
			"arity", int64(-1),
			"beginIndex", int64(0),
			"lastIndex", int64(0),
			"step", int64(0),
		},
	}}, false},
	{"Invalid command", "INVALID_CMD",
		[]any{},
		false,
	},
	{"Combination of valid and Invalid command", "SET INVALID_CMD", []interface{}{[]interface{}{
		"set",
		[]interface{}{
			"summary",
			string("SET puts a new <key, value> pair in db as in the args\n\t\targs must contain key and value.\n\t\targs can also contain multiple options -\n\t\tEX or ex which will set the expiry time(in secs) for the key\n\t\tReturns encoded error response if at least a <key, value> pair is not part of args\n\t\tReturns encoded error response if expiry tme value in not integer\n\t\tReturns encoded OK RESP once new entry is added\n\t\tIf the key already exists then the value will be overwritten and expiry will be discarded"),
			"arity", int64(-3),
			"beginIndex", int64(1),
			"lastIndex", int64(0),
			"step", int64(0),
		}}}, false},
	{"Combination of multiple valid commands", "SET GET", []interface{}{[]interface{}{
		"set",
		[]interface{}{
			"summary",
			string("SET puts a new <key, value> pair in db as in the args\n\t\targs must contain key and value.\n\t\targs can also contain multiple options -\n\t\tEX or ex which will set the expiry time(in secs) for the key\n\t\tReturns encoded error response if at least a <key, value> pair is not part of args\n\t\tReturns encoded error response if expiry tme value in not integer\n\t\tReturns encoded OK RESP once new entry is added\n\t\tIf the key already exists then the value will be overwritten and expiry will be discarded"),
			"arity", int64(-3),
			"beginIndex", int64(1),
			"lastIndex", int64(0),
			"step", int64(0),
		}},
		[]interface{}{"get",
			[]interface{}{
				"summary",
				string("GET returns the value for the queried key in args\n\t\tThe key should be the only param in args\n\t\tThe RESP value of the key is encoded and then returned\n\t\tGET returns RespNIL if key is expired or it does not exist"),
				"arity", int64(2),
				"beginIndex", int64(1),
				"lastIndex", int64(0),
				"step", int64(0),
			},
		}}, false},
}

func TestCommandDocs(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	for _, tc := range getDocsTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result := client.FireString("COMMAND DOCS " + tc.inCmd)
			if !tc.skipExpectedMatch {
				assert.Equal(t, tc.expected, result)
			} else {
				assert.NotNil(t, result)
				ok := result.GetVStr() != ""
				assert.True(t, ok)
				assert.True(t, len(result.GetVStr()) > 0)
			}
		})
	}
}

func BenchmarkCommandDocs(b *testing.B) {
	client := getLocalConnection()
	defer client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range getDocsTestCases {
			client.FireString("COMMAND DOCS " + tc.inCmd)
		}
	}
}
