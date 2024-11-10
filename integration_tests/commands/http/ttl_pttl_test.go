package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTTLPTTL(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name        string
		commands    []HTTPCommand
		expected    []interface{}
		assert_type []string
		delay       []time.Duration
	}{
		// TTL Simple Value
		{
			name: "TTL Simple Value",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo", "ex": 5}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:    []interface{}{"OK", "bar", "bar", float64(5)},
			assert_type: []string{"equal", "equal", "equal", "assert"},
			delay:       []time.Duration{0, 0, 0, 0},
		},
		// PTTL Simple Value
		{
			name: "PTTL Simple Value",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo", "px": 5000}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo"}},
				{Command: "PTTL", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:    []interface{}{"OK", "bar", "bar", float64(5000)},
			assert_type: []string{"equal", "equal", "equal", "assert"},
			delay:       []time.Duration{0, 0, 0, 0},
		},
		// TTL & PTTL Non-Existent Key
		{
			name: "TTL & PTTL Non-Existent Key",
			commands: []HTTPCommand{
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "PTTL", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:    []interface{}{float64(-2), float64(-2)},
			assert_type: []string{"equal", "equal"},
			delay:       []time.Duration{0, 0},
		},
		// TTL & PTTL without Expiry
		{
			name: "TTL & PTTL without Expiry",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "GET", Body: map[string]interface{}{"key": "foo"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "PTTL", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:    []interface{}{"OK", "bar", float64(-1), float64(-1)},
			assert_type: []string{"equal", "equal", "equal", "equal"},
			delay:       []time.Duration{0, 0, 0, 0},
		},
		// TTL & PTTL with Persist
		{
			name: "TTL & PTTL with Persist",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo", "value": "persist"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "PTTL", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:    []interface{}{"OK", "bar", float64(-1), float64(-1)},
			assert_type: []string{"equal", "equal", "equal", "equal"},
			delay:       []time.Duration{0, 0, 0, 0},
		},
		// TTL & PTTL with Expire and Expired Key
		{
			name: "TTL & PTTL with Expire and Expired Key",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo", "ex": 5}},
				{Command: "GET", Body: map[string]interface{}{"key": "foo"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "PTTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "PTTL", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:    []interface{}{"OK", "bar", "bar", float64(5), float64(5000), float64(-2), float64(-2)},
			assert_type: []string{"equal", "equal", "equal", "assert", "assert", "equal", "equal"},
			delay:       []time.Duration{0, 0, 0, 0, 0, 5 * time.Second, 0},
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
					assert.True(t, result.(float64) <= tc.expected[i].(float64), "Expected %v to be less than or equal to %v", result, tc.expected[i])
				}
			}
		})
	}

}
