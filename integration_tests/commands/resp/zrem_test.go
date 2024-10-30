package resp

import (
	"testing"
	"time"

	testifyAssert "github.com/stretchr/testify/assert"
)

func TestZREM(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

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
			cmds:   []string{"ZADD myzset 1 one", "ZREM wrong_myzset one"},
			expect: []interface{}{int64(1), int64(0)},
			delays: []time.Duration{0, 0},
		},
		{
			name:   "ZREM with non-existent element",
			cmds:   []string{"ZADD myzset 1 one", "ZREM myzset two"},
			expect: []interface{}{int64(1), int64(0)},
			delays: []time.Duration{0, 0},
		},
		{
			name:   "ZREM with sorted set holding single element",
			cmds:   []string{"ZADD myzset 1 one", "ZREM myzset one"},
			expect: []interface{}{int64(1), int64(1)},
			delays: []time.Duration{0, 0},
		},
		{
			name:   "ZREM with sorted set holding multiple elements",
			cmds:   []string{"ZADD myzset 1 one 2 two 3 three 4 four", "ZREM myzset four five", "ZREM myzset one two"},
			expect: []interface{}{int64(4), int64(1), int64(2)},
			delays: []time.Duration{0, 0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			FireCommand(conn, "DEL myzset string_key")
			for i, cmd := range tc.cmds {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result := FireCommand(conn, cmd)
				testifyAssert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}
