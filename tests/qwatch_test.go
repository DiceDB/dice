package tests

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

func TestQWATCH(t *testing.T) {
	publisher := getLocalConnection()
	subscriber := getLocalConnection()

	rp := fireCommandAndGetRESPParser(subscriber, "QWATCH \"SELECT $key, $value FROM `match:100:*` ORDER BY $value DESC LIMIT 3\"")
	if rp == nil {
		t.Fail()
	}

	messages := []int{11, 33, 22, 0}

	// Read first message (OK)
	v, err := rp.DecodeOne()
	assert.NilError(t, err)
	assert.Equal(t, "OK", v.(string))

	for idx, msg := range messages {
		fireCommand(publisher, fmt.Sprintf("SET match:100:user:%d %d", idx, msg))

		// Check if the update is received by the subscriber.
		v, err := rp.DecodeOne()
		assert.NilError(t, err)

		// Message format: [key, op, message]
		// Ensure the update matches the expected value.
		update := v.([]interface{})
		// print the update object.
		fmt.Println(update)
		// assert.Equal(t, fmt.Sprintf("match:100:user:%d", idx), update[0].(string), "unexpected key")
		// assert.Equal(t, msg, update[1].(string), "unexpected operation")
	}
}
