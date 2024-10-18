package resp

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestGetDel(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
		delays []time.Duration
	}{
		{
			name:   "GetDel",
			cmds:   []string{"SET k v", "GETDEL k", "GETDEL k", "GET k"},
			expect: []interface{}{"OK", "v", "(nil)", "(nil)"},
			delays: []time.Duration{0, 0, 0, 0},
		},
		{
			name:   "GetDel with expiration, checking if key exist and is already expired, then it should return null",
			cmds:   []string{"GETDEL k", "SET k v EX 2", "GETDEL k"},
			expect: []interface{}{"(nil)", "OK", "(nil)"},
			delays: []time.Duration{0, 0, 3 * time.Second},
		},
		{
			name: "GetDel with expiration, checking if key exist and is not yet expired, then it should return its " +
				"value",
			cmds:   []string{"SET k v EX 40", "GETDEL k"},
			expect: []interface{}{"OK", "v"},
			delays: []time.Duration{0, 2 * time.Second},
		},
		{
			name: "GetDel with invalid command",
			cmds: []string{"GETDEL", "GETDEL k v"},
			expect: []interface{}{"ERR wrong number of arguments for 'getdel' command",
				"ERR wrong number of arguments for 'getdel' command"},
			delays: []time.Duration{0, 0},
		},
		{
			name:   "Getdel with value created from Setbit",
			cmds:   []string{"SETBIT k 1 1", "GET k", "GETDEL k", "GET k"},
			expect: []interface{}{int64(0), "@", "@", "(nil)"},
			delays: []time.Duration{0, 0, 0, 0},
		},
		{
			name:   "GetDel with Set object should return wrong type error",
			cmds:   []string{"SADD myset member1", "GETDEL myset"},
			expect: []interface{}{int64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			delays: []time.Duration{0, 0},
		},
		{
			name:   "GetDel with JSON object should return wrong type error",
			cmds:   []string{"JSON.SET k $ 1", "GETDEL k", "JSON.GET k"},
			expect: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value", "1"},
			delays: []time.Duration{0, 0, 0},
		},
	}

	for _, tc := range testCases {
		FireCommand(conn, "del k")
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s, expected this %s, "+
					"got this %s", cmd, tc.expect[i], result)
			}
		})
	}
}
