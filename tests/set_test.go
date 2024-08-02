package tests

import (
	"testing"
	"time"

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

func TestSetWithKeepTTLFlag(t *testing.T) {
	conn := getLocalConnection()
	for _, tcase := range []DTestCase{
		{
			InCmds: []string{"SET k v EX 6"},
			Out:    []interface{}{"OK"},
		},
		{
			InCmds: []string{"SET k vv KEEPTTL", "GET k"},
			Out:    []interface{}{"OK", "vv"},
		},
		{
			InCmds: []string{"GET k"},
			Out:    []interface{}{"nil"},
		},
	} {
		for i := 0; i < len(tcase.InCmds); i++ {
			cmd := tcase.InCmds[i]
			out := tcase.Out[i]
			// fmt.Println(fireCommand(conn, cmd))
			assert.Equal(t, out, fireCommand(conn, cmd), "Value mismatch for cmd %s\n.", cmd)
			time.Sleep(3 * time.Second)
		}
	}
}
