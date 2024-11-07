package http

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExistsHttp(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name          string
		commands      []HTTPCommand
		expected      []interface{}
		delay         []time.Duration
		errorExpected bool
	}{
		{
			name: "Test EXISTS command",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "key", "value": "value"}},
				{Command: "EXISTS", Body: map[string]interface{}{"key": "key"}},
				{Command: "EXISTS", Body: map[string]interface{}{"key": "key2"}},
			},
			expected:      []interface{}{"OK", float64(1), float64(0)},
			delay:         []time.Duration{0, 0, 0},
			errorExpected: false,
		},
		{
			name: "Test EXISTS command with multiple keys",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "key", "value": "value"}},
				{Command: "SET", Body: map[string]interface{}{"key": "key2", "value": "value2"}},
				{Command: "EXISTS", Body: map[string]interface{}{"keys": []interface{}{"key", "key2", "key3"}}},
				{Command: "EXISTS", Body: map[string]interface{}{"keys": []interface{}{"key", "key2", "key3", "key4"}}},
				{Command: "DEL", Body: map[string]interface{}{"key": "key"}},
				{Command: "EXISTS", Body: map[string]interface{}{"keys": []interface{}{"key", "key2", "key3", "key4"}}},
			},
			expected: []interface{}{"OK", "OK", float64(2), float64(2), float64(1), float64(1)},
			delay:    []time.Duration{0, 0, 0, 0, 0, 0},
		},
		{
			name: "Test EXISTS an expired key",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "key", "value": "value", "ex": 2}},
				{Command: "EXISTS", Body: map[string]interface{}{"key": "key"}},
				{Command: "EXISTS", Body: map[string]interface{}{"key": "key"}},
			},
			expected:      []interface{}{"OK", float64(1), float64(0)},
			delay:         []time.Duration{0, 0, 2 * time.Second},
			errorExpected: false,
		},
		{
			name: "Test EXISTS with multiple keys and expired key",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "value", "value": "value", "ex": 2}},
				{Command: "SET", Body: map[string]interface{}{"key": "key2", "value": "value2"}},
				{Command: "SET", Body: map[string]interface{}{"key": "key3", "value": "value3"}},
				{Command: "EXISTS", Body: map[string]interface{}{"keys": []interface{}{"key", "key2", "key3"}}},
				{Command: "EXISTS", Body: map[string]interface{}{"keys": []interface{}{"key", "key2", "key3"}}},
			},
			expected:      []interface{}{"OK", "OK", "OK", float64(2), float64(2)},
			delay:         []time.Duration{0, 0, 0, 0, 2 * time.Second},
			errorExpected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset keys before each test
			for _, key := range []string{"k", "key", "key2", "key3", "key4"} {
				exec.FireCommand(HTTPCommand{
					Command: "DEL",
					Body:    map[string]interface{}{"key": key},
				})
			}

			for i, cmd := range tc.commands {
				if tc.delay[i] > 0 {
					time.Sleep(tc.delay[i])
				}
				result, err := exec.FireCommand(cmd)
				if err != nil {
					log.Printf("Error executing command: %v", err)
				}
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}
