package tests

import (
	"fmt"
	"sync"
	"testing"

	"gotest.tools/v3/assert"
)

func TestQWATCH(t *testing.T) {
	var wg sync.WaitGroup
	runTestServer(&wg)

	publisher := getLocalConnection()
	subscriber := getLocalConnection()

	rp := fireCommandAndGetRESPParser(subscriber, "QWATCH \"SELECT $key, $value FROM `match:100:*` ORDER BY $value DESC LIMIT 10\"")
	if rp == nil {
		t.Fail()
	}

	messages := []string{"val1", "val2", "val3", "val4", "val5", "val6", "val7", "val8", "val9", "val10"}

	// Read first message (OK)
	v, err := rp.DecodeOne()
	assert.NilError(t, err)
	assert.Equal(t, "OK", v.(string))

	for idx, msg := range messages {
		fireCommand(publisher, fmt.Sprintf("SET match:100:%d %s", idx, msg))

		// Check if the update is received by the subscriber.
		v, err := rp.DecodeOne()
		assert.NilError(t, err)

		// Message format: [key, op, message]
		// Ensure the update matches the expected value.
		update := v.([]interface{})
		assert.Equal(t, "key:k1", update[0].(string), "unexpected key")
		assert.Equal(t, "op:SET", update[1].(string), "unexpected operation")
		assert.Equal(t, msg, update[2].(string), "unexpected message")
	}

	fireCommand(publisher, "ABORT")
	wg.Wait()
}
