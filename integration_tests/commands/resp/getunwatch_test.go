package resp

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dicedb-go"
	"github.com/stretchr/testify/assert"
)

const (
	getUnwatchKey = "getunwatchkey"
	fingerprint   = "3557732805"
)

type getUnwatchTestCase struct {
	key string
	val string
}

var getUnwatchTestCases = []getUnwatchTestCase{
	{getUnwatchKey, "value1"},
	{getUnwatchKey, "value2"},
	{getUnwatchKey, "value3"},
	{getUnwatchKey, "value4"},
}

func TestGETUNWATCH(t *testing.T) {
	publisher := getLocalConnection()
	subscribers := []net.Conn{getLocalConnection(), getLocalConnection(), getLocalConnection()}

	defer func() {
		if err := publisher.Close(); err != nil {
			t.Errorf("Error closing publisher connection: %v", err)
		}
		for _, sub := range subscribers {
			if err := sub.Close(); err != nil {
				t.Errorf("Error closing subscriber connection: %v", err)
			}
		}
	}()

	FireCommand(publisher, fmt.Sprintf("DEL %s", getUnwatchKey))

	// Fire SET command to set the key
	res := FireCommand(publisher, fmt.Sprintf("SET %s %s", getUnwatchKey, "value"))
	assert.Equal(t, "OK", res)

	// subscribe for updates
	respParsers := make([]*clientio.RESPParser, len(subscribers))
	for i, subscriber := range subscribers {
		rp := fireCommandAndGetRESPParser(subscriber, fmt.Sprintf("GET.WATCH %s", getUnwatchKey))
		assert.NotNil(t, rp)
		respParsers[i] = rp

		v, err := rp.DecodeOne()
		assert.Nil(t, err)
		castedValue, ok := v.([]interface{})
		if !ok {
			t.Errorf("Type assertion to []interface{} failed for value: %v", v)
		}
		assert.Equal(t, 3, len(castedValue))
	}

	//	Fire updates to the key using the publisher, then check if the subscribers receive the updates in the push-response form (i.e. array of three elements, with third element being the value)
	for _, tc := range getUnwatchTestCases {
		res := FireCommand(publisher, fmt.Sprintf("SET %s %s", tc.key, tc.val))
		assert.Equal(t, "OK", res)

		for _, rp := range respParsers {
			v, err := rp.DecodeOne()
			assert.Nil(t, err)
			castedValue, ok := v.([]interface{})
			if !ok {
				t.Errorf("Type assertion to []interface{} failed for value: %v", v)
			}
			assert.Equal(t, 3, len(castedValue))
			assert.Equal(t, "GET", castedValue[0])
			assert.Equal(t, fingerprint, castedValue[1])
			assert.Equal(t, tc.val, castedValue[2])
		}
	}

	// unsubscribe from updates
	for _, subscriber := range subscribers {
		rp := fireCommandAndGetRESPParser(subscriber, fmt.Sprintf("GET.UNWATCH %s", fingerprint))
		assert.NotNil(t, rp)

		v, err := rp.DecodeOne()
		assert.NoError(t, err)
		castedValue, ok := v.(string)
		if !ok {
			t.Errorf("Type assertion to string failed for value: %v", v)
		}
		assert.Equal(t, castedValue, "OK")
	}

	// Test updates are not sent after unsubscribing
	for _, tc := range getUnwatchTestCases[2:] {
		res := FireCommand(publisher, fmt.Sprintf("SET %s %s", tc.key, tc.val))
		assert.Equal(t, "OK", res)

		for _, rp := range respParsers {
			responseChan := make(chan interface{})
			errChan := make(chan error)

			go func() {
				v, err := rp.DecodeOne()
				if err != nil {
					errChan <- err
				} else {
					responseChan <- v
				}
			}()

			select {
			case v := <-responseChan:
				t.Errorf("Unexpected response after unwatch: %v", v)
			case err := <-errChan:
				t.Errorf("Error while decoding: %v", err)
			case <-time.After(100 * time.Millisecond):
				// This is the expected behavior - no response within the timeout
			}
		}
	}
}

func TestGETUNWATCHWithSDK(t *testing.T) {
	publisher := getLocalSdk()
	subscribers := []WatchSubscriber{{client: getLocalSdk()}, {client: getLocalSdk()}, {client: getLocalSdk()}}

	publisher.Del(context.Background(), getUnwatchKey)

	// subscribe for updates
	channels := make([]<-chan *dicedb.WatchResult, len(subscribers))
	for i, subscriber := range subscribers {
		watch := subscriber.client.WatchConn(context.Background())
		subscribers[i].watch = watch
		assert.NotNil(t, watch)
		firstMsg, err := watch.Watch(context.Background(), "GET", getUnwatchKey)
		assert.Nil(t, err)
		assert.Equal(t, firstMsg.Command, "GET")
		assert.Equal(t, firstMsg.Fingerprint, fingerprint)
		channels[i] = watch.Channel()
	}

	// Fire updates and validate receipt
	err := publisher.Set(context.Background(), getUnwatchKey, "check", 0).Err()
	assert.Nil(t, err)

	for _, channel := range channels {
		v := <-channel
		assert.Equal(t, "GET", v.Command)           // command
		assert.Equal(t, fingerprint, v.Fingerprint) // Fingerprint
		assert.Equal(t, "check", v.Data.(string))   // data
	}

	// unsubscribe from updates
	for _, subscriber := range subscribers {
		err := subscriber.watch.Unwatch(context.Background(), "GET", fingerprint)
		assert.Nil(t, err)
	}

	// fire updates and validate that they are not received
	err = publisher.Set(context.Background(), getUnwatchKey, "final", 0).Err()
	assert.Nil(t, err)
	for _, channel := range channels {
		go func(ch <-chan *dicedb.WatchResult) {
			select {
			case v := <-ch:
				assert.Fail(t, fmt.Sprintf("%v", v))
			case <-time.After(100 * time.Millisecond):
				// This is the expected behavior - no response within the timeout
			}
		}(channel)
	}
}
