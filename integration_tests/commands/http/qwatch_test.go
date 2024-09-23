package http

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var qWatchQuery = "SELECT $key, $value WHERE $key LIKE \"match:100:*\" AND $value > 10 ORDER BY $value DESC LIMIT 3"

func TestQWatch(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "QWATCH Register Bad Request",
			commands: []HTTPCommand{
				{Command: "QWATCH", Body: map[string]interface{}{}},
			},
			expected: []interface{}{
				[]interface{}{},
			},
			errorExpected: true,
		},
		{
			name: "QWATCH Register",
			commands: []HTTPCommand{
				{Command: "QWATCH", Body: map[string]interface{}{"query": qWatchQuery}},
			},
			expected: []interface{}{
				[]interface{}{
					"qwatch",
					"SELECT $key, $value WHERE $key like 'match:100:*' and $value > 10 ORDER BY $value desc LIMIT 3",
					// Empty array, as the initial result will be empty
					[]interface{}{},
				},
			},
			errorExpected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"key": "match:100:user"},
			})

			for i, cmd := range tc.commands {
				result, err := exec.FireCommand(cmd)
				if tc.errorExpected {
					assert.NotNil(t, err)
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}
}
