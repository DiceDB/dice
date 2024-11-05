package async

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiscard(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
	}{
		{
			name:   "Discard commands in a txn",
			cmds:   []string{"MULTI", "SET key1 value1", "DISCARD"},
			expect: []interface{}{"OK", "QUEUED", "OK"},
		},
		{
			name:   "Throw error if Discard used outside a txn",
			cmds:   []string{"DISCARD"},
			expect: []interface{}{"ERR DISCARD without MULTI"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			//deleteTestKeys([]string{"key1"}, store)
			FireCommand(conn, "DEL key1")
			for i, cmd := range tc.cmds {
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}
