package tests

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

type DDelayTestCase struct {
	InCmds []string
	Out    []interface{}
	Delay  []time.Duration
}

func TestGet(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	for _, tcase := range []DDelayTestCase{
		{
			InCmds: []string{"SET k v EX 4", "GET k", "GET k"},
			Out:    []interface{}{"OK", "v", "(nil)"},
			Delay:  []time.Duration{0, 0, 5 * time.Second},
		},
	} {
		for i := 0; i < len(tcase.InCmds); i++ {
			if tcase.Delay[i] > 0 {
				time.Sleep(tcase.Delay[i])
			}
			cmd := tcase.InCmds[i]
			out := tcase.Out[i]
			assert.Equal(t, out, fireCommand(conn, cmd), "Value mismatch for cmd %s\n.", cmd)
		}
	}
}
