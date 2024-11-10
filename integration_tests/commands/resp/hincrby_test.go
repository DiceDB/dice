package resp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHINCRBY(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	FireCommand(conn, "FLUSHDB")

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
		delays []time.Duration
	}{
		{
			name:   "HINCRBY on non-existing key",
			cmds:   []string{"HINCRBY key field1 10"},
			expect: []interface{}{int64(10)},
			delays: []time.Duration{0},
		},
		{
			name:   "HINCRBY on existing key",
			cmds:   []string{"HINCRBY key field1 5"},
			expect: []interface{}{int64(15)},
			delays: []time.Duration{0},
		},
		{
			name:   "HINCRBY on non-integer value",
			cmds:   []string{"HSET keys field value", "HINCRBY keys field 1"},
			expect: []interface{}{int64(1), "ERR hash value is not an integer"},
			delays: []time.Duration{0, 0, 0},
		},
		{
			name:   "HINCRBY on non-hashmap key",
			cmds:   []string{"SET key value", "HINCRBY key value 10"},
			expect: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			delays: []time.Duration{0, 0, 0},
		},
		{
			name:   "HINCRBY overflow",
			cmds:   []string{"HSET new-key value 9000000000000000000", "HINCRBY new-key value 1000000000000000000"},
			expect: []interface{}{int64(1), "ERR increment or decrement would overflow"},
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
