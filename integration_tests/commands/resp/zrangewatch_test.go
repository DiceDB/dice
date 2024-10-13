package resp

import (
	"context"
	"fmt"
	"github.com/dicedb/dice/internal/clientio"
	dicedb "github.com/dicedb/dicedb-go"
	"gotest.tools/v3/assert"
	"net"
	"testing"
	"time"
)

type zrangeWatchTestCase struct {
	key    string
	score  float64
	val    string
	result any
}

const (
	zrangeWatchKey   = "zrangewatchkey"
	zrangeWatchQuery = "ZRANGE.WATCH %s 0 -1 REV WITHSCORES"
)

var zrangeWatchTestCases = []zrangeWatchTestCase{
	{zrangeWatchKey, 1.0, "member1", []any{"member1", "1"}},
	{zrangeWatchKey, 2.0, "member2", []any{"member2", "2", "member1", "1"}},
	{zrangeWatchKey, 3.0, "member3", []any{"member3", "3", "member2", "2", "member1", "1"}},
	{zrangeWatchKey, 4.0, "member4", []any{"member4", "4", "member3", "3", "member2", "2", "member1", "1"}},
}

func TestZRANGEWATCH(t *testing.T) {
	publisher := getLocalConnection()
	subscribers := []net.Conn{getLocalConnection(), getLocalConnection(), getLocalConnection()}

	FireCommand(publisher, fmt.Sprintf("DEL %s", zrangeWatchKey))

	defer func() {
		if err := publisher.Close(); err != nil {
			t.Errorf("Error closing publisher connection: %v", err)
		}
		for _, sub := range subscribers {
			time.Sleep(100 * time.Millisecond)
			if err := sub.Close(); err != nil {
				t.Errorf("Error closing subscriber connection: %v", err)
			}
		}
	}()

	respParsers := make([]*clientio.RESPParser, len(subscribers))
	for i, subscriber := range subscribers {
		rp := fireCommandAndGetRESPParser(subscriber, fmt.Sprintf(zrangeWatchQuery, zrangeWatchKey))
		assert.Assert(t, rp != nil)
		respParsers[i] = rp

		v, err := rp.DecodeOne()
		assert.NilError(t, err)
		castedValue, ok := v.([]interface{})
		if !ok {
			t.Errorf("Type assertion to []interface{} failed for value: %v", v)
		}
		assert.Equal(t, 3, len(castedValue))
	}

	// Fire updates to the sorted set and check if the subscribers receive the updates in the push-response form
	for _, tc := range zrangeWatchTestCases {
		FireCommand(publisher, fmt.Sprintf("ZADD %s %f %s", tc.key, tc.score, tc.val))

		for _, rp := range respParsers {
			v, err := rp.DecodeOne()
			assert.NilError(t, err)
			castedValue, ok := v.([]interface{})
			if !ok {
				t.Errorf("Type assertion to []interface{} failed for value: %v", v)
			}
			assert.Equal(t, 3, len(castedValue))
			assert.Equal(t, "ZRANGE", castedValue[0])
			assert.Equal(t, "2491069200", castedValue[1])
			assert.DeepEqual(t, tc.result, castedValue[2])
		}
	}
}

type zrangeWatchSDKTestCase struct {
	key    string
	score  float64
	val    string
	result []dicedb.Z
}

var zrangeWatchSDKTestCases = []zrangeWatchSDKTestCase{
	{zrangeWatchKey, 1.0, "member1", []dicedb.Z{
		{Score: 1, Member: "member1"},
	}},
	{zrangeWatchKey, 2.0, "member2", []dicedb.Z{
		{Score: 2, Member: "member2"},
		{Score: 1, Member: "member1"},
	}},
	{zrangeWatchKey, 3.0, "member3", []dicedb.Z{
		{Score: 3, Member: "member3"},
		{Score: 2, Member: "member2"},
		{Score: 1, Member: "member1"},
	}},
	{zrangeWatchKey, 4.0, "member4", []dicedb.Z{
		{Score: 4, Member: "member4"},
		{Score: 3, Member: "member3"},
		{Score: 2, Member: "member2"},
		{Score: 1, Member: "member1"},
	}},
}

func TestZRANGEWATCHWithSDK(t *testing.T) {
	publisher := getLocalSdk()
	subscribers := []WatchSubscriber{{client: getLocalSdk()}, {client: getLocalSdk()}, {client: getLocalSdk()}}

	publisher.Del(context.Background(), zrangeWatchKey)

	channels := make([]<-chan *dicedb.WatchResult, len(subscribers))
	for i, subscriber := range subscribers {
		watch := subscriber.client.WatchConn(context.Background())
		subscribers[i].watch = watch
		assert.Assert(t, watch != nil)
		firstMsg, err := watch.Watch(context.Background(), "ZRANGE", zrangeWatchKey, "0", "-1", "REV", "WITHSCORES")
		assert.NilError(t, err)
		assert.Equal(t, firstMsg.Command, "ZRANGE")
		assert.Equal(t, firstMsg.Fingerprint, "2491069200")
		channels[i] = watch.Channel()
	}

	for _, tc := range zrangeWatchSDKTestCases {
		err := publisher.ZAdd(context.Background(), tc.key, dicedb.Z{
			Score:  tc.score,
			Member: tc.val,
		}).Err()
		assert.NilError(t, err)

		for _, channel := range channels {
			v := <-channel

			assert.Equal(t, "ZRANGE", v.Command)         // command
			assert.Equal(t, "2491069200", v.Fingerprint) // Fingerprint
			assert.DeepEqual(t, tc.result, v.Data)       // data
		}
	}
}

func TestZRANGEWATCHWithSDK2(t *testing.T) {
	publisher := getLocalSdk()
	subscribers := []WatchSubscriber{{client: getLocalSdk()}, {client: getLocalSdk()}, {client: getLocalSdk()}}

	publisher.Del(context.Background(), zrangeWatchKey)

	channels := make([]<-chan *dicedb.WatchResult, len(subscribers))
	for i := range subscribers {
		conn := subscribers[i].client.WatchConn(context.Background())
		subscribers[i].watch = conn
		firstMsg, err := conn.ZRangeWatch(context.Background(), zrangeWatchKey, "0", "-1", "REV", "WITHSCORES")
		assert.NilError(t, err)
		assert.Equal(t, firstMsg.Command, "ZRANGE")
		assert.Equal(t, firstMsg.Fingerprint, "2491069200")
		channels[i] = conn.Channel()
	}

	for _, tc := range zrangeWatchSDKTestCases {
		err := publisher.ZAdd(context.Background(), tc.key, dicedb.Z{
			Score:  tc.score,
			Member: tc.val,
		}).Err()
		assert.NilError(t, err)

		for _, channel := range channels {
			v := <-channel

			assert.Equal(t, "ZRANGE", v.Command)
			assert.Equal(t, "2491069200", v.Fingerprint)
			assert.DeepEqual(t, tc.result, v.Data)
		}
	}
}
