package tests

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestGet(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	for _, tcase := range []DTestCase{
		{
			InCmds: []string{"SET k v EX 4", "GET k", "SLEEP 5", "GET k"},
			Out:    []interface{}{"OK", "v", "OK", "(nil)"},
		},
	} {
		for i := 0; i < len(tcase.InCmds); i++ {
			cmd := tcase.InCmds[i]
			out := tcase.Out[i]
			assert.Equal(t, out, fireCommand(conn, cmd), "Value mismatch for cmd %s\n.", cmd)
		}
	}
}
