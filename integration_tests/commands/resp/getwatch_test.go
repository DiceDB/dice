package resp

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dicedb-go"
	"github.com/stretchr/testify/assert"
)

const (
	getWatchKey = "getwatchkey"
)

type getWatchTestCase struct {
	key         string
	fingerprint string
	val         string
}

var getWatchTestCases = []getWatchTestCase{
	{getWatchKey, "2714318480", "value1"},
	{getWatchKey, "2714318480", "value2"},
	{getWatchKey, "2714318480", "value3"},
	{getWatchKey, "2714318480", "value4"},
}

func TestGETWATCH(t *testing.T) {
	publisher := getLocalConnection()
	subscribers := []net.Conn{getLocalConnection(), getLocalConnection(), getLocalConnection()}

	defer func() {
		err := ClosePublisherSubscribers(publisher, subscribers)
		assert.Nil(t, err)
	}()

	FireCommand(publisher, fmt.Sprintf("DEL %s", getWatchKey))

	// Fire a SET command to set a key
	res := FireCommand(publisher, fmt.Sprintf("SET %s %s", getWatchKey, "value"))
	assert.Equal(t, "OK", res)

	respParsers := make([]*clientio.RESPParser, len(subscribers))
	for i, subscriber := range subscribers {
		// subscribe to updates
		rp := fireCommandAndGetRESPParser(subscriber, fmt.Sprintf("GET.WATCH %s", getWatchKey))
		assert.True(t, rp != nil)
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
	for _, tc := range getWatchTestCases {
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
			assert.Equal(t, "2714318480", castedValue[1])
			assert.Equal(t, tc.val, castedValue[2])
		}
	}

	// unsubscribe from updates
	unsubscribeFromWatchUpdates(t, subscribers, "GET", "2714318480")
}

func TestGETWATCHWithSDK(t *testing.T) {
	publisher := getLocalSdk()
	subscribers := []WatchSubscriber{{client: getLocalSdk()}, {client: getLocalSdk()}, {client: getLocalSdk()}}

	defer func() {
		err := ClosePublisherSubscribersSDK(publisher, subscribers)
		assert.Nil(t, err)
	}()

	publisher.Del(context.Background(), getWatchKey)

	channels := make([]<-chan *dicedb.WatchResult, len(subscribers))
	for i, subscriber := range subscribers {
		// subscribe to updates
		watch := subscriber.client.WatchConn(context.Background())
		subscribers[i].watch = watch
		assert.True(t, watch != nil)

		firstMsg, err := watch.Watch(context.Background(), "GET", getWatchKey)
		assert.Nil(t, err)
		assert.Equal(t, "2714318480", firstMsg.Fingerprint)

		channels[i] = watch.Channel()
	}

	for _, tc := range getWatchTestCases {
		err := publisher.Set(context.Background(), tc.key, tc.val, 0).Err()
		assert.Nil(t, err)

		for _, channel := range channels {
			v := <-channel
			assert.Equal(t, "GET", v.Command)            // command
			assert.Equal(t, "2714318480", v.Fingerprint) // Fingerprint
			assert.Equal(t, tc.val, v.Data.(string))     // data
		}
	}

	// unsubscribe from updates
	unsubscribeFromWatchUpdatesSDK(t, subscribers, "GET", "2714318480")
}

func TestGETWATCHWithSDK2(t *testing.T) {
	publisher := getLocalSdk()
	subscribers := []WatchSubscriber{{client: getLocalSdk()}, {client: getLocalSdk()}, {client: getLocalSdk()}}

	defer func() {
		err := ClosePublisherSubscribersSDK(publisher, subscribers)
		assert.Nil(t, err)
	}()

	publisher.Del(context.Background(), getWatchKey)

	channels := make([]<-chan *dicedb.WatchResult, len(subscribers))
	for i, subscriber := range subscribers {
		watch := subscriber.client.WatchConn(context.Background())
		subscribers[i].watch = watch
		assert.True(t, watch != nil)

		// subscribe to updates
		firstMsg, err := watch.GetWatch(context.Background(), getWatchKey)
		assert.Nil(t, err)
		assert.Equal(t, "2714318480", firstMsg.Fingerprint)
		channels[i] = watch.Channel()
	}

	for _, tc := range getWatchTestCases {
		err := publisher.Set(context.Background(), tc.key, tc.val, 0).Err()
		assert.Nil(t, err)

		for _, channel := range channels {
			v := <-channel
			assert.Equal(t, "GET", v.Command)            // command
			assert.Equal(t, "2714318480", v.Fingerprint) // Fingerprint
			assert.Equal(t, tc.val, v.Data.(string))     // data
		}
	}

	// unsubscribe from updates
	unsubscribeFromWatchUpdatesSDK(t, subscribers, "GET", "2714318480")
}

var getWatchWithLabelTestCases = []getWatchTestCase{
	{"k1", "1207366008", "k1-initial"},
	{"k2", "605425024", "k2-initial"},
}

type getWatchUpdates struct {
	key string
	val string
}

var getWatchWithLabelUpdates = []getWatchUpdates{
	{"k1", "k1-firstupdate"},
	{"k1", "k1-secondupdate"},
}

