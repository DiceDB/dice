package async

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandCount(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	t.Run("Command count should be positive", func(t *testing.T) {
		commandCount := getCommandCount(conn)
		assert.True(t, commandCount > 0,
			fmt.Sprintf("Unexpected number of CLI commands found. expected greater than 0, %d found", commandCount))
	})
}

func getCommandCount(connection net.Conn) int64 {
	responseValue := FireCommand(connection, "COMMAND COUNT")
	if responseValue == nil {
		return -1
	}
	return responseValue.(int64)
}

func BenchmarkCountCommand(b *testing.B) {
	conn := getLocalConnection()
	defer conn.Close()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		commandCount := getCommandCount(conn)
		if commandCount <= 0 {
			b.Fail()
		}
	}
}
