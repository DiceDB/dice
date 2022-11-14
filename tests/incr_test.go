package tests

import (
	"fmt"
	"sync"
	"testing"
)

type tcase struct {
	op  string
	val int64
}

func TestINCR(t *testing.T) {
	var wg sync.WaitGroup
	go runTestServer(&wg)
	conn := getLocalConnection()

	for _, tc := range []tcase{
		{"s1", 0},
		{"i1", 1},
		{"i1", 1},
	} {

		switch tc.op[0] {
		case 's':
			cmd := fmt.Sprintf("SET Key %d", tc.val)
			fireCommand(conn, cmd)
		case 'i':
			cmd := fmt.Sprintf("INCR Key")
			fireCommand(conn, cmd)
		}
	}

	cmd := fmt.Sprintf("GET Key")
	r := fireCommand(conn, cmd)
	if r != "2" {
		t.Fail()
	}

	fireCommand(conn, "ABORT")
	wg.Wait()
}
