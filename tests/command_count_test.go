package tests

import (
	"fmt"
	"testing"
	"gotest.tools/v3/assert"
)

func TestCommandCount(t *testing.T) {
	connection := getLocalConnection()
	responseValue := fireCommand(connection, "COMMAND COUNT")
	responseCount  := responseValue.(int64)
	if responseValue == nil {
		t.Fail()
	}

	assert.Assert(t, responseCount > 0,
		fmt.Sprintf("Unexpected number of CLI commands found. expected greater than 0, %d found", responseCount))
}

func getCommandCount(t *testing.B) {
	connection := getLocalConnection()
	responseValue := fireCommand(connection, "COMMAND COUNT")
	responseCount  := responseValue.(int64)
	if responseValue == nil || responseCount <= 0 {
		t.Fail()
	}
}

func BenchmarkCountCommand(b *testing.B) {
	for n := 0; n < b.N; n++ {
		getCommandCount(b)
	}
}