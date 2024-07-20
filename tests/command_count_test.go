package tests

import (
	"fmt"
	"net"
	"testing"
	"gotest.tools/v3/assert"
)

func TestCommandCount(t *testing.T) {
	connection := getLocalConnection()
	commandCount := getCommandCount(connection)
	if commandCount <= 0 {
		t.Fail()
	}
	assert.Assert(t, commandCount > 0,
		fmt.Sprintf("Unexpected number of CLI commands found. expected greater than 0, %d found", commandCount))
}

func getCommandCount(connection net.Conn) int64 {
	responseValue := fireCommand(connection, "COMMAND COUNT")
	if responseValue == nil {
		return -1
	}
	return responseValue.(int64)
}

func BenchmarkCountCommand(b *testing.B) {
	connection := getLocalConnection()
	for n := 0; n < b.N; n++ {
		commandCount := getCommandCount(connection)
		if commandCount <= 0 {
			b.Fail()
		}
	}
}