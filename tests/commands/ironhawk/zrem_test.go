package ironhawk

import (
	"errors"
	"testing"
)

func TestZREM(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "Call ZREM with bad arguments",
			commands: []string{"ZREM", "ZREM key"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'ZREM' command"), errors.New("wrong number of arguments for 'ZREM' command"),
			},
		},
		{
			name:     "Call ZREM with non-existing key",
			commands: []string{"ZREM nonExistingKey member1"},
			expected: []interface{}{
				int64(0),
			},
		},
		{
			name: "Call ZREM on a key which is not a sorted set",
			commands: []string{
				"SET key value",
				"ZREM key member1",
			},
			expected: []interface{}{
				"OK",
				errors.New("wrongtype operation against a key holding the wrong kind of value"),
			},
		},
		{
			name: "Call ZREM with existing key and members",
			commands: []string{
				"ZADD key 1 member1",
				"ZADD key 2 member2",
				"ZADD key 3 member3",
				"ZREM key member1 member2",
			},
			expected: []interface{}{
				int64(1), // member1 added
				int64(1), // member2 added
				int64(1), // member3 added
				int64(2), // member1,member2 removed
			},
		},
	}
	runTestcases(t, client, testCases)
}
