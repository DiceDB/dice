package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRandomKey(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		delays   []time.Duration
	}{

		{
			name: "Random Key",
			commands: []HTTPCommand{
				{Command: "FLUSHDB"},
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "RANDOMKEY"},
			},
			expected: []interface{}{"OK", "OK", "k1"},
			delays:   []time.Duration{0, 0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result, err := exec.FireCommand(cmd)
				assert.Nil(t, err)
				assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}
