package tests

import (
	"fmt"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func TestCommandLIST(t *testing.T) {
	commands := []string{
		"ABORT", "BFADD", "BFEXISTS", "BFINFO", "BFINIT", "BGREWRITEAOF", "CLIENT",
		"DEL", "DISCARD", "EXEC", "EXPIRE", "GET", "INCR", "INFO", "LATENCY", "COMMAND_LIST", "LRU",
		"MULTI", "PING", "QINTINS", "QINTLEN", "QINTPEEK", "QINTREM", "QREFINS", "QREFLEN",
		"QREFPEEK", "QREFREM", "QWATCH", "SET", "SLEEP", "STACKINTLEN", "STACKINTPEEK",
		"STACKINTPOP", "STACKINTPUSH", "STACKREFLEN", "STACKREFPEEK", "STACKREFPOP",
		"STACKREFPUSH", "SUBSCRIBE", "TTL", "HELLO",
	}

	conn := getLocalConnection()

	responseValue := fireCommand(conn, "COMMAND LIST")
	if responseValue == nil {
		t.Fail()
	}

	assert.Assert(t, len(strings.Split(responseValue.(string), ",")) == len(commands),
		fmt.Sprintf("Unexpected number of CLI commands found. %d expected, %d found. Update TestCommandLIST if new commands have been added.",
			len(commands), len(strings.Split(responseValue.(string), ","))))

	for _, expectedCmd := range commands {
		contains := strings.Contains(responseValue.(string), expectedCmd)
		assert.Assert(t, contains, fmt.Sprintf("Expected command not found: %s", expectedCmd))
	}
}

func BenchmarkListCommand200(t *testing.B) {
	conn := getLocalConnection()
	for i := 0; i < 200; i++ {
		rp := fireCommandAndGetRESPParser(conn, "COMMAND LIST")
		if rp == nil {
			t.Fail()
		}
	}
}

func BenchmarkListCommand2000(t *testing.B) {
	conn := getLocalConnection()
	for i := 0; i < 2000; i++ {
		rp := fireCommandAndGetRESPParser(conn, "COMMAND LIST")
		if rp == nil {
			t.Fail()
		}
	}
}
