package resp

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLInsert(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name   string
		cmds   []string
		expect []any
	}{
		{
			name:   "LINSERT before",
			cmds:   []string{"LPUSH k v1 v2 v3 v4", "LINSERT k before v2 e1", "LINSERT k before v1 e2", "LINSERT k before v4 e3", "LRANGE k 0 6"},
			expect: []any{int64(4), int64(5), int64(6), int64(7), []any{"e3", "v4", "v3", "e1", "v2", "e2", "v1"}},
		},
		{
			name:   "LINSERT after",
			cmds:   []string{"LINSERT k after v2 e4", "LINSERT k after v1 e5", "LINSERT k after v4 e6", "LRANGE k 0 10"},
			expect: []any{int64(8), int64(9), int64(10), []any{"e3", "v4", "e6", "v3", "e1", "v2", "e4", "e2", "v1", "e5"}},
		},
		{
			name:   "LINSERT wrong number of args",
			cmds:   []string{"LINSERT k before e1"},
			expect: []any{"-wrong number of arguments for LINSERT"},
		},
		{
			name:   "LINSERT wrong type",
			cmds:   []string{"SET k1 val1", "LINSERT k1 before val1 val2"},
			expect: []any{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result := FireCommand(conn, cmd)
				// assert.DeepEqual(t, tc.expect[i], result)
				assert.EqualValues(t, tc.expect[i], result)
			}
		})
	}

	deqCleanUp(conn, "k")
}

func TestLRange(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name   string
		cmds   []string
		expect []any
	}{
		{
			name:   "LRANGE with +ve start stop",
			cmds:   []string{"LPUSH k v1 v2 v3 v4", "LINSERT k before v2 e1", "LINSERT k before v1 e2", "LINSERT k before v4 e3", "LRANGE k 0 6"},
			expect: []any{int64(4), int64(5), int64(6), int64(7), []any{"e3", "v4", "v3", "e1", "v2", "e2", "v1"}},
		},
		{
			name:   "LRANGE with -ve start stop",
			cmds:   []string{"LRANGE k -100 -2"},
			expect: []any{[]any{"e3", "v4", "v3", "e1", "v2", "e2"}},
		},
		{
			name:   "LRANGE wrong number of args",
			cmds:   []string{"LRANGE k -100"},
			expect: []any{"-wrong number of arguments for LRANGE"},
		},
		{
			name:   "LRANGE wrong type",
			cmds:   []string{"SET k1 val1", "LRANGE k1 0 100"},
			expect: []any{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result := FireCommand(conn, cmd)
				// assert.DeepEqual(t, tc.expect[i], result)
				assert.EqualValues(t, tc.expect[i], result)
			}
		})
	}

	deqCleanUp(conn, "k")
}

func deqCleanUp(conn net.Conn, key string) {
	for {
		result := FireCommand(conn, "LPOP "+key)
		if result == "(nil)" {
			break
		}
	}
}
