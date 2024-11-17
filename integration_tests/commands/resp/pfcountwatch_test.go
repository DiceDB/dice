package resp

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/dicedb/dice/internal/clientio"
	dicedb "github.com/dicedb/dicedb-go"
	"gotest.tools/v3/assert"
)

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

func TestPFCOUNTWATCH(t *testing.T) {
	publisher := getLocalConnection()
	subscribers := setupSubscribers(3)

	FireCommand(publisher, fmt.Sprintf("DEL %s", pfcountWatchKey))

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

	res := FireCommand(publisher, fmt.Sprintf("PFADD %s %s", pfcountWatchKey, "randomvalue"))
	assert.Equal(t, int64(1), res)

	respParsers := setUpRespParsers(t, subscribers)

	t.Run("Basic PFCount Operations", func(t *testing.T) {
		testPFCountAdd(t, publisher, respParsers)
	},
	)

	t.Run("PFCount Operations including PFMerge", func(t *testing.T) {
		testPFCountMerge(t, publisher, respParsers)
	},
	)
}

func setupSubscribers(count int) []net.Conn {
	subscribers := make([]net.Conn, 0, count)
	for i := 0; i < count; i++ {
		conn := getLocalConnection()
		subscribers = append(subscribers, conn)
	}
	return subscribers
}

