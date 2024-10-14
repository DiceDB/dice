package http

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

func TestCommandCount(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "Command count should not support any argument",
			commands: []HTTPCommand{
				{Command: "COMMAND/COUNT", Body: map[string]interface{}{"key": ""}},
			},
			expected: []interface{}{"ERR wrong number of arguments for 'command|count' command"},
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.DeepEqual(t, tc.expected[i], result)

			}

		})
	}
}

func TestCommandCount1(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	t.Run("Command count should be positive", func(t *testing.T) {
		commandCount := getCommandCount(exec)
		assert.Assert(t, commandCount > 0,
			fmt.Sprintf("Unexpected number of CLI commands found. expected greater than 0, %f found", commandCount))
	})
}

func getCommandCount(exec *HTTPCommandExecutor) float64 {
	cmd := HTTPCommand{Command: "COMMAND/COUNT", Body: map[string]interface{}{"key": ""}}
	responseValue, _ := exec.FireCommand(cmd)
	if responseValue == nil {
		return -1
	}
	return responseValue.(float64)
}

func BenchmarkCountCommand(b *testing.B) {
	exec := NewHTTPCommandExecutor()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		commandCount := getCommandCount(exec)
		if commandCount <= 0 {
			b.Fail()
		}
	}
}
