package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// This file may contain test cases for checking error messages across all commands
func TestErrorsForSetData(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	setErrorMsg := "WRONGTYPE Operation against a key holding the wrong kind of value"
	testCases := []struct {
		name       string
		commands   []HTTPCommand
		expected   []interface{}
		delays     []time.Duration
		assertType []string
	}{
		{
			name: "GET a key holding a set",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "member": "bar"}},
				{Command: "GET", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{float64(1), setErrorMsg},
			assertType: []string{"equal", "equal"},
			delays:     []time.Duration{0, 0},
		},
		{
			name: "GETDEL a key holding a set",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "member": "bar"}},
				{Command: "GETDEL", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{float64(1), setErrorMsg},
			assertType: []string{"equal", "equal"},
			delays:     []time.Duration{0, 0},
		},
		{
			name: "INCR a key holding a set",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "member": "bar"}},
				{Command: "INCR", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{float64(1), setErrorMsg},
			assertType: []string{"equal", "equal"},
			delays:     []time.Duration{0, 0},
		},
		{
			name: "DECR a key holding a set",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "member": "bar"}},
				{Command: "DECR", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{float64(1), setErrorMsg},
			assertType: []string{"equal", "equal"},
			delays:     []time.Duration{0, 0},
		},
		{
			name: "BIT operations on a key holding a set",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "member": "bar"}},
				{Command: "GETBIT", Body: map[string]interface{}{"key": "foo", "offset": 1}},
				{Command: "BITCOUNT", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{float64(1), setErrorMsg, setErrorMsg},
			assertType: []string{"equal", "equal", "equal"},
			delays:     []time.Duration{0, 0, 0},
		},
		{
			name: "GETEX a key holding a set",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "member": "bar"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{float64(1), setErrorMsg},
			assertType: []string{"equal", "equal"},
			delays:     []time.Duration{0, 0},
		},
		{
			name: "GETSET a key holding a set",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "member": "bar"}},
				{Command: "GETSET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
			},
			expected:   []interface{}{float64(1), setErrorMsg},
			assertType: []string{"equal", "equal"},
			delays:     []time.Duration{0, 0},
		},
		{
			name: "LPUSH, LPOP, RPUSH, RPOP a key holding a set",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "member": "bar"}},
				{Command: "LPUSH", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "foo"}},
				{Command: "RPUSH", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{float64(1), setErrorMsg, setErrorMsg, setErrorMsg, setErrorMsg},
			assertType: []string{"equal", "equal", "equal", "equal", "equal"},
			delays:     []time.Duration{0, 0, 0, 0, 0},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "foo"}})
			for i, cmd := range tc.commands {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result, _ := exec.FireCommand(cmd)
				switch tc.assertType[i] {
				case "equal":
					assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s", cmd)
				}
			}
		})
	}
}
