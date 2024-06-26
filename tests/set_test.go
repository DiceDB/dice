package tests

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"

	"gotest.tools/v3/assert"
)

func TestSet(t *testing.T) {
	var wg sync.WaitGroup
	runTestServer(&wg)

	var b []byte
	var buf = bytes.NewBuffer(b)
	conn := getLocalConnection()
	for i := 1; i < 100; i++ {
		buf.WriteByte('a')
		cmd := fmt.Sprintf("SET k%d %s", i, buf.String())
		fireCommand(conn, cmd)
	}

	for i := 1; i < 100; i++ {
		cmd := fmt.Sprintf("GET k%d", i)
		v := fireCommand(conn, cmd)
		expectedValue := strings.Repeat("a", i)
		assert.Equal(t, expectedValue, v.(string), "Value mismatch for key k%d", i)
	}

	fireCommand(conn, "ABORT")
	wg.Wait()
}
