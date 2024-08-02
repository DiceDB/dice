package tests

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestKeys(t *testing.T) {
	conn := getLocalConnection()
	for _, tcase := range []DTestCase{
		{
			InCmds: []string{"SET k v", "KEYS k"},
			Out:    []interface{}{"OK", []interface{}{"k"}},
		},
		{
			InCmds: []string{"SET good v", "SET great v", "KEYS g*"},
			Out:    []interface{}{"OK", "OK", []interface{}{"good", "great"}},
		},
		{
			InCmds: []string{"SET good v", "SET great v", "KEYS g?od", "KEYS g?eat"},
			Out:    []interface{}{"OK", "OK", []interface{}{"good"}, []interface{}{"great"}},
		},
		{
			InCmds: []string{"SET hallo v", "SET hbllo v", "SET hello v", "KEYS h[^e]llo"},
			Out:    []interface{}{"OK", "OK", "OK", []interface{}{"hallo", "hbllo"}},
		},
	} {
		for i := 0; i < len(tcase.InCmds); i++ {
			cmd := tcase.InCmds[i]
			out := tcase.Out[i]
			assert.DeepEqual(t, out, fireCommand(conn, cmd))
		}
	}
}
