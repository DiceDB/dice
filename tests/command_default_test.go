package tests

import (
	"fmt"
	"net"
	"testing"

	"gotest.tools/v3/assert"
)

func TestCommandDefault(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	t.Run("Command should not be empty", func(t *testing.T) {
		commands := getCommandDefault(conn)
		assert.Assert(t, len(commands) > 0,
			fmt.Sprintf("Unexpected number of CLI commands found. expected greater than 0, %d found", len(commands)))
	})
}

func getCommandDefault(connection net.Conn) []interface{} {
	responseValue := fireCommand(connection, "COMMAND")
	if responseValue == nil {
		return nil
	}
	var cmds []interface{}
	for _, v := range responseValue.([]interface{}) {
		cmds = append(cmds, v)
	}
	return cmds
}

func BenchmarkCommandDefault(b *testing.B) {
	conn := getLocalConnection()
	defer conn.Close()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		commands := getCommandDefault(conn)
		if len(commands) <= 0 {
			b.Fail()
		}
	}
}
