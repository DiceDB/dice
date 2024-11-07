package websocket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestZCARD(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	DeleteKey(t, conn, exec, "string_key")
	DeleteKey(t, conn, exec, "myzset1")
	DeleteKey(t, conn, exec, "myzset2")
	DeleteKey(t, conn, exec, "myzset3")

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
		delays []time.Duration
	}{
		{
			name: "ZCARD with wrong number of arguments",
			cmds: []string{"ZCARD", "ZCARD myzset field"},
			expect: []interface{}{"ERR wrong number of arguments for 'zcard' command",
				"ERR wrong number of arguments for 'zcard' command"},
			delays: []time.Duration{0, 0},
		},
		{
			name: "ZCARD with wrong type of key",
			cmds: []string{"SET string_key string_value", "ZCARD string_key"},
			expect: []interface{}{"OK",
				"WRONGTYPE Operation against a key holding the wrong kind of value"},
			delays: []time.Duration{0, 0},
		},
		{
			name:   "ZCARD with non-existent key",
			cmds:   []string{"ZADD myzset1 1 one", "ZCARD wrong_myzset1"},
			expect: []interface{}{float64(1), float64(0)},
			delays: []time.Duration{0, 0},
		},
		{
			name:   "ZCARD with sorted set holding single element",
			cmds:   []string{"ZADD myzset2 1 one", "ZCARD myzset2"},
			expect: []interface{}{float64(1), float64(1)},
			delays: []time.Duration{0, 0},
		},
		{
			name:   "ZCARD with sorted set holding multiple elements",
			cmds:   []string{"ZADD myzset3 1 one 2 two", "ZCARD myzset3", "ZADD myzset3 3 three", "ZCARD myzset3", "ZREM myzset3 two", "ZCARD myzset3"},
			expect: []interface{}{float64(2), float64(2), float64(1), float64(3), float64(1), float64(2)},
			delays: []time.Duration{0, 0, 0, 0, 0, 0},
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
