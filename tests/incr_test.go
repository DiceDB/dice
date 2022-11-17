package tests

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
)

type tcase struct {
	op  string
	key string
	val int64
}

func TestINCR(t *testing.T) {
	var wg sync.WaitGroup
	go runTestServer(&wg)
	conn := getLocalConnection()

	for _, tc := range []tcase{
		{"s", "key1", 0},
		{"i", "key1", 1},
		{"i", "key1", 2},
		{"i", "key2", 1},
		{"g", "key1", 2},
		{"g", "key2", 1},
	} {

		switch tc.op[0] {
		case 's':
			cmd := fmt.Sprintf("SET %s %d", tc.key, tc.val)
			fireCommand(conn, cmd)
		case 'i':
			cmd := fmt.Sprintf("INCR %s", tc.key)
			fireCommand(conn, cmd)
		case 'g':
			cmd := fmt.Sprintf("GET %s", tc.key)
			r := fireCommand(conn, cmd)
			assertResult(t, r, strconv.FormatInt(tc.val, 10))

		}
	}

	fireCommand(conn, "ABORT")
	wg.Wait()
}

func assertResult(t *testing.T, r interface{}, expected string) {
	if r != expected {
		t.Fail()
	}
}
