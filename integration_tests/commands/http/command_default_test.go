package http

import (
	"fmt"
	"testing"

	"github.com/dicedb/dice/internal/eval"
	"github.com/stretchr/testify/assert"
)

func TestCommandDefault(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	commands := getCommandDefault(exec)
	t.Run("Command should not be empty", func(t *testing.T) {
		assert.True(t, len(commands) > 0,
			fmt.Sprintf("Unexpected number of CLI commands found. expected greater than 0, %d found", len(commands)))
	})

	t.Run("Command count matches", func(t *testing.T) {
		assert.True(t, len(commands) == len(eval.DiceCmds),
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
	cmds = append(cmds, responseValue.([]interface{})...)
	return cmds
}

func BenchmarkCommandDefault(b *testing.B) {
	exec := NewHTTPCommandExecutor()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		commands := getCommandDefault(exec)
		if len(commands) <= 0 {
			b.Fail()
		}
	}
}
