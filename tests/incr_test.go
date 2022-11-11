package tests

import (
	"fmt"
	"sync"
	"testing"
)

func TestINCRWhenKeyAlreadyExists(t *testing.T) {
	var wg sync.WaitGroup
	go runTestServer(&wg)

	conn := getLocalConnection()

	cmd := fmt.Sprintf("SET key %s", "1")
	fireCommand(conn, cmd)

	cmd = fmt.Sprintf("INCR key")

	value := fireCommand(conn, cmd)

	if value.(int64) != 2 {
		t.Fail()
	}

	fireCommand(conn, "ABORT")
	wg.Wait()
}

func TestINCRWhenKeyDoesNotExist(t *testing.T) {
	var wg sync.WaitGroup
	go runTestServer(&wg)

	conn := getLocalConnection()

	cmd := fmt.Sprintf("INCR key")

	value := fireCommand(conn, cmd)

	if value.(int64) != 1 {
		t.Fail()
	}

	fireCommand(conn, "ABORT")
	wg.Wait()
}

func TestINCRWhenMultipleCommandsFired(t *testing.T) {
	var wg sync.WaitGroup
	go runTestServer(&wg)

	conn := getLocalConnection()

	cmd := fmt.Sprintf("SET k%d %s", 1, "0")
	fireCommand(conn, cmd)

	for i := 1; i <= 5; i++ {
		cmd := fmt.Sprintf("INCR k%d", 1)
		fireCommand(conn, cmd)
	}

	cmd = fmt.Sprintf("GET k1")
	value := fireCommand(conn, cmd)
	if value != "5" {
		t.Fail()
	}

	fireCommand(conn, "ABORT")
	wg.Wait()
}
