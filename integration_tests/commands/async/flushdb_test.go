package async

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFLUSHDB(t *testing.T) {
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
			name:     "FLUSHDB",
			setup:    []string{"MSET k1 v1 k2 v2 k3 v3"},
			commands: []string{"FLUSHDB", "DBSIZE"},
			expected: []interface{}{"OK", int64(0)},
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
