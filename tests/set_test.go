package tests

import (
	"fmt"
	"net"
	"testing"

	"gotest.tools/v3/assert"
)

func deleteKeys(conn net.Conn, keysToDelete []string) {
	for _, key := range keysToDelete {
		fireCommand(conn, fmt.Sprintf("DEL %s", key))
	}
}

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

func TestSetWithXX(t *testing.T) {
	conn := getLocalConnection()
	deleteKeys(conn, []string{"k"})
	for _, tcase := range []DTestCase{
		{
			InCmds: []string{"SET k v XX"},
			Out:    []interface{}{"nil"},
		},
		{
			InCmds: []string{"SET k v1", "SET k v2 XX", "GET k"},
			Out:    []interface{}{"OK", "OK", "v2"},
		},
		{
			InCmds: []string{"SET k v1", "SET k v2 XX", "SET k v3 XX", "GET k"},
			Out:    []interface{}{"OK", "OK", "OK", "v3"},
		},
		{
			InCmds: []string{"SET k v1", "SET k v2 XX", "DEL k", "GET k", "SET k v XX"},
			Out:    []interface{}{"OK", "OK", "1", "nil", "nil"},
		},
	} {
		for i := 0; i < len(tcase.InCmds); i++ {
			cmd := tcase.InCmds[i]
			out := tcase.Out[i]
			assert.Equal(t, out, fireCommand(conn, cmd), "Value mismatch for cmd %s\n.", cmd)
		}
	}
}
