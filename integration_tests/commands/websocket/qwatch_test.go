package websocket

import (
	"testing"

	testifyAssert "github.com/stretchr/testify/assert"
)

func TestQWatch(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()

	testCases := []struct {
		name   string
		cmds   []string
		expect interface{}
	}{
		{
			name:   "Wrong number of arguments",
			cmds:   []string{"QWATCH "},
			expect: "ERR wrong number of arguments for 'qwatch' command",
		},
		{
			name:   "Invalid query",
			cmds:   []string{"QWATCH \"SELECT \""},
			expect: "error parsing SQL statement: syntax error at position 8",
		},
		{
			name:   "Successful register",
			cmds:   []string{`QWATCH "SELECT $key, $value WHERE $key like 'k?'"`},
			expect: []interface{}{"qwatch", "SELECT $key, $value WHERE $key like 'k?'", []interface{}{}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, cmd := range tc.cmds {
				result := exec.FireCommand(conn, cmd)
				if _, ok := tc.expect.(string); ok {
					// compare strings
					testifyAssert.Equal(t, tc.expect, result, "Value mismatch for cmd %s", cmd)
				} else {
					// compare lists
					testifyAssert.ElementsMatch(t, tc.expect, result, "Value mismatch for cmd %s", cmd)
				}
			}
		})
	}
}
