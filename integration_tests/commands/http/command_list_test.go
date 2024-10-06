package http

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

func TestCommandList(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "Command list should not be empty",
			commands: []HTTPCommand{
				{Command: "COMMAND/LIST", Body: map[string]interface{}{"key": ""}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				var commandList []string
				for _, v := range result.([]interface{}) {
					commandList = append(commandList, v.(string))
				}

				assert.Assert(t, len(commandList) > 0,
					fmt.Sprintf("Unexpected number of CLI commands found. expected greater than 0, %d found", len(commandList)))
			}

		})
	}
}
