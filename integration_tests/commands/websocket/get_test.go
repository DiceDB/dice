package websocket

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestGet(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
		delays []time.Duration
	}{
		{
			name:   "Get with expiration",
			cmds:   []string{"SET k v EX 4", "GET k", "GET k"},
			expect: []interface{}{"OK", "v", "(nil)"},
			delays: []time.Duration{0, 0, 5 * time.Second},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result := exec.FireCommand(conn, cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}
