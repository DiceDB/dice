// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

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
		err := ClosePublisherSubscribers(publisher, subscribers)
		assert.Nil(t, err)
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
			assert.Equal(t, "426696421", castedValue[1])
			assert.Equal(t, tc.val, castedValue[2])
		}
	}

	// unsubscribe from updates
	unsubscribeFromWatchUpdates(t, subscribers, "GET", "426696421")

	// Test updates are not sent after unsubscribing
	for _, tc := range getUnwatchTestCases[2:] {
		res := FireCommand(publisher, fmt.Sprintf("SET %s %s", tc.key, tc.val))
		assert.Equal(t, "OK", res)

		for _, rp := range respParsers {
			responseChan := make(chan interface{})
			errChan := make(chan error)

			go func() {
				v, err := rp.DecodeOne()
				select {
				case errChan <- err:
				case responseChan <- v:
				case <-time.After(200 * time.Millisecond):
					// if test goroutine returns, this one must exit too
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	defer func() {
		err := ClosePublisherSubscribersSDK(publisher, subscribers)
		assert.Nil(t, err)
	}()

	publisher.Del(ctx, getUnwatchKey)

	// subscribe for updates
	channels := make([]<-chan *dicedb.WatchResult, len(subscribers))
	for i, subscriber := range subscribers {
		watch := subscriber.client.WatchConn(ctx)
		subscribers[i].watch = watch
		assert.NotNil(t, watch)
		firstMsg, err := watch.Watch(ctx, "GET", getUnwatchKey)
		assert.Nil(t, err)
		assert.Equal(t, firstMsg.Command, "GET")
		assert.Equal(t, "426696421", firstMsg.Fingerprint)
		channels[i] = watch.Channel()
	}

	// Fire updates and validate receipt
	err := publisher.Set(ctx, getUnwatchKey, "check", 0).Err()
	assert.Nil(t, err)

	for _, channel := range channels {
		v := <-channel
		assert.Equal(t, "GET", v.Command)           // command
		assert.Equal(t, "426696421", v.Fingerprint) // Fingerprint
		assert.Equal(t, "check", v.Data.(string))   // data
	}

	// unsubscribe from updates
	unsubscribeFromWatchUpdatesSDK(t, subscribers, "GET", "426696421")

	// fire updates and validate that they are not received
	err = publisher.Set(ctx, getUnwatchKey, "final", 0).Err()
	assert.Nil(t, err)
	for _, channel := range channels {
		select {
		case v := <-channel:
			assert.Fail(t, fmt.Sprintf("%v", v))
		case <-time.After(100 * time.Millisecond):
			// This is the expected behavior - no response within the timeout
		}
	}
}
