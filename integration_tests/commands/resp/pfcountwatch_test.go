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
	dest_key_1    string
	dest_value_1  []string
	dest_key_2 	  string
	dest_value_2  []string
	result 		  int64
}

const (
	pfcountWatchKey   = "hllkey"
	pfcountWatchQuery = "PFCOUNT.WATCH %s"
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
	subscribers := []net.Conn{getLocalConnection(), getLocalConnection(), getLocalConnection()}

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

	// Add elements with PFADD
	for _, tc := range pfcountWatchTestCases {
		res := FireCommand(publisher, fmt.Sprintf("PFADD %s %s", tc.key, tc.val))
		assert.Equal(t, int64(1), res)
		for _, rp := range respParsers {
			v, err := rp.DecodeOne()
			assert.NilError(t, err)
			castedValue, ok := v.([]interface{})
			if !ok {
				t.Errorf("Type assertion to []interface{} failed for value: %v", v)
			}
			assert.Equal(t, 3, len(castedValue))
			assert.Equal(t, "PFCOUNT", castedValue[0])
			assert.Equal(t, "1580567186", castedValue[1])
			assert.DeepEqual(t, tc.result, castedValue[2])
		}
	}

	// Test for PFMERGE Case
	for _, tc := range pfcountWatchhWithPFMergeTestCases {
		FireCommand(publisher, fmt.Sprintf("DEL %s %s", tc.dest_key_1, tc.dest_key_2))
		FireCommand(publisher, fmt.Sprintf("PFADD %s %s", tc.dest_key_1, tc.dest_value_1))
		FireCommand(publisher, fmt.Sprintf("PFADD %s %s", tc.dest_key_2, tc.dest_value_2))
		FireCommand(publisher, fmt.Sprintf("PFMERGE %s %s %s", pfcountWatchKey, tc.dest_key_1, tc.dest_key_2))
		assert.Equal(t, int64(1), res)
		for _, rp := range respParsers {
			v, err := rp.DecodeOne()
			assert.NilError(t, err)
			castedValue, ok := v.([]interface{})
			if !ok {
				t.Errorf("Type assertion to []interface{} failed for value: %v", v)
			}
			assert.Equal(t, 3, len(castedValue))
			assert.Equal(t, "PFCOUNT", castedValue[0])
			assert.Equal(t, "1580567186", castedValue[1])
			assert.DeepEqual(t, tc.result, castedValue[2])
		}
	}
}

type pfcountWatchSDKTestCase struct {
	key    string
	val    any
	result int64
}

type pfcountWatchSDKWithPFMergeTestCase struct {
	dest_key_1    string
	dest_value_1  []string
	dest_key_2 	  string
	dest_value_2  []string
	result 		  int64
}

var PFCountWatchSDKTestCases = []pfcountWatchSDKTestCase{
	{pfcountWatchKey, "value1", int64(1)},
	{pfcountWatchKey, "value2", int64(2)},
	{pfcountWatchKey, "value3", int64(3)},
	{pfcountWatchKey, "value4", int64(4)},
}

var pfcountWatchSDKhWithPFMergeTestCases = []pfcountWatchSDKWithPFMergeTestCase{
	{"DEST_KEY_1", []string{"a", "b", "c", "d", "e", "f"}, "DEST_KEY_2", []string{"f", "g"}, int64(11)},
	{"DEST_KEY_3", []string{"a", "b", "c", "d", "e", "f"}, "DEST_KEY_4", []string{"a1", "b1", "c1", "d1", "e1", "f1"}, int64(17)},
	{"DEST_KEY_5", []string{"f", "g"}, "DEST_KEY_6", []string{"f", "g"}, int64(17)},
}

func TestPFCountWATCHWithSDK(t *testing.T) {
	publisher := getLocalSdk()
	subscribers := []WatchSubscriber{{client: getLocalSdk()}, {client: getLocalSdk()}, {client: getLocalSdk()}}

	publisher.Del(context.Background(), pfcountWatchKey)

	channels := make([]<-chan *dicedb.WatchResult, len(subscribers))
	for i, subscriber := range subscribers {
		watch := subscriber.client.WatchConn(context.Background())
		subscribers[i].watch = watch
		assert.Assert(t, watch != nil)
		firstMsg, err := watch.Watch(context.Background(), "PFCOUNT", pfcountWatchKey)
		assert.NilError(t, err)
		assert.Equal(t, firstMsg.Command, "PFCOUNT")
		channels[i] = watch.Channel()
	}

	for _, tc := range PFCountWatchSDKTestCases {
		err := publisher.PFAdd(context.Background(), tc.key, tc.val).Err()
		assert.NilError(t, err)

		for _, channel := range channels {
			v := <-channel

			assert.Equal(t, "PFCOUNT", v.Command)         // command
			assert.Equal(t, "1580567186", v.Fingerprint) // Fingerprint
			assert.DeepEqual(t, tc.result, v.Data)       // data
		}
	}

	for _, tc := range pfcountWatchSDKhWithPFMergeTestCases {
		publisher.Del(context.Background(), tc.dest_key_1, tc.dest_key_2)
		publisher.PFAdd(context.Background(), tc.dest_key_1, tc.dest_value_1).Err()
		publisher.PFAdd(context.Background(), tc.dest_key_2, tc.dest_value_2).Err()
		publisher.PFMerge(context.Background(), pfcountWatchKey, tc.dest_key_1, tc.dest_key_2).Err()

		for _, channel := range channels {
			v := <-channel

			assert.Equal(t, "PFCOUNT", v.Command)         // command
			assert.Equal(t, "1580567186", v.Fingerprint) // Fingerprint
			assert.DeepEqual(t, tc.result, v.Data)       // data
		}
	}
}