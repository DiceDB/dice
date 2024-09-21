package commands

import (
	"fmt"
	"net"
	"testing"

	"gotest.tools/v3/assert"
)

func TestCommandDocs(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	t.Run("Command docs should not be empty", func(t *testing.T) {
		commandDocs := getCommandDocs(conn)
		assert.Assert(t, len(commandDocs) > 0,
			fmt.Sprintf("Unexpected number of CLI commands found. expected greater than 0, %d found", len(commandDocs)))
	})
}

func getCommandDocs(connection net.Conn) []interface{} {
	responseValue := FireCommand(connection, "COMMAND DOCS")
	if responseValue == nil {
		return nil
	}

	var cmds []interface{}
	for _, v := range responseValue.([]interface{}) {
		cmds = append(cmds, v.(interface{}))
	}
	return cmds
}

func BenchmarkCommandDocs(b *testing.B) {
	conn := getLocalConnection()
	defer conn.Close()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		commandDocs := getCommandDocs(conn)
		if len(commandDocs) <= 0 {
			b.Fail()
		}
	}
}
