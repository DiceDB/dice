// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

type pfcountWatchTestCase struct {
	key    string
	val    any
	result int64
}

type pfcountWatchWithPFMergeTestCase struct {
	destKey1   string
	destValue1 []string
	destKey2   string
	destValue2 []string
	result     int64
}

const (
	pfcountCommand          = "PFCOUNT"
	pfcountWatchKey         = "hllkey"
	pfcountWatchQuery       = "PFCOUNT.WATCH %s"
	pfcountWatchFingerPrint = "1832643469"
)

var pfcountWatchTestCases = []pfcountWatchTestCase{
	{pfcountWatchKey, "value1", int64(2)},
	{pfcountWatchKey, "value2", int64(3)},
	{pfcountWatchKey, "value3", int64(4)},
	{pfcountWatchKey, "value4", int64(5)},
}

var pfcountWatchhWithPFMergeTestCases = []pfcountWatchWithPFMergeTestCase{
	{"DEST_KEY_1", []string{"a", "b", "c", "d", "e", "f"}, "DEST_KEY_2", []string{"f", "g"}, int64(13)},
	{"DEST_KEY_3", []string{"a", "b", "c", "d", "e", "f"}, "DEST_KEY_4", []string{"a1", "b1", "c1", "d1", "e1", "f1"}, int64(19)},
	{"DEST_KEY_5", []string{"f", "g"}, "DEST_KEY_6", []string{"f", "g"}, int64(19)},
}

// func TestPFCOUNTWATCH(t *testing.T) {
// 	publisher := getLocalConnection()
// 	subscribers := setupSubscribers(3)
// 	defer func() {
// 		err := ClosePublisherSubscribers(publisher, subscribers)
// 		assert.Nil(t, err)
// 	}()

// 	FireCommand(publisher, fmt.Sprintf("DEL %s", pfcountWatchKey))

// 	res := FireCommand(publisher, fmt.Sprintf("PFADD %s %s", pfcountWatchKey, "randomvalue"))
// 	assert.Equal(t, int64(1), res)

// 	respParsers := setUpRespParsers(t, subscribers)

// 	t.Run("Basic PFCount Operations", func(t *testing.T) {
// 		testPFCountAdd(t, publisher, respParsers)
// 	},
// 	)

// 	t.Run("PFCount Operations including PFMerge", func(t *testing.T) {
// 		testPFCountMerge(t, publisher, respParsers)
// 	},
// 	)

// 	unsubscribeFromWatchUpdates(t, subscribers, pfcountCommand, pfcountWatchFingerPrint)
// }

// func setupSubscribers(count int) []net.Conn {
// 	subscribers := make([]net.Conn, 0, count)
// 	for i := 0; i < count; i++ {
// 		client := getLocalConnection()
// 		subscribers = append(subscribers, conn)
// 	}
// 	return subscribers
// }

// func testPFCountAdd(t *testing.T, publisher net.Conn, respParsers []*clientio.RESPParser) {
// 	for _, tc := range pfcountWatchTestCases {
// 		res := FireCommand(publisher, fmt.Sprintf("PFADD %s %s", tc.key, tc.val))
// 		assert.Equal(t, int64(1), res)

// 		verifyWatchResults(t, respParsers, tc.result)
// 	}
// }

// func testPFCountMerge(t *testing.T, publisher net.Conn, respParsers []*clientio.RESPParser) {
// 	for _, tc := range pfcountWatchhWithPFMergeTestCases {
// 		FireCommand(publisher, fmt.Sprintf("DEL %s %s", tc.destKey1, tc.destKey2))
// 		FireCommand(publisher, fmt.Sprintf("PFADD %s %s", tc.destKey1, tc.destValue1))
// 		FireCommand(publisher, fmt.Sprintf("PFADD %s %s", tc.destKey2, tc.destValue2))
// 		FireCommand(publisher, fmt.Sprintf("PFMERGE %s %s %s", pfcountWatchKey, tc.destKey1, tc.destKey2))

// 		verifyWatchResults(t, respParsers, tc.result)
// 	}
// }

// func verifyWatchResults(t *testing.T, respParsers []*clientio.RESPParser, expected int64) {
// 	for _, rp := range respParsers {
// 		v, err := rp.DecodeOne()
// 		assert.Nil(t, err)
// 		castedValue, ok := v.([]interface{})
// 		if !ok {
// 			t.Errorf("Type assertion to []interface{} failed for value: %v", v)
// 		}
// 		assert.Equal(t, 3, len(castedValue))
// 		assert.Equal(t, pfcountCommand, castedValue[0])
// 		assert.Equal(t, pfcountWatchFingerPrint, castedValue[1])
// 		assert.Equal(t, int64(expected), castedValue[2])
// 	}
// }

// const (
// 	pfcountCommandSDK          = "PFCOUNT"
// 	pfcountWatchKeySDK         = "hllkey"
// 	pfcountWatchFingerPrintSDK = "1832643469"
// 	defaultTimeout             = 5 * time.Second
// )

