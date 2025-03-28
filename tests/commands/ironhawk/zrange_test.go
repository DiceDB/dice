package ironhawk

import (
	"errors"
	"testing"
)

func TestZrange(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "Test ZRANGE command with bad params",
			commands: []string{"ZRANGE", "ZRANGE key", "ZRANGE key 1"},
			expected: []interface{}{errors.New("wrong number of arguments for 'ZRANGE' command"), errors.New("wrong number of arguments for 'ZRANGE' command"), errors.New("wrong number of arguments for 'ZRANGE' command")},
		},
		{
			name:     "Test ZRANGE command non numeric start and stop",
			commands: []string{"ZRANGE key a b", "ZRANGE key 1 b"},
			expected: []interface{}{errors.New("value is not an integer or a float"), errors.New("value is not an integer or a float")},
		},
		{
			name:     "Test ZRANGE command non existent key",
			commands: []string{"ZRANGE key 1 2"},
			expected: []interface{}{nil},
		},
		{
			name:     "Test ZRANGE on non sorted set key",
			commands: []string{"SET key value", "ZRANGE key 1 2"},
			expected: []interface{}{"OK", errors.New("wrongtype operation against a key holding the wrong kind of value")},
		},
	}

	runTestcases(t, client, testCases)
}
