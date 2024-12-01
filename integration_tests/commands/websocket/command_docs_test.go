package websocket

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandDocs(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
	}{
		{
			name: "Set command",
			cmds: []string{"COMMAND DOCS SET"},
			expect: []interface{}{[]interface{}{[]interface{}{
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
			cmds: []string{"COMMAND DOCS GET"},
			expect: []interface{}{[]interface{}{[]interface{}{
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
			cmds: []string{"COMMAND DOCS PING"},
			expect: []interface{}{[]interface{}{[]interface{}{
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
			name:   "Invalid command",
			cmds:   []string{"COMMAND DOCS INVALID_CMD"},
			expect: []any{[]any{}},
		},
		{
			name: "Combination of valid and Invalid command",
			cmds: []string{"COMMAND DOCS SET INVALID_CMD"},
			expect: []interface{}{[]interface{}{[]interface{}{
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
			cmds: []string{"COMMAND DOCS SET GET"},
			expect: []interface{}{[]interface{}{[]interface{}{
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
			for i, cmd := range tc.cmds {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}
