package http

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetEx(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	Etime5 := strconv.FormatInt(time.Now().Unix()+5, 10)
	Etime10 := strconv.FormatInt(time.Now().Unix()+10, 10)

	testCases := []struct {
		name       string
		commands   []HTTPCommand
		expected   []interface{}
		assertType []string
		delay      []time.Duration
	}{
		{
			name: "GetEx Simple Value",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{"OK", "bar", "bar", float64(-1)},
			assertType: []string{"equal", "equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0},
		},
		{
			name: "GetEx Non-Existent Key",
			commands: []HTTPCommand{
				{Command: "GETEX", Body: map[string]interface{}{"key": "nonexistent"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "nonexistent"}},
			},
			expected:   []interface{}{nil, float64(-2)},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
		{
			name: "GetEx with EX option",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo", "ex": 2}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{"OK", "bar", float64(-1), "bar", float64(2), nil, float64(-2)},
			assertType: []string{"equal", "equal", "equal", "equal", "assert", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0, 0, 2 * time.Second, 0},
		},
		{
			name: "GetEx with PX option",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo", "px": 2000}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{"OK", "bar", float64(-1), "bar", float64(2), nil, float64(-2)},
			assertType: []string{"equal", "equal", "equal", "equal", "assert", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0, 0, 2 * time.Second, 0},
		},

		{
			name: "GetEx with EX option and invalid value",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo", "ex": -1}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{"OK", "bar", float64(-1), "ERR invalid expire time in 'getex' command", float64(-1), "bar", float64(-1)},
			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0, 0, 0, 0},
		},
		{
			name: "GetEx with PX option and invalid value",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo", "px": -1}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{"OK", "bar", float64(-1), "ERR invalid expire time in 'getex' command", float64(-1), "bar", float64(-1)},
			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0, 0, 0, 0},
		},
		{
			name: "GetEx with EXAT option",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo", "exat": Etime5}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{"OK", "bar", float64(-1), "bar", float64(5), nil, float64(-2)},
			assertType: []string{"equal", "equal", "equal", "equal", "assert", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0, 0, 5 * time.Second, 0},
		},
		{
			name: "GetEx with PXAT option",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo", "pxat": Etime10 + "000"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{"OK", "bar", float64(-1), "bar", float64(10), nil, float64(-2)},
			assertType: []string{"equal", "equal", "equal", "equal", "assert", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0, 0, 10 * time.Second, 0},
		},
		{
			name: "GetEx with EXAT option and invalid value",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo", "exat": 123123}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{"OK", "bar", float64(-1), "bar", nil, float64(-2)},
			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0, 0, 0},
		},
		{
			name: "GetEx with PXAT option and invalid value",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo", "pxat": 123123}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{"OK", "bar", float64(-1), "bar", nil, float64(-2)},
			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0, 0, 0},
		},
		{
			name: "GetEx with PERSIST option",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar", "ex": 10}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo", "persist": true}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{"OK", float64(10), "bar", float64(-1)},
			assertType: []string{"equal", "assert", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0},
		},
		{
			name: "GetEx with multiple expiry options",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo", "ex": 2, "px": 123123}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{"OK", "ERR syntax error", float64(-1), "bar"},
			assertType: []string{"equal", "equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0},
		},
		{
			name: "GetEx with persist and ex options",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo", "ex": 2}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo", "persist": true, "ex": 2}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{"OK", "bar", float64(2), "ERR syntax error", float64(2)},
			assertType: []string{"equal", "equal", "assert", "equal", "assert"},
			delay:      []time.Duration{0, 0, 0, 0, 0},
		},
		{
			name: "GetEx with persist and px options",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo", "px": 2000}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "foo", "px": 2000, "persist": true}},
				{Command: "TTL", Body: map[string]interface{}{"key": "foo"}},
			},
			expected:   []interface{}{"OK", "bar", float64(2), "ERR syntax error", float64(2)},
			assertType: []string{"equal", "equal", "assert", "equal", "assert"},
			delay:      []time.Duration{0, 0, 0, 0, 0},
		},
		{
			name: "GetEx with key holding JSON type",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "KEY", "path": "$", "value": "1"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "KEY"}},
			},
			expected:   []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
		{
			name: "GetEx with key holding JSON type with multiple set commands",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "MJSONKEY", "path": "$", "value": "{\"a\":2}"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "MJSONKEY"}},
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "MJSONKEY", "path": "$.a", "value": "3"}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "MJSONKEY"}},
			},
			expected: []interface{}{
				"OK",
				"WRONGTYPE Operation against a key holding the wrong kind of value",
				"OK",
				"WRONGTYPE Operation against a key holding the wrong kind of value",
			},
			assertType: []string{"equal", "equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0},
		},
		{
			name: "GetEx with key holding SET type",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "SKEY", "members": []interface{}{1, 2, 3}}},
				{Command: "GETEX", Body: map[string]interface{}{"key": "SKEY"}},
			},
			expected:   []interface{}{float64(3), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "foo"}})

			for i, cmd := range tc.commands {
				if tc.delay[i] > 0 {
					time.Sleep(tc.delay[i])
				}
				result, err := exec.FireCommand(cmd)
				assert.Nil(t, err)
				if tc.assertType[i] == "equal" {
					assert.Equal(t, tc.expected[i], result)
				} else if tc.assertType[i] == "assert" {
					assert.True(t, result.(float64) <= tc.expected[i].(float64), "Expected %v to be less than or equal to %v", result, tc.expected[i])
				}
			}
		})
	}
}