func TestGETWATCHWithLabelWithSDK(t *testing.T) {
	publisher := getLocalSdk()
	subscribers := []WatchSubscriber{{client: getLocalSdk()}, {client: getLocalSdk()}, {client: getLocalSdk()}}

	defer func() {
		err := ClosePublisherSubscribersSDK(publisher, subscribers)
		assert.Nil(t, err)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// delete keys if they already exist
	publisher.Del(ctx, getWatchWithLabelTestCases[0].key)
	publisher.Del(ctx, getWatchWithLabelTestCases[1].key)

	channels := make([]<-chan *dicedb.WatchResult, len(subscribers))

	// set initial values
	publisher.Set(ctx, getWatchWithLabelTestCases[0].key, getWatchWithLabelTestCases[0].val, 0)
	publisher.Set(ctx, getWatchWithLabelTestCases[1].key, getWatchWithLabelTestCases[1].val, 0)

	// subscribe first key
	for i, subscriber := range subscribers {
		watch := subscriber.client.WatchConn(ctx)
		assert.True(t, watch != nil)
		subscribers[i].watch = watch

		// subscribe to updates
		firstMsg, err := watch.GetWatch(ctx, getWatchWithLabelTestCases[0].key)

		assert.Nil(t, err)
		assert.NotNil(t, firstMsg)

		assert.Equal(t, getWatchWithLabelTestCases[0].fingerprint, firstMsg.Fingerprint)

		val, ok := firstMsg.Data.(string)
		assert.True(t, ok)
		assert.Equal(t, getWatchWithLabelTestCases[0].val, val)

		// Get the channel after calling GetWatch
		channels[i] = watch.Channel()
	}

	// Cocurrently do the following:
	// 1. Update already subscribed key
	// 2. Subscribe new key

	wg := sync.WaitGroup{}

	// 1. update already subscribed key
	wg.Add(1)
	go func() {
		defer wg.Done()
		publisher.Set(context.Background(), getWatchWithLabelTestCases[0].key, "k1-firstupdate", 0)
		time.Sleep(100 * time.Millisecond)
		publisher.Set(context.Background(), getWatchWithLabelTestCases[0].key, "k1-secondupdate", 0)
	}()

	// 2. subscribe new key
	wg.Add(1)
	go func() {
		defer wg.Done()
		watch := subscribers[0].watch

		firstMsg, err := watch.GetWatch(ctx, getWatchWithLabelTestCases[1].key)

		assert.Nil(t, err)
		assert.NotNil(t, firstMsg)
		assert.Equal(t, getWatchWithLabelTestCases[1].fingerprint, firstMsg.Fingerprint)

		val, ok := firstMsg.Data.(string)
		assert.True(t, ok)
		assert.Equal(t, getWatchWithLabelTestCases[1].val, val)
	}()

	wg.Wait()

	// check if the subscribers received the updates
	for _, channel := range channels {
		for i := 0; i < 2; i++ {
			select {
			case v := <-channel:
				assert.NotNil(t, v)

				assert.Equal(t, "GET", v.Command)
				assert.Equal(t, getWatchWithLabelTestCases[0].fingerprint, v.Fingerprint)

				val, ok := v.Data.(string)
				assert.True(t, ok)
				assert.Equal(t, getWatchWithLabelUpdates[i].val, val)
			case <-ctx.Done():
				t.Errorf("Timeout waiting for update %d", i)
			}
		}
	}

	// unsubscribe from updates
	unsubscribeFromWatchUpdatesSDK(t, subscribers, "GET", getWatchWithLabelTestCases[0].fingerprint)
	unsubscribeFromWatchUpdatesSDK(t, subscribers, "GET", getWatchWithLabelTestCases[1].fingerprint)
}

func TestGETWATCHWhenClientUnsubscribe(t *testing.T) {
	publisher := getLocalSdk()
	subscribers := []WatchSubscriber{{client: getLocalSdk()}, {client: getLocalSdk()}, {client: getLocalSdk()}}

	defer func() {
		err := ClosePublisherSubscribersSDK(publisher, subscribers)
		assert.Nil(t, err)
	}()

	publisher.Del(context.Background(), getWatchKey)

	channels := make([]<-chan *dicedb.WatchResult, len(subscribers))
	for i, subscriber := range subscribers {
		// subscribe to updates
		watch := subscriber.client.WatchConn(context.Background())
		subscribers[i].watch = watch
		assert.True(t, watch != nil)

		firstMsg, err := watch.Watch(context.Background(), "GET", getWatchKey)
		assert.Nil(t, err)
		assert.Equal(t, "2714318480", firstMsg.Fingerprint)

		channels[i] = watch.Channel()
	}

	for _, tc := range getWatchTestCases {
		err := publisher.Set(context.Background(), tc.key, tc.val, 0).Err()
		assert.Nil(t, err)

		for _, channel := range channels {
			v := <-channel
			assert.Equal(t, "GET", v.Command)            // command
			assert.Equal(t, "2714318480", v.Fingerprint) // Fingerprint
			assert.Equal(t, tc.val, v.Data.(string))     // data
		}
	}

	// unsubscribe first 2 clients from updates
	unsubscribeFromWatchUpdatesSDK(t, subscribers[:2], "GET", "2714318480")
	// Trigger an update and verify only the third subscriber receives it
	err := publisher.Set(context.Background(), getWatchKey, "new-updated-val", 0).Err()
	assert.Nil(t, err)

	// Assert other unsubscribed clients don't get any updates
	for _, channel := range channels[:2] {
		select {
		case v := <-channel:
			assert.Fail(t, fmt.Sprintf("Unexpected update received: %v", v))
		case <-time.After(100 * time.Millisecond):
			// This is the expected behavior - no response within the timeout
		}
	}

	lastMsg := <-channels[2]
	assert.Equal(t, "GET", lastMsg.Command)
	assert.Equal(t, "2714318480", lastMsg.Fingerprint)
	assert.Equal(t, "new-updated-val", lastMsg.Data.(string))

	// Unsubscribe all clients
	unsubscribeFromWatchUpdatesSDK(t, subscribers, "GET", getWatchKey)
}
