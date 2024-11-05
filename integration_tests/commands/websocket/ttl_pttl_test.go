package websocket

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestTTLPTTL(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	defer conn.Close()

	testCases := []struct {
		name       string
		commands   []string
		expected   []interface{}
		assertType []string
		delay      []time.Duration
	}{
		{
			name:       "TTL Simple Value",
			commands:   []string{"SET foo bar", "GETEX foo ex 5", "GETEX foo", "TTL foo"},
			expected:   []interface{}{"OK", "bar", "bar", float64(5)},
			assertType: []string{"equal", "equal", "equal", "assert"},
			delay:      []time.Duration{0, 0, 0, 0},
		},
		{
			name:       "PTTL Simple Value",
			commands:   []string{"SET foo bar", "GETEX foo px 5000", "GETEX foo", "PTTL foo"},
			expected:   []interface{}{"OK", "bar", "bar", float64(5000)},
			assertType: []string{"equal", "equal", "equal", "assert"},
			delay:      []time.Duration{0, 0, 0, 0},
		},
		{
			name:       "TTL & PTTL Non-Existent Key",
			commands:   []string{"TTL foo", "PTTL foo"},
			expected:   []interface{}{float64(-2), float64(-2)},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
		{
			name:       "TTL & PTTL without Expiry",
			commands:   []string{"SET foo bar", "GET foo", "TTL foo", "PTTL foo"},
			expected:   []interface{}{"OK", "bar", float64(-1), float64(-1)},
			assertType: []string{"equal", "equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0},
		},
		{
			name:       "TTL & PTTL with Persist",
			commands:   []string{"SET foo bar", "GETEX foo persist", "TTL foo", "PTTL foo"},
			expected:   []interface{}{"OK", "bar", float64(-1), float64(-1)},
			assertType: []string{"equal", "equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0},
		},
		{
			name:       "TTL & PTTL with Expire and Expired Key",
			commands:   []string{"SET foo bar", "GETEX foo ex 5", "GET foo", "TTL foo", "PTTL foo", "TTL foo", "PTTL foo"},
			expected:   []interface{}{"OK", "bar", "bar", float64(5), float64(5000), float64(-2), float64(-2)},
			assertType: []string{"equal", "equal", "equal", "assert", "assert", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0, 0, 5 * time.Second, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			DeleteKey(t, conn, exec, "foo")
			for i, cmd := range tc.commands {
				if tc.delay[i] > 0 {
					time.Sleep(tc.delay[i])
				}
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				if err != nil {
					t.Fatalf("Error executing command: %v", err)
				}
				if tc.assertType[i] == "equal" {
					assert.DeepEqual(t, tc.expected[i], result)
				} else if tc.assertType[i] == "assert" {
					assert.Assert(t, result.(float64) <= tc.expected[i].(float64), "Expected %v to be less than or equal to %v", result, tc.expected[i])
				}
			}
		})
	}
}
