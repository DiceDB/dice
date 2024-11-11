package async

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetSet(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
		delays []time.Duration
	}{
		{
			name:   "GETSET with INCR",
			cmds:   []string{"INCR mycounter", "GETSET mycounter \"0\"", "GET mycounter"},
			expect: []interface{}{int64(1), int64(1), int64(0)},
			delays: []time.Duration{0, 0, 0},
		},
		{
			name:   "GETSET with SET",
			cmds:   []string{"SET mykey \"Hello\"", "GETSET mykey \"world\"", "GET mykey"},
			expect: []interface{}{"OK", "Hello", "world"},
			delays: []time.Duration{0, 0, 0},
		},
		{
			name:   "GETSET with TTL",
			cmds:   []string{"SET k v EX 60", "GETSET k v1", "TTL k"},
			expect: []interface{}{"OK", "v", int64(-1)},
			delays: []time.Duration{0, 0, 0},
		},
		{
			name:   "GETSET error when key exists but does not hold a string value",
			cmds:   []string{"LPUSH k1 \"somevalue\"", "GETSET k1 \"v1\""},
			expect: []interface{}{int64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			delays: []time.Duration{0, 0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}
