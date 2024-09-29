package async

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestObjectCommand(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name        string
		commands    []string
		expected   []interface{}
		assertType []string
		delay      []time.Duration
	}{
		{
			name:       "Object Idletime",
			commands:   []string{"SET foo bar", "OBJECT IDLETIME foo", "OBJECT IDLETIME foo", "TOUCH foo", "OBJECT IDLETIME foo"},
			expected:   []interface{}{"OK", int64(2), int64(3), int64(1), int64(0)},
			assertType: []string{"equal", "assert", "assert", "equal", "assert"},
			delay:      []time.Duration{0, 2 * time.Second, 3 * time.Second, 0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// deleteTestKeys([]string{"foo"}, store)
			FireCommand(conn, "DEL foo")

			for i, cmd := range tc.commands {
				if tc.delay[i] != 0 {
					time.Sleep(tc.delay[i])
				}
				result := FireCommand(conn, cmd)
				if tc.assertType[i] == "equal" {
					assert.DeepEqual(t, tc.expected[i], result)
				} else {
					assert.Assert(t, result.(int64) >= tc.expected[i].(int64), "Expected %v to be less than or equal to %v", result, tc.expected[i])
				}
			}
		})
	}
}
