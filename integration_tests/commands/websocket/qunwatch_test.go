package websocket

import (
	"testing"

	testifyAssert "github.com/stretchr/testify/assert"
)

func TestQUnwatch(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()

	testCases := []struct {
		name   string
		cmds   []string
		expect interface{}
	}{
		{
			name:   "Wrong number of arguments",
			cmds:   []string{"QUNWATCH "},
			expect: "ERR wrong number of arguments for 'qunwatch' command",
		},
		{
			name:   "Invalid query",
			cmds:   []string{"QUNWATCH \"SELECT \""},
			expect: "error parsing SQL statement: syntax error at position 8",
		},
		{
			name:   "Successful unregister",
			cmds:   []string{`QUNWATCH "SELECT $key, $value WHERE $key like 'k?'"`},
			expect: "OK",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, cmd := range tc.cmds {
				result := exec.FireCommand(conn, cmd)
				if _, ok := tc.expect.(string); ok {
					// compare strings
					testifyAssert.Equal(t, tc.expect, result, "Value mismatch for cmd %s", cmd)
				}
			}
		})
	}
}
