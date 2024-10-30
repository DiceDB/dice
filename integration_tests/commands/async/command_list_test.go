package async

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandList(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	t.Run("Command list should not be empty", func(t *testing.T) {
		commandList := getCommandList(conn)
		assert.True(t, len(commandList) > 0,
			fmt.Sprintf("Unexpected number of CLI commands found. expected greater than 0, %d found", len(commandList)))
	})
}

func getCommandList(connection net.Conn) []string {
	responseValue := FireCommand(connection, "COMMAND LIST")
	if responseValue == nil {
		return nil
	}

	var cmds []string
	for _, v := range responseValue.([]interface{}) {
		cmds = append(cmds, v.(string))
	}
	return cmds
}

func BenchmarkCommandList(b *testing.B) {
	conn := getLocalConnection()
	defer conn.Close()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		commandList := getCommandList(conn)
		if len(commandList) <= 0 {
			b.Fail()
		}
	}
}
