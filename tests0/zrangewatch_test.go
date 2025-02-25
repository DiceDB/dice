// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

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

// func TestZRANGEWATCH(t *testing.T) {
// 	publisher := getLocalConnection()
// 	subscribers := []net.Conn{getLocalConnection(), getLocalConnection(), getLocalConnection()}
// 	defer func() {
// 		err := ClosePublisherSubscribers(publisher, subscribers)
// 		assert.Nil(t, err)
// 	}()

// 	FireCommand(publisher, fmt.Sprintf("DEL %s", zrangeWatchKey))

// 	respParsers := make([]*clientio.RESPParser, len(subscribers))
// 	for i, subscriber := range subscribers {
// 		rp := fireCommandAndGetRESPParser(subscriber, fmt.Sprintf(zrangeWatchQuery, zrangeWatchKey))
// 		assert.NotNil(t, rp)
// 		respParsers[i] = rp

// 		v, err := rp.DecodeOne()
// 		assert.Nil(t, err)
// 		castedValue, ok := v.([]interface{})
// 		if !ok {
// 			t.Errorf("Type assertion to []interface{} failed for value: %v", v)
// 		}
// 		assert.Equal(t, 3, len(castedValue))
// 	}

// 	// Fire updates to the sorted set and check if the subscribers receive the updates in the push-response form
// 	for _, tc := range zrangeWatchTestCases {
// 		FireCommand(publisher, fmt.Sprintf("ZADD %s %f %s", tc.key, tc.score, tc.val))

// 		for _, rp := range respParsers {
// 			v, err := rp.DecodeOne()
// 			assert.Nil(t, err)
// 			castedValue, ok := v.([]interface{})
// 			if !ok {
// 				t.Errorf("Type assertion to []interface{} failed for value: %v", v)
// 			}
// 			assert.Equal(t, 3, len(castedValue))
// 			assert.Equal(t, "ZRANGE", castedValue[0])
// 			assert.Equal(t, "1178068413", castedValue[1])
// 			assert.Equal(t, tc.result, castedValue[2])
// 		}
// 	}

// 	unsubscribeFromWatchUpdates(t, subscribers, "ZRANGE", "1178068413")
// }

// type zrangeWatchSDKTestCase struct {
// 	key    string
// 	score  float64
// 	val    string
// 	result []dicedb.Z
// }

// var zrangeWatchSDKTestCases = []zrangeWatchSDKTestCase{
// 	{zrangeWatchKey, 1.0, "member1", []dicedb.Z{
// 		{Score: 1, Member: "member1"},
// 	}},
// 	{zrangeWatchKey, 2.0, "member2", []dicedb.Z{
// 		{Score: 2, Member: "member2"},
// 		{Score: 1, Member: "member1"},
// 	}},
// 	{zrangeWatchKey, 3.0, "member3", []dicedb.Z{
// 		{Score: 3, Member: "member3"},
// 		{Score: 2, Member: "member2"},
// 		{Score: 1, Member: "member1"},
// 	}},
// 	{zrangeWatchKey, 4.0, "member4", []dicedb.Z{
// 		{Score: 4, Member: "member4"},
// 		{Score: 3, Member: "member3"},
// 		{Score: 2, Member: "member2"},
// 		{Score: 1, Member: "member1"},
// 	}},
// }

// func TestZRANGEWATCHWithSDK(t *testing.T) {
// 	publisher := getLocalSdk()
// 	subscribers := []WatchSubscriber{{client: getLocalSdk()}, {client: getLocalSdk()}, {client: getLocalSdk()}}
// 	defer func() {
// 		err := ClosePublisherSubscribersSDK(publisher, subscribers)
// 		assert.Nil(t, err)
// 	}()

// 	publisher.Del(context.Background(), zrangeWatchKey)

// 	channels := make([]<-chan *dicedb.WatchResult, len(subscribers))
// 	for i, subscriber := range subscribers {
// 		watch := subscriber.client.WatchConn(context.Background())
// 		subscribers[i].watch = watch
// 		assert.NotNil(t, watch)
// 		firstMsg, err := watch.Watch(context.Background(), "ZRANGE", zrangeWatchKey, "0", "-1", "REV", "WITHSCORES")
// 		assert.Nil(t, err)
// 		assert.Equal(t, firstMsg.Command, "ZRANGE")
// 		assert.Equal(t, firstMsg.Fingerprint, "1178068413")
// 		channels[i] = watch.Channel()
// 	}

// 	for _, tc := range zrangeWatchSDKTestCases {
// 		err := publisher.ZAdd(context.Background(), tc.key, dicedb.Z{
// 			Score:  tc.score,
// 			Member: tc.val,
// 		}).Err()
// 		assert.Nil(t, err)

// 		for _, channel := range channels {
// 			v := <-channel

// 			assert.Equal(t, "ZRANGE", v.Command)         // command
// 			assert.Equal(t, "1178068413", v.Fingerprint) // Fingerprint
// 			assert.Equal(t, tc.result, v.Data)           // data
// 		}
// 	}

// 	unsubscribeFromWatchUpdatesSDK(t, subscribers, "ZRANGE", "1178068413")
// }

// func TestZRANGEWATCHWithSDK2(t *testing.T) {
// 	publisher := getLocalSdk()
// 	subscribers := []WatchSubscriber{{client: getLocalSdk()}, {client: getLocalSdk()}, {client: getLocalSdk()}}
// 	defer func() {
// 		err := ClosePublisherSubscribersSDK(publisher, subscribers)
// 		assert.Nil(t, err)
// 	}()

// 	publisher.Del(context.Background(), zrangeWatchKey)

// 	channels := make([]<-chan *dicedb.WatchResult, len(subscribers))
// 	for i := range subscribers {
// 		conn := subscribers[i].client.WatchConn(context.Background())
// 		subscribers[i].watch = conn
// 		firstMsg, err := conn.ZRangeWatch(context.Background(), zrangeWatchKey, "0", "-1", "REV", "WITHSCORES")
// 		assert.Nil(t, err)
// 		assert.Equal(t, firstMsg.Command, "ZRANGE")
// 		assert.Equal(t, firstMsg.Fingerprint, "1178068413")
// 		channels[i] = conn.Channel()
// 	}

// 	for _, tc := range zrangeWatchSDKTestCases {
// 		err := publisher.ZAdd(context.Background(), tc.key, dicedb.Z{
// 			Score:  tc.score,
// 			Member: tc.val,
// 		}).Err()
// 		assert.Nil(t, err)

// 		for _, channel := range channels {
// 			v := <-channel

// 			assert.Equal(t, "ZRANGE", v.Command)
// 			assert.Equal(t, "1178068413", v.Fingerprint)
// 			assert.Equal(t, tc.result, v.Data)
// 		}
// 	}

// 	unsubscribeFromWatchUpdatesSDK(t, subscribers, "ZRANGE", "1178068413")
// }
