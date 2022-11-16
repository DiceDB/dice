package tests

import (
	"fmt"
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
		{"i", "key1", 1},
		{"i", "key2", 1},
	} {

		switch tc.op[0] {
		case 's':
			cmd := fmt.Sprintf("SET %s %d", tc.key, tc.val)
			fireCommand(conn, cmd)
		case 'i':
			cmd := fmt.Sprintf("INCR %s", tc.key)
			fireCommand(conn, cmd)
		}
	}

	cmd := fmt.Sprintf("GET key1")
	r := fireCommand(conn, cmd)
	expected := "2"
	assertResult(t, r, expected)

	cmd = fmt.Sprintf("GET key2")
	r = fireCommand(conn, cmd)
	expected = "1"
	assertResult(t, r, expected)

	fireCommand(conn, "ABORT")
	wg.Wait()
}

func assertResult(t *testing.T, r interface{}, expected string) {
	if r != expected {
		t.Fail()
	}
}
