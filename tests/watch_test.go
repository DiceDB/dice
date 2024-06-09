package tests

import (
	"fmt"
	"log"
	"sync"
	"testing"
)

func TestWatch(t *testing.T) {
	var wg sync.WaitGroup
	go runTestServer(&wg)

	publisher := getLocalConnection()
	subscriber := getLocalConnection()

	rp := fireCommandAndGetRESPParser(subscriber, "SUBSCRIBE k1")
	if rp == nil {
		t.Fail()
	}

	messages := []string{"val1", "val2", "val3", "val4", "val5", "val6", "val7", "val8", "val9", "val10"}

	// Read first message (OK)
	v, err := rp.DecodeOne()
	if err != nil {
		t.Fail()
	}
	if v.(string) != "OK" {
		log.Fatalf("expected OK, got %s", v.(string))
	}

	for _, msg := range messages {
		fireCommand(publisher, fmt.Sprintf("SET k1 %s", msg))

		// Check if the update is received by the subscriber.
		v, err := rp.DecodeOne()
		if err != nil {
			t.Fail()
		}

		// Message format: [key, op, message]
		// Ensure the update matches the expected value.
		if v.([]interface{})[0].(string) != "key:k1" {
			log.Fatalf("expected %s, got %s", msg, v.([]interface{})[1].(string))
		}
		if v.([]interface{})[1].(string) != "op:SET" {
			log.Fatalf("expected %s, got %s", msg, v.([]interface{})[1].(string))
		}
		if v.([]interface{})[2].(string) != msg {
			log.Fatalf("expected %s, got %s", msg, v.([]interface{})[2].(string))
		}
	}

	fireCommand(publisher, "ABORT")
	wg.Wait()
}
