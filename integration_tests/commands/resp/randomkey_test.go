package resp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRandomKey(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
		delays []time.Duration
	}{
		{
			name: "Random Key",
			cmds: []string{
				"FLUSHDB",
				"SET k1 v1",
				"RANDOMKEY",
			},
			expect: []interface{}{"OK", "OK", "k1"},
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