package tests

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

func TestCommandCOUNT(t *testing.T) {
	commandsCount := 40

	subscriber := getLocalConnection()

	responseValue := fireCommand(subscriber, "COUNT")
	responseCount  := responseValue.(int64)
	if responseValue == nil {
		t.Fail()
	}

	assert.Assert(t, responseCount == int64(commandsCount),
		fmt.Sprintf("Unexpected number of CLI commands found. %d expected, %d found", commandsCount, responseCount))
}

func call(howmany int, t *testing.B) {
	subscriber := getLocalConnection()
	for i := 0; i < howmany; i++ {
		rp := fireCommandAndGetRESPParser(subscriber, "COUNT")
		if rp == nil {
			t.Fail()
		}
	}
}

func BenchmarkCountCommand200(t *testing.B) {
	call(200, t)
}

func BenchmarkCountCommand2000(t *testing.B) {
	call(2000, t)
}