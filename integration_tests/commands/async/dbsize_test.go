package async

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestDBSIZE(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name     string
		setup    []string
		commands []string
		expected []interface{}
		delay    []time.Duration
		cleanUp  []string
	}{
		{
			name:     "DBSIZE",
			setup:    []string{"FLUSHDB", "MSET k1 v1 k2 v2 k3 v3"},
			commands: []string{"DBSIZE"},
			expected: []interface{}{int64(3)},
			delay:    []time.Duration{0},
			cleanUp:  []string{"DEL k1 k2 k3"},
		},
		{
			name:     "DBSIZE with repeative keys in MSET/SET",
			setup:    []string{"MSET k1 v1 k2 v2 k3 v3 k1 v3", "SET k2 v22"},
			commands: []string{"DBSIZE"},
			expected: []interface{}{int64(3)},
			delay:    []time.Duration{0},
			cleanUp:  []string{"DEL k1 k2 k3"},
		},
		{
			name:     "DBSIZE with expired keys",
			setup:    []string{"MSET k1 v1 k2 v2 k3 v3", "SET k3 v3 ex 1"},
			commands: []string{"DBSIZE", "DBSIZE"},
			expected: []interface{}{int64(3), int64(2)},
			delay:    []time.Duration{0, 2 * time.Second},
			cleanUp:  []string{"DEL k1 k2 k3"},
		},
		{
			name:     "DBSIZE with deleted keys",
			setup:    []string{"MSET k1 v1 k2 v2 k3 v3"},
			commands: []string{"DBSIZE", "DEL k1 k2", "DBSIZE"},
			expected: []interface{}{int64(3), int64(2), int64(1)},
			delay:    []time.Duration{0, 0, 0},
			cleanUp:  []string{"DEL k1 k2 k3"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			for _, cmd := range tc.setup {
				result := FireCommand(conn, cmd)
				assert.Equal(t, "OK", result, "Setup Failed")
			}

			for i, cmd := range tc.commands {
				if tc.delay[i] > 0 {
					time.Sleep(tc.delay[i])
				}
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result)
			}

			for _, cmd := range tc.cleanUp {
				FireCommand(conn, cmd)
			}
		})
	}
}
