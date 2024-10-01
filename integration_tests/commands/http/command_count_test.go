package http

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

func TestCommandCount(t *testing.T) {
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