func setUpRespParsers(t *testing.T, subscribers []net.Conn) []*clientio.RESPParser {
	respParsers := make([]*clientio.RESPParser, len(subscribers))
	for i, subscriber := range subscribers {
		rp := fireCommandAndGetRESPParser(subscriber, fmt.Sprintf(pfcountWatchQuery, pfcountWatchKey))
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
	return respParsers
}

func testPFCountAdd(t *testing.T, publisher net.Conn, respParsers []*clientio.RESPParser) {
	for _, tc := range pfcountWatchTestCases {
		res := FireCommand(publisher, fmt.Sprintf("PFADD %s %s", tc.key, tc.val))
		assert.Equal(t, int64(1), res)

		verifyWatchResults(t, respParsers, tc.result)
	}
}

func testPFCountMerge(t *testing.T, publisher net.Conn, respParsers []*clientio.RESPParser) {
	for _, tc := range pfcountWatchhWithPFMergeTestCases {
		FireCommand(publisher, fmt.Sprintf("DEL %s %s", tc.destKey1, tc.destKey2))
		FireCommand(publisher, fmt.Sprintf("PFADD %s %s", tc.destKey1, tc.destValue1))
		FireCommand(publisher, fmt.Sprintf("PFADD %s %s", tc.destKey2, tc.destValue2))
		FireCommand(publisher, fmt.Sprintf("PFMERGE %s %s %s", pfcountWatchKey, tc.destKey1, tc.destKey2))

		verifyWatchResults(t, respParsers, tc.result)
	}
}

func verifyWatchResults(t *testing.T, respParsers []*clientio.RESPParser, expected int64) {
	for _, rp := range respParsers {
		v, err := rp.DecodeOne()
		assert.NilError(t, err)
		castedValue, ok := v.([]interface{})
		if !ok {
			t.Errorf("Type assertion to []interface{} failed for value: %v", v)
		}
		assert.Equal(t, 3, len(castedValue))
		assert.Equal(t, pfcountCommand, castedValue[0])
		assert.Equal(t, pfcountWatchFingerPrint, castedValue[1])
		assert.DeepEqual(t, expected, castedValue[2])
	}
}

const (
	pfcountCommandSDK          = "PFCOUNT"
	pfcountWatchKeySDK         = "hllkey"
	pfcountWatchQuerySDK       = "PFCOUNT.WATCH %s"
	pfcountWatchFingerPrintSDK = "1832643469"
	defaultTimeout             = 5 * time.Second
)

type pfcountWatchSDKTestCase struct {
	key    string
	val    any
	result int64
}

type pfcountWatchSDKWithPFMergeTestCase struct {
	destKey1   string
	destValue1 []string
	destKey2   string
	destValue2 []string
	result     int64
}

var PFCountWatchSDKTestCases = []pfcountWatchSDKTestCase{
	{pfcountWatchKeySDK, "value1", int64(1)},
	{pfcountWatchKeySDK, "value2", int64(2)},
	{pfcountWatchKeySDK, "value3", int64(3)},
	{pfcountWatchKeySDK, "value4", int64(4)},
}

var pfcountWatchSDKhWithPFMergeTestCases = []pfcountWatchSDKWithPFMergeTestCase{
	{"DEST_KEY_1", []string{"a", "b", "c", "d", "e", "f"}, "DEST_KEY_2", []string{"f", "g"}, int64(11)},
	{"DEST_KEY_3", []string{"a", "b", "c", "d", "e", "f"}, "DEST_KEY_4", []string{"a1", "b1", "c1", "d1", "e1", "f1"}, int64(17)},
	{"DEST_KEY_5", []string{"f", "g"}, "DEST_KEY_6", []string{"f", "g"}, int64(17)},
}

func TestPFCountWATCHWithSDK(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	publisher := getLocalSdk()
	subscribers := setupSubscribersSDK(3)
	defer cleanupSubscribersSDK(subscribers)

	publisher.Del(ctx, pfcountWatchKey)

	channels := setUpWatchChannelsSDK(t, ctx, subscribers)

	t.Run("Basic PFCount Operations", func(t *testing.T) {
		testPFCountAddSDK(t, ctx, channels, publisher)
	},
	)

	t.Run("PFCount Operations including PFMerge", func(t *testing.T) {
		testPFCountMergeSDK(t, ctx, channels, publisher)
	},
	)
}

func setupSubscribersSDK(count int) []WatchSubscriber {
	subscribers := make([]WatchSubscriber, count)
	for i := range subscribers {
		subscribers[i].client = getLocalSdk()
	}
	return subscribers
}

func cleanupSubscribersSDK(subscribers []WatchSubscriber) {
	for _, sub := range subscribers {
		if sub.watch != nil {
			sub.watch.Close()
		}
	}
}

func setUpWatchChannelsSDK(t *testing.T, ctx context.Context, subscribers []WatchSubscriber) []<-chan *dicedb.WatchResult {
	channels := make([]<-chan *dicedb.WatchResult, len(subscribers))
	for i, subscriber := range subscribers {
		watch := subscriber.client.WatchConn(ctx)
		subscribers[i].watch = watch
		assert.Assert(t, watch != nil)
		firstMsg, err := watch.Watch(ctx, pfcountCommandSDK, pfcountWatchKey)
		assert.NilError(t, err)
		assert.Equal(t, firstMsg.Command, pfcountCommandSDK)
		channels[i] = watch.Channel()
	}
	return channels
}

func testPFCountAddSDK(t *testing.T, ctx context.Context, channels []<-chan *dicedb.WatchResult, publisher *dicedb.Client) {
	for _, tc := range PFCountWatchSDKTestCases {
		err := publisher.PFAdd(ctx, tc.key, tc.val).Err()
		assert.NilError(t, err)

		verifyWatchResultsSDK(t, channels, tc.result)
	}
}

func testPFCountMergeSDK(t *testing.T, ctx context.Context, channels []<-chan *dicedb.WatchResult, publisher *dicedb.Client) {
	for _, tc := range pfcountWatchSDKhWithPFMergeTestCases {
		publisher.Del(ctx, tc.destKey1, tc.destKey2)
		publisher.PFAdd(ctx, tc.destKey1, tc.destValue1).Err()
		publisher.PFAdd(ctx, tc.destKey2, tc.destValue2).Err()
		publisher.PFMerge(ctx, pfcountWatchKey, tc.destKey1, tc.destKey2).Err()

		verifyWatchResultsSDK(t, channels, tc.result)
	}
}

func verifyWatchResultsSDK(t *testing.T, channels []<-chan *dicedb.WatchResult, expected int64) {
	for _, channel := range channels {
		select {
		case v := <-channel:
			assert.Equal(t, pfcountCommandSDK, v.Command)           // command
			assert.Equal(t, pfcountWatchFingerPrint, v.Fingerprint) // Fingerprint
			assert.DeepEqual(t, expected, v.Data)                   // data
		case <-time.After(defaultTimeout):
			t.Fatal("timeout waiting for watch result")
		}
	}
}
