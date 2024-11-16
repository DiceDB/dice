package resp

import (
	"context"
	"fmt"
	"net"
	"sort"
	"testing"
	"time"

	"github.com/dicedb/dice/internal/clientio"
	dicedb "github.com/dicedb/dicedb-go"
	"gotest.tools/v3/assert"
)

type smembersWatchTestCase struct {
    key    string
    val    string
    result any
}

const (
    smembersCommand          = "SMEMBERS"
    smembersWatchKey         = "smemberswatchkey"
    smembersWatchQuery       = "SMEMBERS.WATCH %s"
    smembersWatchFingerPrint = "3660753939"
)

var smembersWatchTestCases = []smembersWatchTestCase{
    {smembersWatchKey, "member1", []any{"member1"}},
    {smembersWatchKey, "member2", []any{"member1", "member2"}},
    {smembersWatchKey, "member3", []any{"member1", "member2", "member3"}},
}

func TestSMEMBERSWATCH(t *testing.T) {
    publisher := getLocalConnection()
    subscribers := setupSubscribers(3)

    FireCommand(publisher, fmt.Sprintf("DEL %s", smembersWatchKey))

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

    respParsers := setUpSmembersRespParsers(t, subscribers)

    t.Run("Basic Set Operations", func(t *testing.T) {
        testSetOperations(t, publisher, respParsers)
    })
}

func sortSlice(v any) any {
    switch v := v.(type) {
    case []interface{}:
        sorted := make([]interface{}, len(v))
        copy(sorted, v)
        sort.Slice(sorted, func(i, j int) bool {
            return sorted[i].(string) < sorted[j].(string)
        })
        return sorted
    case []string:
        sorted := make([]string, len(v))
        copy(sorted, v)
        sort.Strings(sorted)
        return sorted
    default:
        return v
    }
}

func setUpSmembersRespParsers(t *testing.T, subscribers []net.Conn) []*clientio.RESPParser {
    respParsers := make([]*clientio.RESPParser, len(subscribers))
    for i, subscriber := range subscribers {
        rp := fireCommandAndGetRESPParser(subscriber, fmt.Sprintf(smembersWatchQuery, smembersWatchKey))
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

func testSetOperations(t *testing.T, publisher net.Conn, respParsers []*clientio.RESPParser) {
    for _, tc := range smembersWatchTestCases {
        res := FireCommand(publisher, fmt.Sprintf("SADD %s %s", tc.key, tc.val))
        assert.Equal(t, int64(1), res)
        verifySmembersWatchResults(t, respParsers, tc.result)
    }
}

func verifySmembersWatchResults(t *testing.T, respParsers []*clientio.RESPParser, expected any) {
    for _, rp := range respParsers {
        v, err := rp.DecodeOne()
        assert.NilError(t, err)
        castedValue, ok := v.([]interface{})
        if !ok {
            t.Errorf("Type assertion to []interface{} failed for value: %v", v)
        }
        assert.Equal(t, 3, len(castedValue))
        assert.Equal(t, smembersCommand, castedValue[0])
        assert.Equal(t, smembersWatchFingerPrint, castedValue[1])
        assert.DeepEqual(t, sortSlice(expected), sortSlice(castedValue[2]))
    }
}

type smembersWatchSDKTestCase struct {
    key    string
    val    string
    result []string
}

var smembersWatchSDKTestCases = []smembersWatchSDKTestCase{
    {smembersWatchKey, "member1", []string{"member1"}},
    {smembersWatchKey, "member2", []string{"member1", "member2"}},
    {smembersWatchKey, "member3", []string{"member1", "member2", "member3"}},
}

func TestSMEMBERSWATCHWithSDK(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
    defer cancel()

    publisher := getLocalSdk()
    subscribers := setupSubscribersSDK(3)
    defer cleanupSubscribersSDK(subscribers)

    publisher.Del(ctx, smembersWatchKey)

    channels := setUpSmembersWatchChannelsSDK(t, ctx, subscribers)

    t.Run("Basic Set Operations", func(t *testing.T) {
        testSetOperationsSDK(t, ctx, channels, publisher)
    })
}

func setUpSmembersWatchChannelsSDK(t *testing.T, ctx context.Context, subscribers []WatchSubscriber) []<-chan *dicedb.WatchResult {
    channels := make([]<-chan *dicedb.WatchResult, len(subscribers))
    for i, subscriber := range subscribers {
        watch := subscriber.client.WatchConn(ctx)
        subscribers[i].watch = watch
        assert.Assert(t, watch != nil)
        firstMsg, err := watch.Watch(ctx, smembersCommand, smembersWatchKey)
        assert.NilError(t, err)
        assert.Equal(t, firstMsg.Command, smembersCommand)
        channels[i] = watch.Channel()
    }
    return channels
}

func testSetOperationsSDK(t *testing.T, ctx context.Context, channels []<-chan *dicedb.WatchResult, publisher *dicedb.Client) {
    for _, tc := range smembersWatchSDKTestCases {
        err := publisher.SAdd(ctx, tc.key, tc.val).Err()
        assert.NilError(t, err)
        verifySmembersWatchResultsSDK(t, channels, tc.result)
    }
}

func verifySmembersWatchResultsSDK(t *testing.T, channels []<-chan *dicedb.WatchResult, expected []string) {
    for _, channel := range channels {
        select {
        case v := <-channel:
            assert.Equal(t, smembersCommand, v.Command)
            assert.Equal(t, smembersWatchFingerPrint, v.Fingerprint)
            
            received, ok := v.Data.([]interface{})
            if !ok {
                t.Fatalf("Expected []interface{}, got %T", v.Data)
            }
            
            receivedStrings := make([]string, len(received))
            for i, item := range received {
                str, ok := item.(string)
                if !ok {
                    t.Fatalf("Expected string, got %T", item)
                }
                receivedStrings[i] = str
            }
            
            assert.DeepEqual(t, sortSlice(expected), sortSlice(receivedStrings))
        case <-time.After(defaultTimeout):
            t.Fatal("timeout waiting for watch result")
        }
    }
}