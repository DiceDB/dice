package tests

import (
	"strconv"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestExpireat(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name     string
		command  []string
		expected []interface{}
		delay    []time.Duration
	}{
		{
			name:     "Test EXPIREAT",
			command:  []string{"SET key value", "EXISTS key", "EXPIREAT key " + strconv.FormatInt(time.Now().Add(1*time.Second).Unix(), 10), "EXISTS key"},
			expected: []interface{}{"OK", int64(1), int64(1), int64(0)},
			delay:    []time.Duration{0, 0, 0, 2 * time.Second},
		},
	}
	for _, tcase := range testCases {
		t.Run(tcase.name, func(t *testing.T) {
			for i := 0; i < len(tcase.command); i++ {
				if tcase.delay[i] > 0 {
					time.Sleep(tcase.delay[i])
				}
				cmd := tcase.command[i]
				out := tcase.expected[i]
				assert.Equal(t, out, fireCommand(conn, cmd), "Value mismatch for cmd %s\n.", cmd)
			}
		})
	}
}