// type pfcountWatchSDKTestCase struct {
// 	key    string
// 	val    any
// 	result int64
// }

// type pfcountWatchSDKWithPFMergeTestCase struct {
// 	destKey1   string
// 	destValue1 []string
// 	destKey2   string
// 	destValue2 []string
// 	result     int64
// }

// var PFCountWatchSDKTestCases = []pfcountWatchSDKTestCase{
// 	{pfcountWatchKeySDK, "value1", int64(1)},
// 	{pfcountWatchKeySDK, "value2", int64(2)},
// 	{pfcountWatchKeySDK, "value3", int64(3)},
// 	{pfcountWatchKeySDK, "value4", int64(4)},
// }

// var pfcountWatchSDKhWithPFMergeTestCases = []pfcountWatchSDKWithPFMergeTestCase{
// 	{"DEST_KEY_1", []string{"a", "b", "c", "d", "e", "f"}, "DEST_KEY_2", []string{"f", "g"}, int64(11)},
// 	{"DEST_KEY_3", []string{"a", "b", "c", "d", "e", "f"}, "DEST_KEY_4", []string{"a1", "b1", "c1", "d1", "e1", "f1"}, int64(17)},
// 	{"DEST_KEY_5", []string{"f", "g"}, "DEST_KEY_6", []string{"f", "g"}, int64(17)},
// }

// func TestPFCountWATCHWithSDK(t *testing.T) {
// 	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
// 	defer cancel()

// 	publisher := getLocalSdk()
// 	subscribers := setupSubscribersSDK(3)
// 	defer func() {
// 		err := ClosePublisherSubscribersSDK(publisher, subscribers)
// 		assert.Nil(t, err)
// 	}()

// 	publisher.Del(ctx, pfcountWatchKey)

// 	channels := setUpWatchChannelsSDK(t, ctx, subscribers)

// 	t.Run("Basic PFCount Operations", func(t *testing.T) {
// 		testPFCountAddSDK(t, ctx, channels, publisher)
// 	},
// 	)

// 	t.Run("PFCount Operations including PFMerge", func(t *testing.T) {
// 		testPFCountMergeSDK(t, ctx, channels, publisher)
// 	},
// 	)

// 	unsubscribeFromWatchUpdatesSDK(t, subscribers, pfcountCommandSDK, pfcountWatchFingerPrintSDK)
// }

// func setupSubscribersSDK(count int) []WatchSubscriber {
// 	subscribers := make([]WatchSubscriber, count)
// 	for i := range subscribers {
// 		subscribers[i].client = getLocalSdk()
// 	}
// 	return subscribers
// }

// func setUpWatchChannelsSDK(t *testing.T, ctx context.Context, subscribers []WatchSubscriber) []<-chan *dicedb.WatchResult {
// 	channels := make([]<-chan *dicedb.WatchResult, len(subscribers))
// 	for i, subscriber := range subscribers {
// 		watch := subscriber.client.WatchConn(ctx)
// 		subscribers[i].watch = watch
// 		assert.NotNil(t, watch)
// 		firstMsg, err := watch.Watch(ctx, pfcountCommandSDK, pfcountWatchKey)
// 		assert.Nil(t, err)
// 		assert.Equal(t, firstMsg.Command, pfcountCommandSDK)
// 		channels[i] = watch.Channel()
// 	}
// 	return channels
// }

// func testPFCountAddSDK(t *testing.T, ctx context.Context, channels []<-chan *dicedb.WatchResult, publisher *dicedb.Client) {
// 	for _, tc := range PFCountWatchSDKTestCases {
// 		err := publisher.PFAdd(ctx, tc.key, tc.val).Err()
// 		assert.Nil(t, err)

// 		verifyWatchResultsSDK(t, channels, tc.result)
// 	}
// }

// func testPFCountMergeSDK(t *testing.T, ctx context.Context, channels []<-chan *dicedb.WatchResult, publisher *dicedb.Client) {
// 	for _, tc := range pfcountWatchSDKhWithPFMergeTestCases {
// 		publisher.Del(ctx, tc.destKey1, tc.destKey2)
// 		publisher.PFAdd(ctx, tc.destKey1, tc.destValue1).Err()
// 		publisher.PFAdd(ctx, tc.destKey2, tc.destValue2).Err()
// 		publisher.PFMerge(ctx, pfcountWatchKey, tc.destKey1, tc.destKey2).Err()

// 		verifyWatchResultsSDK(t, channels, tc.result)
// 	}
// }

// func verifyWatchResultsSDK(t *testing.T, channels []<-chan *dicedb.WatchResult, expected int64) {
// 	for _, channel := range channels {
// 		select {
// 		case v := <-channel:
// 			assert.Equal(t, pfcountCommandSDK, v.Command)           // command
// 			assert.Equal(t, pfcountWatchFingerPrint, v.Fingerprint) // Fingerprint
// 			assert.Equal(t, int64(expected), v.Data)                // data
// 		case <-time.After(defaultTimeout):
// 			t.Fatal("timeout waiting for watch result")
// 		}
// 	}
// }
