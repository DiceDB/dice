package http

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestCommandDocs(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "Set command",
			commands: []HTTPCommand{
				{Command: "COMMAND/DOCS", Body: map[string]interface{}{"key": "SET"}},
			},
			expected: []interface{}{[]interface{}{[]interface{}{
				string("set"),
				[]interface{}{
					string("summary"),
					string("SET puts a new <key, value> pair in db as in the args\n\t\targs must contain key and value.\n\t\targs can also contain multiple options -\n\t\tEX or ex which will set the expiry time(in secs) for the key\n\t\tReturns encoded error response if at least a <key, value> pair is not part of args\n\t\tReturns encoded error response if expiry tme value in not integer\n\t\tReturns encoded OK RESP once new entry is added\n\t\tIf the key already exists then the value will be overwritten and expiry will be discarded"),
					string("arity"), float64(-3),
					string("beginIndex"), float64(1),
					string("lastIndex"), float64(0),
					string("step"), float64(0),
				},
			},
			}},
		},
		{
			name: "Get command",
			commands: []HTTPCommand{
				{Command: "COMMAND/DOCS", Body: map[string]interface{}{"key": "GET"}},
			},
			expected: []interface{}{[]interface{}{[]interface{}{
				"get",
				[]interface{}{
					string("summary"),
					string("GET returns the value for the queried key in args\n\t\tThe key should be the only param in args\n\t\tThe RESP value of the key is encoded and then returned\n\t\tGET returns RespNIL if key is expired or it does not exist"),
					string("arity"), float64(2),
					string("beginIndex"), float64(1),
					string("lastIndex"), float64(0),
					string("step"), float64(0),
				},
			}}},
		},
		{
			name: "PING command",
			commands: []HTTPCommand{
				{Command: "COMMAND/DOCS", Body: map[string]interface{}{"key": "PING"}},
			},
			expected: []interface{}{[]interface{}{[]interface{}{
				"ping",
				[]interface{}{
					string("summary"),
					string(`PING returns with an encoded "PONG" If any message is added with the ping command,the message will be returned.`),
					string("arity"), float64(-1),
					string("beginIndex"), float64(0),
					string("lastIndex"), float64(0),
					string("step"), float64(0),
				},
			}}},
		},
		{
			name: "Invalid command",
			commands: []HTTPCommand{
				{Command: "COMMAND/DOCS", Body: map[string]interface{}{"key": "INVALID_CMD"}},
			},
			expected: []any{[]any{}},
		},
		{
			name: "Combination of valid and Invalid command",
			commands: []HTTPCommand{
				{Command: "COMMAND/DOCS", Body: map[string]interface{}{"keys": []interface{}{"SET", "INVALID_CMD"}}},
			},
			expected: []interface{}{[]interface{}{[]interface{}{
				"set",
				[]interface{}{
					string("summary"),
					string("SET puts a new <key, value> pair in db as in the args\n\t\targs must contain key and value.\n\t\targs can also contain multiple options -\n\t\tEX or ex which will set the expiry time(in secs) for the key\n\t\tReturns encoded error response if at least a <key, value> pair is not part of args\n\t\tReturns encoded error response if expiry tme value in not integer\n\t\tReturns encoded OK RESP once new entry is added\n\t\tIf the key already exists then the value will be overwritten and expiry will be discarded"),
					string("arity"), float64(-3),
					string("beginIndex"), float64(1),
					string("lastIndex"), float64(0),
					string("step"), float64(0),
				}},
			}},
		},
		{
			name: "Combination of multiple valid commands",
			commands: []HTTPCommand{
				{Command: "COMMAND/DOCS", Body: map[string]interface{}{"keys": []interface{}{"SET", "GET"}}},
			},
			expected: []interface{}{[]interface{}{[]interface{}{
				"set",
				[]interface{}{
					string("summary"),
					string("SET puts a new <key, value> pair in db as in the args\n\t\targs must contain key and value.\n\t\targs can also contain multiple options -\n\t\tEX or ex which will set the expiry time(in secs) for the key\n\t\tReturns encoded error response if at least a <key, value> pair is not part of args\n\t\tReturns encoded error response if expiry tme value in not integer\n\t\tReturns encoded OK RESP once new entry is added\n\t\tIf the key already exists then the value will be overwritten and expiry will be discarded"),
					string("arity"), float64(-3),
					string("beginIndex"), float64(1),
					string("lastIndex"), float64(0),
					string("step"), float64(0),
				}},
				[]interface{}{"get",
					[]interface{}{
						string("summary"),
						string("GET returns the value for the queried key in args\n\t\tThe key should be the only param in args\n\t\tThe RESP value of the key is encoded and then returned\n\t\tGET returns RespNIL if key is expired or it does not exist"),
						string("arity"), float64(2),
						string("beginIndex"), float64(1),
						string("lastIndex"), float64(0),
						string("step"), float64(0),
					},
				}}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}
