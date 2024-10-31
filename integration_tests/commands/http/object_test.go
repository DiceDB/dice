package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOBJECT(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name        string
		commands    []HTTPCommand
		expected    []interface{}
		assert_type []string
		delay       []time.Duration
	}{
		{
			name: "Object Idletime",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "OBJECT", Body: map[string]interface{}{"values": []interface{}{"IDLETIME", "foo"}}},
				{Command: "OBJECT", Body: map[string]interface{}{"values": []interface{}{"IDLETIME", "foo"}}},
				{Command: "TOUCH", Body: map[string]interface{}{"keys": []interface{}{"foo"}}},
				{Command: "OBJECT", Body: map[string]interface{}{"values": []interface{}{"IDLETIME", "foo"}}},
			},
			expected:    []interface{}{"OK", float64(2), float64(3), float64(1), float64(0)},
			assert_type: []string{"equal", "assert", "assert", "equal", "assert"},
			delay:       []time.Duration{0, 2 * time.Second, 3 * time.Second, 0, 0},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Ensure key is deleted before the test
			exec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"key": "foo"},
			})

			for i, cmd := range tc.commands {
				if tc.delay[i] > 0 {
					time.Sleep(tc.delay[i])
				}
				result, _ := exec.FireCommand(cmd)
				if tc.assert_type[i] == "equal" {
					assert.Equal(t, tc.expected[i], result)
				} else if tc.assert_type[i] == "assert" {
					assert.True(t, result.(float64) >= tc.expected[i].(float64), "Expected %v to be less than or equal to %v", result, tc.expected[i])
				}
			}
		})
	}
}
