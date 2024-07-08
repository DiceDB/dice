package tests

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestSet(t *testing.T) {
	conn := getLocalConnection()
	for _, tcase := range []DTestCase{
		{
			InCmds: []string{"SET k v", "GET k"},
			Out:    []interface{}{"OK", "v"},
		},
	} {
		for i := 0; i < len(tcase.InCmds); i++ {
			cmd := tcase.InCmds[i]
			out := tcase.Out[i]
			assert.Equal(t, out, fireCommand(conn, cmd), "Value mismatch for cmd %s\n.", cmd)
		}
	}
}
