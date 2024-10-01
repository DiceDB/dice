package http

import (
	"fmt"
	"testing"

	"github.com/dicedb/dice/internal/eval"
	"gotest.tools/v3/assert"
)

func TestCommandDefault(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	commands := getCommandDefault(exec)
	t.Run("Command should not be empty", func(t *testing.T) {
		assert.Assert(t, len(commands) > 0,
			fmt.Sprintf("Unexpected number of CLI commands found. expected greater than 0, %d found", len(commands)))
	})

	t.Run("Command count matches", func(t *testing.T) {
		assert.Assert(t, len(commands) == len(eval.DiceCmds),
			fmt.Sprintf("Unexpected number of CLI commands found. expected %d, %d found", len(eval.DiceCmds), len(commands)))
	})
}

func getCommandDefault(exec *HTTPCommandExecutor) []interface{} {
	cmd := HTTPCommand{Command: "COMMAND", Body: map[string]interface{}{"values": []interface{}{}}}
	responseValue, _ := exec.FireCommand(cmd)
	if responseValue == nil {
		return nil
	}
	var cmds []interface{}
	for _, v := range responseValue.([]interface{}) {
		cmds = append(cmds, v)
	}
	return cmds
}
