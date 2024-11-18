package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandCount(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name          string
		commands      []HTTPCommand
		expected      []interface{}
		errorExpected bool
		assertType    []string
	}{
		{
			name: "Command count should be greather than zero",
			commands: []HTTPCommand{
				{Command: "COMMAND/COUNT"},
			},
			expected:   []interface{}{float64(0)},
			assertType: []string{"greater"},
		},
		{
			name: "Command count should not support any argument",
			commands: []HTTPCommand{
				{Command: "COMMAND/COUNT", Body: map[string]interface{}{"key": ""}},
			},
			expected:   []interface{}{"ERR wrong number of arguments for 'command|count' command"},
			assertType: []string{"equal"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for c, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				switch tc.assertType[c] {
				case "equal":
					assert.Equal(t, tc.expected[c], result)
				case "greater":
					assert.True(t, result.(float64) >= tc.expected[c].(float64))
				}
			}

		})
	}
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
