package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZRANK(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	// Clean up before and after tests
	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "myset"}})
	defer exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "myset"}})

	// Initialize the sorted set with members and their scores
	exec.FireCommand(HTTPCommand{
		Command: "ZADD",
		Body: map[string]interface{}{
			"key": "myset",
			"key_values": map[string]interface{}{
				"1": "member1",
				"2": "member2",
				"3": "member3",
				"4": "member4",
				"5": "member5",
			},
		},
	})

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
	}{
		{
			name: "ZRANK of existing member",
			commands: []HTTPCommand{
				{
					Command: "ZRANK",
					Body: map[string]interface{}{
						"key":   "myset",
						"value": "member1",
					},
				},
			},
			expected: []interface{}{float64(0)},
		},
		{
			name: "ZRANK of non-existing member",
			commands: []HTTPCommand{
				{
					Command: "ZRANK",
					Body: map[string]interface{}{
						"key":   "myset",
						"value": "member6",
					},
				},
			},
			expected: []interface{}{nil},
		},
		{
			name: "ZRANK with WITHSCORE option for existing member",
			commands: []HTTPCommand{
				{
					Command: "ZRANK",
					Body: map[string]interface{}{
						"key":       "myset",
						"value":     "member3",
						"withscore": true,
					},
				},
			},
			expected: []interface{}{[]interface{}{float64(2), "3"}},
		},
		{
			name: "ZRANK with WITHSCORE option for non-existing member",
			commands: []HTTPCommand{
				{
					Command: "ZRANK",
					Body: map[string]interface{}{
						"key":       "myset",
						"value":     "member6",
						"withscore": true,
					},
				},
			},
			expected: []interface{}{nil},
		},
		{
			name: "ZRANK on non-existing key",
			commands: []HTTPCommand{
				{
					Command: "ZRANK",
					Body: map[string]interface{}{
						"key":   "nonexistingset",
						"value": "member1",
					},
				},
			},
			expected: []interface{}{nil},
		},
		{
			name: "ZRANK with wrong number of arguments",
			commands: []HTTPCommand{
				{
					Command: "ZRANK",
					Body: map[string]interface{}{
						"key": "myset",
					},
				},
			},
			expected: []interface{}{"ERR wrong number of arguments for 'zrank' command"},
		},
		{
			name: "ZRANK with invalid option",
			commands: []HTTPCommand{
				{
					Command: "ZRANK",
					Body: map[string]interface{}{
						"key": "myset",
						"values": []interface{}{
							"member1",
							"invalidoption",
						},
					},
				},
			},
			expected: []interface{}{"ERR syntax error"},
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}
