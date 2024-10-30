package websocket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHLEN(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	DeleteKey(t, conn, exec, "string_key")
	DeleteKey(t, conn, exec, "key_hLen1")
	DeleteKey(t, conn, exec, "key_hLen2")
	DeleteKey(t, conn, exec, "key_hLen3")

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
		delays []time.Duration
	}{

		{
			name:   "HLEN with wrong number of args",
			cmds:   []string{"HLEN"},
			expect: []interface{}{"ERR wrong number of arguments for 'hlen' command"},
			delays: []time.Duration{0},
		},
		{
			name:   "HLEN with non-existent key",
			cmds:   []string{"HLEN non_existent_key"},
			expect: []interface{}{float64(0)},
			delays: []time.Duration{0},
		},
		{
			name:   "HLEN with non-hash",
			cmds:   []string{"SET string_key string_value", "HLEN string_key"},
			expect: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			delays: []time.Duration{0, 0},
		},
		{
			name:   "HLEN with empty hash",
			cmds:   []string{"HSET key_hLen1 field value", "HDEL key_hLen1 field", "HLEN key_hLen1"},
			expect: []interface{}{float64(1), float64(1), float64(0)},
			delays: []time.Duration{0, 0, 0},
		},
		{
			name:   "HLEN with single field",
			cmds:   []string{"HSET key_hLen2 field1 value1", "HLEN key_hLen2"},
			expect: []interface{}{float64(1), float64(1)},
			delays: []time.Duration{0, 0},
		},
		{
			name:   "HLEN with multiple fields",
			cmds:   []string{"HSET key_hLen3 field1 value1 field2 value2 field3 value3", "HLEN key_hLen3"},
			expect: []interface{}{float64(3), float64(3)},
			delays: []time.Duration{0, 0},
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
