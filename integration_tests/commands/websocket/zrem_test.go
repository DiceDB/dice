package websocket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestZREM(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	DeleteKey(t, conn, exec, "string_key")
	DeleteKey(t, conn, exec, "myzset1")
	DeleteKey(t, conn, exec, "myzset2")
	DeleteKey(t, conn, exec, "myzset3")
	DeleteKey(t, conn, exec, "myzset4")

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
		delays []time.Duration
	}{
		{
			name: "ZREM with wrong number of arguments",
			cmds: []string{"ZREM", "ZREM myzset"},
			expect: []interface{}{"ERR wrong number of arguments for 'zrem' command",
				"ERR wrong number of arguments for 'zrem' command"},
			delays: []time.Duration{0, 0},
		},
		{
			name: "ZREM with wrong type of key",
			cmds: []string{"SET string_key string_value", "ZREM string_key string_value"},
			expect: []interface{}{"OK",
				"WRONGTYPE Operation against a key holding the wrong kind of value"},
			delays: []time.Duration{0, 0},
		},
		{
			name:   "ZREM with non-existent key",
			cmds:   []string{"ZADD myzset1 1 one", "ZREM wrong_myzset1 one"},
			expect: []interface{}{float64(1), float64(0)},
			delays: []time.Duration{0, 0},
		},
		{
			name:   "ZREM with non-existent element",
			cmds:   []string{"ZADD myzset2 1 one", "ZREM myzset2 two"},
			expect: []interface{}{float64(1), float64(0)},
			delays: []time.Duration{0, 0},
		},
		{
			name:   "ZREM with sorted set holding single element",
			cmds:   []string{"ZADD myzset3 1 one", "ZREM myzset3 one"},
			expect: []interface{}{float64(1), float64(1)},
			delays: []time.Duration{0, 0},
		},
		{
			name:   "ZREM with sorted set holding multiple elements",
			cmds:   []string{"ZADD myzset4 1 one 2 two 3 three 4 four", "ZREM myzset4 four five", "ZREM myzset4 one two"},
			expect: []interface{}{float64(4), float64(1), float64(2)},
			delays: []time.Duration{0, 0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			for i, cmd := range tc.cmds {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}
