package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTouch(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		delay    []time.Duration
	}{
		{
			name: "Touch Simple Value",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "OBJECT", Body: map[string]interface{}{"values": []interface{}{"IDLETIME", "foo"}}},
				{Command: "TOUCH", Body: map[string]interface{}{"key": "foo"}},
				{Command: "OBJECT", Body: map[string]interface{}{"values": []interface{}{"IDLETIME", "foo"}}},
			},
			expected: []interface{}{"OK", float64(2), float64(1), float64(0)},
			delay:    []time.Duration{0, 2 * time.Second, 0, 0},
		},
		// Touch Multiple Existing Keys
		{
			name: "Touch Multiple Existing Keys",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SET", Body: map[string]interface{}{"key": "foo1", "value": "bar"}},
				{Command: "TOUCH", Body: map[string]interface{}{"keys": []interface{}{"foo", "foo1"}}},
			},
			expected: []interface{}{"OK", "OK", float64(2)},
			delay:    []time.Duration{0, 0, 0},
		},
		// Touch Multiple Existing and Non-Existing Keys
		{
			name: "Touch Multiple Existing and Non-Existing Keys",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "TOUCH", Body: map[string]interface{}{"keys": []interface{}{"foo", "foo1"}}},
			},
			expected: []interface{}{"OK", float64(1)},
			delay:    []time.Duration{0, 0},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"key": "foo"},
			})
			exec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"key": "foo1"},
			})

			for i, cmd := range tc.commands {
				if tc.delay[i] > 0 {
					time.Sleep(tc.delay[i])
				}
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}

}
