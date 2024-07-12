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
		"DEL", "DISCARD", "EXEC", "EXPIRE", "GET", "INCR", "INFO", "LATENCY", "LIST", "LRU",
		"MULTI", "PING", "QINTINS", "QINTLEN", "QINTPEEK", "QINTREM", "QREFINS", "QREFLEN",
		"QREFPEEK", "QREFREM", "QWATCH", "SET", "SLEEP", "STACKINTLEN", "STACKINTPEEK",
		"STACKINTPOP", "STACKINTPUSH", "STACKREFLEN", "STACKREFPEEK", "STACKREFPOP",
		"STACKREFPUSH", "SUBSCRIBE", "TTL",
	}

	subscriber := getLocalConnection()

	rp := fireCommandAndGetRESPParser(subscriber, "LIST")
	if rp == nil {
		t.Fail()
	}

	// Read first message (OK)
	v, err := rp.DecodeOne()
	assert.NilError(t, err)

	for _, expectedCmd := range commands {
		contains := strings.Contains(v.(string), expectedCmd)
		assert.Assert(t, contains, fmt.Sprintf("Expected command not found: %s", expectedCmd))
	}
}

func call(howmany int, t *testing.B) {
	subscriber := getLocalConnection()
	for i := 0; i < howmany; i++ {
		rp := fireCommandAndGetRESPParser(subscriber, "LIST")
		if rp == nil {
			t.Fail()
		}
	}
}

func BenchmarkListCommand20(t *testing.B) {
	call(20, t)
}

func BenchmarkListCommand200(t *testing.B) {
	call(200, t)
}

func BenchmarkListCommand2000(t *testing.B) {
	call(2000, t)
}
