package async

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTouch(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name       string
		commands   []string
		expected   []interface{}
		assertType []string
		delay      []time.Duration
	}{
		{
			name:       "Touch Simple Value",
			commands:   []string{"SET foo bar", "OBJECT IDLETIME foo", "TOUCH foo", "OBJECT IDLETIME foo"},
			expected:   []interface{}{"OK", int64(2), int64(1), int64(0)},
			assertType: []string{"equal", "assert", "equal", "assert"},
			delay:      []time.Duration{0, 2 * time.Second, 0, 0},
		},
		{
			name:       "Touch Multiple Existing Keys",
			commands:   []string{"SET foo bar", "SET foo1 bar", "TOUCH foo foo1"},
			expected:   []interface{}{"OK", "OK", int64(2)},
			assertType: []string{"equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0},
		},
		{
			name:       "Touch Multiple Existing and Non-Existing Keys",
			commands:   []string{"SET foo bar", "TOUCH foo foo1"},
			expected:   []interface{}{"OK", int64(1)},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// deleteTestKeys([]string{"foo", "foo1"}, store)
			FireCommand(conn, "DEL foo")
			FireCommand(conn, "DEL foo1")
			for i, cmd := range tc.commands {
				if tc.delay[i] != 0 {
					time.Sleep(tc.delay[i])
				}
				result := FireCommand(conn, cmd)
				if tc.assertType[i] == "equal" {
					assert.Equal(t, tc.expected[i], result)
				} else {
					assert.True(t, result.(int64) >= tc.expected[i].(int64), "Expected %v to be less than or equal to %v", result, tc.expected[i])
				}
			}
		})
	}
}
