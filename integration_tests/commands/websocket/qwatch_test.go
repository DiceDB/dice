package websocket

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
		// TODO - once following query is registered, websocket will also attempt sending updates
		// while keys are set for other tests in this package
		// Add unregister test case to handle this scenario once qunwatch support is added
		{
			name:   "Successful register",
			cmds:   []string{`QWATCH "SELECT $key, $value WHERE $key like 'qwatch-test-key?'"`},
			expect: []interface{}{"qwatch", "SELECT $key, $value WHERE $key like 'qwatch-test-key?'", []interface{}{}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, cmd := range tc.cmds {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				if _, ok := tc.expect.(string); ok {
					// compare strings
					assert.Equal(t, tc.expect, result, "Value mismatch for cmd %s", cmd)
				} else {
					// compare lists
					assert.ElementsMatch(t, tc.expect, result, "Value mismatch for cmd %s", cmd)
				}
			}
		})
	}
}
