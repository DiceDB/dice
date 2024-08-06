package tests

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestExists(t *testing.T) {
	conn := getLocalConnection()
	testCases := []struct {
		command  []string
		expected []interface{}
		delay    []time.Duration
	}{
		{
			command:  []string{"SET key value", "EXISTS key", "EXISTS key2", "EXISTS key key2", "SET key2 value2", "EXISTS key key2", "EXISTS key key2 key3","SET key3 value ex 4", "EXISTS key3", "EXISTS key3 key2", "EXISTS key3 key2 key", "DEL key", "EXISTS key key2"},
			expected: []interface{}{"OK", int64(1), int64(0), int64(1), "OK", int64(2), int64(2),"OK", int64(1), int64(1), int64(2), int64(1), int64(1)},
			delay:    []time.Duration{0, 0, 0, 0, 0, 0, 0, 0, 0, 5 * time.Second, 0, 0, 0},
		},
	}
	for _, tcase := range testCases {
		for i := 0; i < len(tcase.command); i++ {
			cmd := tcase.command[i]
			out := tcase.expected[i]
			if tcase.delay[i] > 0 {
				time.Sleep(tcase.delay[i])
			}
			assert.Equal(t, out, fireCommand(conn, cmd), "Value mismatch for cmd %s\n.", cmd)
		}
	}
}
