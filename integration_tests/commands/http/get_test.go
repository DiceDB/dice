package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		delays   []time.Duration
	}{
		{
			name: "Get with expiration",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v", "ex": 4}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
				{Command: "SLEEP", Body: map[string]interface{}{"key": 5}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "v", "OK", nil},
			delays:   []time.Duration{0, 0, 5 * time.Second, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}
